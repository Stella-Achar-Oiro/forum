package handlers

import (
	"log"
	"net/http"

	"github.com/Stella-Achar-Oiro/forum/models"
	"github.com/Stella-Achar-Oiro/forum/utils"
)

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Nickname  string `json:"nickname"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Age       int    `json:"age"`
	Gender    string `json:"gender"`
}

// LoginRequest represents a user login request
type LoginRequest struct {
	Identifier string `json:"identifier"` // Email or nickname
	Password   string `json:"password"`
}

// RegisterHandler handles user registration
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received registration request")
	if r.Method != http.MethodPost {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req RegisterRequest
	if err := utils.DecodeJSONBody(r, &req); err != nil {
		log.Printf("Failed to decode request body: %v", err)
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	log.Printf("Registration request received for user: %s", req.Nickname)

	// Validate the request
	if req.Nickname == "" || req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" || req.Age <= 0 || req.Gender == "" {
		log.Println("Invalid registration data: missing required fields")
		utils.RespondWithError(w, http.StatusBadRequest, "All fields are required")
		return
	}

	// Create a new user
	user := models.User{
		Nickname:  req.Nickname,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Age:       req.Age,
		Gender:    req.Gender,
	}

	userID, err := models.CreateUser(user, req.Password)
	if err != nil {
		log.Printf("Failed to create user: %v", err)
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	log.Printf("User created successfully with ID: %d", userID)

	// Create a session and set the cookie
	cookie, err := utils.CreateSessionCookie(userID)
	if err != nil {
		log.Printf("Failed to create session: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	http.SetCookie(w, &cookie)

	// Get the newly created user
	newUser, err := models.GetUserByID(userID)
	if err != nil {
		log.Printf("Failed to get created user: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get user")
		return
	}

	log.Println("Registration completed successfully")
	utils.RespondWithSuccess(w, http.StatusCreated, "User registered successfully", newUser.ToPublic())
}

// LoginHandler handles user login
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req LoginRequest
	if err := utils.DecodeJSONBody(r, &req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate the request
	if req.Identifier == "" || req.Password == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Identifier and password are required")
		return
	}

	// Authenticate the user
	user, err := models.GetUserByCredentials(req.Identifier, req.Password)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Create a session and set the cookie
	cookie, err := utils.CreateSessionCookie(user.ID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	http.SetCookie(w, &cookie)

	// Update user's online status
	_ = models.UpdateUserOnlineStatus(user.ID, true)

	utils.RespondWithSuccess(w, http.StatusOK, "Login successful", user.ToPublic())
}

// LogoutHandler handles user logout
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get the current user
	user, err := utils.GetUserFromRequest(r)
	if err == nil {
		// Delete the user's session
		cookie, err := r.Cookie(utils.SessionCookieName)
		if err == nil {
			_ = models.DeleteSession(cookie.Value)
		}

		// Update user's online status
		_ = models.UpdateUserOnlineStatus(user.ID, false)
	}

	// Clear the session cookie
	clearCookie := utils.ClearSessionCookie()
	http.SetCookie(w, &clearCookie)

	utils.RespondWithSuccess(w, http.StatusOK, "Logout successful", nil)
}

// GetCurrentUserHandler returns the current logged-in user
func GetCurrentUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	user, err := utils.GetUserFromRequest(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, "User retrieved successfully", user.ToPublic())
}
