package models

import (
    "database/sql"
    "errors"
    "fmt"
    "time"
)

type Comment struct {
    ID        int64     `json:"id"`
    PostID    int64     `json:"post_id"`
    UserID    int64     `json:"user_id"`
    Content   string    `json:"content"`
    CreatedAt time.Time `json:"created_at"`
    Author    *User     `json:"author,omitempty"`
}

type CommentCreate struct {
    PostID  int64  `json:"post_id"`
    Content string `json:"content"`
}

type CommentUpdate struct {
    Content string `json:"content"`
}

var (
    ErrCommentNotFound = errors.New("comment not found")
    ErrInvalidComment  = errors.New("invalid comment data")
)

// CreateComment creates a new comment
func CreateComment(db *sql.DB, userID int64, comment *CommentCreate) (*Comment, error) {
    // Validate comment
    if err := validateComment(comment); err != nil {
        return nil, err
    }

    // Verify post exists
    exists, err := postExists(db, comment.PostID)
    if err != nil {
        return nil, err
    }
    if !exists {
        return nil, errors.New("post not found")
    }

    query := `
        INSERT INTO comments (post_id, user_id, content)
        VALUES (?, ?, ?)
    `

    result, err := db.Exec(query, comment.PostID, userID, comment.Content)
    if err != nil {
        return nil, fmt.Errorf("error creating comment: %w", err)
    }

    id, err := result.LastInsertId()
    if err != nil {
        return nil, fmt.Errorf("error getting comment id: %w", err)
    }

    return GetCommentByID(db, id)
}

// GetCommentByID retrieves a comment by its ID
func GetCommentByID(db *sql.DB, id int64) (*Comment, error) {
    query := `
        SELECT c.id, c.post_id, c.user_id, c.content, c.created_at,
               u.id, u.nickname, u.email
        FROM comments c
        JOIN users u ON c.user_id = u.id
        WHERE c.id = ?
    `

    var comment Comment
    var author User

    err := db.QueryRow(query, id).Scan(
        &comment.ID,
        &comment.PostID,
        &comment.UserID,
        &comment.Content,
        &comment.CreatedAt,
        &author.ID,
        &author.Nickname,
        &author.Email,
    )

    if err == sql.ErrNoRows {
        return nil, ErrCommentNotFound
    }
    if err != nil {
        return nil, fmt.Errorf("error querying comment: %w", err)
    }

    comment.Author = &author
    return &comment, nil
}

// GetCommentsByPostID retrieves all comments for a post
func GetCommentsByPostID(db *sql.DB, postID int64, page, pageSize int) ([]*Comment, error) {
    query := `
        SELECT c.id, c.post_id, c.user_id, c.content, c.created_at,
               u.id, u.nickname, u.email
        FROM comments c
        JOIN users u ON c.user_id = u.id
        WHERE c.post_id = ?
        ORDER BY c.created_at DESC
        LIMIT ? OFFSET ?
    `

    offset := (page - 1) * pageSize
    rows, err := db.Query(query, postID, pageSize, offset)
    if err != nil {
        return nil, fmt.Errorf("error querying comments: %w", err)
    }
    defer rows.Close()

    var comments []*Comment
    for rows.Next() {
        var comment Comment
        var author User

        err := rows.Scan(
            &comment.ID,
            &comment.PostID,
            &comment.UserID,
            &comment.Content,
            &comment.CreatedAt,
            &author.ID,
            &author.Nickname,
            &author.Email,
        )
        if err != nil {
            return nil, fmt.Errorf("error scanning comment: %w", err)
        }

        comment.Author = &author
        comments = append(comments, &comment)
    }

    if err = rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating comments: %w", err)
    }

    return comments, nil
}

// UpdateComment updates a comment
func UpdateComment(db *sql.DB, id int64, userID int64, update *CommentUpdate) (*Comment, error) {
    // Verify comment exists and belongs to user
    comment, err := GetCommentByID(db, id)
    if err != nil {
        return nil, err
    }
    if comment.UserID != userID {
        return nil, errors.New("unauthorized")
    }

    // Update comment
    query := "UPDATE comments SET content = ? WHERE id = ? AND user_id = ?"
    _, err = db.Exec(query, update.Content, id, userID)
    if err != nil {
        return nil, fmt.Errorf("error updating comment: %w", err)
    }

    return GetCommentByID(db, id)
}

// DeleteComment deletes a comment
func DeleteComment(db *sql.DB, id int64, userID int64) error {
    result, err := db.Exec("DELETE FROM comments WHERE id = ? AND user_id = ?", id, userID)
    if err != nil {
        return fmt.Errorf("error deleting comment: %w", err)
    }

    rows, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("error checking deleted rows: %w", err)
    }

    if rows == 0 {
        return ErrCommentNotFound
    }

    return nil
}

// Helper functions

func validateComment(comment *CommentCreate) error {
    if comment.Content == "" || comment.PostID == 0 {
        return ErrInvalidComment
    }
    return nil
}

func postExists(db *sql.DB, postID int64) (bool, error) {
    var exists bool
    query := "SELECT EXISTS(SELECT 1 FROM posts WHERE id = ?)"
    err := db.QueryRow(query, postID).Scan(&exists)
    if err != nil {
        return false, fmt.Errorf("error checking post existence: %w", err)
    }
    return exists, nil
}