package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/planforever/nexusacg/internal/service"
)

type FollowHandler struct {
	svc *service.FollowService
}

func NewFollowHandler(r *gin.RouterGroup, svc *service.FollowService, authMW gin.HandlerFunc) {
	h := &FollowHandler{svc: svc}

	users := r.Group("/users")
	users.Use(authMW)
	users.POST("/:id/follow", h.Follow)
	users.DELETE("/:id/follow", h.Unfollow)
	users.GET("/:id/followers", h.Followers)
	users.GET("/:id/following", h.Following)
	users.GET("/:id/isfollowing", h.IsFollowing)
}

func (h *FollowHandler) Follow(c *gin.Context) {
	targetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		BadRequest(c, "invalid user id")
		return
	}
	userID, _ := c.Get("user_id")
	followerID, _ := uuid.Parse(userID.(string))
	if err := h.svc.Follow(followerID, targetID); err != nil {
		BadRequest(c, err.Error())
		return
	}
	Success(c, gin.H{"followed": true})
}

func (h *FollowHandler) Unfollow(c *gin.Context) {
	targetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		BadRequest(c, "invalid user id")
		return
	}
	userID, _ := c.Get("user_id")
	followerID, _ := uuid.Parse(userID.(string))
	if err := h.svc.Unfollow(followerID, targetID); err != nil {
		InternalError(c, err.Error())
		return
	}
	Success(c, gin.H{"followed": false})
}

func (h *FollowHandler) Followers(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		BadRequest(c, "invalid user id")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	result, err := h.svc.GetFollowers(userID, page, pageSize)
	if err != nil {
		InternalError(c, err.Error())
		return
	}
	Success(c, result)
}

func (h *FollowHandler) Following(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		BadRequest(c, "invalid user id")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	result, err := h.svc.GetFollowing(userID, page, pageSize)
	if err != nil {
		InternalError(c, err.Error())
		return
	}
	Success(c, result)
}

func (h *FollowHandler) IsFollowing(c *gin.Context) {
	userID, _ := c.Get("user_id")
	followerID, _ := uuid.Parse(userID.(string))
	targetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		BadRequest(c, "invalid user id")
		return
	}
	isFollowing, err := h.svc.IsFollowing(followerID, targetID)
	if err != nil {
		InternalError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": gin.H{"is_following": isFollowing}})
}
