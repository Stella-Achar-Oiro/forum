package handlers

import (
	"encoding/json"
	"net/http"
	"real-time-forum/backend/database"
	"real-time-forum/backend/models"
	"time"

	"github.com/gofrs/uuid"
)

// Response structures
type ErrorResponse struct {
    Error string `json:"error"`
}

type AuthResponse struct {
    Token     string      `json:"token"`
    User      *models.User `json:"user"`
    ExpiresAt time.Time   `json:"expires_at"`
}

// Session represents a user session
type Session struct {
    Token     string    `json:"token"`
    UserID    int64     `json:"user_id"`
    ExpiresAt time.Time `json:"expires_at"`
}

// In-memory session store (should be replaced with a proper session store in production)
var sessions = make(map[string]Session)

// RegisterHandler handles user registration
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
    // Only allow POST method
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // Parse registration data
    var reg models.UserRegistration
    if err := json.NewDecoder(r.Body).Decode(&reg); err != nil {
        sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Create user
    user, err := models.CreateUser(database.GetDB(), &reg)
    if err != nil {
        switch err {
        case models.ErrUserExists:
            sendErrorResponse(w, "User already exists", http.StatusConflict)
        case models.ErrInvalidInput:
            sendErrorResponse(w, "Invalid input data", http.StatusBadRequest)
        default:
            sendErrorResponse(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    // Create session
    session, err := createSession(user.ID)
    if err != nil {
        sendErrorResponse(w, "Error creating session", http.StatusInternalServerError)
        return
    }

    // Send response
    sendAuthResponse(w, session.Token, user)
}

// LoginHandler handles user login
func LoginHandler(w http.ResponseWriter, r *http.Request) {
    // Only allow POST method
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // Parse login data
    var login models.UserLogin
    if err := json.NewDecoder(r.Body).Decode(&login); err != nil {
        sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Authenticate user
    user, err := models.AuthenticateUser(database.GetDB(), &login)
    if err != nil {
        switch err {
        case models.ErrUserNotFound, models.ErrInvalidCredentials:
            sendErrorResponse(w, "Invalid credentials", http.StatusUnauthorized)
        default:
            sendErrorResponse(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    // Create session
    session, err := createSession(user.ID)
    if err != nil {
        sendErrorResponse(w, "Error creating session", http.StatusInternalServerError)
        return
    }

    // Send response
    sendAuthResponse(w, session.Token, user)
}

// LogoutHandler handles user logout
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
    // Only allow POST method
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // Get token from Authorization header
    token := r.Header.Get("Authorization")
    if token == "" {
        sendErrorResponse(w, "No token provided", http.StatusBadRequest)
        return
    }

    // Remove session
    delete(sessions, token)

    // Send success response
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "message": "Successfully logged out",
    })
}

// Helper functions

func createSession(userID int64) (Session, error) {
    // Generate UUID for session token
    token, err := uuid.NewV4()
    if err != nil {
        return Session{}, err
    }

    // Create session with 24-hour expiry
    session := Session{
        Token:     token.String(),
        UserID:    userID,
        ExpiresAt: time.Now().Add(24 * time.Hour),
    }

    // Store session
    sessions[session.Token] = session

    return session, nil
}

func sendErrorResponse(w http.ResponseWriter, message string, status int) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

func sendAuthResponse(w http.ResponseWriter, token string, user *models.User) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(AuthResponse{
        Token:     token,
        User:      user,
        ExpiresAt: sessions[token].ExpiresAt,
    })
}

// GetSession retrieves a session by token
func GetSession(token string) (Session, bool) {
    session, exists := sessions[token]
    if !exists {
        return Session{}, false
    }

    // Check if session has expired
    if time.Now().After(session.ExpiresAt) {
        delete(sessions, token)
        return Session{}, false
    }

    return session, true
}