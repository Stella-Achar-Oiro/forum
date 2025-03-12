package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Stella-Achar-Oiro/forum/models"
	"github.com/Stella-Achar-Oiro/forum/utils"
)

type UserActivity struct {
	CreatedPosts []models.Post     `json:"createdPosts"`
	Reactions    []models.Reaction `json:"reactions"`
	Comments     []models.Comment  `json:"comments"`
}

// GetUserActivityHandler handles retrieving all activity for the current user
func GetUserActivityHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := utils.GetUserFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user's created posts
	createdPosts, err := models.GetPostsByUserID(user.ID)
	if err != nil {
		http.Error(w, "Failed to get user posts", http.StatusInternalServerError)
		return
	}

	// Get user's reactions (likes/dislikes)
	reactions, err := models.GetUserReactions(user.ID)
	if err != nil {
		http.Error(w, "Failed to get user reactions", http.StatusInternalServerError)
		return
	}

	// Get user's comments
	comments, err := models.GetUserComments(user.ID)
	if err != nil {
		http.Error(w, "Failed to get user comments", http.StatusInternalServerError)
		return
	}

	activity := UserActivity{
		CreatedPosts: createdPosts,
		Reactions:    reactions,
		Comments:     comments,
	}

	response := map[string]interface{}{
		"success": true,
		"message": "User activity retrieved successfully",
		"data":    activity,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
