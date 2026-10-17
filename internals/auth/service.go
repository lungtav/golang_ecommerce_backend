package auth

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lungtav/golang_ecommerce_backend/internals/db"
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
	Refresh(ctx context.Context, req RefreshRequest) (string, string, error)
}

type svc struct {
	queries     *db.Queries
	mailer      *mailer.Mailer
	jwtSecret   string
	frontendUrl string
	pool        *pgxpool.Pool
}

func NewService(queries *db.Queries, mailer *mailer.Mailer, jwtSecret, frontendUrl string, pool *pgxpool.Pool) Service {
	return &svc{
		queries:     queries,
		mailer:      mailer,
		jwtSecret:   jwtSecret,
		frontendUrl: frontendUrl,
		pool:        pool,
	}
}

var (
	ErrEmailTaken         = errors.New("email is existing")
	ErrInvalidCredentials = errors.New("incorrect email or password")
	ErrServerError        = errors.New("internal server error")
	ErrNoUser             = errors.New("no user associated with this email")
	ErrTooSoon            = errors.New("a full minute before another token can be assigned")
	ErrInvalidToken       = errors.New("token is invalid")
	ErrUserVerified       = errors.New("this user is already verified")
	ErrUserNotVerified    = errors.New("email has to be verified before login")
)

const (
	VerificationTTL      = 15 * time.Minute
	VerificationCooldown = 1 * time.Minute
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
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return db.User{}, ErrEmailTaken
		}

		slog.Error("create user failed", "error", err)
		return db.User{}, ErrServerError
	}

	//create link
	rawToken, hashedToken, err := tokens.GenerateToken()

	if err != nil {
		slog.Error("generate verification token failed", "error", err)
		return db.User{}, ErrServerError
	}

	//save hashed token to db
	if err := s.queries.CreateVerificationToken(ctx, db.CreateVerificationTokenParams{TokenHash: hashedToken, UserID: user.ID}); err != nil {
		return db.User{}, ErrServerError
	}

	url := s.frontendUrl + `/verify-email?token=` + rawToken

	//send email
	go func() {
		if _, err := s.mailer.SendEmailVerification(req.Name, email, url); err != nil {
			slog.Error("send verification email failed", "error", err)
		}
	}()

	return user, nil
}

func (s *svc) ResendVerification(ctx context.Context, req ResendVerificationRequest) error {
	email := normalize.Email(req.Email)

	user, err := s.queries.GetUserByEmail(ctx, email)

	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNoUser
	}

	if err != nil {
		slog.Error("get user by email failed", "error", err)
		return ErrServerError
	}

	if user.EmailVerifiedAt.Valid {
		return ErrUserVerified
	}

	token, err := s.queries.GetVerificationToken(ctx, user.ID)

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		slog.Error("get verification token failed", "error", err)
		return ErrServerError
	}

	//check cooldown
	if token.CreatedAt.Valid && time.Since(token.CreatedAt.Time) < VerificationCooldown {
		return ErrTooSoon
	}

	//create new token
	raw, hash, err := tokens.GenerateToken()

	if err != nil {
		slog.Error("create verification token failed", "error", err)
		return ErrServerError
	}

	url := s.frontendUrl + `/verify-email?token=` + raw

	tx, err := s.pool.Begin(ctx)

	if err != nil {
		slog.Error("begin tx failed", "error", err)
		return ErrServerError
	}
	defer tx.Rollback(ctx)

	qtx:=s.queries.WithTx(tx)


	//delete the old token if one exists
	if err := qtx.DeleteVerificationToken(ctx, user.ID); err != nil {
		slog.Error("delete verification tokens failed", "error", err)
		return ErrServerError
	}

	if err := qtx.CreateVerificationToken(ctx, db.CreateVerificationTokenParams{
		TokenHash: hash,
		UserID:    user.ID,
	}); err != nil {
		slog.Error("create verification token failed", "error", err)
		return ErrServerError
	}

	if err := tx.Commit(ctx); err!=nil{
		slog.Error("failed to commit transaction", "error", err)
		return ErrServerError
	}

	//send to the user
	go func() {
		if _, err := s.mailer.SendEmailVerification(user.Name, email, url); err != nil {
			slog.Error("send verification email failed", "error", err)
		}
	}()
	return nil
}

