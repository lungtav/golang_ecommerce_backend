package carts

type AddToCartRequest struct {
	ProductId        string `json:"product_id" validate:"required"`
	ProductVariantId string `json:"product_variant_id" validate:"required"`
	Quantity         int    `json:"quantity" validate:"required,min=1"`
}
