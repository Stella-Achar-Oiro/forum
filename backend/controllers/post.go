// backend/controllers/post.go
package controllers

import (
    "database/sql"
    "encoding/json"
    "net/http"
    "strconv"
    "forum/backend/models"
)

type PostController struct {
    DB *sql.DB
}

type CreatePostRequest struct {
    Title    string `json:"title"`
    Content  string `json:"content"`
    Category string `json:"category"`
}

type CreateCommentRequest struct {
    Content string `json:"content"`
}

// CreatePost handles new post creation
func (c *PostController) CreatePost(w http.ResponseWriter, r *http.Request, userID int) {
    // Only allow POST method
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    var req CreatePostRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    // Validate request
    if req.Title == "" || req.Content == "" || req.Category == "" {
        http.Error(w, "Title, content, and category are required", http.StatusBadRequest)
        return
    }
    
    // Create post
    post := models.Post{
        UserID:   userID,
        Title:    req.Title,
        Content:  req.Content,
        Category: req.Category,
    }
    
    // Save to database
    postID, err := models.CreatePost(c.DB, post)
    if err != nil {
        http.Error(w, "Error creating post", http.StatusInternalServerError)
        return
    }
    
    // Get complete post with user data
    post, err = models.GetPostByID(c.DB, int(postID))
    if err != nil {
        http.Error(w, "Error retrieving post", http.StatusInternalServerError)
        return
    }
    
    // Return post data
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(post)
}

// GetAllPosts retrieves all posts
func (c *PostController) GetAllPosts(w http.ResponseWriter, r *http.Request) {
    // Only allow GET method
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    // Get posts from database
    posts, err := models.GetAllPosts(c.DB)
    if err != nil {
        http.Error(w, "Error retrieving posts", http.StatusInternalServerError)
        return
    }
    
    // Return posts
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(posts)
}

// GetPost retrieves a specific post with its comments
func (c *PostController) GetPost(w http.ResponseWriter, r *http.Request) {
    // Only allow GET method
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    // Get post ID from URL
    postIDStr := r.URL.Query().Get("id")
    if postIDStr == "" {
        http.Error(w, "Post ID is required", http.StatusBadRequest)
        return
    }
    
    postID, err := strconv.Atoi(postIDStr)
    if err != nil {
        http.Error(w, "Invalid post ID", http.StatusBadRequest)
        return
    }
    
    // Get post with comments
    post, err := models.GetPostByID(c.DB, postID)
    if err != nil {
        http.Error(w, "Post not found", http.StatusNotFound)
        return
    }
    
    // Return post with comments
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(post)
}

// CreateComment adds a comment to a post
func (c *PostController) CreateComment(w http.ResponseWriter, r *http.Request, userID int) {
    // Only allow POST method
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    // Get post ID from URL
    postIDStr := r.URL.Query().Get("postId")
    if postIDStr == "" {
        http.Error(w, "Post ID is required", http.StatusBadRequest)
        return
    }
    
    postID, err := strconv.Atoi(postIDStr)
    if err != nil {
        http.Error(w, "Invalid post ID", http.StatusBadRequest)
        return
    }
    
    var req CreateCommentRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    // Validate request
    if req.Content == "" {
        http.Error(w, "Comment content is required", http.StatusBadRequest)
        return
    }
    
    // Create comment
    comment := models.Comment{
        PostID:  postID,
        UserID:  userID,
        Content: req.Content,
    }
    
    // Save to database
    commentID, err := models.CreateComment(c.DB, comment)
    if err != nil {
        http.Error(w, "Error creating comment", http.StatusInternalServerError)
        return
    }
    
    // Get comments for post
    comments, err := models.GetCommentsByPostID(c.DB, postID)
    if err != nil {
        http.Error(w, "Error retrieving comments", http.StatusInternalServerError)
        return
    }
    
    // Find the new comment
    var newComment models.Comment
    for _, c := range comments {
        if c.ID == int(commentID) {
            newComment = c
            break
        }
    }
    
    // Return comment data
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(newComment)
}