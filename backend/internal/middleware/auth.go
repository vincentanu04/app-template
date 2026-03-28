package middleware

import (
	"context"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type userIDContextKeyType struct{}

var userIDContextKey = userIDContextKeyType{}

const CookieName = "access_token"

// Auth validates the JWT cookie on all requests except public routes.
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Public routes — bypass auth
		switch r.URL.Path {
		case "/api/auth/register", "/api/auth/login", "/api/auth/logout", "/api/health":
			next.ServeHTTP(w, r)
			return
		}

		cookie, err := r.Cookie(CookieName)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		token, err := jwt.ParseWithClaims(
			cookie.Value,
			&jwt.RegisteredClaims{},
			func(t *jwt.Token) (interface{}, error) {
				return []byte(os.Getenv("JWT_SECRET")), nil
			},
		)
		if err != nil || !token.Valid {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		claims := token.Claims.(*jwt.RegisteredClaims)
		ctx := context.WithValue(r.Context(), userIDContextKey, claims.Subject)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// UserIDFromContext extracts the authenticated user ID from the request context.
func UserIDFromContext(ctx context.Context) uuid.UUID {
	userID, _ := ctx.Value(userIDContextKey).(string)
	id, _ := uuid.Parse(userID)
	return id
}
