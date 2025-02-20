package models

import (
    "database/sql"
    "errors"
    "fmt"
    "time"
)

type Post struct {
    ID        int64     `json:"id"`
    UserID    int64     `json:"user_id"`
    Title     string    `json:"title"`
    Content   string    `json:"content"`
    Category  string    `json:"category"`
    CreatedAt time.Time `json:"created_at"`
    Author    *User     `json:"author,omitempty"`
}

type PostCreate struct {
    Title    string `json:"title"`
    Content  string `json:"content"`
    Category string `json:"category"`
}

type PostUpdate struct {
    Title    *string `json:"title,omitempty"`
    Content  *string `json:"content,omitempty"`
    Category *string `json:"category,omitempty"`
}

var (
    ErrPostNotFound = errors.New("post not found")
    ErrInvalidPost  = errors.New("invalid post data")
)

// CreatePost creates a new post in the database
func CreatePost(db *sql.DB, userID int64, post *PostCreate) (*Post, error) {
    if err := validatePost(post); err != nil {
        return nil, err
    }

    query := `
        INSERT INTO posts (user_id, title, content, category)
        VALUES (?, ?, ?, ?)
    `

    result, err := db.Exec(query, userID, post.Title, post.Content, post.Category)
    if err != nil {
        return nil, fmt.Errorf("error creating post: %w", err)
    }

    id, err := result.LastInsertId()
    if err != nil {
        return nil, fmt.Errorf("error getting post id: %w", err)
    }

    return GetPostByID(db, id)
}

// GetPostByID retrieves a post by its ID, including author information
func GetPostByID(db *sql.DB, id int64) (*Post, error) {
    query := `
        SELECT p.id, p.user_id, p.title, p.content, p.category, p.created_at,
               u.id, u.nickname, u.email
        FROM posts p
        JOIN users u ON p.user_id = u.id
        WHERE p.id = ?
    `

    var post Post
    var author User

    err := db.QueryRow(query, id).Scan(
        &post.ID,
        &post.UserID,
        &post.Title,
        &post.Content,
        &post.Category,
        &post.CreatedAt,
        &author.ID,
        &author.Nickname,
        &author.Email,
    )

    if err == sql.ErrNoRows {
        return nil, ErrPostNotFound
    }
    if err != nil {
        return nil, fmt.Errorf("error querying post: %w", err)
    }

    post.Author = &author
    return &post, nil
}

// GetPosts retrieves a list of posts with pagination
func GetPosts(db *sql.DB, page, pageSize int) ([]*Post, error) {
    query := `
        SELECT p.id, p.user_id, p.title, p.content, p.category, p.created_at,
               u.id, u.nickname, u.email
        FROM posts p
        JOIN users u ON p.user_id = u.id
        ORDER BY p.created_at DESC
        LIMIT ? OFFSET ?
    `

    offset := (page - 1) * pageSize
    rows, err := db.Query(query, pageSize, offset)
    if err != nil {
        return nil, fmt.Errorf("error querying posts: %w", err)
    }
    defer rows.Close()

    var posts []*Post
    for rows.Next() {
        var post Post
        var author User

        err := rows.Scan(
            &post.ID,
            &post.UserID,
            &post.Title,
            &post.Content,
            &post.Category,
            &post.CreatedAt,
            &author.ID,
            &author.Nickname,
            &author.Email,
        )
        if err != nil {
            return nil, fmt.Errorf("error scanning post: %w", err)
        }

        post.Author = &author
        posts = append(posts, &post)
    }

    if err = rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating posts: %w", err)
    }

    return posts, nil
}

// UpdatePost updates an existing post
func UpdatePost(db *sql.DB, id int64, userID int64, update *PostUpdate) (*Post, error) {
    // First check if post exists and belongs to user
    existing, err := GetPostByID(db, id)
    if err != nil {
        return nil, err
    }
    if existing.UserID != userID {
        return nil, errors.New("unauthorized")
    }

    // Build update query dynamically based on provided fields
    query := "UPDATE posts SET"
    args := []interface{}{}
    setFields := []string{}

    if update.Title != nil {
        setFields = append(setFields, " title = ?")
        args = append(args, *update.Title)
    }
    if update.Content != nil {
        setFields = append(setFields, " content = ?")
        args = append(args, *update.Content)
    }
    if update.Category != nil {
        setFields = append(setFields, " category = ?")
        args = append(args, *update.Category)
    }

    if len(setFields) == 0 {
        return existing, nil // Nothing to update
    }

    // Add WHERE clause
    query = query + fmt.Sprintf("%s WHERE id = ? AND user_id = ?", setFields[0])
    args = append(args, id, userID)

    _, err = db.Exec(query, args...)
    if err != nil {
        return nil, fmt.Errorf("error updating post: %w", err)
    }

    return GetPostByID(db, id)
}

// DeletePost deletes a post
func DeletePost(db *sql.DB, id int64, userID int64) error {
    result, err := db.Exec("DELETE FROM posts WHERE id = ? AND user_id = ?", id, userID)
    if err != nil {
        return fmt.Errorf("error deleting post: %w", err)
    }

    rows, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("error checking deleted rows: %w", err)
    }

    if rows == 0 {
        return ErrPostNotFound
    }

    return nil
}

// Helper functions

func validatePost(post *PostCreate) error {
    if post.Title == "" || post.Content == "" || post.Category == "" {
        return ErrInvalidPost
    }
    // Add more validation as needed
    return nil
}

// GetPostsByCategory retrieves posts filtered by category
func GetPostsByCategory(db *sql.DB, category string, page, pageSize int) ([]*Post, error) {
    query := `
        SELECT p.id, p.user_id, p.title, p.content, p.category, p.created_at,
               u.id, u.nickname, u.email
        FROM posts p
        JOIN users u ON p.user_id = u.id
        WHERE p.category = ?
        ORDER BY p.created_at DESC
        LIMIT ? OFFSET ?
    `

    offset := (page - 1) * pageSize
    rows, err := db.Query(query, category, pageSize, offset)
    if err != nil {
        return nil, fmt.Errorf("error querying posts by category: %w", err)
    }
    defer rows.Close()

    var posts []*Post
    for rows.Next() {
        var post Post
        var author User

        err := rows.Scan(
            &post.ID,
            &post.UserID,
            &post.Title,
            &post.Content,
            &post.Category,
            &post.CreatedAt,
            &author.ID,
            &author.Nickname,
            &author.Email,
        )
        if err != nil {
            return nil, fmt.Errorf("error scanning post: %w", err)
        }

        post.Author = &author
        posts = append(posts, &post)
    }

    if err = rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating posts: %w", err)
    }

    return posts, nil
}