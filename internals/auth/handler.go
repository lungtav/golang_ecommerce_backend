package auth

import (
	"encoding/json"
	"errors"

	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lungtav/golang_ecommerce_backend/internals/platform/response"
)

type handler struct {
	service Service
}

func NewHandler(service Service) *handler {
	return &handler{service: service}
}

func (h *handler) SignUp(w http.ResponseWriter, r *http.Request) {
	var req SignUpRequest

	//decode
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("an error occured while decoding the request", "error", err)
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	//validate
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid input")
		return
	}

	user, err := h.service.SignUp(r.Context(), req)

	if err != nil {
		switch {
		case errors.Is(err, ErrEmailTaken):
			response.Error(w, http.StatusConflict, err.Error())

		default:
			response.Error(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	formattedUserResponse := struct {
		Name              string             `json:"name"`
		Email             string             `json:"email"`
		Email_verified_at pgtype.Timestamptz `json:"email_verified_at"`
	}{
		Name:              user.Name,
		Email:             user.Email,
		Email_verified_at: user.EmailVerifiedAt,
	}

	response.JSON(w, 201, formattedUserResponse)
}
