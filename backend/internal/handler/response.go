package handler

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{Code: 0, Message: "ok", Data: data})
}

func Error(c *gin.Context, status int, msg string) {
	c.JSON(status, Response{Code: status, Message: msg})
}

func BadRequest(c *gin.Context, msg string) {
	Error(c, http.StatusBadRequest, msg)
}

func Unauthorized(c *gin.Context, msg string) {
	Error(c, http.StatusUnauthorized, msg)
}

func NotFound(c *gin.Context, msg string) {
	Error(c, http.StatusNotFound, msg)
}

// InternalError logs the full error but returns a generic message to the client
// to prevent leaking internal implementation details.
func InternalError(c *gin.Context, msg string) {
	log.Printf("internal error: %s", msg)
	safeMsg := "服务器内部错误"
	// Only show safe, user-facing messages
	if strings.HasPrefix(msg, "failed to create order") {
		safeMsg = "订单创建失败，请重试"
	} else if strings.HasPrefix(msg, "failed to create post") {
		safeMsg = "帖子发布失败"
	} else if strings.HasPrefix(msg, "failed to create product") {
		safeMsg = "商品创建失败"
	}
	Error(c, http.StatusInternalServerError, safeMsg)
}
