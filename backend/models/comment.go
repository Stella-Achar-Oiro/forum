// backend/models/comment.go
package models

import (
    "database/sql"
    "time"
)

type Comment struct {
    ID        int       `json:"id"`
    PostID    int       `json:"postId"`
    UserID    int       `json:"userId"`
    Content   string    `json:"content"`
    CreatedAt time.Time `json:"createdAt"`
    User      User      `json:"user"`
}

// CreateComment adds a new comment to a post
func CreateComment(db *sql.DB, comment Comment) (int64, error) {
    query := `INSERT INTO comments (post_id, user_id, content) VALUES (?, ?, ?)`
    
    result, err := db.Exec(query, comment.PostID, comment.UserID, comment.Content)
    if err != nil {
        return 0, err
    }
    
    return result.LastInsertId()
}

// GetCommentsByPostID retrieves all comments for a specific post
func GetCommentsByPostID(db *sql.DB, postID int) ([]Comment, error) {
    query := `
    SELECT c.id, c.post_id, c.user_id, c.content, c.created_at,
           u.id, u.nickname, u.email
    FROM comments c
    JOIN users u ON c.user_id = u.id
    WHERE c.post_id = ?
    ORDER BY c.created_at ASC`
    
    rows, err := db.Query(query, postID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var comments []Comment
    for rows.Next() {
        var comment Comment
        var user User
        
        err := rows.Scan(
            &comment.ID, &comment.PostID, &comment.UserID, &comment.Content, &comment.CreatedAt,
            &user.ID, &user.Nickname, &user.Email,
        )
        if err != nil {
            return nil, err
        }
        
        comment.User = user
        comments = append(comments, comment)
    }
    
    return comments, nil
}