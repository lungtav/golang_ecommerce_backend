package orders

import (
	"errors"
	"net/http"

	"github.com/lungtav/golang_ecommerce_backend/internals/platform/middleware"
	"github.com/lungtav/golang_ecommerce_backend/internals/platform/response"
)

type handler struct {
	service Service
}

func NewHandler(service Service) *handler {
	return &handler{service: service}
}

func (h *handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	userId, ok := middleware.UserId(r.Context())

	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauhorixed")
		return
	}

	OrderResult, err := h.service.CreateOrder(r.Context(), userId)

	if err != nil {
		switch {
		case errors.Is(err, ErrEmptyCart):
			response.Error(w, http.StatusConflict, "no items in the cart")
			return
		case errors.Is(err, ErrNotEnoughStock):
			response.Error(w, http.StatusConflict, "your requested quantity is more than is available")
			return
		default:
			response.Error(w, http.StatusInternalServerError, "internal server error")
			return
		}
	}

	response.JSON(w, http.StatusCreated, map[string]any{"order": OrderResult})
}
