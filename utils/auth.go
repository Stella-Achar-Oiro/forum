package utils

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Stella-Achar-Oiro/forum/models"
	"golang.org/x/crypto/bcrypt"
)

const (
	// SessionCookieName is the name of the cookie that stores the session ID
	SessionCookieName = "forum_session"
	// SessionExpirationHours is the number of hours before a session expires
	SessionExpirationHours = 24 // 24 hours
	// StateTokenCookieName is the name of the cookie that stores the OAuth state token
	StateTokenCookieName = "oauth_state"
	// CSRFTokenCookieName is the name of the cookie that stores the CSRF token
	CSRFTokenCookieName = "forum_csrf_token"
	// CSRFHeaderName is the name of the header that should contain the CSRF token
	CSRFHeaderName = "X-CSRF-Token"
	// CSRFTokenExpirationHours is the number of hours before a CSRF token expires
	CSRFTokenExpirationHours = 1 // 1 hour
	// InactivityTimeoutMinutes is the number of minutes of inactivity before a session is considered expired
	InactivityTimeoutMinutes = 30 // 30 minutes
	// SessionInactivityMinutes is the number of minutes of inactivity before a session is considered inactive
	SessionInactivityMinutes = 30
)

// Custom type for context keys to avoid string collisions
type contextKey string

// Context keys
const userIDKey contextKey = "userID"

// CreateSessionCookie creates a session cookie for the user
func CreateSessionCookie(userID int) (http.Cookie, error) {
	// Create a new session
	session, err := models.CreateSession(userID, SessionExpirationHours)
	if err != nil {
		return http.Cookie{}, err
	}

	// Create a cookie to store the session ID
	cookie := http.Cookie{
		Name:     SessionCookieName,
		Value:    session.ID,
		Path:     "/",
		Expires:  session.ExpiresAt,
		HttpOnly: true, // HttpOnly to prevent JS access (XSS protection)
		Secure:   true, // Only send over HTTPS
		SameSite: http.SameSiteStrictMode,
	}

	return cookie, nil
}

// GetUserFromRequest gets the user from the request's session cookie and validates
// that the session is not expired due to inactivity
func GetUserFromRequest(r *http.Request) (models.User, error) {
	// Get the session cookie
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		return models.User{}, errors.New("no session cookie found")
	}

	// Get the session
	session, err := models.GetSessionByID(cookie.Value)
	if err != nil {
		return models.User{}, errors.New("invalid session")
	}

	// Check if session has expired
	if session.ExpiresAt.Before(time.Now()) {
		// Delete the expired session
		_ = models.DeleteSession(session.ID)
		return models.User{}, errors.New("session expired")
	}

	// Check for inactivity timeout
	if session.LastActive.Add(time.Minute * InactivityTimeoutMinutes).Before(time.Now()) {
		// Delete the inactive session
		_ = models.DeleteSession(session.ID)
		return models.User{}, errors.New("session expired due to inactivity")
	}

	// Update last active timestamp
	_ = models.UpdateSessionActivity(session.ID)

	// Get the user
	user, err := models.GetUserByID(session.UserID)
	if err != nil {
		return models.User{}, errors.New("user not found")
	}

	return user, nil
}

// ClearSessionCookie creates a cookie that clears the session
func ClearSessionCookie() http.Cookie {
	return http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
}

// GenerateRandomString generates a random string of the specified length
func GenerateRandomString(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:length]
}

