package categories

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lungtav/golang_ecommerce_backend/internals/db"
)

type Service interface {
	CreateCategory(ctx context.Context, name string) error
}

type svc struct {
	queries *db.Queries
}

func NewService(queries *db.Queries) Service {
	return &svc{queries: queries}
}

var (
	ErrConflictError = errors.New("category exists")
)

func (s *svc) CreateCategory(ctx context.Context, name string) error {
	var pgErr *pgconn.PgError

	if err := s.queries.CreateCategory(ctx, strings.ToLower(name)); err != nil {
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrConflictError
		}
		slog.Error("create category failed", "error", err)
		return err
	}

	return nil
}
