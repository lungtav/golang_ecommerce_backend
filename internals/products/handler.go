package products

import (
	"errors"
	"net/http"

	"github.com/lungtav/golang_ecommerce_backend/internals/platform/request"
	"github.com/lungtav/golang_ecommerce_backend/internals/platform/response"
)

type handler struct {
	service Service
}

func NewHandler(service Service) *handler {
	return &handler{service: service}
}

func (h *handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req CreateProductsRequest

	if !request.Decode(w, r, &req, 1<<10) {
		return
	}

	product, err := h.service.CreateProduct(r.Context(), req)

	if err != nil {
		switch {
		case errors.Is(err, ErrNoStore):
			response.Error(w, http.StatusNotFound, "store doesnt exists")
			return
		default:
			response.Error(w, http.StatusInternalServerError, "internal server error")
			return
		}
	}

	response.JSON(w, http.StatusCreated, map[string]any{"product": product})
}
