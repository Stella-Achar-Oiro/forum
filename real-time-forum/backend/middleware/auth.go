package middleware

import (
	"context"
	"net/http"

	"real-time-forum/backend/database"
	"real-time-forum/backend/handlers"
	"real-time-forum/backend/models"
)

type contextKey string

const UserContextKey contextKey = "user"

// AuthMiddleware checks for valid session token and adds user to request context
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get token from Authorization header
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Validate session
		session, valid := handlers.GetSession(token)
		if !valid {
			http.Error(w, "Invalid or expired session", http.StatusUnauthorized)
			return
		}

		// Get user from database
		user, err := models.GetUserByID(database.GetDB(), session.UserID)
		if err != nil {
			http.Error(w, "User not found", http.StatusUnauthorized)
			return
		}

		// Add user to request context
		ctx := context.WithValue(r.Context(), UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserFromContext retrieves user from request context
func GetUserFromContext(ctx context.Context) (*models.User, bool) {
	user, ok := ctx.Value(UserContextKey).(*models.User)
	return user, ok
}

// RequireAuth is a middleware that ensures a route is only accessible to authenticated users
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := GetUserFromContext(r.Context())
		if !ok || user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	}
}
