package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"project03/internal/domain/service"
)

type contextKey struct{}

var userIDKey contextKey

func NewUserAuthenticator(
	userService service.UserService,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			login, password, ok := r.BasicAuth()
			if !ok {
				writeUnauthorized(w)
				return
			}

			user, err := userService.Authenticate(
				r.Context(),
				login,
				password,
			)
			if err != nil {
				writeUnauthorized(w)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, user.ID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(userIDKey).(uuid.UUID)
	return userID, ok
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="tictactoe"`)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"error":"authentication required"}` + "\n"))
}