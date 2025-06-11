package server

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/JustinLi007/whatdoing-server/internal/database"
	"github.com/JustinLi007/whatdoing-server/internal/migrations"
)

type Server struct {
	port int
	db   database.DbService
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

	newServer := Server{
		port: 8000,
		db:   db,
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
