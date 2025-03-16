// backend/controllers/profile.go
package controllers

import (
    "database/sql"
    "encoding/json"
    "net/http"
    "strconv"
    
    "forum/backend/models"
)

type ProfileController struct {
    DB *sql.DB
}

// GetProfile retrieves a user's profile
func (c *ProfileController) GetProfile(w http.ResponseWriter, r *http.Request) {
    // Only allow GET method
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    // Get user ID from query string
    userIDStr := r.URL.Query().Get("userId")
    if userIDStr == "" {
        http.Error(w, "User ID is required", http.StatusBadRequest)
        return
    }
    
    userID, err := strconv.Atoi(userIDStr)
    if err != nil {
        http.Error(w, "Invalid user ID", http.StatusBadRequest)
        return
    }
    
    // Get user profile
    profile, err := models.GetUserProfile(c.DB, userID)
    if err != nil {
        http.Error(w, "Error retrieving profile", http.StatusInternalServerError)
        return
    }
    
    // Get user basic info
    user, err := models.GetUserByID(c.DB, userID)
    if err != nil {
        http.Error(w, "User not found", http.StatusNotFound)
        return
    }
    
    // Combine data
    response := map[string]interface{}{
        "user": user,
        "profile": profile,
    }
    
    // Return profile data
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

// UpdateProfile updates a user's profile
func (c *ProfileController) UpdateProfile(w http.ResponseWriter, r *http.Request, userID int) {
    // Only allow POST method
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    var profile models.UserProfile
    if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    // Ensure the user can only update their own profile
    profile.UserID = userID
    
    // Update profile
    err := models.UpdateUserProfile(c.DB, profile)
    if err != nil {
        http.Error(w, "Error updating profile", http.StatusInternalServerError)
        return
    }
    
    // Get updated profile
    updatedProfile, err := models.GetUserProfile(c.DB, userID)
    if err != nil {
        http.Error(w, "Error retrieving updated profile", http.StatusInternalServerError)
        return
    }
    
    // Return updated profile
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(updatedProfile)
}