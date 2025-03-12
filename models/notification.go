package models

import (
	"database/sql"
	"time"
)

type NotificationType string

const (
	NotificationTypeLike    NotificationType = "like"
	NotificationTypeDislike NotificationType = "dislike"
	NotificationTypeComment NotificationType = "comment"
)

type Notification struct {
	ID        int64            `json:"id"`
	UserID    int64            `json:"userId"`
	Type      NotificationType `json:"type"`
	PostID    int64            `json:"postId"`
	CommentID *int64           `json:"commentId,omitempty"`
	ActorID   int64            `json:"actorId"` // The user who triggered the notification
	Message   string           `json:"message"`
	Read      bool             `json:"read"`
	CreatedAt time.Time        `json:"createdAt"`
}

// CreateNotification creates a new notification in the database
func CreateNotification(db *sql.DB, notification *Notification) error {
	query := `
		INSERT INTO notifications (user_id, type, post_id, comment_id, actor_id, message, read, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := db.Exec(query,
		notification.UserID,
		notification.Type,
		notification.PostID,
		notification.CommentID,
		notification.ActorID,
		notification.Message,
		notification.Read,
		time.Now())
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	notification.ID = id
	return nil
}

// GetUserNotifications retrieves all notifications for a user
func GetUserNotifications(db *sql.DB, userID int64) ([]Notification, error) {
	query := `
		SELECT id, user_id, type, post_id, comment_id, actor_id, message, read, created_at
		FROM notifications
		WHERE user_id = ?
		ORDER BY created_at DESC`

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []Notification
	for rows.Next() {
		var n Notification
		var commentID sql.NullInt64
		err := rows.Scan(
			&n.ID,
			&n.UserID,
			&n.Type,
			&n.PostID,
			&commentID,
			&n.ActorID,
			&n.Message,
			&n.Read,
			&n.CreatedAt)
		if err != nil {
			return nil, err
		}
		if commentID.Valid {
			n.CommentID = &commentID.Int64
		}
		notifications = append(notifications, n)
	}

	return notifications, nil
}

// MarkNotificationAsRead marks a notification as read
func MarkNotificationAsRead(db *sql.DB, notificationID int64) error {
	query := `UPDATE notifications SET read = true WHERE id = ?`
	_, err := db.Exec(query, notificationID)
	return err
}

// GetUnreadNotificationCount gets the count of unread notifications for a user
func GetUnreadNotificationCount(db *sql.DB, userID int64) (int, error) {
	query := `SELECT COUNT(*) FROM notifications WHERE user_id = ? AND read = false`
	var count int
	err := db.QueryRow(query, userID).Scan(&count)
	return count, err
}
