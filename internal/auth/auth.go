package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"forum/internal/models"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	// SessionDuration defines how long a session token remains valid
	SessionDuration = 24 * time.Hour
)

var (
	ErrUserExists     = errors.New("user already exists")
	ErrInvalidCreds   = errors.New("invalid credentials")
	ErrInvalidSession = errors.New("invalid session")
)

type contextKey string

const userContextKey contextKey = "user"

// WithUser adds the user to the context
func WithUser(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

// UserFromContext retrieves the user from the context
func UserFromContext(ctx context.Context) (*models.User, bool) {
	user, ok := ctx.Value(userContextKey).(*models.User)
	return user, ok
}

// RegisterUser creates a new user account
func RegisterUser(db *sql.DB, input models.UserInput) error {
	// Check if user exists
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = ? OR username = ?)",
		input.Email, input.Username).Scan(&exists)
	if err != nil {
		return fmt.Errorf("error checking user existence: %v", err)
	}
	if exists {
		return ErrUserExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("error hashing password: %v", err)
	}

	// Create user
	_, err = db.Exec(`
		INSERT INTO users (username, email, password_hash)
		VALUES (?, ?, ?)`,
		input.Username, input.Email, string(hashedPassword))
	if err != nil {
		return fmt.Errorf("error creating user: %v", err)
	}

	return nil
}

// LoginUser authenticates a user and creates a new session
func LoginUser(db *sql.DB, email, password string) (string, error) {
	var user models.User
	err := db.QueryRow(`
		SELECT id, username, email, password_hash
		FROM users
		WHERE email = ?`,
		email).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrInvalidCreds
		}
		return "", fmt.Errorf("error fetching user: %v", err)
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", ErrInvalidCreds
	}

	// Create session
	sessionToken := uuid.New().String()
	expiresAt := time.Now().Add(SessionDuration)

	_, err = db.Exec(`
		INSERT INTO sessions (user_id, session_token, expires_at)
		VALUES (?, ?, ?)`,
		user.ID, sessionToken, expiresAt)
	if err != nil {
		return "", fmt.Errorf("error creating session: %v", err)
	}

	return sessionToken, nil
}

// ValidateSession checks if a session token is valid and returns the associated user
func ValidateSession(db *sql.DB, sessionToken string) (*models.User, error) {
	var user models.User
	err := db.QueryRow(`
		SELECT u.id, u.username, u.email, u.created_at
		FROM users u
		JOIN sessions s ON u.id = s.user_id
		WHERE s.session_token = ? AND s.expires_at > ?`,
		sessionToken, time.Now()).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrInvalidSession
		}
		return nil, fmt.Errorf("error validating session: %v", err)
	}

	return &user, nil
}

// LogoutUser invalidates a session token
func LogoutUser(db *sql.DB, sessionToken string) error {
	_, err := db.Exec("DELETE FROM sessions WHERE session_token = ?", sessionToken)
	if err != nil {
		return fmt.Errorf("error logging out user: %v", err)
	}
	return nil
}
