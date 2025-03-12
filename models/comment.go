package models

import (
	"database/sql"
	"log"
	"time"

	"github.com/Stella-Achar-Oiro/forum/database"
)

// Comment represents a comment on a post
type Comment struct {
	ID           int                   `json:"id"`
	PostID       int                   `json:"postId"`
	UserID       int                   `json:"userId"`
	ParentID     *int                  `json:"parentId,omitempty"`
	Content      string                `json:"content"`
	CreatedAt    time.Time             `json:"createdAt"`
	User         UserForPublic         `json:"user"`
	LikeCount    int                   `json:"likeCount"`
	DislikeCount int                   `json:"dislikeCount"`
	Replies      []CommentWithReaction `json:"replies,omitempty"`
}

// CommentWithReaction extends Comment with the user's reaction
type CommentWithReaction struct {
	Comment
	UserReaction string `json:"userReaction"` // "like", "dislike" or ""
}

// CreateCommentRequest represents a request to create a comment
type CreateCommentRequest struct {
	Content  string `json:"content"`
	ParentID *int   `json:"parentId,omitempty"`
}

// CreateComment creates a new comment in the database
func CreateComment(comment Comment) (int, error) {
	result, err := database.DB.Exec(
		"INSERT INTO comments (post_id, user_id, parent_id, content) VALUES (?, ?, ?, ?)",
		comment.PostID, comment.UserID, comment.ParentID, comment.Content,
	)
	if err != nil {
		return 0, err
	}

	// Get the comment ID
	commentID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(commentID), nil
}

// GetCommentsByPostID retrieves all comments for a post
func GetCommentsByPostID(postID int, currentUserID int) ([]CommentWithReaction, error) {
	log.Printf("Getting comments for post %d with currentUserID %d", postID, currentUserID)

	// First, get all top-level comments (no parent_id)
	rows, err := database.DB.Query(`
		SELECT c.id, c.post_id, c.user_id, c.parent_id, c.content, c.created_at,
		       u.id, u.nickname, u.first_name, u.last_name, u.is_online
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.post_id = ? AND c.parent_id IS NULL
		ORDER BY c.created_at ASC
	`, postID)
	if err != nil {
		log.Printf("Error querying comments: %v", err)
		return nil, err
	}
	defer rows.Close()

	var comments []CommentWithReaction
	for rows.Next() {
		var comment CommentWithReaction
		var parentID sql.NullInt64
		if err := rows.Scan(
			&comment.ID, &comment.PostID, &comment.UserID, &parentID, &comment.Content, &comment.CreatedAt,
			&comment.User.ID, &comment.User.Nickname, &comment.User.FirstName, &comment.User.LastName, &comment.User.IsOnline,
		); err != nil {
			log.Printf("Error scanning comment row: %v", err)
			return nil, err
		}
		if parentID.Valid {
			comment.ParentID = &[]int{int(parentID.Int64)}[0]
		}
		comments = append(comments, comment)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error after iterating rows: %v", err)
		return nil, err
	}

	log.Printf("Found %d top-level comments", len(comments))

	// For each comment, get reactions and replies
	for i := range comments {
		// Get reaction counts
		log.Printf("Getting reactions for comment %d", comments[i].ID)
		var likes, dislikes int
		err = database.DB.QueryRow(`
			SELECT 
				(SELECT COUNT(*) FROM comment_reactions WHERE comment_id = ? AND reaction_type = 'like'),
				(SELECT COUNT(*) FROM comment_reactions WHERE comment_id = ? AND reaction_type = 'dislike')
		`, comments[i].ID, comments[i].ID).Scan(&likes, &dislikes)
		if err == nil {
			comments[i].LikeCount = likes
			comments[i].DislikeCount = dislikes
		} else {
			log.Printf("Error getting reaction counts (using defaults): %v", err)
		}

		// Get user's own reaction if they are logged in
		if currentUserID > 0 {
			var reactionType string
			err = database.DB.QueryRow(`
				SELECT reaction_type FROM comment_reactions
				WHERE user_id = ? AND comment_id = ?
			`, currentUserID, comments[i].ID).Scan(&reactionType)
			if err == nil {
				comments[i].UserReaction = reactionType
			} else {
				log.Printf("No reaction found for user %d on comment %d", currentUserID, comments[i].ID)
			}
		}

		// Get replies
		replies, err := GetCommentReplies(comments[i].ID, currentUserID)
		if err != nil {
			log.Printf("Error getting replies for comment %d: %v", comments[i].ID, err)
		} else {
			comments[i].Replies = replies
		}
	}

	log.Printf("Successfully retrieved all comments, reactions, and replies")
	return comments, nil
}

