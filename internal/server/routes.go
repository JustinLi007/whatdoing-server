package server

import (
	"github.com/go-chi/chi/v5"
)

func (s *Server) RegisterRoutes() *chi.Mux {
	r := chi.NewRouter()
	r.Use(s.middleware.Cors)

	r.Post("/users/login", s.handlerUsers.Login)
	r.Post("/users/signup", s.handlerUsers.SignUp)
	r.Group(func(r chi.Router) {
		r.Use(s.middleware.RequireJwt)
		r.Use(s.middleware.RequireUser)

		r.Get("/users/{userId}", s.handlerUsers.GetUserById)
	})

	r.Get("/contents/anime", s.handlerAnime.GetAllAnime)
	r.Get("/contents/anime/{contentId}", s.handlerAnime.GetAnime)
	r.Group(func(r chi.Router) {
		r.Use(s.middleware.RequireJwt)
		r.Use(s.middleware.RequireUser)

		r.Post("/contents/anime", s.handlerAnime.NewAnime)
		r.Post("/contents/anime/{contentId}", s.handlerAnime.UpdateAnime)
	})

	return r
}
