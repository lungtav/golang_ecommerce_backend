package auth

import (
	"time"
)

type SignUpRequest struct {
	Email           string `json:"email" validate:"required,email"`
	Name            string `json:"name" validate:"required,min=2"`
	Password        string `json:"password" validate:"required,min=8,max=72"`
	ConfirmPassword string `json:"confirm_password" validate:"eqfield=Password"`
}

type SignInRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

type SignInResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
}

type UserResponse struct {
	Name              string     `json:"name"`
	Email             string     `json:"email"`
	EmailVerifiedAt *time.Time `json:"email_verified_at"`
}

type ResendVerificationRequest struct {
	Email string `json:"email" validate:"required,email,max=254"`
}

type VerifyEmailRequest struct {
	Token string `json:"token" validate:"required"`
}
