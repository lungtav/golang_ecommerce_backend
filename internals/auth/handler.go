package auth

import (
	"errors"
	"time"

	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/lungtav/golang_ecommerce_backend/internals/platform/request"
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

	if !request.Decode(w, r, &req, 5<<10) {
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
		Name:            user.Name,
		Email:           user.Email,
		EmailVerifiedAt: verifiedAt,
	})
}

func (h *handler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	var req ResendVerificationRequest

	if !request.Decode(w, r, &req, 1<<10) {
		return
	}

	if err := h.service.ResendVerification(r.Context(), req); err != nil {
		switch {
		case errors.Is(err, ErrNoUser):
			response.Error(w, http.StatusForbidden, "only a signed up user can request another verification email")
		case errors.Is(err, ErrTooSoon):
			response.Error(w, http.StatusTooEarly, "wait a full minute before requesting another token")
		case errors.Is(err, ErrUserVerified):
			response.Error(w, http.StatusConflict, "user is already verified")
		default:
			response.Error(w, http.StatusInternalServerError, "internal server error")
			slog.Error("failed to resend verification", "error", err)
		}
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "another verification has been sent"})
}

func (h *handler) VerifyEmail(w http.ResponseWriter, r *http.Request) {

	var req VerifyEmailRequest

	req.Token = r.URL.Query().Get("token")

	//validate
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid input")
		return
	}

	if err := h.service.VerifyEmail(r.Context(), req); err != nil {
		switch {
		case errors.Is(err, ErrInvalidToken):
			response.Error(w, http.StatusBadRequest, "invalid or expired token")
		default:
			response.Error(w, http.StatusInternalServerError, "internal server error")
			slog.Error("error verifying email", "error", err)
		}
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "email verification successful"})
}

func (h *handler) SignIn(w http.ResponseWriter, r *http.Request) {
	var req SignInRequest

	if !request.Decode(w, r, &req, 5<<10) {
		return
	}

	user, accessToken, refreshToken, err := h.service.SignIn(r.Context(), req)

	if err != nil {
		switch {
		case errors.Is(err, ErrNoUser):
			response.Error(w, http.StatusUnauthorized, "invalid email or password")
		case errors.Is(err, ErrInvalidCredentials):
			response.Error(w, http.StatusUnauthorized, "invalid email or password")
		case errors.Is(err, ErrUserNotVerified):
			response.Error(w, http.StatusForbidden, "verify your email first")
		default:
			slog.Error("sign in failed", "error", err)
			response.Error(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	//refresh and access token sent in res body as theres no frontend
	response.JSON(w, http.StatusOK, SignInResponse{
		User: UserResponse{
			Name:            user.Name,
			Email:           user.Email,
			EmailVerifiedAt: &user.EmailVerifiedAt.Time,
		},
		Tokens: RefreshResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
	})
}

func (h *handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if !request.Decode(w, r, &req, 1<<10) {
		return
	}

	accessToken, refreshToken, err := h.service.Refresh(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidToken):
			response.Error(w, http.StatusUnauthorized, "invalid or expired refresh token")
		default:
			slog.Error("refresh failed", "error", err)
			response.Error(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	response.JSON(w, http.StatusOK, RefreshResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

func (h *handler) Logout(w http.ResponseWriter, r *http.Request) {
	var req LogoutRequest
	if !request.Decode(w, r, &req, 1<<10) {
		return
	}

	if err := h.service.Logout(r.Context(), req); err != nil {
		slog.Error("logout failed", "error", err)
		response.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}