// SetStateToken sets the OAuth state token in a cookie
func SetStateToken(w http.ResponseWriter, state string) {
	http.SetCookie(w, &http.Cookie{
		Name:     StateTokenCookieName,
		Value:    state,
		Path:     "/",
		Expires:  time.Now().Add(15 * time.Minute),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
}

// VerifyStateToken verifies the OAuth state token from the cookie
func VerifyStateToken(r *http.Request, state string) bool {
	cookie, err := r.Cookie(StateTokenCookieName)
	if err != nil {
		return false
	}
	return cookie.Value == state
}

// GenerateCSRFToken generates a new CSRF token and sets it in a cookie
func GenerateCSRFToken(w http.ResponseWriter) string {
	// Generate a random token
	token := GenerateRandomString(32)

	// Set token in cookie
	http.SetCookie(w, &http.Cookie{
		Name:     CSRFTokenCookieName,
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(time.Hour * CSRFTokenExpirationHours),
		HttpOnly: false,                // Needs to be accessible from JavaScript
		Secure:   true,                 // Only send over HTTPS
		SameSite: http.SameSiteLaxMode, // Allow cross-origin requests for form submissions
	})

	return token
}

// VerifyCSRFToken verifies that the CSRF token in the request matches the one in the cookie
func VerifyCSRFToken(r *http.Request) bool {
	// Get token from cookie
	cookie, err := r.Cookie(CSRFTokenCookieName)
	if err != nil {
		return false
	}

	// Get token from header
	headerToken := r.Header.Get(CSRFHeaderName)
	if headerToken == "" {
		// If not in header, check form
		headerToken = r.FormValue("csrf_token")
	}

	// Compare tokens
	return headerToken != "" && cookie.Value == headerToken
}

// CSRFProtection is middleware that enforces CSRF protection for non-GET/HEAD/OPTIONS requests
func CSRFProtection(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip CSRF check for GET, HEAD, OPTIONS methods
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		// Verify CSRF token for all other methods
		if !VerifyCSRFToken(r) {
			http.Error(w, "Invalid CSRF token", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash compares a password with a hash to see if they match
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// SessionManager handles session creation, validation, and expiration
type SessionManager struct {
	// Session configuration
	MaxAge      int
	Secure      bool
	HttpOnly    bool
	SameSite    http.SameSite
	SessionName string
}

// NewSessionManager creates a new session manager with default settings
func NewSessionManager() *SessionManager {
	return &SessionManager{
		MaxAge:      SessionExpirationHours * 60 * 60, // Convert hours to seconds
		Secure:      true,
		HttpOnly:    true,
		SameSite:    http.SameSiteStrictMode,
		SessionName: SessionCookieName,
	}
}

// CreateSession creates a new session and returns the session cookie
func (sm *SessionManager) CreateSession(userID int) (http.Cookie, error) {
	session, err := models.CreateSession(userID, SessionExpirationHours)
	if err != nil {
		return http.Cookie{}, err
	}

	return http.Cookie{
		Name:     sm.SessionName,
		Value:    session.ID,
		Path:     "/",
		MaxAge:   sm.MaxAge,
		HttpOnly: sm.HttpOnly,
		Secure:   sm.Secure,
		SameSite: sm.SameSite,
	}, nil
}

// CreateSessionWithResponse creates a new session for a user and sets a session cookie in the response
func (sm *SessionManager) CreateSessionWithResponse(w http.ResponseWriter, userID int) (string, error) {
	// Create a new session in the database
	session, err := models.CreateSession(userID, SessionExpirationHours)
	if err != nil {
		return "", err
	}

	// Set the session cookie
	cookie := &http.Cookie{
		Name:     SessionCookieName,
		Value:    session.ID,
		Path:     "/",
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, cookie)

	// Generate and set a CSRF token for this session
	csrfToken, err := sm.GenerateCSRFTokenWithResponse(w)
	if err != nil {
		return "", err
	}

	return csrfToken, nil
}

// GetUserFromSession gets the user from the session ID
func (sm *SessionManager) GetUserFromSession(sessionID string) (models.User, error) {
	// Get the session
	session, err := models.GetSessionByID(sessionID)
	if err != nil {
		return models.User{}, errors.New("invalid session")
	}

	// Check if session has expired
	if session.ExpiresAt.Before(time.Now()) {
		// Delete the expired session
		_ = models.DeleteSession(session.ID)
		return models.User{}, errors.New("session expired")
	}

	// Check for inactivity timeout
	if session.LastActive.Add(time.Minute * InactivityTimeoutMinutes).Before(time.Now()) {
		// Delete the inactive session
		_ = models.DeleteSession(session.ID)
		return models.User{}, errors.New("session expired due to inactivity")
	}

	// Update last active timestamp
	_ = models.UpdateSessionActivity(session.ID)

	// Get the user
	user, err := models.GetUserByID(session.UserID)
	if err != nil {
		return models.User{}, errors.New("user not found")
	}

	return user, nil
}

// GetUserSession retrieves a user session based on the session cookie
func (sm *SessionManager) GetUserSession(r *http.Request) (models.Session, error) {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		return models.Session{}, err
	}

	sessionID := cookie.Value
	session, err := models.GetSessionByID(sessionID)
	if err != nil {
		return models.Session{}, err
	}

	// Update the session's last active time
	err = models.UpdateSessionActivity(sessionID)
	if err != nil {
		log.Printf("Failed to update session activity: %v", err)
		// Continue anyway, this is not critical
	}

	return session, nil
}

// IsAuthenticated checks if the user has a valid session
func (sm *SessionManager) IsAuthenticated(r *http.Request) bool {
	_, err := sm.GetUserSession(r)
	return err == nil
}

// RequireAuthentication middleware ensures the user is authenticated
func (sm *SessionManager) RequireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := sm.GetUserSession(r)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Add the user ID to the request context
		ctx := context.WithValue(r.Context(), userIDKey, session.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GenerateCSRFTokenWithResponse generates a new CSRF token and sets it as a cookie in the response
func (sm *SessionManager) GenerateCSRFTokenWithResponse(w http.ResponseWriter) (string, error) {
	// Generate a random token
	tokenBytes := make([]byte, 32)
	_, err := rand.Read(tokenBytes)
	if err != nil {
		return "", err
	}
	token := base64.StdEncoding.EncodeToString(tokenBytes)

	// Set the CSRF token cookie
	expiration := time.Now().Add(time.Hour * CSRFTokenExpirationHours)
	cookie := &http.Cookie{
		Name:     CSRFTokenCookieName,
		Value:    token,
		Path:     "/",
		Expires:  expiration,
		HttpOnly: false, // Must be accessible from JavaScript
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, cookie)

	return token, nil
}

// VerifyCSRFTokenFromHeader verifies that the CSRF token in the request matches the one in the cookie
func (sm *SessionManager) VerifyCSRFTokenFromHeader(r *http.Request) bool {
	// Get the token from the header
	headerToken := r.Header.Get(CSRFHeaderName)
	if headerToken == "" {
		return false
	}

	// Get the token from the cookie
	cookie, err := r.Cookie(CSRFTokenCookieName)
	if err != nil {
		return false
	}

	// Compare the tokens
	return headerToken == cookie.Value
}

// CSRFProtectionMiddleware middleware enforces CSRF protection for authenticated requests
func (sm *SessionManager) CSRFProtectionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip CSRF protection for GET, HEAD, OPTIONS, TRACE requests
		if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" || r.Method == "TRACE" {
			next.ServeHTTP(w, r)
			return
		}

		// Verify CSRF token for state-changing requests
		if !sm.VerifyCSRFTokenFromHeader(r) {
			http.Error(w, "Invalid CSRF Token", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RefreshSession updates the last active time and extends expiration if needed
func (sm *SessionManager) RefreshSession(sessionID string) error {
	return models.UpdateSessionActivity(sessionID)
}

// ExtendSession extends the current session's expiration time
func (sm *SessionManager) ExtendSession(w http.ResponseWriter, r *http.Request) error {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		return err
	}

	sessionID := cookie.Value
	err = models.ExtendSession(sessionID, SessionExpirationHours)
	if err != nil {
		return err
	}

	// Update the cookie expiration time
	session, err := models.GetSessionByID(sessionID)
	if err != nil {
		return err
	}

	newCookie := &http.Cookie{
		Name:     SessionCookieName,
		Value:    sessionID,
		Path:     "/",
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, newCookie)

	return nil
}

// DeleteSession invalidates a session
func (sm *SessionManager) DeleteSession(sessionID string) error {
	return models.DeleteSession(sessionID)
}

// LogoutUser logs a user out by deleting their session
func (sm *SessionManager) LogoutUser(w http.ResponseWriter, r *http.Request) error {
	// Get the session ID from the cookie
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		return err
	}

	// Delete the session from the database
	err = models.DeleteSession(cookie.Value)
	if err != nil {
		return err
	}

	// Delete the session cookie
	sessionCookie := &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, sessionCookie)

	// Delete the CSRF token cookie
	csrfCookie := &http.Cookie{
		Name:     CSRFTokenCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: false,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, csrfCookie)

	return nil
}

// CleanupSessions removes expired and inactive sessions
func (sm *SessionManager) CleanupSessions() error {
	// Delete expired sessions
	if err := models.DeleteExpiredSessions(); err != nil {
		return fmt.Errorf("failed to delete expired sessions: %w", err)
	}

	// Delete inactive sessions
	if err := models.DeleteInactiveSessions(SessionInactivityMinutes); err != nil {
		return fmt.Errorf("failed to delete inactive sessions: %w", err)
	}

	return nil
}

// GetUserIDFromRequest extracts the user ID from the request context
func GetUserIDFromRequest(r *http.Request) (int, bool) {
	userID, ok := r.Context().Value(userIDKey).(int)
	return userID, ok
}
