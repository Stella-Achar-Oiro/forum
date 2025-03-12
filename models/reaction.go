package models

import (
	"database/sql"
	"log"
	"time"

	"github.com/Stella-Achar-Oiro/forum/database"
)

// Reaction represents a user's reaction to a post or comment
type Reaction struct {
	ID           int       `json:"id"`
	UserID       int       `json:"userId"`
	PostID       *int      `json:"postId,omitempty"`
	CommentID    *int      `json:"commentId,omitempty"`
	ReactionType string    `json:"reactionType"` // "like" or "dislike"
	CreatedAt    time.Time `json:"createdAt"`
}

// GetUserReactions retrieves all reactions made by a user
func GetUserReactions(userID int) ([]Reaction, error) {
	log.Printf("Getting reactions for user %d", userID)

	// Combine post and comment reactions using a subquery
	query := `
		SELECT type, user_id, target_id, comment_id, reaction_type, created_at
		FROM (
			SELECT 'post' as type, pr.user_id, p.id as target_id, NULL as comment_id, 
				   CASE WHEN pr.is_like = 1 THEN 'like' ELSE 'dislike' END as reaction_type,
				   pr.created_at
			FROM post_reactions pr
			JOIN posts p ON pr.post_id = p.id
			WHERE pr.user_id = ?
			UNION ALL
			SELECT 'comment' as type, cr.user_id, NULL as post_id, c.id as comment_id,
				   cr.reaction_type,
				   cr.created_at
			FROM comment_reactions cr
			JOIN comments c ON cr.comment_id = c.id
			WHERE cr.user_id = ?
		) AS combined
		ORDER BY created_at DESC`

	log.Printf("Executing query: %s", query)
	rows, err := database.DB.Query(query, userID, userID)
	if err != nil {
		log.Printf("Error executing query: %v", err)
		return nil, err
	}
	defer rows.Close()

	var reactions []Reaction
	for rows.Next() {
		var r Reaction
		var typ string
		var targetID int
		var commentID sql.NullInt64
		var createdAtStr string

		err := rows.Scan(&typ, &r.UserID, &targetID, &commentID, &r.ReactionType, &createdAtStr)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			return nil, err
		}

		if typ == "post" {
			r.PostID = &targetID
		} else {
			if commentID.Valid {
				commentIDInt := int(commentID.Int64)
				r.CommentID = &commentIDInt
			}
		}

		// Parse timestamp
		r.CreatedAt, err = time.Parse("2006-01-02 15:04:05", createdAtStr)
		if err != nil {
			// Try alternative format
			r.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
			if err != nil {
				log.Printf("Failed to parse timestamp '%s': %v", createdAtStr, err)
				// If both formats fail, use current time as fallback
				r.CreatedAt = time.Now()
			}
		}

		reactions = append(reactions, r)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error after iterating rows: %v", err)
		return nil, err
	}

	log.Printf("Successfully retrieved %d reactions", len(reactions))
	return reactions, nil
}
