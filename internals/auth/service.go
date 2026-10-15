package auth

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/lungtav/golang_ecommerce_backend/internals/db"
	"github.com/lungtav/golang_ecommerce_backend/internals/platform/env"
	"github.com/lungtav/golang_ecommerce_backend/internals/platform/mailer"
	"github.com/lungtav/golang_ecommerce_backend/internals/platform/normalize"
	"github.com/lungtav/golang_ecommerce_backend/internals/platform/password"
	"github.com/lungtav/golang_ecommerce_backend/internals/platform/tokens"
)

type Service interface {
	SignUp(ctx context.Context, req SignUpRequest) (db.User, error)
	ResendVerification(ctx context.Context, req ResendVerificationRequest) error
	VerifyEmail(ctx context.Context, req VerifyEmailRequest) error
}

type svc struct {
	queries *db.Queries
	mailer  *mailer.Mailer
}

func NewService(queries *db.Queries, mailer *mailer.Mailer) Service {
	return &svc{
		queries: queries,
		mailer:  mailer,
	}
}

var ErrEmailTaken = errors.New("email is existing")
var ErrServerError = errors.New("internal server error")
var ErrNoUser = errors.New("no user associated with this email")
var ErrTooSoon = errors.New("a full minute before another token can be assigned")
var ErrInvalidToken = errors.New("token is invalid")
var ErrUserVerified = errors.New("this user is already verified")

const (
	VerificationTTL      = 15 * time.Minute // token lifetime
	VerificationCooldown = 1 * time.Minute  // min gap between resends
)

func (s *svc) SignUp(ctx context.Context, req SignUpRequest) (db.User, error) {

	email := normalize.Email(req.Email)
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

	//create link
	rawToken, hashedToken, err := tokens.GenerateVerificationToken()

	//save hashed token to db
	s.queries.CreateVerficationToken(ctx, db.CreateVerficationTokenParams{TokenHash: hashedToken, UserID: user.ID})

	url := env.Get("FRONTEND_URL", "") + `/verify-email?token=` + rawToken

	//send email
	s.mailer.SendEmailVerification(req.Name, email, url)

	return user, nil
}

func (s *svc) ResendVerification(ctx context.Context, req ResendVerificationRequest) error {
	email := normalize.Email(req.Email)

	user, err := s.queries.GetUserByEmail(ctx, email)

	if user.EmailVerifiedAt.Valid {
		return  ErrUserVerified
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNoUser
	}

	if err != nil {
		return ErrServerError
	}

	token, err := s.queries.GetVerificationToken(ctx, user.ID)

	if err != nil {
		return ErrServerError
	}

	//check cooldown

	if time.Since(token.CreatedAt.Time) < VerificationCooldown {
		return ErrTooSoon
	}

	//delete the old token
	if err := s.queries.DeleteVerificationToken(ctx, token.TokenHash); err != nil {
		return ErrServerError
	}

	//create new token
	raw, hash, err := tokens.GenerateVerificationToken()

	url := env.Get("FRONTEND_URL", "") + `/verify-email?token=` + raw

	if err := s.queries.CreateVerficationToken(ctx, db.CreateVerficationTokenParams{
		TokenHash: hash,
		UserID:    user.ID,
	}); err != nil {
		return ErrServerError
	}

	//send to the user
	s.mailer.SendEmailVerification(user.Name, email, url)

	return nil
}

func (s *svc) VerifyEmail(ctx context.Context, req VerifyEmailRequest) error {
	hashedToken := tokens.HashToken(req.Token)

	storedToken, err := s.queries.GetVerificationTokenByHash(ctx, hashedToken)

	if errors.Is(err, pgx.ErrNoRows) {
		return ErrInvalidToken
	}

	if err != nil {
		return ErrServerError
	}

	//if expired return
	if time.Now().After(storedToken.ExpiresAt.Time) {
		return ErrInvalidToken
	}

	//if used, return
	if storedToken.UsedAt.Valid {
		return ErrInvalidToken
	}

	//mark the token as used
	if err=s.queries.UpdateVerificationTokensById(ctx, storedToken.ID); err!=nil{
		return ErrServerError
	}

	//verify email
	if err := s.queries.UpdateUserById(ctx, storedToken.UserID); err != nil {
		return ErrServerError
	}

	return nil
}
