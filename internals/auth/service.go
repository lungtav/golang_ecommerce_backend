package auth

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/lungtav/golang_ecommerce_backend/internals/db"
	"github.com/lungtav/golang_ecommerce_backend/internals/platform/password"
)

type Service interface {
	SignUp(ctx context.Context, req SignUpRequest) (db.User, error)
}

type svc struct {
	queries *db.Queries
}

func NewService(queries *db.Queries) Service {
	return &svc{queries: queries}
}

var ErrEmailTaken = errors.New("email is existing")
var ErrServerError = errors.New("internal server error")

func (s *svc) SignUp(ctx context.Context, req SignUpRequest) (db.User, error) {

	email := strings.ToLower(strings.TrimSpace(req.Email))
	_, err := s.queries.GetUserByEmail(ctx, email)

	if err == nil {
		return db.User{}, ErrEmailTaken
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return db.User{}, ErrServerError
	}

	//hash password
	hashedPassword, err := password.Hash(req.Password)

	if err != nil {
		slog.Error("hash password failed", "error", err)
		return db.User{}, ErrServerError
	}

	//create user
	user, err := s.queries.CreateUser(ctx, db.CreateUserParams{Email: email, PasswordHash: hashedPassword, Name: req.Name})

	if err != nil {
		slog.Error("create user failed", "error", err)
		return db.User{}, ErrServerError
	}

	return user, nil
}
