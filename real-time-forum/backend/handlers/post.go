package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"real-time-forum/backend/database"
	"real-time-forum/backend/middleware"
	"real-time-forum/backend/models"
)

// PostsHandler handles all post-related requests
func PostsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleGetPosts(w, r)
	case http.MethodPost:
		handleCreatePost(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// SpecificPostHandler handles requests for specific posts
func SpecificPostHandler(w http.ResponseWriter, r *http.Request) {
	// Extract post ID from URL
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(parts[len(parts)-1], 10, 64)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		handleGetPost(w, r, id)
	case http.MethodPut:
		handleUpdatePost(w, r, id)
	case http.MethodDelete:
		handleDeletePost(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetPosts handles retrieving posts with pagination and category filtering
func handleGetPosts(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10 // Default page size
	}

	category := r.URL.Query().Get("category")

	var posts []*models.Post
	var err error

	if category != "" {
		posts, err = models.GetPostsByCategory(database.GetDB(), category, page, pageSize)
	} else {
		posts, err = models.GetPosts(database.GetDB(), page, pageSize)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

// handleCreatePost handles post creation
func handleCreatePost(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var postCreate models.PostCreate
	if err := json.NewDecoder(r.Body).Decode(&postCreate); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create post
	post, err := models.CreatePost(database.GetDB(), user.ID, &postCreate)
	if err != nil {
		switch err {
		case models.ErrInvalidPost:
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(post)
}

// handleGetPost handles retrieving a specific post
func handleGetPost(w http.ResponseWriter, r *http.Request, id int64) {
	post, err := models.GetPostByID(database.GetDB(), id)
	if err != nil {
		switch err {
		case models.ErrPostNotFound:
			http.Error(w, "Post not found", http.StatusNotFound)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}

// handleUpdatePost handles updating a specific post
func handleUpdatePost(w http.ResponseWriter, r *http.Request, id int64) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var postUpdate models.PostUpdate
	if err := json.NewDecoder(r.Body).Decode(&postUpdate); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	post, err := models.UpdatePost(database.GetDB(), id, user.ID, &postUpdate)
	if err != nil {
		switch err {
		case models.ErrPostNotFound:
			http.Error(w, "Post not found", http.StatusNotFound)
		case models.ErrInvalidPost:
			http.Error(w, "Invalid post data", http.StatusBadRequest)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}

// handleDeletePost handles deleting a specific post
func handleDeletePost(w http.ResponseWriter, r *http.Request, id int64) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err := models.DeletePost(database.GetDB(), id, user.ID)
	if err != nil {
		switch err {
		case models.ErrPostNotFound:
			http.Error(w, "Post not found", http.StatusNotFound)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
