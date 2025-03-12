package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/Stella-Achar-Oiro/forum/database"
	"github.com/Stella-Achar-Oiro/forum/models"
	"github.com/Stella-Achar-Oiro/forum/utils"
	"github.com/gorilla/mux"
)

func HandlePostReaction(w http.ResponseWriter, r *http.Request, reactionType string) {
	user, err := utils.GetUserFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	postIDStr, ok := vars["id"]
	if !ok {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}

	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Get the post to check ownership
	post, err := models.GetPostByID(int(postID))
	if err != nil {
		http.Error(w, "Failed to get post", http.StatusInternalServerError)
		return
	}

	// Create or update the reaction
	err = models.AddPostReaction(int(postID), user.ID, reactionType == "like")
	if err != nil {
		http.Error(w, "Failed to process reaction", http.StatusInternalServerError)
		return
	}

	// Create notification for post owner if it's not their own post
	if post.UserID != user.ID {
		notificationType := models.NotificationTypeLike
		if reactionType == "dislike" {
			notificationType = models.NotificationTypeDislike
		}

		notification := &models.Notification{
			UserID:  int64(post.UserID),
			Type:    notificationType,
			PostID:  postID,
			ActorID: int64(user.ID),
			Message: fmt.Sprintf("%s %s your post", user.Nickname, reactionType+"d"),
			Read:    false,
		}

		err = models.CreateNotification(database.DB, notification)
		if err != nil {
			// Log the error but don't fail the request
			log.Printf("Failed to create notification: %v", err)
		}
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Reaction processed successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
