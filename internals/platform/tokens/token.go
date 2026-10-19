package tokens

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const AccessTokenTTL = time.Minute * 15
const RefreshTokenTTL = time.Hour * 24 * 5

func GenerateToken() (raw, hash string, err error) {
	buf := make([]byte, 32)

	if _, err = rand.Read(buf); err != nil {
		return "", "", err
	}

	raw = base64.RawURLEncoding.EncodeToString(buf)
	hash = HashToken(raw)
	return raw, hash, nil
}

func HashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func GenerateAccessToken(userID string, secret string) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenTTL)),
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(secret))
}

func ParseAccessToken(raw, secret string) (*jwt.RegisteredClaims, error){
	if raw == "" || secret == "" {
		return nil, errors.New("token and secret are required")
	}

	claims := &jwt.RegisteredClaims{}

	token, err := jwt.ParseWithClaims(
		raw,
		claims,
		func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, errors.New("unexpected signing method")
			}

			return []byte(secret), nil
		},
		jwt.WithValidMethods([]string{
			jwt.SigningMethodHS256.Alg(),
		}),
	)

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid access token")
	}

	
	if claims.Subject == "" {
		return nil, errors.New("missing token subject")
	}

	if claims.ExpiresAt == nil {
		return nil, errors.New("missing token expiration")
	}

	return claims, nil
}