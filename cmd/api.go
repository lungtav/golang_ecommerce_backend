package main

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/lungtav/golang_ecommerce_backend/internals/auth"
	"github.com/lungtav/golang_ecommerce_backend/internals/db"
	"github.com/lungtav/golang_ecommerce_backend/internals/platform/mailer"
)

func (app *application) mount() http.Handler {

	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	queries := db.New(app.db)

	mailer := mailer.New(app.config.resendApiKey)

	authService := auth.NewService(queries, mailer)
	authHandler := auth.NewHandler(authService)

	r.Route("/auth", func(r chi.Router) {
		r.Post("/sign-up", authHandler.SignUp)
	})

	return r
}

func (app *application) run(r http.Handler) error {
	srv := http.Server{
		Addr:         app.config.addr,
		Handler:      r,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 30,
		IdleTimeout:  time.Minute,
	}

	slog.Info("server is running on port", "addr", app.config.addr)

	return srv.ListenAndServe()
}
