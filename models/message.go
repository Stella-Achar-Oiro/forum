package models

import (
	"time"

	"github.com/Stella-Achar-Oiro/forum/database"
)

// Message represents a private message between users
type Message struct {
	ID         int           `json:"id"`
	SenderID   int           `json:"senderId"`
	ReceiverID int           `json:"receiverId"`
	Content    string        `json:"content"`
	IsRead     bool          `json:"isRead"`
	CreatedAt  time.Time     `json:"createdAt"`
	Sender     UserForPublic `json:"sender"`
	Receiver   UserForPublic `json:"receiver,omitempty"`
}

// MessageUser represents a user with whom the current user has exchanged messages
type MessageUser struct {
	User        UserForPublic `json:"user"`
	LastMessage Message       `json:"lastMessage"`
	UnreadCount int           `json:"unreadCount"`
}

// CreateMessage creates a new message in the database
func CreateMessage(message Message) (int, error) {
	result, err := database.DB.Exec(
		"INSERT INTO messages (sender_id, receiver_id, content) VALUES (?, ?, ?)",
		message.SenderID, message.ReceiverID, message.Content,
	)
	if err != nil {
		return 0, err
	}

	// Get the message ID
	messageID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(messageID), nil
}

// GetMessagesByUsers retrieves messages between two users with pagination
func GetMessagesByUsers(userID1, userID2, limit, offset int) ([]Message, error) {
	rows, err := database.DB.Query(`
		SELECT m.id, m.sender_id, m.receiver_id, m.content, m.is_read, m.created_at,
		       s.id, s.nickname, s.first_name, s.last_name, s.is_online,
		       r.id, r.nickname, r.first_name, r.last_name, r.is_online
		FROM messages m
		JOIN users s ON m.sender_id = s.id
		JOIN users r ON m.receiver_id = r.id
		WHERE (m.sender_id = ? AND m.receiver_id = ?) OR (m.sender_id = ? AND m.receiver_id = ?)
		ORDER BY m.created_at DESC
		LIMIT ? OFFSET ?
	`, userID1, userID2, userID2, userID1, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var message Message
		if err := rows.Scan(
			&message.ID, &message.SenderID, &message.ReceiverID, &message.Content, &message.IsRead, &message.CreatedAt,
			&message.Sender.ID, &message.Sender.Nickname, &message.Sender.FirstName, &message.Sender.LastName, &message.Sender.IsOnline,
			&message.Receiver.ID, &message.Receiver.Nickname, &message.Receiver.FirstName, &message.Receiver.LastName, &message.Receiver.IsOnline,
		); err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Mark messages as read if the current user is the receiver
	_, err = database.DB.Exec("UPDATE messages SET is_read = 1 WHERE receiver_id = ? AND sender_id = ?", userID1, userID2)
	if err != nil {
		return nil, err
	}

	// Reverse the order to show oldest first
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// GetMessageUsers retrieves all users with whom the current user has exchanged messages
func GetMessageUsers(userID int) ([]MessageUser, error) {
	rows, err := database.DB.Query(`
		SELECT DISTINCT 
			other_user.id, other_user.nickname, other_user.first_name, other_user.last_name, other_user.is_online,
			(
				SELECT COUNT(*) FROM messages 
				WHERE receiver_id = ? AND sender_id = other_user.id AND is_read = 0
			) as unread_count,
			last_msg.id, last_msg.sender_id, last_msg.receiver_id, last_msg.content, last_msg.is_read, last_msg.created_at,
			sender.id, sender.nickname, sender.first_name, sender.last_name, sender.is_online
		FROM (
			SELECT DISTINCT 
				CASE WHEN sender_id = ? THEN receiver_id ELSE sender_id END as other_id,
				MAX(id) as last_message_id
			FROM messages
			WHERE sender_id = ? OR receiver_id = ?
			GROUP BY other_id
		) AS msg
		JOIN users other_user ON other_user.id = msg.other_id
		JOIN messages last_msg ON last_msg.id = msg.last_message_id
		JOIN users sender ON last_msg.sender_id = sender.id
		ORDER BY last_msg.created_at DESC
	`, userID, userID, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messageUsers []MessageUser
	for rows.Next() {
		var messageUser MessageUser
		if err := rows.Scan(
			&messageUser.User.ID, &messageUser.User.Nickname, &messageUser.User.FirstName, &messageUser.User.LastName, &messageUser.User.IsOnline,
			&messageUser.UnreadCount,
			&messageUser.LastMessage.ID, &messageUser.LastMessage.SenderID, &messageUser.LastMessage.ReceiverID,
			&messageUser.LastMessage.Content, &messageUser.LastMessage.IsRead, &messageUser.LastMessage.CreatedAt,
			&messageUser.LastMessage.Sender.ID, &messageUser.LastMessage.Sender.Nickname,
			&messageUser.LastMessage.Sender.FirstName, &messageUser.LastMessage.Sender.LastName, &messageUser.LastMessage.Sender.IsOnline,
		); err != nil {
			return nil, err
		}
		messageUsers = append(messageUsers, messageUser)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return messageUsers, nil
}
