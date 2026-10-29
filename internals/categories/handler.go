package categories

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

func (h *handler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req CategoryRequest

	if !request.Decode(w, r, &req, 1<<10) {
		return
	}

	if err := h.service.CreateCategory(r.Context(), req.Name); err != nil {
		switch {
		case errors.Is(err, ErrConflictError):
			response.Error(w, http.StatusConflict, "category already exist")
			return
		default:
			response.Error(w, http.StatusInternalServerError, "internal server error")
			return
		}
	}

	response.JSON(w, http.StatusCreated, map[string]string{"message": "catgoery created successfully"})
}
