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
		r.Get("/users/session", s.handlerUsers.CheckSession)
		r.Delete("/users/session", s.handlerUsers.Logout)
	})

	r.Group(func(r chi.Router) {
		r.Use(s.middleware.RequireJwt)
		r.Use(s.middleware.RequireUser)
		r.Post("/tokens/refresh", s.handlerJwt.RefreshJwt)
	})

	r.Get("/anime/{contentId}", s.handlerAnime.GetAnime)
	r.Group(func(r chi.Router) {
		r.Use(s.middleware.RequireJwt)
		r.Get("/anime", s.handlerAnime.GetAllAnime)
	})
	r.Group(func(r chi.Router) {
		r.Use(s.middleware.RequireJwt)
		r.Use(s.middleware.RequireUser)
		r.Post("/anime", s.handlerAnime.NewAnime)
		r.Put("/anime/{contentId}", s.handlerAnime.UpdateAnime)
	})

	r.Group(func(r chi.Router) {
		r.Use(s.middleware.RequireJwt)
		r.Use(s.middleware.RequireUser)
		r.Post("/progress/anime", s.handlerProgressAnime.AddToLibrary)
		r.Put("/progress/anime", s.handlerProgressAnime.SetProgress)
		r.Delete("/progress/anime", s.handlerProgressAnime.RemoveProgress)
		r.Get("/progress/anime", s.handlerProgressAnime.GetProgress)
	})

	return r
}
