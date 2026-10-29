package products

type CreateProductsRequest struct {
	StoreId     string           `json:"store_id" validate:"required"`
	Name        string           `json:"name" validate:"required,min=2"`
	Status      string           `json:"status" validate:"required,oneof=draft active archive"`
	Description string           `json:"description"`
	Variants    []VariantRequest `json:"variants" validate:"required,min=1,dive"`
	CategoryId  string           `json:"category_id" validate:"required"`
}

type VariantRequest struct {
	Size        string `json:"size" validate:"omitempty,oneof=S M L XL"`
	Color       string `json:"color" validate:"max=10"`
	PriceInKobo int    `json:"price_in_kobo" validate:"required,gt=0"`
	Inventory   int    `json:"inventory" validate:"required,min=1"`
}
