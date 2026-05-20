package handler

import (
	"fmt"
	"log"
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/planforever/nexusacg/internal/storage"
)

var allowedImageMIME = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

var allowedVideoMIME = map[string]bool{
	"video/mp4":       true,
	"video/webm":      true,
	"video/quicktime": true,
}

type UploadHandler struct {
	store storage.Storage
}

func NewUploadHandler(r *gin.RouterGroup, store storage.Storage, authMW, adminMW gin.HandlerFunc) {
	h := &UploadHandler{store: store}

	upload := r.Group("/upload")
	upload.Use(authMW)
	upload.POST("", h.Upload)
	upload.POST("/video", h.UploadVideo)

	// Delete requires admin to prevent arbitrary file deletion
	admin := upload.Group("")
	admin.Use(adminMW)
	admin.DELETE("/:filename", h.Delete)
}

// UploadFile godoc
// @Summary Upload file (image)
// @Description Upload an image file (jpg, jpeg, png, gif, webp, max 10MB)
// @Tags upload
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Image file to upload"
// @Success 200 {object} map[string]string "Upload response with URL"
// @Failure 400 {object} map[string]string "Error response"
// @Router /upload [post]
func (h *UploadHandler) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		BadRequest(c, "missing file field")
		return
	}

	if err := validateFileMIME(file, allowedImageMIME); err != nil {
		BadRequest(c, err.Error())
		return
	}

	url, err := h.store.Upload(file)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	Success(c, gin.H{"url": url})
}

// UploadVideo godoc
// @Summary Upload video file
// @Description Upload a video file (mp4, webm, mov, max 50MB)
// @Tags upload
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Video file to upload"
// @Success 200 {object} map[string]string "Upload response with URL"
// @Failure 400 {object} map[string]string "Error response"
// @Router /upload/video [post]
func (h *UploadHandler) UploadVideo(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		log.Printf("upload/video: formfile error: %v", err)
		BadRequest(c, "missing file field")
		return
	}
	log.Printf("upload/video: received %s size=%d ct=%s", file.Filename, file.Size, file.Header.Get("Content-Type"))

	if err := validateFileMIME(file, allowedVideoMIME); err != nil {
		log.Printf("upload/video: MIME rejected: %v", err)
		BadRequest(c, err.Error())
		return
	}

	url, err := h.store.Upload(file)
	if err != nil {
		log.Printf("upload/video: store error: %v", err)
		BadRequest(c, err.Error())
		return
	}

	log.Printf("upload/video: ok url=%s", url)
	Success(c, gin.H{"url": url})
}

// DeleteFile godoc
// @Summary Delete uploaded file
// @Description Delete an uploaded file by filename (requires auth)
// @Tags upload
// @Produce json
// @Param filename path string true "Filename to delete"
// @Success 200 {object} map[string]bool "Delete response"
// @Failure 400 {object} map[string]string "Error response"
// @Failure 404 {object} map[string]string "File not found"
// @Security BearerAuth
// @Router /upload/{filename} [delete]
func (h *UploadHandler) Delete(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		BadRequest(c, "missing filename")
		return
	}

	url := "/uploads/" + filename
	if err := h.store.Delete(url); err != nil {
		NotFound(c, "file not found")
		return
	}

	Success(c, gin.H{"deleted": true})
}

func validateFileMIME(file *multipart.FileHeader, allowed map[string]bool) error {
	src, err := file.Open()
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer src.Close()

	buf := make([]byte, 512)
	n, err := src.Read(buf)
	if err != nil {
		return fmt.Errorf("read file header: %w", err)
	}

	detected := http.DetectContentType(buf[:n])
	if !allowed[detected] {
		return fmt.Errorf("invalid file content type: %s", detected)
	}
	return nil
}
