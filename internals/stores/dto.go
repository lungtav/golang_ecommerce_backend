package stores

type StoreRequest struct {
	Name        string `json:"name" validate:"required,max=100"`
	Description string `json:"description"`
}
