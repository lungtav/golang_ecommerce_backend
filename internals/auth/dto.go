package auth

import "github.com/jackc/pgx/v5/pgtype"

type SignUpRequest struct {
	Email           string `json:"email" validate:"required,email"`
	Name            string `json:"name" validate:"required,min=2"`
	Password        string `json:"password" validate:"required,min=8,max=72"`
	ConfirmPassword string `json:"confirm_password" validate:"eqfield=Password"`
}

type UserResponse struct {
	Name              string             `json:"name"`
	Email             string             `json:"email"`
	Email_verified_at pgtype.Timestamptz `json:"email_verified_at"`
}

type ResendVerificationRequest struct {
	Email string `json:"email" validate:"required,email"`
}
