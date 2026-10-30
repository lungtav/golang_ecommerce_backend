package carts

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
	AddToCart(ctx context.Context, req AddToCartRequest, userId string) error
	GetCartItems(ctx context.Context, userId string) ([]db.GetCartItemsRow, error)
}

type svc struct {
	queries *db.Queries
	pool    *pgxpool.Pool
}

func NewService(queries *db.Queries, pool *pgxpool.Pool) Service {
	return &svc{queries: queries, pool: pool}
}

var (
	ErrNoProduct      = errors.New("product to be added to cart doesnt exist")
	ErrOutOfStock     = errors.New("product sold out")
	ErrNotEnoughStock = errors.New("quantity available is lesser than you wish to purchase")
	ErrNoCartItem     = errors.New("no item currently in cart")
)

func (s *svc) AddToCart(ctx context.Context, req AddToCartRequest, userId string) error {

	//check if products id exists
	var variantID pgtype.UUID
	var productID pgtype.UUID

	if err := variantID.Scan(req.ProductVariantId); err != nil {
		slog.Error("convert variant id type to uuid", "error", err)
		return err
	}
	if err := productID.Scan(req.ProductId); err != nil {
		slog.Error("convert product id type to uuid", "error", err)
		return err
	}

	if _, err := s.queries.GetProductVariant(ctx, db.GetProductVariantParams{
		ID:        variantID,
		ProductID: productID,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoProduct
		}
		slog.Error("get pproduct variant failed", "error", err)
		return err
	}

	//get users cart
	var cartID pgtype.UUID
	var userID pgtype.UUID

	if err := userID.Scan(userId); err != nil {
		slog.Error("failed to convert user id to uuid pgtype", "error", err)
	}

	cartID, err := s.queries.GetCart(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			cartID, err = s.queries.CreateCart(ctx, userID)
			if err != nil {
				slog.Error("failed to create cart while adding to cart", "error", err)
				return err
			}
		}

		slog.Error("get cart failed", "error", err)
		return err
	}

	//get the stock
	stock, err := s.queries.GetInventory(ctx, variantID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoProduct
		}
		slog.Error("get inventory failed", "error", err)
		return err
	}

	q, err := s.queries.GetCartItemQuantity(ctx, db.GetCartItemQuantityParams{
		CartID:           cartID,
		ProductVariantID: variantID,
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			//add to cart items
			if err = s.queries.AddToCart(ctx, db.AddToCartParams{
				ProductID:        productID,
				ProductVariantID: variantID,
				CartID:           cartID,
				Quantity:         int32(req.Quantity),
			}); err != nil {
				slog.Error("add to cart failed", "error", err)
				return err
			}
			return nil
		}
		slog.Error("get inventory failed", "error", err)
		return err
	}

	newQuantity := q + int32(req.Quantity)

	if newQuantity > stock {
		return ErrNotEnoughStock
	}

	//update cart quantity
	i, err := s.queries.UpdateCartItem(ctx, db.UpdateCartItemParams{
		Quantity:         newQuantity,
		CartID:           cartID,
		ProductVariantID: variantID,
	})

	if err != nil {
		slog.Error("couldnt update cart item", "error", err)
		return err
	}

	if i == 1 {
		return nil
	}
	return nil
}

func (s *svc) GetCartItems(ctx context.Context, userId string) ([]db.GetCartItemsRow, error) {
	var userID pgtype.UUID

	if err := userID.Scan(userId); err != nil {
		slog.Error("failed to convert user id to uuid pgtype", "error", err)
		return []db.GetCartItemsRow{}, err
	}

	cart, err := s.queries.GetCartItems(ctx, userID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []db.GetCartItemsRow{}, ErrNoCartItem
		}

		slog.Error("get cart item failed", "error", err)
		return []db.GetCartItemsRow{}, err
	}

	return cart, nil
}