func (s *svc) VerifyEmail(ctx context.Context, req VerifyEmailRequest) error {
	hashedToken := tokens.HashToken(req.Token)

	storedToken, err := s.queries.GetVerificationTokenByHash(ctx, hashedToken)

	if errors.Is(err, pgx.ErrNoRows) {
		return ErrInvalidToken
	}

	if err != nil {
		slog.Error("get verification token by hash failed", "error", err)
		return ErrServerError
	}

	//if expired return	
	if !storedToken.ExpiresAt.Valid || time.Now().After(storedToken.ExpiresAt.Time) {
		return ErrInvalidToken
	}

	//if used, return
	if storedToken.UsedAt.Valid {
		return ErrInvalidToken
	}
	tx, err := s.pool.Begin(ctx)

	if err != nil {
		slog.Error("begin tx failed", "error", err)
		return ErrServerError
	}

	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)

	//mark the token as used
	n, err := qtx.UpdateVerificationTokensById(ctx, storedToken.ID)
	if err != nil {
		slog.Error("failed to mark verification token as used", "error", err)
		return ErrServerError
	}

	if n == 0 {
		return ErrInvalidToken
	}

	//verify email
	if err := qtx.MarkEmailVerified(ctx, storedToken.UserID); err != nil {
		slog.Error("mark email as verified failed", "error", err)
		return ErrServerError
	}

	if err := tx.Commit(ctx); err != nil {
		slog.Error("failed to commit transaction", "error", err)
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
		slog.Error("get user by email failed", "error", err)
		return db.GetUserByEmailRow{}, "", "", ErrServerError
	}

	//compare password
	isPasswordMatch := password.Compare(req.Password, user.PasswordHash)

	if !isPasswordMatch {
		return db.GetUserByEmailRow{}, "", "", ErrInvalidCredentials
	}

	if !user.EmailVerifiedAt.Valid {
		return db.GetUserByEmailRow{}, "", "", ErrUserNotVerified
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

func (s *svc) Refresh(ctx context.Context, req RefreshRequest) (string, string, error) {
	hash := tokens.HashToken(req.RefreshToken)

	stored, err := s.queries.GetRefreshToken(ctx, hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", ErrInvalidToken
	}
	if err != nil {
		slog.Error("get refresh token failed", "error", err)
		return "", "", err
	}

	if stored.RevokedAt.Valid {
		return "", "", ErrInvalidToken
	}
	if !stored.ExpiresAt.Valid || time.Now().After(stored.ExpiresAt.Time) {
		return "", "", ErrInvalidToken
	}

	// new tokens are generated before the transaction, no DB involved
	newRaw, newHash, err := tokens.GenerateToken()
	if err != nil {
		slog.Error("generate refresh token failed", "error", err)
		return "", "", err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		slog.Error("begin tx failed", "error", err)
		return "", "", err
	}
	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)

	// revoke the old token; n == 0 means another request already used it
	n, err := qtx.RevokeRefreshToken(ctx, stored.ID)
	if err != nil {
		slog.Error("revoke refresh token failed", "error", err)
		return "", "", err
	}
	if n == 0 {
		return "", "", ErrInvalidToken
	}

	err = qtx.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
		TokenHash: newHash,
		UserID:    stored.UserID,
	})
	if err != nil {
		slog.Error("create refresh token failed", "error", err)
		return "", "", err
	}

	if err := tx.Commit(ctx); err != nil {
		slog.Error("commit failed", "error", err)
		return "", "", err
	}

	accessToken, err := tokens.GenerateAccessToken(stored.UserID.String(), s.jwtSecret)
	if err != nil {
		slog.Error("generate access token failed", "error", err)
		return "", "", err
	}

	return accessToken, newRaw, nil
}