package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type AccessTokenValidator interface {
	GetUserIDFromToken(token string) (uuid.UUID, error)
}

type contextKey struct{}

var userIDKey contextKey

func NewUserAuthenticator(
	accessToken AccessTokenValidator,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			parts := strings.Fields(authHeader)

			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				writeUnauthorized(w)
				return
			}

			token := parts[1]

			userID, err := accessToken.GetUserIDFromToken(token)

			if err != nil {
				writeUnauthorized(w)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(userIDKey).(uuid.UUID)
	return userID, ok
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Bearer`)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"error":"authentication required"}` + "\n"))
}
