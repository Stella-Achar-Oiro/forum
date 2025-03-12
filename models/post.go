package models

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/Stella-Achar-Oiro/forum/database"
)

// Post represents a post in the forum
type Post struct {
	ID         int        `json:"id"`
	UserID     int        `json:"userId"`
	Title      string     `json:"title"`
	Content    string     `json:"content"`
	ImageURL   string     `json:"imageUrl,omitempty"`
	VideoURL   string     `json:"videoUrl,omitempty"`
	GiphyID    string     `json:"giphyId,omitempty"`
	GiphyURL   string     `json:"giphyUrl,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
	Likes      int        `json:"likes"`
	Dislikes   int        `json:"dislikes"`
	Categories []Category `json:"categories"`
}

// PostWithUser represents a post with user information
type PostWithUser struct {
	Post
	Author UserForPublic `json:"author"`
}

// Category represents a post category
type Category struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// PostWithReaction extends Post with the user's reaction
type PostWithReaction struct {
	Post
	UserReaction string `json:"userReaction"` // "like", "dislike" or ""
}

// CreatePost creates a new post in the database
func CreatePost(post Post) (int, error) {
	log.Printf("Creating post with title: %s", post.Title)

	// Format timestamps in ISO 8601 format
	now := time.Now().UTC().Format(time.RFC3339)

	result, err := database.DB.Exec(
		"INSERT INTO posts (user_id, title, content, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
		post.UserID, post.Title, post.Content, now, now,
	)
	if err != nil {
		log.Printf("Failed to create post: %v", err)
		return 0, err
	}

	postID, err := result.LastInsertId()
	if err != nil {
		log.Printf("Failed to get last insert ID: %v", err)
		return 0, err
	}

	log.Printf("Successfully created post with ID: %d", postID)
	return int(postID), nil
}

// GetPostByID retrieves a post by its ID
func GetPostByID(id int) (PostWithUser, error) {
	log.Printf("Getting post by ID: %d", id)

	var post PostWithUser
	var createdAtStr, updatedAtStr string

	// Join with users table to get author information
	err := database.DB.QueryRow(`
		SELECT p.id, p.user_id, p.title, p.content, p.image_url, p.video_url, 
			   p.giphy_id, p.giphy_url, p.created_at, p.updated_at, p.likes, p.dislikes,
			   u.id, u.nickname, u.first_name, u.last_name, u.is_online
		FROM posts p
		JOIN users u ON p.user_id = u.id
		WHERE p.id = ?`,
		id,
	).Scan(
		&post.ID, &post.UserID, &post.Title, &post.Content, &post.ImageURL,
		&post.VideoURL, &post.GiphyID, &post.GiphyURL, &createdAtStr, &updatedAtStr,
		&post.Likes, &post.Dislikes,
		&post.Author.ID, &post.Author.Nickname, &post.Author.FirstName, &post.Author.LastName,
		&post.Author.IsOnline,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Post not found with ID: %d", id)
			return PostWithUser{}, err
		}
		log.Printf("Failed to get post: %v", err)
		return PostWithUser{}, err
	}

	// Parse timestamps
	formats := []string{
		time.RFC3339,                // ISO 8601 format
		"2006-01-02 15:04:05",       // SQLite default format
		"2006-01-02T15:04:05Z07:00", // Another common format
	}

	var parseErr error
	for _, format := range formats {
		post.CreatedAt, parseErr = time.Parse(format, createdAtStr)
		if parseErr == nil {
			break
		}
	}
	if parseErr != nil {
		log.Printf("Failed to parse created_at timestamp: %v", parseErr)
		return PostWithUser{}, parseErr
	}

	for _, format := range formats {
		post.UpdatedAt, parseErr = time.Parse(format, updatedAtStr)
		if parseErr == nil {
			break
		}
	}
	if parseErr != nil {
		log.Printf("Failed to parse updated_at timestamp: %v", parseErr)
		return PostWithUser{}, parseErr
	}

	// Get categories for the post
	categories, err := GetPostCategories(id)
	if err != nil {
		log.Printf("Failed to get post categories: %v", err)
		return PostWithUser{}, err
	}
	post.Categories = categories

	return post, nil
}

// GetPosts retrieves posts with optional filtering and pagination
func GetPosts(userID int, categoryID int, page, pageSize int) ([]PostWithUser, error) {
	log.Printf("Getting posts (userID: %d, categoryID: %d, page: %d)", userID, categoryID, page)

	query := `
		SELECT DISTINCT p.id, p.user_id, p.title, p.content, p.image_url, p.created_at, p.updated_at, p.likes, p.dislikes,
			   u.id, u.nickname, u.first_name, u.last_name, u.is_online
		FROM posts p
		JOIN users u ON p.user_id = u.id`

	var whereConditions []string
	var args []interface{}

	if userID > 0 {
		whereConditions = append(whereConditions, "p.user_id = ?")
		args = append(args, userID)
	}

	if categoryID > 0 {
		query += " JOIN post_categories pc ON p.id = pc.post_id"
		whereConditions = append(whereConditions, "pc.category_id = ?")
		args = append(args, categoryID)
	}

	if len(whereConditions) > 0 {
		query += " WHERE " + whereConditions[0]
		for i := 1; i < len(whereConditions); i++ {
			query += " AND " + whereConditions[i]
		}
	}

	query += " ORDER BY p.created_at DESC LIMIT ? OFFSET ?"
	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		log.Printf("Failed to get posts: %v", err)
		return nil, err
	}
	defer rows.Close()

	var posts []PostWithUser
	for rows.Next() {
		var post PostWithUser
		var createdAtStr, updatedAtStr string

		err := rows.Scan(
			&post.ID, &post.UserID, &post.Title, &post.Content, &post.ImageURL, &createdAtStr, &updatedAtStr,
			&post.Likes, &post.Dislikes,
			&post.Author.ID, &post.Author.Nickname, &post.Author.FirstName, &post.Author.LastName,
			&post.Author.IsOnline,
		)
		if err != nil {
			log.Printf("Failed to scan post row: %v", err)
			return nil, err
		}

		// Parse timestamps
		formats := []string{
			time.RFC3339,                // ISO 8601 format
			"2006-01-02 15:04:05",       // SQLite default format
			"2006-01-02T15:04:05Z07:00", // Another common format
		}

		var parseErr error
		for _, format := range formats {
			post.CreatedAt, parseErr = time.Parse(format, createdAtStr)
			if parseErr == nil {
				break
			}
		}
		if parseErr != nil {
			log.Printf("Failed to parse created_at timestamp: %v", parseErr)
			return nil, parseErr
		}

		for _, format := range formats {
			post.UpdatedAt, parseErr = time.Parse(format, updatedAtStr)
			if parseErr == nil {
				break
			}
		}
		if parseErr != nil {
			log.Printf("Failed to parse updated_at timestamp: %v", parseErr)
			return nil, parseErr
		}

		// Get categories for the post
		categories, err := GetPostCategories(post.ID)
		if err != nil {
			log.Printf("Failed to get categories for post %d: %v", post.ID, err)
			return nil, err
		}
		post.Categories = categories

		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating post rows: %v", err)
		return nil, err
	}

	return posts, nil
}

// AddPostReaction adds a like or dislike to a post
func AddPostReaction(postID, userID int, isLike bool) error {
	log.Printf("Adding reaction to post %d by user %d (isLike: %v)", postID, userID, isLike)

	// Start a transaction
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Check if user has already reacted to this post
	var existingReaction bool
	err = tx.QueryRow(
		"SELECT CASE WHEN is_like = ? THEN 1 ELSE 0 END FROM post_reactions WHERE post_id = ? AND user_id = ?",
		isLike, postID, userID,
	).Scan(&existingReaction)

	if err == nil {
		if existingReaction {
			// User has already made this reaction
			return nil
		}
		// User has made the opposite reaction, remove it
		_, err = tx.Exec("DELETE FROM post_reactions WHERE post_id = ? AND user_id = ?", postID, userID)
		if err != nil {
			return err
		}
	} else if err != sql.ErrNoRows {
		return err
	}

	// Add the new reaction
	_, err = tx.Exec(
		"INSERT INTO post_reactions (post_id, user_id, is_like) VALUES (?, ?, ?)",
		postID, userID, isLike,
	)
	if err != nil {
		return err
	}

	// Update the post's like/dislike count
	_, err = tx.Exec(`
		UPDATE posts SET 
			likes = (SELECT COUNT(*) FROM post_reactions WHERE post_id = ? AND is_like = 1),
			dislikes = (SELECT COUNT(*) FROM post_reactions WHERE post_id = ? AND is_like = 0)
		WHERE id = ?`,
		postID, postID, postID,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetAllCategories retrieves all categories
func GetAllCategories() ([]Category, error) {
	rows, err := database.DB.Query("SELECT id, name, description FROM categories")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var category Category
		if err := rows.Scan(&category.ID, &category.Name, &category.Description); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

// CreatePostTx creates a new post in the database within a transaction
func CreatePostTx(tx *sql.Tx, post Post) (int, error) {
	log.Printf("Creating post with title: %s", post.Title)

	// Format timestamps in ISO 8601 format
	now := time.Now().UTC().Format(time.RFC3339)

	result, err := tx.Exec(
		`INSERT INTO posts (
			user_id, title, content, image_url, video_url, giphy_id, giphy_url, 
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		post.UserID, post.Title, post.Content, post.ImageURL, post.VideoURL,
		post.GiphyID, post.GiphyURL, now, now,
	)
	if err != nil {
		log.Printf("Failed to create post: %v", err)
		return 0, err
	}

	postID, err := result.LastInsertId()
	if err != nil {
		log.Printf("Failed to get last insert ID: %v", err)
		return 0, err
	}

	log.Printf("Successfully created post with ID: %d", postID)
	return int(postID), nil
}

