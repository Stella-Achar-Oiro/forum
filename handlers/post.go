package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/Stella-Achar-Oiro/forum/database"
	"github.com/Stella-Achar-Oiro/forum/models"
	"github.com/Stella-Achar-Oiro/forum/utils"
	"github.com/gorilla/mux"
)

// CreatePostRequest represents a request to create a new post
type CreatePostRequest struct {
	Title      string `json:"title"`
	Content    string `json:"content"`
	Categories []int  `json:"categories"` // Category IDs
}

// ReactionRequest represents a request to add a reaction to a post
type ReactionRequest struct {
	IsLike bool `json:"isLike"`
}

// CreatePostHandler handles the creation of a new post
func CreatePostHandler(w http.ResponseWriter, r *http.Request) {
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

	// Parse multipart form
	err = r.ParseMultipartForm(utils.MaxFileSize)
	if err != nil {
		log.Printf("Failed to parse multipart form: %v", err)
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Get form values
	title := r.FormValue("title")
	content := r.FormValue("content")
	categoriesStr := r.FormValue("categories")
	giphyID := r.FormValue("giphyId")

	// Validate required fields
	if title == "" || content == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Title and content are required")
		return
	}

	// Parse categories
	var categories []int
	if categoriesStr != "" {
		err = json.Unmarshal([]byte(categoriesStr), &categories)
		if err != nil {
			log.Printf("Failed to parse categories: %v", err)
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid categories format")
			return
		}
	}

	// Initialize post
	post := models.Post{
		UserID:  user.ID,
		Title:   title,
		Content: content,
	}

	// Handle file uploads (image or video)
	file, header, err := r.FormFile("media")
	if err == nil {
		defer file.Close()
		mediaURL, err := utils.SaveUploadedFile(header)
		if err != nil {
			log.Printf("Failed to save uploaded file: %v", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to save media")
			return
		}

		// Determine if it's an image or video based on content type
		contentType := header.Header.Get("Content-Type")
		if strings.Contains(utils.AllowedImageTypes, contentType) {
			post.ImageURL = mediaURL
		} else if strings.Contains(utils.AllowedVideoTypes, contentType) {
			post.VideoURL = mediaURL
		}
	}

	// Handle Giphy integration
	if giphyID != "" {
		// Get Giphy API key from environment variable
		giphyAPIKey := os.Getenv("GIPHY_API_KEY")
		if giphyAPIKey == "" {
			log.Printf("Giphy API key not found")
			utils.RespondWithError(w, http.StatusInternalServerError, "Giphy integration not configured")
			return
		}

		// Get GIF details from Giphy
		gif, err := utils.GetGiphyGIF(giphyID, giphyAPIKey)
		if err != nil {
			log.Printf("Failed to get Giphy GIF: %v", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get Giphy GIF")
			return
		}

		post.GiphyID = gif.ID
		post.GiphyURL = gif.URL
	}

	// Start a transaction for post creation and category association
	tx, err := database.DB.Begin()
	if err != nil {
		log.Printf("Failed to start transaction: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create post")
		return
	}
	defer tx.Rollback()

	// Create the post within the transaction
	postID, err := models.CreatePostTx(tx, post)
	if err != nil {
		log.Printf("Failed to create post: %v", err)
		if post.ImageURL != "" {
			_ = utils.DeleteUploadedFile(post.ImageURL)
		}
		if post.VideoURL != "" {
			_ = utils.DeleteUploadedFile(post.VideoURL)
		}
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create post")
		return
	}

	// Associate categories with the post
	if len(categories) > 0 {
		err = models.AssociatePostCategories(tx, postID, categories)
		if err != nil {
			log.Printf("Failed to associate categories: %v", err)
			if post.ImageURL != "" {
				_ = utils.DeleteUploadedFile(post.ImageURL)
			}
			if post.VideoURL != "" {
				_ = utils.DeleteUploadedFile(post.VideoURL)
			}
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to associate categories")
			return
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		if post.ImageURL != "" {
			_ = utils.DeleteUploadedFile(post.ImageURL)
		}
		if post.VideoURL != "" {
			_ = utils.DeleteUploadedFile(post.VideoURL)
		}
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create post")
		return
	}

	// Get the created post with user information and categories
	newPost, err := models.GetPostByID(postID)
	if err != nil {
		log.Printf("Failed to get created post: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get post")
		return
	}

	utils.RespondWithSuccess(w, http.StatusCreated, "Post created successfully", newPost)
}

// GetPostsHandler handles retrieving posts with optional filtering
func GetPostsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse query parameters
	query := r.URL.Query()

	userID := 0
	if userIDStr := query.Get("userId"); userIDStr != "" {
		var err error
		userID, err = strconv.Atoi(userIDStr)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid user ID")
			return
		}
	}

	categoryID := 0
	if categoryIDStr := query.Get("categoryId"); categoryIDStr != "" {
		var err error
		categoryID, err = strconv.Atoi(categoryIDStr)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid category ID")
			return
		}
	}

	page := 1
	if pageStr := query.Get("page"); pageStr != "" {
		var err error
		page, err = strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid page number")
			return
		}
	}

	pageSize := 10 // Default page size
	if pageSizeStr := query.Get("pageSize"); pageSizeStr != "" {
		var err error
		pageSize, err = strconv.Atoi(pageSizeStr)
		if err != nil || pageSize < 1 || pageSize > 50 {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid page size")
			return
		}
	}

	// Get posts
	posts, err := models.GetPosts(userID, categoryID, page, pageSize)
	if err != nil {
		log.Printf("Failed to get posts: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get posts")
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Posts retrieved successfully", posts)
}

