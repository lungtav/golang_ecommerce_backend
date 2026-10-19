package stores

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lungtav/golang_ecommerce_backend/internals/db"
)

type Service interface {
	CreateStore(ctx context.Context, req CreateStoreRequest, ownerId string) (db.CreateStoreRow, error)
}

type svc struct {
	queries *db.Queries
}

func NewService(queries *db.Queries) Service{
	return &svc{queries: queries}
}

var(
	ErrConflictError= errors.New("store with this name already exists, choose another")
)

func (s *svc) CreateStore(ctx context.Context, req CreateStoreRequest, ownerId string) (db.CreateStoreRow, error){

	var id pgtype.UUID

	if err:=id.Scan(ownerId); err!=nil{
		slog.Error("uuid parse failed", "error", err)
		return db.CreateStoreRow{}, err
	}


	//insert into db
	store, err:= s.queries.CreateStore(ctx, db.CreateStoreParams{
		Name: req.Name,
		OwnerID: id,
	})

	var pgErr *pgconn.PgError

	if err!=nil{
		if errors.As(err, &pgErr) && pgErr.Code == "23505"{
		return db.CreateStoreRow{}, ErrConflictError
		}

		slog.Error("create store failed", "error", err)
		return db.CreateStoreRow{}, err
	}

	return store, nil
}