// AssociatePostCategories associates categories with a post
func AssociatePostCategories(tx *sql.Tx, postID int, categoryIDs []int) error {
	log.Printf("Associating categories %v with post %d", categoryIDs, postID)

	// Verify that all categories exist
	for _, categoryID := range categoryIDs {
		var exists bool
		err := tx.QueryRow("SELECT EXISTS(SELECT 1 FROM categories WHERE id = ?)", categoryID).Scan(&exists)
		if err != nil {
			log.Printf("Failed to check category existence: %v", err)
			return err
		}
		if !exists {
			log.Printf("Category %d does not exist", categoryID)
			return fmt.Errorf("category %d does not exist", categoryID)
		}
	}

	// Associate categories with the post
	for _, categoryID := range categoryIDs {
		_, err := tx.Exec(
			"INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)",
			postID, categoryID,
		)
		if err != nil {
			log.Printf("Failed to associate category %d with post %d: %v", categoryID, postID, err)
			return err
		}
	}

	log.Printf("Successfully associated categories with post %d", postID)
	return nil
}

// GetPostCategories retrieves all categories associated with a post
func GetPostCategories(postID int) ([]Category, error) {
	log.Printf("Getting categories for post %d", postID)

	rows, err := database.DB.Query(`
		SELECT c.id, c.name, c.description
		FROM categories c
		JOIN post_categories pc ON c.id = pc.category_id
		WHERE pc.post_id = ?`,
		postID,
	)
	if err != nil {
		log.Printf("Failed to get post categories: %v", err)
		return nil, err
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var category Category
		if err := rows.Scan(&category.ID, &category.Name, &category.Description); err != nil {
			log.Printf("Failed to scan category row: %v", err)
			return nil, err
		}
		categories = append(categories, category)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating category rows: %v", err)
		return nil, err
	}

	return categories, nil
}

