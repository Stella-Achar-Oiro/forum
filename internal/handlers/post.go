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

// PostHandler handles all post-related routes
type PostHandler struct {
	db *sql.DB
}

// NewPostHandler creates a new PostHandler
func NewPostHandler(db *sql.DB) *PostHandler {
	return &PostHandler{db: db}
}

// CreatePost handles the creation of a new post
func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var input models.PostInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Start transaction
	tx, err := h.db.Begin()
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Insert post
	var postID int64
	err = tx.QueryRow(`
		INSERT INTO posts (title, content, user_id)
		VALUES (?, ?, ?)
		RETURNING id`,
		input.Title, input.Content, user.ID).Scan(&postID)
	if err != nil {
		http.Error(w, "Error creating post", http.StatusInternalServerError)
		return
	}

	// Insert categories
	for _, category := range input.Categories {
		var categoryID int64
		// Get or create category
		err = tx.QueryRow(`
			INSERT INTO categories (name)
			VALUES (?)
			ON CONFLICT (name) DO UPDATE SET name = name
			RETURNING id`,
			category).Scan(&categoryID)
		if err != nil {
			http.Error(w, "Error with categories", http.StatusInternalServerError)
			return
		}

		// Link post to category
		_, err = tx.Exec(`
			INSERT INTO post_categories (post_id, category_id)
			VALUES (?, ?)`,
			postID, categoryID)
		if err != nil {
			http.Error(w, "Error linking categories", http.StatusInternalServerError)
			return
		}
	}

	if err = tx.Commit(); err != nil {
		http.Error(w, "Error committing transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// GetPosts handles retrieving posts with optional filtering
func (h *PostHandler) GetPosts(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	userID := r.URL.Query().Get("user_id")
	likedBy := r.URL.Query().Get("liked_by")

	query := `
		SELECT DISTINCT
			p.id,
			p.title,
			p.content,
			p.user_id,
			u.username,
			p.created_at,
			COALESCE(
				(SELECT COUNT(*) FROM likes WHERE post_id = p.id AND is_like = true),
				0
			) as likes,
			COALESCE(
				(SELECT COUNT(*) FROM likes WHERE post_id = p.id AND is_like = false),
				0
			) as dislikes
		FROM posts p
		JOIN users u ON p.user_id = u.id`

	var args []interface{}
	var conditions []string

	if category != "" {
		query += ` JOIN post_categories pc ON p.id = pc.post_id
				  JOIN categories c ON pc.category_id = c.id`
		conditions = append(conditions, "c.name = ?")
		args = append(args, category)
	}

	if userID != "" {
		conditions = append(conditions, "p.user_id = ?")
		args = append(args, userID)
	}

	if likedBy != "" {
		query += ` JOIN likes l ON p.id = l.post_id`
		conditions = append(conditions, "l.user_id = ? AND l.is_like = true")
		args = append(args, likedBy)
	}

	if len(conditions) > 0 {
		query += " WHERE " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			query += " AND " + conditions[i]
		}
	}

	query += " ORDER BY p.created_at DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		http.Error(w, "Error fetching posts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&post.UserID,
			&post.Username,
			&post.CreatedAt,
			&post.Likes,
			&post.Dislikes,
		)
		if err != nil {
			http.Error(w, "Error scanning posts", http.StatusInternalServerError)
			return
		}

		// Get categories for the post
		categoryRows, err := h.db.Query(`
			SELECT c.name
			FROM categories c
			JOIN post_categories pc ON c.id = pc.category_id
			WHERE pc.post_id = ?`,
			post.ID)
		if err != nil {
			http.Error(w, "Error fetching categories", http.StatusInternalServerError)
			return
		}
		defer categoryRows.Close()

		var categories []string
		for categoryRows.Next() {
			var category string
			if err := categoryRows.Scan(&category); err != nil {
				http.Error(w, "Error scanning categories", http.StatusInternalServerError)
				return
			}
			categories = append(categories, category)
		}
		post.Categories = categories

		posts = append(posts, post)
	}

	json.NewEncoder(w).Encode(posts)
}

// LikePost handles liking or disliking a post
func (h *PostHandler) LikePost(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	postID, err := strconv.ParseInt(r.URL.Query().Get("post_id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	isLike := r.URL.Query().Get("like") == "true"

	// Check if user has already liked/disliked
	var existingID int64
	err = h.db.QueryRow(`
		SELECT id FROM likes
		WHERE user_id = ? AND post_id = ?`,
		user.ID, postID).Scan(&existingID)

	if err == sql.ErrNoRows {
		// Create new like
		_, err = h.db.Exec(`
			INSERT INTO likes (user_id, post_id, is_like)
			VALUES (?, ?, ?)`,
			user.ID, postID, isLike)
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
