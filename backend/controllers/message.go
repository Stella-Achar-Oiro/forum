// backend/controllers/message.go
package controllers

import (
    "database/sql"
    "encoding/json"
    "net/http"
    "strconv"
    "time"
    "forum/backend/models"
)

type MessageController struct {
    DB *sql.DB
}

type SendMessageRequest struct {
    ReceiverID int    `json:"receiverId"`
    Content    string `json:"content"`
}

// SendMessage handles sending a new message
func (c *MessageController) SendMessage(w http.ResponseWriter, r *http.Request, senderID int) {
    // Only allow POST method
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    var req struct {
        ReceiverID int    `json:"receiverId"`
        Content    string `json:"content"`
        ImageURL   string `json:"imageUrl"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    // Validate request
    if req.ReceiverID <= 0 || (req.Content == "" && req.ImageURL == "") {
        http.Error(w, "Receiver ID and either content or image are required", http.StatusBadRequest)
        return
    }
    
    // Create message
    message := models.Message{
        SenderID:   senderID,
        ReceiverID: req.ReceiverID,
        Content:    req.Content,
        ImageURL:   req.ImageURL,
        CreatedAt:  time.Now(),
    }
    
    // Save to database
    messageID, err := models.CreateMessage(c.DB, message)
    if err != nil {
        http.Error(w, "Error sending message", http.StatusInternalServerError)
        return
    }
    
    // Get sender info for response
    sender, err := models.GetUserByID(c.DB, senderID)
    if err != nil {
        http.Error(w, "Error retrieving sender info", http.StatusInternalServerError)
        return
    }
    
    message.ID = int(messageID)
    message.Sender = sender
    
    // Return message data
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(message)
}

// GetMessages retrieves messages between two users with pagination
func (c *MessageController) GetMessages(w http.ResponseWriter, r *http.Request, userID int) {
    // Only allow GET method
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    // Get other user ID from URL
    otherIDStr := r.URL.Query().Get("userId")
    if otherIDStr == "" {
        http.Error(w, "Other user ID is required", http.StatusBadRequest)
        return
    }
    
    otherID, err := strconv.Atoi(otherIDStr)
    if err != nil {
        http.Error(w, "Invalid user ID", http.StatusBadRequest)
        return
    }
    
    // Get pagination parameters
    limitStr := r.URL.Query().Get("limit")
    offsetStr := r.URL.Query().Get("offset")
    
    limit := 10 // Default limit
    if limitStr != "" {
        if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
            limit = l
        }
    }
    
    offset := 0 // Default offset
    if offsetStr != "" {
        if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
            offset = o
        }
    }
    
    // Get messages
    messages, err := models.GetMessagesBetweenUsers(c.DB, userID, otherID, limit, offset)
    if err != nil {
        http.Error(w, "Error retrieving messages", http.StatusInternalServerError)
        return
    }
    
    // Return messages
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(messages)
}

// GetChats retrieves recent chats for the current user
func (c *MessageController) GetChats(w http.ResponseWriter, r *http.Request, userID int) {
    // Only allow GET method
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    // Get users with chat history
    recentUsers, err := models.GetRecentChats(c.DB, userID)
    if err != nil {
        http.Error(w, "Error retrieving chats", http.StatusInternalServerError)
        return
    }
    
    // Get users with no messages (for alphabetical list)
    otherUsers, err := models.GetUsersWithNoMessages(c.DB, userID)
    if err != nil {
        http.Error(w, "Error retrieving users", http.StatusInternalServerError)
        return
    }
    
    // Combine both lists
    allUsers := append(recentUsers, otherUsers...)
    
    // Return chat list
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(allUsers)
}