package handler

import (
	"strconv"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/planforever/nexusacg/internal/model"
	"github.com/planforever/nexusacg/internal/service"
)

type MessageHandler struct{ svc *service.MessageService }

func NewMessageHandler(r *gin.RouterGroup, svc *service.MessageService, authMW gin.HandlerFunc) {
	h := &MessageHandler{svc: svc}
	private := r.Group("/messages")
	private.Use(authMW)
	private.POST("", h.Send)
	private.GET("/conversations", h.Conversations)
	private.GET("/:user_id", h.GetMessages)
}

func (h *MessageHandler) Send(c *gin.Context) {
	userID, _ := c.Get("user_id")
	senderID, _ := uuid.Parse(userID.(string))
	var input struct {
		ReceiverID string `json:"receiver_id"`
		Content    string `json:"content"`
	}
	if err := c.ShouldBindJSON(&input); err != nil || input.Content == "" {
		BadRequest(c, "invalid input")
		return
	}
	receiverID, _ := uuid.Parse(input.ReceiverID)
	msg := &model.Message{SenderID: senderID, ReceiverID: receiverID, Content: input.Content}
	if err := h.svc.Send(msg); err != nil {
		BadRequest(c, err.Error())
		return
	}
	Success(c, msg)
}

func (h *MessageHandler) Conversations(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, _ := uuid.Parse(userID.(string))
	infos, err := h.svc.GetConversations(uid)
	if err != nil {
		InternalError(c, err.Error())
		return
	}
	if infos == nil { infos = []service.ConversationInfo{} }
	Success(c, gin.H{"items": infos})
}

func (h *MessageHandler) GetMessages(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, _ := uuid.Parse(userID.(string))
	otherID, _ := uuid.Parse(c.Param("user_id"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	msgs, err := h.svc.GetMessages(uid, otherID, page, pageSize)
	if err != nil {
		InternalError(c, err.Error())
		return
	}
	if msgs == nil { msgs = []model.Message{} }
	Success(c, gin.H{"items": msgs})
}
