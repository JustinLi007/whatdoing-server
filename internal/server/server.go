package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/JustinLi007/whatdoing-server/internal/api"
	"github.com/JustinLi007/whatdoing-server/internal/database"
	"github.com/JustinLi007/whatdoing-server/internal/middleware"
	"github.com/JustinLi007/whatdoing-server/migrations"
)

type Server struct {
	port                    int
	db                      database.DbService
	middleware              *middleware.Middleware
	handlerUsers            api.HandlerUsers
	handlerJwt              api.HandlerJwt
	handlerAnime            api.HandlerAnime
	handlerAnimeNames       api.HandlerAnimeNames
	handlerUserLibraryAnime api.HandlerUserLibraryAnime
}

func NewServer(ctx context.Context) *http.Server {
	db, err := database.NewDb()
	if err != nil {
		log.Fatalf("error: Server NewServer NewDb: %v", err)
	}

	err = db.MigrateFS(migrations.Fs, ".")
	if err != nil {
		log.Fatalf("error: Server NewServer MigrateFS: %v", err)
	}

	// dbs
	dbsUsers := database.NewDbsUsers(db)
	dbsJwt := database.NewDbsJwt(db)
	dbsAnime := database.NewDbsAnime(db)
	dbsRelUsersAnime := database.NewDbsUsersAnime(db)
	dbsAnimeNames := database.NewDbsAnimeNames(db)
	dbsRelAnimeAnimeNames := database.NewDbsRelAnimeAnimeNames(db)
	dbsUserLibrary := database.NewDbsUserLibrary(db)
	dbsRelAnimeUserLibrary := database.NewDbsRelAnimeUserLibrary(db)

	// handlers
	handlerUsers := api.NewHandlerUsers(dbsUsers, dbsJwt)
	handlerJwt := api.NewHandlerJwt(dbsJwt)
	handlerAnime := api.NewHandlerAnime(dbsAnime, dbsRelUsersAnime)
	handlerAnimeNames := api.NewHandlerAnimeNames(dbsAnimeNames, dbsRelAnimeAnimeNames)
	handlerUserLibraryAnime := api.NewHandlerUserLibraryAnime(dbsUserLibrary, dbsRelAnimeUserLibrary)

	// middleware
	middleware := middleware.NewMiddleware(dbsUsers, dbsJwt)

	newServer := Server{
		port:                    8000,
		db:                      db,
		middleware:              middleware,
		handlerUsers:            handlerUsers,
		handlerJwt:              handlerJwt,
		handlerAnime:            handlerAnime,
		handlerAnimeNames:       handlerAnimeNames,
		handlerUserLibraryAnime: handlerUserLibraryAnime,
	}

	mux := newServer.RegisterRoutes()

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", newServer.port),
		Handler:      mux,
		IdleTimeout:  time.Minute,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 30,
	}

	go RoutineRemoveExpiredJwt(ctx, dbsJwt)

	return server
}

func RoutineRemoveExpiredJwt(ctx context.Context, dbsJwt database.DbsJwt) {
	ticker := time.NewTicker(time.Hour)
	for {
		select {
		case <-ticker.C:
			log.Printf("info: Server: Routine: RoutineRemoveExpiredJwt: Execute")
			err := dbsJwt.DeleteExpired()
			if err != nil {
				log.Printf("error: Server: Routine: RoutineRemoveExpiredJwt: DeleteExpired: %v", err)
			}
		case <-ctx.Done():
			log.Printf("info: Server: Routine: RoutineRemoveExpiredJwt: Stop")
			ticker.Stop()
			return
		}
	}
}
