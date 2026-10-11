package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/lungtav/golang_ecommerce_backend/internals/db"
)

type application struct {
	config config
}

func (app *application) mount() http.Handler {
	err:= godotenv.Load()

	if err != nil{
		slog.Error("unable to load env", "error", err)
	}

	r := chi.NewRouter()

	pool, err := pgxpool.New(context.Background(), app.config.db.dsn)

	if err!=nil{
		slog.Error("db connection failed", "error", err)
		os.Exit(1)
	}

	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("root."))
	})

	queries:=db.New(pool)
	

	return r
}

func (app *application) run(r http.Handler) error {
	srv := http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  time.Second * 30,
		WriteTimeout: time.Second * 10,
		IdleTimeout:  time.Minute,
	}

	slog.Info("server is starting....")

	return srv.ListenAndServe()
}

type config struct {
	addr string
	db   dbConfig
}

type dbConfig struct {
	dsn string
}
