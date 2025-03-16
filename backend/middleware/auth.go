// backend/middleware/auth.go
package middleware

import (
    "context"
    "database/sql"
    "net/http"
    "forum/backend/models"
)

// AuthMiddleware checks for valid session and adds user ID to request context
func AuthMiddleware(db *sql.DB, next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Get session cookie
        cookie, err := r.Cookie("session_id")
        if err != nil {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        
        // Get session
        session, err := models.GetSessionByID(db, cookie.Value)
        if err != nil {
            http.Error(w, "Invalid session", http.StatusUnauthorized)
            return
        }
        
        // Add user ID to request context
        ctx := context.WithValue(r.Context(), "userID", session.UserID)
        r = r.WithContext(ctx)
        
        // Call the next handler
        next(w, r)
    }
}

// GetUserID retrieves user ID from request context
func GetUserID(r *http.Request) (int, bool) {
    userID, ok := r.Context().Value("userID").(int)
    return userID, ok
}