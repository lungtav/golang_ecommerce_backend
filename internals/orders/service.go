package orders

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lungtav/golang_ecommerce_backend/internals/db"
)

type Service interface {
	CreateOrder(ctx context.Context) error
}

type svc struct {
	queries *db.Queries
	pool    *pgxpool.Pool
}

func NewService(queries *db.Queries, pool *pgxpool.Pool) Service {
	return &svc{queries: queries, pool: pool}
}

func (s *svc) CreateOrder(ctx context.Context) error {
	return nil
}
