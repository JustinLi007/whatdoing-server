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

	r.Get("/contents/anime", s.handlerAnime.GetAllAnime)
	r.Get("/contents/anime/{contentId}", s.handlerAnime.GetAnime)
	r.Group(func(r chi.Router) {
		r.Use(s.middleware.RequireJwt)
		r.Use(s.middleware.RequireUser)

		r.Post("/contents/anime", s.handlerAnime.NewAnime)
		r.Put("/contents/anime/{contentId}", s.handlerAnime.UpdateAnime)
	})

	r.Group(func(r chi.Router) {
		r.Use(s.middleware.RequireJwt)
		r.Use(s.middleware.RequireUser)

		r.Post("/library/anime", s.handlerUserLibraryAnime.AddToLibrary)
		r.Put("/library/anime/progress", s.handlerUserLibraryAnime.SetProgress)
		r.Put("/library/anime/status", s.handlerUserLibraryAnime.SetStatus)
		r.Delete("/library/anime/progress", s.handlerUserLibraryAnime.RemoveProgress)
		r.Get("/library/anime/progress", s.handlerUserLibraryAnime.GetProgress)
	})

	return r
}
