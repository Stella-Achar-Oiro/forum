package handlers

import (
    "encoding/json"
    "net/http"
    "strconv"
    "strings"
    "forum/database"
    "forum/middleware"
    "forum/models"
)

// CommentsHandler handles comment-related requests
func CommentsHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        handleGetComments(w, r)
    case http.MethodPost:
        handleCreateComment(w, r)
    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}

// SpecificCommentHandler handles requests for specific comments
func SpecificCommentHandler(w http.ResponseWriter, r *http.Request) {
    // Extract comment ID from URL
    parts := strings.Split(r.URL.Path, "/")
    if len(parts) < 3 {
        http.Error(w, "Invalid URL", http.StatusBadRequest)
        return
    }

    id, err := strconv.ParseInt(parts[len(parts)-1], 10, 64)
    if err != nil {
        http.Error(w, "Invalid comment ID", http.StatusBadRequest)
        return
    }

    switch r.Method {
    case http.MethodPut:
        handleUpdateComment(w, r, id)
    case http.MethodDelete:
        handleDeleteComment(w, r, id)
    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}

// handleGetComments retrieves comments for a post
func handleGetComments(w http.ResponseWriter, r *http.Request) {
    // Get post ID from query parameters
    postIDStr := r.URL.Query().Get("post_id")
    if postIDStr == "" {
        http.Error(w, "Post ID is required", http.StatusBadRequest)
        return
    }

    postID, err := strconv.ParseInt(postIDStr, 10, 64)
    if err != nil {
        http.Error(w, "Invalid post ID", http.StatusBadRequest)
        return
    }

    // Parse pagination parameters
    page, _ := strconv.Atoi(r.URL.Query().Get("page"))
    if page < 1 {
        page = 1
    }

    pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
    if pageSize < 1 || pageSize > 50 {
        pageSize = 10
    }

    // Get comments
    comments, err := models.GetCommentsByPostID(database.GetDB(), postID, page, pageSize)
    if err != nil {
        http.Error(w, "Error retrieving comments", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(comments)
}

// handleCreateComment handles comment creation
func handleCreateComment(w http.ResponseWriter, r *http.Request) {
    user, ok := middleware.GetUserFromContext(r.Context())
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    var commentCreate models.CommentCreate
    if err := json.NewDecoder(r.Body).Decode(&commentCreate); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    comment, err := models.CreateComment(database.GetDB(), user.ID, &commentCreate)
    if err != nil {
        switch err {
        case models.ErrInvalidComment:
            http.Error(w, err.Error(), http.StatusBadRequest)
        default:
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(comment)
}

// handleUpdateComment handles updating a specific comment
func handleUpdateComment(w http.ResponseWriter, r *http.Request, id int64) {
    user, ok := middleware.GetUserFromContext(r.Context())
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    var commentUpdate models.CommentUpdate
    if err := json.NewDecoder(r.Body).Decode(&commentUpdate); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    comment, err := models.UpdateComment(database.GetDB(), id, user.ID, &commentUpdate)
    if err != nil {
        switch err {
        case models.ErrCommentNotFound:
            http.Error(w, "Comment not found", http.StatusNotFound)
        case models.ErrInvalidComment:
            http.Error(w, "Invalid comment data", http.StatusBadRequest)
        default:
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(comment)
}

// handleDeleteComment handles deleting a specific comment
func handleDeleteComment(w http.ResponseWriter, r *http.Request, id int64) {
    user, ok := middleware.GetUserFromContext(r.Context())
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    err := models.DeleteComment(database.GetDB(), id, user.ID)
    if err != nil {
        switch err {
        case models.ErrCommentNotFound:
            http.Error(w, "Comment not found", http.StatusNotFound)
        default:
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    w.WriteHeader(http.StatusNoContent)
}