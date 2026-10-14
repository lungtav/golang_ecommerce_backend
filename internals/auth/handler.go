package auth

import (
	"encoding/json"
	"errors"

	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/lungtav/golang_ecommerce_backend/internals/platform/response"
)

type handler struct {
	service Service
}

func NewHandler(service Service) *handler {
	return &handler{service: service}
}

var validate = validator.New()

func (h *handler) SignUp(w http.ResponseWriter, r *http.Request) {
	var req SignUpRequest

	//decode
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	//validate
	if err := validate.Struct(req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid input")
		return
	}

	user, err := h.service.SignUp(r.Context(), req)

	if err != nil {
		switch {
		case errors.Is(err, ErrEmailTaken):
			response.Error(w, http.StatusConflict, "email already exists")

		default:
			slog.Error("sign up failed", "error", err)
			response.Error(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	response.JSON(w, http.StatusCreated, UserResponse{
		Name:              user.Name,
		Email:             user.Email,
		Email_verified_at: user.EmailVerifiedAt,
	})
}

