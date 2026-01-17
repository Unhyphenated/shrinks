package auth

import (
	"strings"
	"errors"
	"net/http"
	"context"

	"github.com/Unhyphenated/shrinks-backend/internal/util"
)

type contextKey string

const claimsContextKey contextKey = "claims"

func GetClaimsFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(claimsContextKey).(*Claims)
	return claims, ok
}

func extractToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("missing Authorization header")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("invalid Authorization header")
	}

	token := parts[1]

	return token, nil
}

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString, err := extractToken(r)
		if err != nil {
			util.WriteError(w, http.StatusUnauthorized, "Unauthorized: " + err.Error())
			return
		}

		claims, err := ValidateToken(tokenString)
		if err != nil {
			util.WriteError(w, http.StatusUnauthorized, "Unauthorized: " + err.Error())
			return
		}

		ctx := context.WithValue(r.Context(), claimsContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString, err := extractToken(r)
		if err == nil {
			// Token exists, try to validate it
			claims, err := ValidateToken(tokenString)
			if err == nil {
				// Valid token, add to context
				ctx := context.WithValue(r.Context(), claimsContextKey, claims)
				r = r.WithContext(ctx)
			}
		}

		next.ServeHTTP(w, r)
	})
}