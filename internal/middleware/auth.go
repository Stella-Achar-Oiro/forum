package middleware

import (
	"database/sql"
	"net/http"

	"forum/internal/auth"
)

const (
	sessionCookie = "session_token"
)

// RequireAuth middleware ensures the user is authenticated
func RequireAuth(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(sessionCookie)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			user, err := auth.ValidateSession(db, cookie.Value)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Store user in request context
			ctx := r.Context()
			ctx = auth.WithUser(ctx, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuth middleware adds user to context if authenticated but doesn't require it
func OptionalAuth(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(sessionCookie)
			if err == nil {
				if user, err := auth.ValidateSession(db, cookie.Value); err == nil {
					ctx := r.Context()
					ctx = auth.WithUser(ctx, user)
					r = r.WithContext(ctx)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
