package server

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (s *Server) RegisterRoutes() *chi.Mux {
	r := chi.NewRouter()

	r.Get("/", s.HandlerHello)

	r.Get("/users/{userId}", s.handlerUsers.GetUserById)
	r.Post("/users", s.handlerUsers.CreateUser)

	return r
}

func (s *Server) HandlerHello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello world")
	return
}
