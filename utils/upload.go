package utils

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	// MaxFileSize is the maximum allowed file size in bytes (50MB for videos)
	MaxFileSize = 50 * 1024 * 1024
	// AllowedImageTypes contains the allowed image MIME types
	AllowedImageTypes = "image/jpeg,image/png,image/gif"
	// AllowedVideoTypes contains the allowed video MIME types
	AllowedVideoTypes = "video/mp4,video/webm,video/quicktime"
	// UploadDir is the base directory for uploads
	UploadDir = "uploads"
	// ImageUploadDir is the directory for image uploads
	ImageUploadDir = "uploads/images"
	// VideoUploadDir is the directory for video uploads
	VideoUploadDir = "uploads/videos"
)

// SaveUploadedFile saves an uploaded file to the specified directory
func SaveUploadedFile(file *multipart.FileHeader) (string, error) {
	// Check file size
	if file.Size > MaxFileSize {
		return "", fmt.Errorf("file size exceeds maximum allowed size of %d bytes", MaxFileSize)
	}

	// Check file type
	contentType := file.Header.Get("Content-Type")
	var uploadDir string

	if strings.Contains(AllowedImageTypes, contentType) {
		uploadDir = ImageUploadDir
	} else if strings.Contains(AllowedVideoTypes, contentType) {
		uploadDir = VideoUploadDir
	} else {
		return "", fmt.Errorf("file type %s is not allowed", contentType)
	}

	// Create upload directory if it doesn't exist
	err := os.MkdirAll(uploadDir, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create upload directory: %v", err)
	}

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	filepath := filepath.Join(uploadDir, filename)

	// Open source file
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %v", err)
	}
	defer src.Close()

	// Create destination file
	dst, err := os.Create(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %v", err)
	}
	defer dst.Close()

	// Copy file contents
	_, err = io.Copy(dst, src)
	if err != nil {
		return "", fmt.Errorf("failed to save file: %v", err)
	}

	// Return the relative URL path
	return "/" + filepath, nil
}

// DeleteUploadedFile deletes an uploaded file
func DeleteUploadedFile(filepath string) error {
	// Remove the leading slash if present
	filepath = strings.TrimPrefix(filepath, "/")

	// Check if the file is in the uploads directory
	if !strings.HasPrefix(filepath, UploadDir) {
		return fmt.Errorf("invalid file path")
	}

	err := os.Remove(filepath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %v", err)
	}

	return nil
}
