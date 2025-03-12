package utils

import (
	"encoding/json"
	"log"
	"net/http"
)

// ApiResponse represents a standard API response format
type ApiResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// RespondWithError sends an error response
func RespondWithError(w http.ResponseWriter, code int, message string) {
	RespondWithJSON(w, code, ApiResponse{
		Success: false,
		Error:   message,
	})
}

// RespondWithJSON sends a JSON response
func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// RespondWithSuccess sends a success response with data
func RespondWithSuccess(w http.ResponseWriter, code int, message string, data interface{}) {
	RespondWithJSON(w, code, ApiResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// DecodeJSONBody decodes a JSON request body into a target structure
func DecodeJSONBody(r *http.Request, target interface{}) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}
