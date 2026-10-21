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
	CreateStore(ctx context.Context, req StoreRequest, ownerId string) (db.CreateStoreRow, error)
	UpdateStore(ctx context.Context,req StoreRequest, storeId string) (error)
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

func (s *svc) CreateStore(ctx context.Context, req StoreRequest, ownerId string) (db.CreateStoreRow, error){

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

func (s *svc) UpdateStore(ctx context.Context, req StoreRequest, Id string) (error){
	var description pgtype.Text
	var storeID pgtype.UUID

	if err:=description.Scan(req.Description); err!=nil{
		slog.Error("failed to parse description", "error", err)
		return err
	}
	if err:=storeID.Scan(Id); err!=nil{
		slog.Error("failed to parse id", "error", err)
		return err
	}
	
	_, err:= s.queries.UpdateStore(ctx, db.UpdateStoreParams{
		Name: req.Name,
		Description: description,
		ID: storeID,
	})

	if err!=nil{
		slog.Error("update store failed", "error", err)
		return err
	}
	
	return nil
}