// backend/models/message.go
package models

import (
    "database/sql"
    "time"
)

type Message struct {
    ID         int       `json:"id"`
    SenderID   int       `json:"senderId"`
    ReceiverID int       `json:"receiverId"`
    Content    string    `json:"content"`
    ImageURL   string    `json:"imageUrl"`
    CreatedAt  time.Time `json:"createdAt"`
    Sender     User      `json:"sender"`
}

// Modify CreateMessage function
func CreateMessage(db *sql.DB, message Message) (int64, error) {
    query := `INSERT INTO messages (sender_id, receiver_id, content, image_url) VALUES (?, ?, ?, ?)`
    
    result, err := db.Exec(query, message.SenderID, message.ReceiverID, message.Content, message.ImageURL)
    if err != nil {
        return 0, err
    }
    
    return result.LastInsertId()
}

// Update GetMessagesBetweenUsers function to include image_url
func GetMessagesBetweenUsers(db *sql.DB, userID1, userID2, limit, offset int) ([]Message, error) {
    query := `
    SELECT m.id, m.sender_id, m.receiver_id, m.content, m.image_url, m.created_at,
           u.id, u.nickname, u.first_name, u.last_name
    FROM messages m
    JOIN users u ON m.sender_id = u.id
    WHERE (m.sender_id = ? AND m.receiver_id = ?) OR (m.sender_id = ? AND m.receiver_id = ?)
    ORDER BY m.created_at DESC
    LIMIT ? OFFSET ?`
    
    rows, err := db.Query(query, userID1, userID2, userID2, userID1, limit, offset)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var messages []Message
    for rows.Next() {
        var message Message
        var sender User
        
        err := rows.Scan(
            &message.ID, &message.SenderID, &message.ReceiverID, &message.Content, &message.ImageURL, &message.CreatedAt,
            &sender.ID, &sender.Nickname, &sender.FirstName, &sender.LastName,
        )
        if err != nil {
            return nil, err
        }
        
        message.Sender = sender
        messages = append(messages, message)
    }
    
    // Reverse the order to get messages in chronological order
    for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
        messages[i], messages[j] = messages[j], messages[i]
    }
    
    return messages, nil
}

// GetRecentChats retrieves a list of users with whom the current user has exchanged messages
func GetRecentChats(db *sql.DB, userID int) ([]User, error) {
    query := `
    SELECT DISTINCT 
        u.id, u.nickname, u.first_name, u.last_name, u.email,
        (SELECT MAX(created_at) FROM messages 
         WHERE (sender_id = ? AND receiver_id = u.id) OR (sender_id = u.id AND receiver_id = ?)) as last_message_time
    FROM users u
    JOIN messages m ON (m.sender_id = u.id AND m.receiver_id = ?) OR (m.receiver_id = u.id AND m.sender_id = ?)
    WHERE u.id != ?
    ORDER BY last_message_time DESC`
    
    rows, err := db.Query(query, userID, userID, userID, userID, userID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var users []User
    for rows.Next() {
        var user User
        var lastMessageTime time.Time
        
        err := rows.Scan(&user.ID, &user.Nickname, &user.FirstName, &user.LastName, &user.Email, &lastMessageTime)
        if err != nil {
            return nil, err
        }
        
        users = append(users, user)
    }
    
    return users, nil
}

// GetUsersWithNoMessages retrieves users with whom current user has no message history
func GetUsersWithNoMessages(db *sql.DB, userID int) ([]User, error) {
    query := `
    SELECT id, nickname, first_name, last_name, email
    FROM users
    WHERE id != ? AND id NOT IN (
        SELECT DISTINCT
            CASE
                WHEN sender_id = ? THEN receiver_id
                WHEN receiver_id = ? THEN sender_id
            END
        FROM messages
        WHERE sender_id = ? OR receiver_id = ?
    )
    ORDER BY nickname ASC`
    
    rows, err := db.Query(query, userID, userID, userID, userID, userID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var users []User
    for rows.Next() {
        var user User
        
        err := rows.Scan(&user.ID, &user.Nickname, &user.FirstName, &user.LastName, &user.Email)
        if err != nil {
            return nil, err
        }
        
        users = append(users, user)
    }
    
    return users, nil
}