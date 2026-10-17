package main

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type application struct {
	config config
	db     *pgxpool.Pool
}

type config struct {
	addr         string
	db           dbConfig
	resendApiKey string
	jwtSecretKey string
	frontendUrl  string
}

type dbConfig struct {
	dsn string
}
