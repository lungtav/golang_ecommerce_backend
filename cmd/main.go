package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/lungtav/golang_ecommerce_backend/internals/platform/env"
)

func main() {

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	slog.SetDefault(logger)

	err := godotenv.Load()

	if err != nil {
		slog.Error("unable to load env", "error", err)
	}

	cfg := config{
		addr: ":8080",
		db: dbConfig{
			dsn: env.Get("GOOSE_DBSTRING", ""),
		},
	}

	pool, err := pgxpool.New(context.Background(), cfg.db.dsn)

	if err != nil {
		slog.Error("db connection failed", "error", err)
		os.Exit(1)
	}

	if err := pool.Ping(context.Background()); err != nil {
		slog.Error("db ping failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	api := application{
		config: cfg,
		db:     pool,
	}

	if err := api.run(api.mount()); err != nil {
		slog.Error("server has failed to start", "error", err)
		os.Exit(1)
	}
}
