package categories

type CategoryRequest struct {
	Name string `json:"name" validate:"required,min=3"`
}
