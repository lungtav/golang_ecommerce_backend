package products

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lungtav/golang_ecommerce_backend/internals/db"
	"github.com/lungtav/golang_ecommerce_backend/internals/platform/sku"
)

type Service interface {
	CreateProduct(ctx context.Context, req CreateProductsRequest)(  db.GetProductRow, error)
}

type svc struct {
	queries *db.Queries
	pool *pgxpool.Pool
}

func NewService (queries *db.Queries, pool *pgxpool.Pool) Service{
	return &svc{queries: queries, pool: pool}
}

var(
	ErrNoStore = errors.New("store doesn't exists")
)

func (s *svc) CreateProduct(ctx context.Context, req CreateProductsRequest)( db.GetProductRow, error){
	var storeId pgtype.UUID

	if err:= storeId.Scan(req.StoreId); err != nil{
		slog.Error("convert id string to uuid type failed", "error", err)
		return  db.GetProductRow{}, err
	}

	//confirm store exists
	if _, err:= s.queries.GetStoreById(ctx, storeId); err!=nil{
		if errors.Is(err, pgx.ErrNoRows){
			return  db.GetProductRow{}, ErrNoStore 
		} else {
			slog.Error("error fetching store", "error", err)
			return  db.GetProductRow{}, err
		}
	}

	tx, err:= s.pool.Begin(ctx)

	if err!=nil{
		slog.Error("create transaction failed", "error", err)
		return  db.GetProductRow{}, err
	}

	defer tx.Rollback(ctx)

	qtx:= s.queries.WithTx(tx)
	//create product
	var description pgtype.Text

	if err =description.Scan(req.Description); err!=nil{
		slog.Error("convert description to pg text failed", "error", err)
		return db.GetProductRow{}, err
	}
	product, err:= qtx.CreateProduct(ctx, db.CreateProductParams{
		StoreID: storeId,
		Name: req.Name,
		Status: req.Status,
		Description: description,
	})

	if err!=nil{
		slog.Error("error creating product", "error", err)
		return  db.GetProductRow{}, err
	}

	//create variant(s)
	var size pgtype.Text
	var color pgtype.Text
	

	for idx, val:= range req.Variants{
		if err=size.Scan(val.Size); err!=nil{
			slog.Error("convert size to pg text failed", "error", err)
			return  db.GetProductRow{}, err
		}
		if err=color.Scan(val.Color); err!=nil{
			slog.Error("convert color to pg text failed", "error", err)
			return  db.GetProductRow{}, err
		}

		sku, err:=sku.Generate()
		if err != nil {
			slog.Error("generate sku failed", "error", err)
			return db.GetProductRow{}, err
		}
		variant,err:= qtx.CreateProductVariant(ctx, db.CreateProductVariantParams{
			ProductID: product.ID,
			Sku: sku,
			Size: size,
			Color: color,
			PriceInKobo: int32(val.PriceInKobo),
		}) 

		if err!=nil{
			slog.Error("failed to create product variant", "index", idx, "error",err)
			return   db.GetProductRow{}, err
		}

		//register inventory
		if err=qtx.CreateInventory(ctx, db.CreateInventoryParams{
			ProductVariantID: variant.ID,
			StoreID:storeId,
			Quantity: int32(val.Inventory),
		}); err!=nil{
			slog.Error("create inventory failed", "error", err)
			return db.GetProductRow{}, err
		}
	}

	productWithVariants, err := qtx.GetProduct(ctx, product.ID)

	if err!=nil{
		slog.Error("error getting product", "error", err)
		return  db.GetProductRow{}, err
	}

	//create product category
	var categoryID pgtype.UUID

	if err:=categoryID.Scan(req.CategoryId); err!= nil{
		slog.Error("convert category id into uuid failed", "error", err)
		return db.GetProductRow{}, err
	}
	if err:= qtx.CreateProductCategory(ctx, db.CreateProductCategoryParams{
		ProductID: product.ID,
		CategoryID:categoryID,
	}); err!=nil{
		slog.Error("failed to create product category")
		return db.GetProductRow{}, err
	}

	if err=tx.Commit(ctx); err!=nil{
		slog.Error("commit transaction failed", "error", err)
		return db.GetProductRow{}, err
	}

	return  productWithVariants, nil
}