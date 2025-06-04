package server

import (
	"fmt"
	"net/http"
	"time"
)

type Server struct {
	port int
}

func NewServer() *http.Server {
	newServer := Server{
		port: 8000,
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
