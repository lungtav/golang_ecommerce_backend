package main

import (
	"log/slog"
	"os"

	"github.com/lungtav/golang_ecommerce_backend/internals/platform/env"
)

func main() {
	cfg := config{
		addr: ":8080",
		db: dbConfig{
			dsn: env.Get("GOOSE_DBSTRING", ""),
		},
	}

	api := application{
		config: cfg,
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	slog.SetDefault(logger)

	if err := api.run(api.mount()); err != nil {
		slog.Error("server has failed to start", "error", err)
		os.Exit(1)
	}
}
