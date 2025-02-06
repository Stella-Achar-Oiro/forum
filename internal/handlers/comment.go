package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"forum/internal/auth"
	"forum/internal/models"
)

// CommentHandler handles all comment-related routes
type CommentHandler struct {
	db *sql.DB
}

// NewCommentHandler creates a new CommentHandler
func NewCommentHandler(db *sql.DB) *CommentHandler {
	return &CommentHandler{db: db}
}

// CreateComment handles the creation of a new comment
func (h *CommentHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var input models.CommentInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Verify post exists
	var exists bool
	err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM posts WHERE id = ?)", input.PostID).Scan(&exists)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	// Create comment
	_, err = h.db.Exec(`
		INSERT INTO comments (content, user_id, post_id)
		VALUES (?, ?, ?)`,
		input.Content, user.ID, input.PostID)
	if err != nil {
		http.Error(w, "Error creating comment", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// GetComments retrieves all comments for a post
func (h *CommentHandler) GetComments(w http.ResponseWriter, r *http.Request) {
	postID, err := strconv.ParseInt(r.URL.Query().Get("post_id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	rows, err := h.db.Query(`
		SELECT 
			c.id,
			c.content,
			c.user_id,
			u.username,
			c.post_id,
			c.created_at,
			COALESCE(
				(SELECT COUNT(*) FROM likes WHERE comment_id = c.id AND is_like = true),
				0
			) as likes,
			COALESCE(
				(SELECT COUNT(*) FROM likes WHERE comment_id = c.id AND is_like = false),
				0
			) as dislikes
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.post_id = ?
		ORDER BY c.created_at DESC`,
		postID)
	if err != nil {
		http.Error(w, "Error fetching comments", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var comment models.Comment
		err := rows.Scan(
			&comment.ID,
			&comment.Content,
			&comment.UserID,
			&comment.Username,
			&comment.PostID,
			&comment.CreatedAt,
			&comment.Likes,
			&comment.Dislikes,
		)
		if err != nil {
			http.Error(w, "Error scanning comments", http.StatusInternalServerError)
			return
		}
		comments = append(comments, comment)
	}

	json.NewEncoder(w).Encode(comments)
}

// LikeComment handles liking or disliking a comment
func (h *CommentHandler) LikeComment(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	commentID, err := strconv.ParseInt(r.URL.Query().Get("comment_id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	isLike := r.URL.Query().Get("like") == "true"

	// Check if user has already liked/disliked
	var existingID int64
	err = h.db.QueryRow(`
		SELECT id FROM likes
		WHERE user_id = ? AND comment_id = ?`,
		user.ID, commentID).Scan(&existingID)

	if err == sql.ErrNoRows {
		// Create new like
		_, err = h.db.Exec(`
			INSERT INTO likes (user_id, comment_id, is_like)
			VALUES (?, ?, ?)`,
			user.ID, commentID, isLike)
	} else if err == nil {
		// Update existing like
		_, err = h.db.Exec(`
			UPDATE likes
			SET is_like = ?
			WHERE id = ?`,
			isLike, existingID)
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("Error handling like: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
