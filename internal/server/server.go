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
	dbsUsers     database.DbsUsers
	handlerUsers api.HandlerUsers
	middleware   *middleware.Middleware
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

	dbsUsers := database.NewDbsUsers(db)
	handlerUsers := api.NewHandlerUsers(dbsUsers)

	middleware := middleware.NewMiddleware()

	newServer := Server{
		port:         8000,
		db:           db,
		dbsUsers:     dbsUsers,
		handlerUsers: handlerUsers,
		middleware:   middleware,
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
