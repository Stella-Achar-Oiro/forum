package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Stella-Achar-Oiro/forum/database"
	"github.com/Stella-Achar-Oiro/forum/models"
	"github.com/Stella-Achar-Oiro/forum/utils"
	"github.com/gorilla/mux"
)

// CreateCommentRequest represents a request to create a comment
type CreateCommentRequest struct {
	Content string `json:"content"`
}

// CommentReactionRequest represents a request to add a reaction to a comment
type CommentReactionRequest struct {
	ReactionType string `json:"reactionType"` // "like" or "dislike"
}

// CreateCommentHandler handles the creation of a new comment
func CreateCommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := utils.GetUserFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var comment models.Comment
	if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	comment.UserID = user.ID
	comment.CreatedAt = time.Now()

	// Create the comment
	commentID, err := models.CreateComment(comment)
	if err != nil {
		http.Error(w, "Failed to create comment", http.StatusInternalServerError)
		return
	}
	comment.ID = commentID

	// Get the post to check if it exists
	post, err := models.GetPostByID(comment.PostID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Post not found")
		return
	}

	if post.UserID != user.ID {
		// Create notification for post owner
		notification := &models.Notification{
			UserID:    int64(post.UserID),
			Type:      models.NotificationTypeComment,
			PostID:    int64(comment.PostID),
			CommentID: func() *int64 { id := int64(comment.ID); return &id }(),
			ActorID:   int64(user.ID),
			Message:   fmt.Sprintf("%s commented on your post", user.Nickname),
			Read:      false,
		}

		err = models.CreateNotification(database.DB, notification)
		if err != nil {
			// Log the error but don't fail the request
			log.Printf("Failed to create notification: %v", err)
		}
	}

	// If this is a reply to another comment, notify that comment's author
	if comment.ParentID != nil && *comment.ParentID > 0 {
		parentComment, err := models.GetCommentByID(*comment.ParentID)
		if err != nil {
			log.Printf("Failed to get parent comment for notification: %v", err)
		} else if parentComment.UserID != user.ID {
			notification := &models.Notification{
				UserID:    int64(parentComment.UserID),
				Type:      models.NotificationTypeComment,
				PostID:    int64(comment.PostID),
				CommentID: func() *int64 { id := int64(comment.ID); return &id }(),
				ActorID:   int64(user.ID),
				Message:   fmt.Sprintf("%s replied to your comment", user.Nickname),
				Read:      false,
			}

			err = models.CreateNotification(database.DB, notification)
			if err != nil {
				log.Printf("Failed to create notification: %v", err)
			}
		}
	}

	// Get the complete comment data with user info
	commentWithUser, err := models.GetCommentByID(commentID)
	if err != nil {
		http.Error(w, "Failed to get comment details", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Comment created successfully",
		"data":    commentWithUser,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetCommentsHandler handles retrieving comments for a post
func GetCommentsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get the post ID from the URL
	vars := mux.Vars(r)
	postIDStr, ok := vars["postId"]
	if !ok {
		utils.RespondWithError(w, http.StatusBadRequest, "Post ID is required")
		return
	}

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	// Get current user ID if authenticated
	var currentUserID int
	user, err := utils.GetUserFromRequest(r)
	if err == nil {
		currentUserID = user.ID
	}

	// Get comments
	comments, err := models.GetCommentsByPostID(postID, currentUserID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve comments")
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Comments retrieved successfully", comments)
}

// AddCommentReactionHandler handles adding a reaction to a comment
func AddCommentReactionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get the current user
	user, err := utils.GetUserFromRequest(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	// Get the comment ID from the URL
	vars := mux.Vars(r)
	commentIDStr, ok := vars["commentId"]
	if !ok {
		utils.RespondWithError(w, http.StatusBadRequest, "Comment ID is required")
		return
	}

	commentID, err := strconv.Atoi(commentIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid comment ID")
		return
	}

	// Parse the request
	var req CommentReactionRequest
	if err := utils.DecodeJSONBody(r, &req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate the reaction type
	if req.ReactionType != "like" && req.ReactionType != "dislike" {
		utils.RespondWithError(w, http.StatusBadRequest, "Reaction type must be 'like' or 'dislike'")
		return
	}

	// Add the reaction
	err = models.AddReactionToComment(user.ID, commentID, req.ReactionType)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to add reaction")
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Reaction added successfully", nil)
}