// GetLikedPosts retrieves posts that a user has liked
func GetLikedPosts(userID int, page, pageSize int) ([]PostWithUser, error) {
	log.Printf("Getting liked posts for user %d (page: %d)", userID, page)

	query := `
		SELECT DISTINCT p.id, p.user_id, p.title, p.content, p.image_url, p.video_url,
			   p.giphy_id, p.giphy_url, p.created_at, p.updated_at, p.likes, p.dislikes,
			   u.id, u.nickname, u.first_name, u.last_name, u.is_online
		FROM posts p
		JOIN users u ON p.user_id = u.id
		JOIN post_reactions pr ON p.id = pr.post_id
		WHERE pr.user_id = ? AND pr.is_like = 1
		ORDER BY p.created_at DESC
		LIMIT ? OFFSET ?`

	rows, err := database.DB.Query(query, userID, pageSize, (page-1)*pageSize)
	if err != nil {
		log.Printf("Failed to get liked posts: %v", err)
		return nil, err
	}
	defer rows.Close()

	var posts []PostWithUser
	for rows.Next() {
		var post PostWithUser
		var createdAtStr, updatedAtStr string
		var imageURL, videoURL, giphyID, giphyURL sql.NullString

		err := rows.Scan(
			&post.ID, &post.UserID, &post.Title, &post.Content, &imageURL, &videoURL,
			&giphyID, &giphyURL, &createdAtStr, &updatedAtStr, &post.Likes, &post.Dislikes,
			&post.Author.ID, &post.Author.Nickname, &post.Author.FirstName, &post.Author.LastName,
			&post.Author.IsOnline,
		)
		if err != nil {
			log.Printf("Failed to scan post row: %v", err)
			return nil, err
		}

		if imageURL.Valid {
			post.ImageURL = imageURL.String
		}
		if videoURL.Valid {
			post.VideoURL = videoURL.String
		}
		if giphyID.Valid {
			post.GiphyID = giphyID.String
		}
		if giphyURL.Valid {
			post.GiphyURL = giphyURL.String
		}

		// Parse timestamps
		formats := []string{
			time.RFC3339,                // ISO 8601 format
			"2006-01-02 15:04:05",       // SQLite default format
			"2006-01-02T15:04:05Z07:00", // Another common format
		}

		var parseErr error
		for _, format := range formats {
			post.CreatedAt, parseErr = time.Parse(format, createdAtStr)
			if parseErr == nil {
				break
			}
		}
		if parseErr != nil {
			log.Printf("Failed to parse created_at timestamp: %v", parseErr)
			return nil, parseErr
		}

		for _, format := range formats {
			post.UpdatedAt, parseErr = time.Parse(format, updatedAtStr)
			if parseErr == nil {
				break
			}
		}
		if parseErr != nil {
			log.Printf("Failed to parse updated_at timestamp: %v", parseErr)
			return nil, parseErr
		}

		// Get categories for the post
		categories, err := GetPostCategories(post.ID)
		if err != nil {
			log.Printf("Failed to get categories for post %d: %v", post.ID, err)
			return nil, err
		}
		post.Categories = categories

		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating post rows: %v", err)
		return nil, err
	}

	return posts, nil
}

