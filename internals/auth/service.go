package auth

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
	SignIn(ctx context.Context, req SignInRequest) (u db.GetUserByEmailRow, accesstoken, refreshToken string, err error)
}

type svc struct {
	queries   *db.Queries
	mailer    *mailer.Mailer
	jwtSecret string
	frontendUrl string
}

func NewService(queries *db.Queries, mailer *mailer.Mailer, jwtSecret string, frontendUrl string) Service {
	return &svc{
		queries:   queries,
		mailer:    mailer,
		jwtSecret: jwtSecret,
		frontendUrl: frontendUrl,
	}
}

var ErrEmailTaken = errors.New("email is existing")
var ErrInvalidCredentials = errors.New("incorrect email or password")
var ErrServerError = errors.New("internal server error")
var ErrNoUser = errors.New("no user associated with this email")
var ErrTooSoon = errors.New("a full minute before another token can be assigned")
var ErrInvalidToken = errors.New("token is invalid")
var ErrUserVerified = errors.New("this user is already verified")
var ErrUserNotVerified = errors.New("email has to be verified before login")

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
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505"{
			return db.User{}, ErrEmailTaken
		}

		slog.Error("create user failed", "error", err)
		return db.User{}, ErrServerError
	}

	//create link
	rawToken, hashedToken, err := tokens.GenerateToken()

	if err!=nil{
		slog.Error("generate verification token failed", "error", err)
		return db.User{}, ErrServerError
	}

	//save hashed token to db
	if err:= s.queries.CreateVerificationToken(ctx, db.CreateVerificationTokenParams{TokenHash: hashedToken, UserID: user.ID}); err!=nil{
		return db.User{}, ErrServerError
	}

	url := s.frontendUrl + `/verify-email?token=` + rawToken

	//send email
	go func(){
		if _, err:=s.mailer.SendEmailVerification(req.Name, email, url); err!=nil{
			slog.Error("send verification email failed", "error", err)
		}
	}()
		
	return user, nil
}

func (s *svc) ResendVerification(ctx context.Context, req ResendVerificationRequest) error {
	email := normalize.Email(req.Email)

	user, err := s.queries.GetUserByEmail(ctx, email)

	if user.EmailVerifiedAt.Valid {
		return ErrUserVerified
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
	raw, hash, err := tokens.GenerateToken()

	url := env.Get("FRONTEND_URL", "") + `/verify-email?token=` + raw

	if err := s.queries.CreateVerificationToken(ctx, db.CreateVerificationTokenParams{
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
	if err = s.queries.UpdateVerificationTokensById(ctx, storedToken.ID); err != nil {
		return ErrServerError
	}

	//verify email
	if err := s.queries.UpdateUserById(ctx, storedToken.UserID); err != nil {
		return ErrServerError
	}

	return nil
}

func (s *svc) SignIn(ctx context.Context, req SignInRequest) (u db.GetUserByEmailRow, accesstoken, refreshToken string, err error) {

	email := normalize.Email(req.Email)
	user, err := s.queries.GetUserByEmail(ctx, email)

	if errors.Is(err, pgx.ErrNoRows) {
		return db.GetUserByEmailRow{}, "", "", ErrNoUser
	}

	if err != nil {
		return db.GetUserByEmailRow{}, "", "", ErrServerError
	}

	if !user.EmailVerifiedAt.Valid {
		return db.GetUserByEmailRow{}, "", "", ErrUserNotVerified
	}

	//compare password
	isPasswordMatch := password.Compare(req.Password, user.PasswordHash)

	if !isPasswordMatch {
		return db.GetUserByEmailRow{}, "", "", ErrInvalidCredentials
	}

	//create refresh and access tokens
	accessToken, err := tokens.GenerateAccessToken(user.ID.String(), s.jwtSecret)
	if err != nil {
		slog.Error("failed to generate access token", "error", err)
		return db.GetUserByEmailRow{}, "", "", ErrServerError
	}

	rawRefreshToken, hashRefreshToken, err := tokens.GenerateToken()
	if err != nil {
		slog.Error("failed to generate refresh token", "error", err)
		return db.GetUserByEmailRow{}, "", "", ErrServerError
	}

	//insert refresh token hash to db
	err = s.queries.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
		TokenHash: hashRefreshToken,
		UserID:    user.ID,
	})

	if err != nil {
		slog.Error("failed to insert refresh token into db", "error", err)
		return db.GetUserByEmailRow{}, "", "", ErrServerError
	}

	return user, accessToken, rawRefreshToken, nil
}
