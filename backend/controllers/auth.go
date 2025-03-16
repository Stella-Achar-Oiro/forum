// backend/controllers/auth.go
package controllers

import (
	"database/sql"
	"encoding/json"
	"forum/backend/models"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type AuthController struct {
	DB *sql.DB
}

type RegisterRequest struct {
	Nickname  string `json:"nickname"`
	Age       int    `json:"age"`
	Gender    string `json:"gender"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

type LoginRequest struct {
	Identifier string `json:"identifier"` // nickname or email
	Password   string `json:"password"`
}

type AuthResponse struct {
	User      models.User `json:"user"`
	SessionID string      `json:"sessionId"`
}

// Register handles user registration
func (c *AuthController) Register(w http.ResponseWriter, r *http.Request) {
	// Only allow POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Invalid request body: %v", err)
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Debug - Log received data
	log.Printf("Register request: %+v", req)

	// Validate request
	if req.Nickname == "" || req.Email == "" || req.Password == "" || req.Age <= 0 ||
		req.Gender == "" || req.FirstName == "" || req.LastName == "" {
		log.Printf("Validation failed: missing required fields")
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Password hashing error: %v", err)
		http.Error(w, "Error processing request: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Create user object
	user := models.User{
		Nickname:  req.Nickname,
		Age:       req.Age,
		Gender:    req.Gender,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
	}

	// Save to database
	userID, err := models.RegisterUser(c.DB, user, string(hashedPassword))
	if err != nil {
		log.Printf("Database error during user registration: %v", err)
		http.Error(w, "Database error: "+err.Error(), http.StatusBadRequest)
		return
	}
	log.Printf("User registered successfully with ID: %d", userID)

	// Create session
	sessionID, err := models.CreateSession(c.DB, int(userID))
	if err != nil {
		log.Printf("Session creation error: %v", err)
		http.Error(w, "Error creating session: "+err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("Session created with ID: %s", sessionID)

	// Get complete user data
	user, err = models.GetUserByID(c.DB, int(userID))
	if err != nil {
		http.Error(w, "Error retrieving user data", http.StatusInternalServerError)
		return
	}

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400 * 7, // 7 days
		SameSite: http.SameSiteLaxMode,
	})

	// Return user and session data
	response := AuthResponse{
		User:      user,
		SessionID: sessionID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Login handles user authentication
func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	// Only allow POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Invalid login request body: %v", err)
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Login attempt for: %s", req.Identifier)

	// Validate request
	if req.Identifier == "" || req.Password == "" {
		log.Printf("Login validation failed: missing identifier or password")
		http.Error(w, "Identifier and password are required", http.StatusBadRequest)
		return
	}

	// Find user by nickname or email
	user, err := models.GetUserByNicknameOrEmail(c.DB, req.Identifier)
	if err != nil {
		log.Printf("User not found for identifier %s: %v", req.Identifier, err)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	log.Printf("User found with ID: %d", user.ID)

	// Verify password
	log.Printf("Stored password hash for user %d: %s", user.ID, user.Password)
	log.Printf("Login attempt password: %s", req.Password)

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		log.Printf("Password verification failed for user %d: %v", user.ID, err)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	log.Printf("Password verified successfully for user %d", user.ID)

	// Create session
	sessionID, err := models.CreateSession(c.DB, user.ID)
	if err != nil {
		log.Printf("Session creation error for user %d: %v", user.ID, err)
		http.Error(w, "Error creating session: "+err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("Session created with ID: %s for user %d", sessionID, user.ID)

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400 * 7, // 7 days
		SameSite: http.SameSiteLaxMode,
	})

	// Clear password before sending to client
	user.Password = ""

	// Return user and session data
	response := AuthResponse{
		User:      user,
		SessionID: sessionID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Logout handles user logout
func (c *AuthController) Logout(w http.ResponseWriter, r *http.Request) {
	// Only allow POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get session cookie
	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "No session found", http.StatusBadRequest)
		return
	}

	// Delete session from database
	err = models.DeleteSession(c.DB, cookie.Value)
	if err != nil {
		http.Error(w, "Error deleting session", http.StatusInternalServerError)
		return
	}

	// Clear cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	// Return success
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
}

// GetCurrentUser retrieves the currently logged in user
func (c *AuthController) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	// Only allow GET method
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Printf("GetCurrentUser request from: %s", r.RemoteAddr)

	// Get session cookie
	cookie, err := r.Cookie("session_id")
	if err != nil {
		log.Printf("No session cookie found: %v", err)
		http.Error(w, "No session found", http.StatusUnauthorized)
		return
	}
	log.Printf("Session cookie found: %s", cookie.Value)

	// Get session
	session, err := models.GetSessionByID(c.DB, cookie.Value)
	if err != nil {
		log.Printf("Invalid session ID %s: %v", cookie.Value, err)
		http.Error(w, "Invalid session", http.StatusUnauthorized)
		return
	}
	log.Printf("Valid session found for user ID: %d", session.UserID)

	// Get user
	user, err := models.GetUserByID(c.DB, session.UserID)
	if err != nil {
		log.Printf("User not found for ID %d: %v", session.UserID, err)
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}
	log.Printf("User found: %s (ID: %d)", user.Nickname, user.ID)

	// Return user data
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
