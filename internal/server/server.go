package server

import (
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
	port         int
	db           database.DbService
	middleware   *middleware.Middleware
	handlerUsers api.HandlerUsers
	handlerJwt   api.HandlerJwt
	handlerAnime api.HandlerAnime
}

func NewServer() *http.Server {
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

	// handlers
	handlerUsers := api.NewHandlerUsers(dbsUsers, dbsJwt)
	handlerJwt := api.NewHandlerJwt(dbsJwt)
	handlerAnime := api.NewHandlerAnime(dbsAnime, dbsRelUsersAnime)

	// middleware
	middleware := middleware.NewMiddleware(dbsUsers, dbsJwt)

	newServer := Server{
		port:         8000,
		db:           db,
		middleware:   middleware,
		handlerUsers: handlerUsers,
		handlerJwt:   handlerJwt,
		handlerAnime: handlerAnime,
	}

	mux := newServer.RegisterRoutes()

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", newServer.port),
		Handler:      mux,
		IdleTimeout:  time.Minute,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 30,
	}

	return server
}
