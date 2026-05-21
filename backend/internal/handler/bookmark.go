package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/planforever/nexusacg/internal/service"
	"strconv"
)

type BookmarkHandler struct{ svc *service.BookmarkService }

func NewBookmarkHandler(r *gin.RouterGroup, svc *service.BookmarkService, authMW gin.HandlerFunc) {
	h := &BookmarkHandler{svc: svc}
	private := r.Group("")
	private.Use(authMW)
	private.POST("/posts/:id/bookmark", h.Bookmark)
	private.DELETE("/posts/:id/bookmark", h.Unbookmark)
	private.GET("/my-bookmarks", h.MyBookmarks)
}

func (h *BookmarkHandler) Bookmark(c *gin.Context) {
	postID, _ := uuid.Parse(c.Param("id"))
	userID, _ := c.Get("user_id")
	uid, _ := uuid.Parse(userID.(string))
	if err := h.svc.Bookmark(uid, postID); err != nil {
		BadRequest(c, err.Error())
		return
	}
	Success(c, gin.H{"bookmarked": true})
}

func (h *BookmarkHandler) Unbookmark(c *gin.Context) {
	postID, _ := uuid.Parse(c.Param("id"))
	userID, _ := c.Get("user_id")
	uid, _ := uuid.Parse(userID.(string))
	_ = h.svc.Unbookmark(uid, postID)
	Success(c, gin.H{"bookmarked": false})
}

func (h *BookmarkHandler) MyBookmarks(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, _ := uuid.Parse(userID.(string))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	result, err := h.svc.GetMyBookmarks(uid, page, pageSize)
	if err != nil {
		InternalError(c, err.Error())
		return
	}
	Success(c, result)
}
