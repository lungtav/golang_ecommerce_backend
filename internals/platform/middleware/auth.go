package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/lungtav/golang_ecommerce_backend/internals/platform/response"
	"github.com/lungtav/golang_ecommerce_backend/internals/platform/tokens"
)

type ctxKey string

const userIdKey ctxKey = "userId"

func Auth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request)  {
			header:= r.Header.Get("Authorization")

			raw, found:= strings.CutPrefix(header, "Bearer ")

			if !found{
				response.Error(w, http.StatusUnauthorized, "missing or invalid token")
				return
			}

			claims, err:=tokens.ParseAccessToken(raw, secret)

			if err!=nil{
				response.Error(w, http.StatusUnauthorized, "inalid or expired token")
				return 
			}

			ctx:= context.WithValue(r.Context(), userIdKey, claims.Subject)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserId (ctx context.Context) (string, bool){
	id, ok:= ctx.Value(userIdKey).(string)
	return id , ok
}