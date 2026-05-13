package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/planforever/nexusacg/internal/storage"
)

type UploadHandler struct {
	store storage.Storage
}

func NewUploadHandler(r *gin.RouterGroup, store storage.Storage, authMW gin.HandlerFunc) {
	h := &UploadHandler{store: store}

	public := r.Group("/upload")
	public.POST("", h.Upload)
	public.POST("/video", h.UploadVideo)

	private := r.Group("/upload")
	private.Use(authMW)
	private.DELETE("/:filename", h.Delete)
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing file field"})
		return
	}

	url, err := h.store.Upload(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": url})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing file field"})
		return
	}

	url, err := h.store.Upload(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": url})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing filename"})
		return
	}

	url := "/uploads/" + filename
	if err := h.store.Delete(url); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": true})
}
