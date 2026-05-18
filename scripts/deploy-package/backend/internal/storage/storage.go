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
var allowedImageExtensions = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
}

const maxImageSize = int64(10 * 1024 * 1024) // 10MB

// Allowed video types and max file size (50MB).
var allowedVideoExtensions = map[string]bool{
	".mp4": true, ".webm": true, ".mov": true,
}

const maxVideoSize = int64(50 * 1024 * 1024) // 50MB

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
	os.MkdirAll(filepath.Join(uploadDir, "videos"), 0o755)
	return &LocalStorage{uploadDir: uploadDir, baseURL: strings.TrimRight(baseURL, "/")}
}

func (s *LocalStorage) Upload(file *multipart.FileHeader) (string, error) {
	ext := strings.ToLower(filepath.Ext(file.Filename))

	// Sanitize: reject filenames with path separators
	cleanName := filepath.Clean(file.Filename)
	if strings.Contains(cleanName, string(filepath.Separator)) || strings.Contains(cleanName, string(filepath.ListSeparator)) {
		return "", fmt.Errorf("invalid filename")
	}

	// Determine file type and apply appropriate limits
	isVideo := allowedVideoExtensions[ext]
	maxSize := maxImageSize
	subdir := ""

	if isVideo {
		maxSize = maxVideoSize
		subdir = "videos"
	} else if !allowedImageExtensions[ext] {
		return "", fmt.Errorf("unsupported file type: %s (images: jpg, jpeg, png, gif, webp; videos: mp4, webm, mov)", ext)
	}

	if file.Size > maxSize {
		limitMB := maxSize / (1024 * 1024)
		return "", fmt.Errorf("file too large: %d bytes (max %dMB)", file.Size, limitMB)
	}

	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("open upload file: %w", err)
	}
	defer src.Close()

	filename := uuid.New().String() + ext
	dstPath := filepath.Join(s.uploadDir, subdir, filename)

	dst, err := os.Create(dstPath)
	if err != nil {
		return "", fmt.Errorf("create upload file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		os.Remove(dstPath)
		return "", fmt.Errorf("write upload file: %w", err)
	}

	if subdir != "" {
		return s.baseURL + "/uploads/" + subdir + "/" + filename, nil
	}
	return s.baseURL + "/uploads/" + filename, nil
}

func (s *LocalStorage) Delete(url string) error {
	filename := filepath.Base(url)
	if filename == url || filename == "." || filename == "/" {
		return fmt.Errorf("invalid file url: %s", url)
	}
	// Check if it's a video file (in videos subdir)
	if strings.Contains(url, "/uploads/videos/") {
		path := filepath.Join(s.uploadDir, "videos", filename)
		return os.Remove(path)
	}
	path := filepath.Join(s.uploadDir, filename)
	return os.Remove(path)
}
