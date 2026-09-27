package server

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type TokenVerifier interface {
	Verify(token string) (uuid.UUID, error)
}

type userIDContextKey struct{}

func RequireAuth(verifier TokenVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			const prefix = "Bearer "

			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, prefix) {
				writeError(w, http.StatusUnauthorized, "unauthorized", "missing bearer token")
				return
			}

			userID, err := verifier.Verify(strings.TrimPrefix(authHeader, prefix))
			if err != nil {
				writeError(w, http.StatusUnauthorized, "unauthorized", "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), userIDContextKey{}, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func userIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(userIDContextKey{}).(uuid.UUID)
	return id, ok
}
