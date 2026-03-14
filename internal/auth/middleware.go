package auth

import (
	"context"
	"net/http"
	"strings"
)

type ContextKey string

const TokenIDKey ContextKey = "tokenID"

func Middleware(tm *TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}

			token, err := tm.ValidateToken(parts[1])
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			if !token.Active {
				http.Error(w, "Token is disabled", http.StatusForbidden)
				return
			}

			if r.Method == http.MethodPost {
				if !tm.CheckRateLimit(token.ID) {
					http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
					return
				}
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, TokenIDKey, token.ID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetTokenID(ctx context.Context) (uint, bool) {
	val, ok := ctx.Value(TokenIDKey).(uint)
	return val, ok
}