// GetPostHandler handles retrieving a single post by ID
func GetPostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get post ID from query parameters
	postIDStr := r.URL.Query().Get("id")
	if postIDStr == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Post ID is required")
		return
	}

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	// Get the post
	post, err := models.GetPostByID(postID)
	if err != nil {
		log.Printf("Failed to get post: %v", err)
		utils.RespondWithError(w, http.StatusNotFound, "Post not found")
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Post retrieved successfully", post)
}

// AddPostReactionHandler handles adding a reaction (like/dislike) to a post
func AddPostReactionHandler(w http.ResponseWriter, r *http.Request) {
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

	// Extract post ID from URL path
	path := r.URL.Path
	postIDStr := path[len("/api/posts/") : len(path)-len("/react")]
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	var req ReactionRequest
	if err := utils.DecodeJSONBody(r, &req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Add the reaction
	err = models.AddPostReaction(postID, user.ID, req.IsLike)
	if err != nil {
		log.Printf("Failed to add reaction: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to add reaction")
		return
	}

	// Get the updated post
	post, err := models.GetPostByID(postID)
	if err != nil {
		log.Printf("Failed to get updated post: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get updated post")
		return
	}

	// Create notification for post owner if it's not their own post
	if post.UserID != user.ID {
		notificationType := models.NotificationTypeLike
		if !req.IsLike {
			notificationType = models.NotificationTypeDislike
		}

		notification := &models.Notification{
			UserID:  int64(post.UserID),
			Type:    notificationType,
			PostID:  int64(postID),
			ActorID: int64(user.ID),
			Message: fmt.Sprintf("%s %s your post", user.Nickname, map[bool]string{true: "liked", false: "disliked"}[req.IsLike]),
			Read:    false,
		}

		err = models.CreateNotification(database.DB, notification)
		if err != nil {
			// Log the error but don't fail the request
			log.Printf("Failed to create notification: %v", err)
		}
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Reaction added successfully", post)
}

// GetCategoriesHandler handles retrieving all categories
func GetCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	categories, err := models.GetAllCategories()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve categories")
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Categories retrieved successfully", categories)
}

// GetPostsByCategoryHandler handles retrieving posts filtered by category
func GetPostsByCategoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get category ID from URL
	vars := mux.Vars(r)
	categoryIDStr, ok := vars["categoryId"]
	if !ok {
		utils.RespondWithError(w, http.StatusBadRequest, "Category ID is required")
		return
	}

	categoryID, err := strconv.Atoi(categoryIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid category ID")
		return
	}

	// Get pagination parameters
	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		var err error
		page, err = strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid page number")
			return
		}
	}

	pageSize := 10
	if pageSizeStr := r.URL.Query().Get("pageSize"); pageSizeStr != "" {
		var err error
		pageSize, err = strconv.Atoi(pageSizeStr)
		if err != nil || pageSize < 1 || pageSize > 50 {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid page size")
			return
		}
	}

	// Get posts by category
	posts, err := models.GetPosts(0, categoryID, page, pageSize)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get posts")
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Posts retrieved successfully", posts)
}

// GetUserCreatedPostsHandler handles retrieving posts created by the logged-in user
func GetUserCreatedPostsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get the current user
	user, err := utils.GetUserFromRequest(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	// Get pagination parameters
	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		var err error
		page, err = strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid page number")
			return
		}
	}

	pageSize := 10
	if pageSizeStr := r.URL.Query().Get("pageSize"); pageSizeStr != "" {
		var err error
		pageSize, err = strconv.Atoi(pageSizeStr)
		if err != nil || pageSize < 1 || pageSize > 50 {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid page size")
			return
		}
	}

	// Get posts created by user
	posts, err := models.GetPosts(user.ID, 0, page, pageSize)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get posts")
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Posts retrieved successfully", posts)
}

// GetUserLikedPostsHandler handles retrieving posts liked by the logged-in user
func GetUserLikedPostsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get the current user
	user, err := utils.GetUserFromRequest(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	// Get pagination parameters
	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		var err error
		page, err = strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid page number")
			return
		}
	}

	pageSize := 10
	if pageSizeStr := r.URL.Query().Get("pageSize"); pageSizeStr != "" {
		var err error
		pageSize, err = strconv.Atoi(pageSizeStr)
		if err != nil || pageSize < 1 || pageSize > 50 {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid page size")
			return
		}
	}

	// Get liked posts
	posts, err := models.GetLikedPosts(user.ID, page, pageSize)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get liked posts")
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Posts retrieved successfully", posts)
}
