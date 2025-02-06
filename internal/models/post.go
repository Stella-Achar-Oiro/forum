package models

import (
	"time"
)

// Post represents a forum post
type Post struct {
	ID         int64     `json:"id"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	UserID     int64     `json:"user_id"`
	Username   string    `json:"username"`
	CreatedAt  time.Time `json:"created_at"`
	Categories []string  `json:"categories"`
	Likes      int       `json:"likes"`
	Dislikes   int       `json:"dislikes"`
}

// PostInput represents the data needed to create a new post
type PostInput struct {
	Title      string   `json:"title"`
	Content    string   `json:"content"`
	Categories []string `json:"categories"`
}

// Comment represents a comment on a post
type Comment struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	UserID    int64     `json:"user_id"`
	Username  string    `json:"username"`
	PostID    int64     `json:"post_id"`
	CreatedAt time.Time `json:"created_at"`
	Likes     int       `json:"likes"`
	Dislikes  int       `json:"dislikes"`
}

// CommentInput represents the data needed to create a new comment
type CommentInput struct {
	Content string `json:"content"`
	PostID  int64  `json:"post_id"`
}

// Category represents a post category
type Category struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Like represents a like or dislike on a post or comment
type Like struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	PostID    *int64    `json:"post_id,omitempty"`
	CommentID *int64    `json:"comment_id,omitempty"`
	IsLike    bool      `json:"is_like"`
	CreatedAt time.Time `json:"created_at"`
}
