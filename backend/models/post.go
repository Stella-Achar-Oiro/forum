// backend/models/post.go
package models

import (
    "database/sql"
    "time"
)

type Post struct {
    ID        int       `json:"id"`
    UserID    int       `json:"userId"`
    Title     string    `json:"title"`
    Content   string    `json:"content"`
    Category  string    `json:"category"`
    CreatedAt time.Time `json:"createdAt"`
    User      User      `json:"user"`
    Comments  []Comment `json:"comments,omitempty"`
}

// CreatePost creates a new post
func CreatePost(db *sql.DB, post Post) (int64, error) {
    query := `INSERT INTO posts (user_id, title, content, category) VALUES (?, ?, ?, ?)`
    
    result, err := db.Exec(query, post.UserID, post.Title, post.Content, post.Category)
    if err != nil {
        return 0, err
    }
    
    return result.LastInsertId()
}

// GetAllPosts retrieves all posts with their authors
func GetAllPosts(db *sql.DB) ([]Post, error) {
    query := `
    SELECT p.id, p.user_id, p.title, p.content, p.category, p.created_at,
           u.id, u.nickname, u.email
    FROM posts p
    JOIN users u ON p.user_id = u.id
    ORDER BY p.created_at DESC`
    
    rows, err := db.Query(query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var posts []Post
    for rows.Next() {
        var post Post
        var user User
        
        err := rows.Scan(
            &post.ID, &post.UserID, &post.Title, &post.Content, &post.Category, &post.CreatedAt,
            &user.ID, &user.Nickname, &user.Email,
        )
        if err != nil {
            return nil, err
        }
        
        post.User = user
        posts = append(posts, post)
    }
    
    return posts, nil
}

// GetPostByID retrieves a post by its ID with comments
func GetPostByID(db *sql.DB, postID int) (Post, error) {
    var post Post
    
    // Get post with author
    postQuery := `
    SELECT p.id, p.user_id, p.title, p.content, p.category, p.created_at,
           u.id, u.nickname, u.email
    FROM posts p
    JOIN users u ON p.user_id = u.id
    WHERE p.id = ?`
    
    row := db.QueryRow(postQuery, postID)
    var user User
    
    err := row.Scan(
        &post.ID, &post.UserID, &post.Title, &post.Content, &post.Category, &post.CreatedAt,
        &user.ID, &user.Nickname, &user.Email,
    )
    if err != nil {
        return post, err
    }
    post.User = user
    
    // Get comments
    comments, err := GetCommentsByPostID(db, postID)
    if err != nil {
        return post, err
    }
    post.Comments = comments
    
    return post, nil
}