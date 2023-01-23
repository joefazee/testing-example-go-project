package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *application) routes() http.Handler {

	mux := chi.NewRouter()
	mux.Use(middleware.Recoverer)
	mux.Use(app.addIpToContext)
	mux.Use(app.Session.LoadAndSave)

	// routes
	mux.Get("/", app.home)
	mux.Post("/login", app.login)

	mux.Route("/user", func(mux chi.Router) {
		mux.Use(app.auth)
		mux.Get("/profile", app.profilePage)
		mux.Post("/upload-profile-pic", app.uploadProfilePicture)
	})

	// register static
	fileServer := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return mux
}
