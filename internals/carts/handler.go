package carts

import (
	"errors"
	"net/http"

	"github.com/lungtav/golang_ecommerce_backend/internals/platform/middleware"
	"github.com/lungtav/golang_ecommerce_backend/internals/platform/request"
	"github.com/lungtav/golang_ecommerce_backend/internals/platform/response"
)

type handler struct {
	service Service
}

func NewHandler(service Service) *handler {
	return &handler{service: service}
}

func (h *handler) AddToCart(w http.ResponseWriter, r *http.Request) {
	var req AddToCartRequest

	if !request.Decode(w, r, &req, 1<<10) {
		return
	}

	userId, ok := middleware.UserId(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauhorixed")
		return
	}

	if err := h.service.AddToCart(r.Context(), req, userId); err != nil {
		switch {
		case errors.Is(err, ErrNoProduct):
			response.Error(w, http.StatusNotFound, "product doesn't exist")
			return
		case errors.Is(err, ErrOutOfStock):
			response.JSON(w, http.StatusConflict, map[string]string{"message": "product out of stock"})
			return
		case errors.Is(err, ErrNotEnoughStock):
			response.JSON(w, http.StatusConflict, map[string]string{"message": "quantity to buy exceeds product stock"})
			return
		default:
			response.Error(w, http.StatusInternalServerError, "internal server error")
		}
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "added to cart"})
}

func (h *handler) GetCartItems(w http.ResponseWriter, r *http.Request) {
	userId, ok := middleware.UserId(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauhorixed")
		return
	}

	cart, err := h.service.GetCartItems(r.Context(), userId)

	if err != nil {
		switch {
		case errors.Is(err, ErrNoCartItem):
			response.JSON(w, http.StatusConflict, map[string]string{"message": "no item in cart"})
			return
		default:
			response.Error(w, http.StatusInternalServerError, "internal server error")
			return
		}
	}

	response.JSON(w, http.StatusOK, map[string]any{"cart": cart})
}
