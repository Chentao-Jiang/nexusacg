package storage

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// Allowed image types and max file size (10MB).
var allowedExtensions = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
}

const maxFileSize = 10 * 1024 * 1024

// Storage defines the interface for file upload backends.
type Storage interface {
	// Upload reads the multipart file and stores it, returning the public URL.
	Upload(file *multipart.FileHeader) (url string, err error)
	// Delete removes the file by its URL path.
	Delete(url string) error
}

// LocalStorage stores files on the local filesystem.
type LocalStorage struct {
	uploadDir string
	baseURL   string
}

func NewLocalStorage(uploadDir, baseURL string) *LocalStorage {
	os.MkdirAll(uploadDir, 0o755)
	return &LocalStorage{uploadDir: uploadDir, baseURL: strings.TrimRight(baseURL, "/")}
}

func (s *LocalStorage) Upload(file *multipart.FileHeader) (string, error) {
	if file.Size > maxFileSize {
		return "", fmt.Errorf("file too large: %d bytes (max %d)", file.Size, maxFileSize)
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedExtensions[ext] {
		return "", fmt.Errorf("unsupported file type: %s (allowed: jpg, jpeg, png, gif, webp)", ext)
	}

	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("open upload file: %w", err)
	}
	defer src.Close()

	filename := uuid.New().String() + ext
	dstPath := filepath.Join(s.uploadDir, filename)

	dst, err := os.Create(dstPath)
	if err != nil {
		return "", fmt.Errorf("create upload file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		os.Remove(dstPath)
		return "", fmt.Errorf("write upload file: %w", err)
	}

	return s.baseURL + "/uploads/" + filename, nil
}

func (s *LocalStorage) Delete(url string) error {
	filename := filepath.Base(url)
	if filename == url || filename == "." || filename == "/" {
		return fmt.Errorf("invalid file url: %s", url)
	}
	path := filepath.Join(s.uploadDir, filename)
	return os.Remove(path)
}
