package auth

import (
	"encoding/json"
	"errors"
	"time"

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
var maxErr *http.MaxBytesError

func (h *handler) SignUp(w http.ResponseWriter, r *http.Request) {
	var req SignUpRequest

	r.Body = http.MaxBytesReader(w, r.Body, 5<<10)

	//decode
	err := json.NewDecoder(r.Body).Decode(&req)

	if errors.As(err, &maxErr) {
		response.Error(w, http.StatusRequestEntityTooLarge, "request body too large")
		return
	}

	if err != nil {
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

	var verifiedAt *time.Time
	if user.EmailVerifiedAt.Valid {
		verifiedAt = &user.EmailVerifiedAt.Time
	}

	response.JSON(w, http.StatusCreated, UserResponse{
		Name:              user.Name,
		Email:             user.Email,
		Email_verified_at: verifiedAt,
	})
}

func (h *handler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	var req ResendVerificationRequest

	//decode
	r.Body = http.MaxBytesReader(w, r.Body, 254)
	err := json.NewDecoder(r.Body).Decode(&req)

	if errors.As(err, &maxErr) {
		response.Error(w, http.StatusRequestEntityTooLarge, "request body too large")
		return
	}

	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	//validate
	if err := validate.Struct(req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid input")
		return
	}

	if err := h.service.ResendVerification(r.Context(), req); err != nil {
		switch {
		case errors.Is(err, ErrNoUser):
			response.Error(w, http.StatusForbidden, "only a signed up user can request another verification email")
			return
		case errors.Is(err, ErrTooSoon):
			response.Error(w, http.StatusTooEarly, "wait a full minute before requesting another token")
			return
		
		case errors.Is(err, ErrUserVerified):
			response.Error(w, http.StatusConflict, "user is already verified")
			return

		default:
			response.Error(w, http.StatusInternalServerError, "internal server error")
			return
		}
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "another verification has been sent"})
}

func (h *handler) VerifyEmail(w http.ResponseWriter, r *http.Request) {

	var req VerifyEmailRequest

	req.Token = r.URL.Query().Get("token")

	//validate
	if err := validate.Struct(req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid input")
		return
	}

	if err := h.service.VerifyEmail(r.Context(), req); err != nil {
		switch {
		case errors.Is(err, ErrInvalidToken):
			response.Error(w, http.StatusBadRequest, "invalid or expired token")
			return
		default:
			response.Error(w, http.StatusInternalServerError, "internal server error")
		}
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "email verification successful"})
}