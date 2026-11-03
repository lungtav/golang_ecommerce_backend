package main

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lungtav/golang_ecommerce_backend/internals/payments"
)

type application struct {
	config config
	db     *pgxpool.Pool
	paystack payments.Initializer
}

type config struct {
	addr         string
	db           dbConfig
	resendApiKey string
	jwtSecretKey string
	frontendUrl  string
	paystackSecretKey string
}

type dbConfig struct {
	dsn string
}