// GetCommentReplies retrieves all replies for a comment
func GetCommentReplies(parentID int, currentUserID int) ([]CommentWithReaction, error) {
	rows, err := database.DB.Query(`
		SELECT c.id, c.post_id, c.user_id, c.parent_id, c.content, c.created_at,
		       u.id, u.nickname, u.first_name, u.last_name, u.is_online
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.parent_id = ?
		ORDER BY c.created_at ASC
	`, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var replies []CommentWithReaction
	for rows.Next() {
		var reply CommentWithReaction
		var parentID sql.NullInt64
		if err := rows.Scan(
			&reply.ID, &reply.PostID, &reply.UserID, &parentID, &reply.Content, &reply.CreatedAt,
			&reply.User.ID, &reply.User.Nickname, &reply.User.FirstName, &reply.User.LastName, &reply.User.IsOnline,
		); err != nil {
			return nil, err
		}
		if parentID.Valid {
			reply.ParentID = &[]int{int(parentID.Int64)}[0]
		}

		// Get reaction counts
		var likes, dislikes int
		err = database.DB.QueryRow(`
			SELECT 
				(SELECT COUNT(*) FROM comment_reactions WHERE comment_id = ? AND reaction_type = 'like'),
				(SELECT COUNT(*) FROM comment_reactions WHERE comment_id = ? AND reaction_type = 'dislike')
		`, reply.ID, reply.ID).Scan(&likes, &dislikes)
		if err == nil {
			reply.LikeCount = likes
			reply.DislikeCount = dislikes
		}

		// Get user's own reaction if they are logged in
		if currentUserID > 0 {
			var reactionType string
			err = database.DB.QueryRow(`
				SELECT reaction_type FROM comment_reactions
				WHERE user_id = ? AND comment_id = ?
			`, currentUserID, reply.ID).Scan(&reactionType)
			if err == nil {
				reply.UserReaction = reactionType
			}
		}

		// Recursively get nested replies
		nestedReplies, err := GetCommentReplies(reply.ID, currentUserID)
		if err == nil && len(nestedReplies) > 0 {
			reply.Replies = nestedReplies
		}

		replies = append(replies, reply)
	}

	return replies, nil
}

// AddReactionToComment adds a like or dislike to a comment
func AddReactionToComment(userID, commentID int, reactionType string) error {
	log.Printf("Adding reaction to comment %d by user %d: %s", commentID, userID, reactionType)

	// Check if a reaction already exists
	var existingReaction string
	err := database.DB.QueryRow("SELECT reaction_type FROM comment_reactions WHERE user_id = ? AND comment_id = ?", userID, commentID).Scan(&existingReaction)

	if err == nil {
		log.Printf("Updating existing reaction from %s to %s", existingReaction, reactionType)
		// Update existing reaction
		_, err = database.DB.Exec("UPDATE comment_reactions SET reaction_type = ? WHERE user_id = ? AND comment_id = ?", reactionType, userID, commentID)
		if err != nil {
			log.Printf("Error updating reaction: %v", err)
		}
		return err
	}

	log.Printf("Creating new reaction")
	// Insert new reaction
	_, err = database.DB.Exec("INSERT INTO comment_reactions (user_id, comment_id, reaction_type) VALUES (?, ?, ?)", userID, commentID, reactionType)
	if err != nil {
		log.Printf("Error inserting reaction: %v", err)
	}
	return err
}

// GetCommentByID retrieves a comment by its ID
func GetCommentByID(commentID int) (CommentWithReaction, error) {
	log.Printf("Getting comment by ID: %d", commentID)

	var comment CommentWithReaction
	var parentID sql.NullInt64

	err := database.DB.QueryRow(`
		SELECT c.id, c.post_id, c.user_id, c.parent_id, c.content, c.created_at,
		       u.id, u.nickname, u.first_name, u.last_name, u.is_online
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.id = ?`,
		commentID,
	).Scan(
		&comment.ID, &comment.PostID, &comment.UserID, &parentID, &comment.Content, &comment.CreatedAt,
		&comment.User.ID, &comment.User.Nickname, &comment.User.FirstName, &comment.User.LastName, &comment.User.IsOnline,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Comment not found with ID: %d", commentID)
			return CommentWithReaction{}, err
		}
		log.Printf("Failed to get comment: %v", err)
		return CommentWithReaction{}, err
	}

	if parentID.Valid {
		comment.ParentID = &[]int{int(parentID.Int64)}[0]
	}

	// Get reaction counts
	var likes, dislikes int
	err = database.DB.QueryRow(`
		SELECT 
			(SELECT COUNT(*) FROM comment_reactions WHERE comment_id = ? AND reaction_type = 'like'),
			(SELECT COUNT(*) FROM comment_reactions WHERE comment_id = ? AND reaction_type = 'dislike')
	`, comment.ID, comment.ID).Scan(&likes, &dislikes)
	if err == nil {
		comment.LikeCount = likes
		comment.DislikeCount = dislikes
	}

	return comment, nil
}

// GetUserComments retrieves all comments made by a user
func GetUserComments(userID int) ([]Comment, error) {
	rows, err := database.DB.Query(`
		SELECT c.id, c.post_id, c.user_id, c.parent_id, c.content, c.created_at,
		       u.id, u.nickname, u.first_name, u.last_name, u.is_online
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.user_id = ?
		ORDER BY c.created_at DESC`,
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var comment Comment
		var parentID sql.NullInt64
		err := rows.Scan(
			&comment.ID, &comment.PostID, &comment.UserID, &parentID, &comment.Content, &comment.CreatedAt,
			&comment.User.ID, &comment.User.Nickname, &comment.User.FirstName, &comment.User.LastName, &comment.User.IsOnline,
		)
		if err != nil {
			return nil, err
		}

		if parentID.Valid {
			comment.ParentID = &[]int{int(parentID.Int64)}[0]
		}

		// Get reaction counts
		var likes, dislikes int
		err = database.DB.QueryRow(`
			SELECT 
				(SELECT COUNT(*) FROM comment_reactions WHERE comment_id = ? AND reaction_type = 'like'),
				(SELECT COUNT(*) FROM comment_reactions WHERE comment_id = ? AND reaction_type = 'dislike')
		`, comment.ID, comment.ID).Scan(&likes, &dislikes)
		if err == nil {
			comment.LikeCount = likes
			comment.DislikeCount = dislikes
		}

		comments = append(comments, comment)
	}

	return comments, nil
}
