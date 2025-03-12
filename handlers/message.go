package handlers

import (
	"net/http"
	"strconv"

	"github.com/Stella-Achar-Oiro/forum/models"
	"github.com/Stella-Achar-Oiro/forum/utils"
	"github.com/gorilla/mux"
)

// CreateMessageRequest represents a request to create a private message
type CreateMessageRequest struct {
	Content string `json:"content"`
}

// CreateMessageHandler handles sending a private message
func CreateMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get the current user
	sender, err := utils.GetUserFromRequest(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	// Get the receiver ID from the URL
	vars := mux.Vars(r)
	receiverIDStr, ok := vars["userId"]
	if !ok {
		utils.RespondWithError(w, http.StatusBadRequest, "Receiver ID is required")
		return
	}

	receiverID, err := strconv.Atoi(receiverIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid receiver ID")
		return
	}

	// Validate that receiver exists
	receiver, err := models.GetUserByID(receiverID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Receiver not found")
		return
	}

	// Parse the request
	var req CreateMessageRequest
	if err := utils.DecodeJSONBody(r, &req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate the request
	if req.Content == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Message content is required")
		return
	}

	// Create the message
	message := models.Message{
		SenderID:   sender.ID,
		ReceiverID: receiver.ID,
		Content:    req.Content,
	}

	messageID, err := models.CreateMessage(message)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to send message")
		return
	}

	// Get the conversation to return updated messages
	messages, err := models.GetMessagesByUsers(sender.ID, receiver.ID, 50, 0)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve messages")
		return
	}

	response := map[string]interface{}{
		"messageId": messageID,
		"messages":  messages,
	}

	utils.RespondWithSuccess(w, http.StatusCreated, "Message sent successfully", response)
}

// GetMessagesHandler handles retrieving messages between two users
func GetMessagesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get the current user
	currentUser, err := utils.GetUserFromRequest(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	// Get the other user's ID from the URL
	vars := mux.Vars(r)
	otherUserIDStr, ok := vars["userId"]
	if !ok {
		utils.RespondWithError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	otherUserID, err := strconv.Atoi(otherUserIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Get pagination parameters
	offsetStr := r.URL.Query().Get("offset")
	limitStr := r.URL.Query().Get("limit")

	offset := 0
	if offsetStr != "" {
		offset, err = strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid offset")
			return
		}
	}

	limit := 50 // Default limit
	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit <= 0 || limit > 100 {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid limit")
			return
		}
	}

	// Get messages
	messages, err := models.GetMessagesByUsers(currentUser.ID, otherUserID, limit, offset)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve messages")
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Messages retrieved successfully", messages)
}

// GetMessageUsersHandler handles retrieving all users with whom the current user has exchanged messages
func GetMessageUsersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get the current user
	currentUser, err := utils.GetUserFromRequest(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	// Get users with messages
	messageUsers, err := models.GetMessageUsers(currentUser.ID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve message users")
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Message users retrieved successfully", messageUsers)
}
