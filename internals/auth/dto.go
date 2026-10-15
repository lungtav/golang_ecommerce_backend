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

type UserResponse struct {
	Name              string     `json:"name"`
	Email             string     `json:"email"`
	Email_verified_at *time.Time `json:"email_verified_at"`
}

type ResendVerificationRequest struct {
	Email string `json:"email" validate:"required,email,max=254"`
}

type VerifyEmailRequest struct {
	Token string `json:"token" validate:"required"`
}
