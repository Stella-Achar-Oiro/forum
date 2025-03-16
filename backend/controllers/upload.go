// backend/controllers/upload.go
package controllers

import (
    "database/sql"
    "encoding/json"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "strconv"
    "strings"
    "time"
    
    "github.com/gofrs/uuid"
)

const (
    MaxUploadSize = 10 * 1024 * 1024 // 10MB
    ImageDir      = "./frontend/uploads/images/"
    AvatarDir     = "./frontend/uploads/avatars/"
)

type UploadController struct {
    DB *sql.DB
}

type UploadResponse struct {
    Filename string `json:"filename"`
    URL      string `json:"url"`
}

// Initialize upload directories
func (c *UploadController) Init() {
    os.MkdirAll(ImageDir, os.ModePerm)
    os.MkdirAll(AvatarDir, os.ModePerm)
}

// UploadImage handles image uploads for messages
func (c *UploadController) UploadImage(w http.ResponseWriter, r *http.Request, userID int) {
    // Only allow POST method
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    // Parse multipart form with max size limit
    r.Body = http.MaxBytesReader(w, r.Body, MaxUploadSize)
    if err := r.ParseMultipartForm(MaxUploadSize); err != nil {
        http.Error(w, "File too large", http.StatusBadRequest)
        return
    }
    
    // Get uploaded file
    file, header, err := r.FormFile("image")
    if err != nil {
        http.Error(w, "Invalid file", http.StatusBadRequest)
        return
    }
    defer file.Close()
    
    // Validate file type
    contentType := header.Header.Get("Content-Type")
    if !strings.HasPrefix(contentType, "image/") {
        http.Error(w, "File is not an image", http.StatusBadRequest)
        return
    }
    
    // Generate unique filename
    ext := filepath.Ext(header.Filename)
    uuid, err := uuid.NewV4()
    if err != nil {
        http.Error(w, "Error generating filename", http.StatusInternalServerError)
        return
    }
    filename := uuid.String() + ext
    
    // Create destination file
    dst, err := os.Create(filepath.Join(ImageDir, filename))
    if err != nil {
        http.Error(w, "Error saving file", http.StatusInternalServerError)
        return
    }
    defer dst.Close()
    
    // Copy file contents
    if _, err := io.Copy(dst, file); err != nil {
        http.Error(w, "Error saving file", http.StatusInternalServerError)
        return
    }
    
    // Return success response
    response := UploadResponse{
        Filename: filename,
        URL:      "/uploads/images/" + filename,
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

// UploadAvatar handles avatar uploads for user profiles
func (c *UploadController) UploadAvatar(w http.ResponseWriter, r *http.Request, userID int) {
    // Only allow POST method
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    // Parse multipart form with max size limit
    r.Body = http.MaxBytesReader(w, r.Body, MaxUploadSize)
    if err := r.ParseMultipartForm(MaxUploadSize); err != nil {
        http.Error(w, "File too large", http.StatusBadRequest)
        return
    }
    
    // Get uploaded file
    file, header, err := r.FormFile("avatar")
    if err != nil {
        http.Error(w, "Invalid file", http.StatusBadRequest)
        return
    }
    defer file.Close()
    
    // Validate file type
    contentType := header.Header.Get("Content-Type")
    if !strings.HasPrefix(contentType, "image/") {
        http.Error(w, "File is not an image", http.StatusBadRequest)
        return
    }
    
    // Generate unique filename
    ext := filepath.Ext(header.Filename)
    filename := "avatar_" + strconv.Itoa(userID) + "_" + strconv.FormatInt(time.Now().Unix(), 10) + ext
    
    // Create destination file
    dst, err := os.Create(filepath.Join(AvatarDir, filename))
    if err != nil {
        http.Error(w, "Error saving file", http.StatusInternalServerError)
        return
    }
    defer dst.Close()
    
    // Copy file contents
    if _, err := io.Copy(dst, file); err != nil {
        http.Error(w, "Error saving file", http.StatusInternalServerError)
        return
    }
    
    // Return success response
    response := UploadResponse{
        Filename: filename,
        URL:      "/uploads/avatars/" + filename,
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}