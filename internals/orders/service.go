package orders

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lungtav/golang_ecommerce_backend/internals/db"
)

type Service interface {
	CreateOrder(ctx context.Context, userId string) (OrderResult, error)
}

type svc struct {
	queries *db.Queries
	pool    *pgxpool.Pool
}

func NewService(queries *db.Queries, pool *pgxpool.Pool) Service {
	return &svc{queries: queries, pool: pool}
}

var (
	ErrNoCart         = errors.New("no cart row exist for user")
	ErrEmptyCart      = errors.New("no items in users cart")
	ErrNotEnoughStock = errors.New("not enough stock")
)

type OrderResult struct {
	ID          pgtype.UUID
	TotalInKobo int64
}

func (s *svc) CreateOrder(ctx context.Context, userId string) (OrderResult, error) {
	var userID pgtype.UUID

	if err := userID.Scan(userId); err != nil {
		slog.Error("change id to pgtype uuid failed", "error", err)
		return OrderResult{}, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		slog.Error("unable to create a transaction", "error", err)
		return OrderResult{}, err
	}

	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)

	//pass cart items into order items
	cartID, err := qtx.GetCart(ctx, userID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return OrderResult{}, ErrEmptyCart
		}
		slog.Error("get cart failed", "error", err)
	}

	items, err := qtx.GetCartItems(ctx, userID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return OrderResult{}, ErrEmptyCart
		}
		slog.Error("get cart items failed", "error", err)
	}

	cartAmount, err := qtx.GetCartTotal(ctx, userID)

	if err != nil {
		slog.Error("get cart total amount failed", "error", err)
		return OrderResult{}, err
	}

	//insert into orders
	orderID, err := qtx.CreateOrder(ctx, db.CreateOrderParams{
		UserID:   userID,
		Status:   "pending",
		Amount:   int32(cartAmount),
		Currency: "NGN",
	})
	if err != nil {
		slog.Error("create order failed", "error", err)
		return OrderResult{}, err
	}

	for _, it := range items {

		n, err := qtx.DecrementInventory(ctx, db.DecrementInventoryParams{
			ProductVariantID: it.ProductVariantID,
			Quantity:         it.Quantity,
		})
		if err != nil {
			slog.Error("decrement inventory failed", "error", err)
			return OrderResult{}, err
		}
		if n == 0 {
			return OrderResult{}, ErrNotEnoughStock
		}

		if err := qtx.CreateOrderItems(ctx, db.CreateOrderItemsParams{
			ProductID:        it.ProductID,
			ProductVariantID: it.ProductVariantID,
			OrderID:          orderID,
			Quantity:         it.Quantity,
		}); err != nil {
			slog.Error("create order items failed", "error", err)
			return OrderResult{}, err
		}
	}

	if err := qtx.ClearCart(ctx, cartID); err != nil {
		slog.Error("clear cart failed", "error", err)
		return OrderResult{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		slog.Error("commit failed", "error", err)
		return OrderResult{}, err
	}

	return OrderResult{
		ID:          orderID,
		TotalInKobo: cartAmount,
	}, nil
}
