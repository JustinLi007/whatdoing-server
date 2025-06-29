package server

import (
	"github.com/go-chi/chi/v5"
)

func (s *Server) RegisterRoutes() *chi.Mux {
	r := chi.NewRouter()
	r.Use(s.middleware.Cors)

	r.Post("/users/login", s.handlerUsers.Login)
	r.Post("/users/signup", s.handlerUsers.SignUp)

	r.Post("/contents", s.handlerAnime.NewAnime)
	r.Get("/contents", s.handlerAnime.GetAllAnime)
	r.Get("/contents/{contentId}", s.handlerAnime.GetAnime)

	r.Group(func(r chi.Router) {
		r.Use(s.middleware.RequireJwt)
		r.Use(s.middleware.RequireUser)

		r.Get("/users/{userId}", s.handlerUsers.GetUserById)
	})

	r.Group(func(r chi.Router) {
		r.Use(s.middleware.RequireJwt)
		r.Use(s.middleware.RequireUser)

		// TODO: this probably shouldn't require an user...
		// r.Post("/contents", s.handlerAnime.NewAnime)
		// r.Get("/contents/{contentId}", s.handlerAnime.GetAnime)
	})

	return r
}