// GetPostsByUserID retrieves all posts created by a specific user
func GetPostsByUserID(userID int) ([]Post, error) {
	rows, err := database.DB.Query(`
		SELECT id, user_id, title, content, image_url, video_url, giphy_id, giphy_url, created_at, updated_at, likes, dislikes
		FROM posts 
		WHERE user_id = ?
		ORDER BY created_at DESC`,
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		var createdAtStr, updatedAtStr string
		var imageURL, videoURL, giphyID, giphyURL sql.NullString

		err := rows.Scan(
			&post.ID, &post.UserID, &post.Title, &post.Content, &imageURL, &videoURL,
			&giphyID, &giphyURL, &createdAtStr, &updatedAtStr, &post.Likes, &post.Dislikes,
		)
		if err != nil {
			return nil, err
		}

		if imageURL.Valid {
			post.ImageURL = imageURL.String
		}
		if videoURL.Valid {
			post.VideoURL = videoURL.String
		}
		if giphyID.Valid {
			post.GiphyID = giphyID.String
		}
		if giphyURL.Valid {
			post.GiphyURL = giphyURL.String
		}

		// Parse timestamps
		formats := []string{
			time.RFC3339,
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05Z07:00",
		}

		for _, format := range formats {
			post.CreatedAt, err = time.Parse(format, createdAtStr)
			if err == nil {
				break
			}
		}

		for _, format := range formats {
			post.UpdatedAt, err = time.Parse(format, updatedAtStr)
			if err == nil {
				break
			}
		}

		// Get categories for the post
		categories, err := GetPostCategories(post.ID)
		if err == nil {
			post.Categories = categories
		}

		posts = append(posts, post)
	}

	return posts, nil
}
