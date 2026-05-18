package handler

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/planforever/nexusacg/internal/service"
)

type DisputeHandler struct {
	svc *service.DisputeService
}

func NewDisputeHandler(r *gin.RouterGroup, svc *service.DisputeService, authMW gin.HandlerFunc, requireAdmin gin.HandlerFunc) {
	h := &DisputeHandler{svc: svc}

	user := r.Group("/disputes")
	user.Use(authMW)
	user.POST("", h.Create)
	user.GET("/my", h.MyDisputes)
	user.GET("/order/:order_no", h.GetByOrder)
	user.POST("/:id/respond", h.Respond)
	user.POST("/:id/messages", h.SendMessage)
	user.GET("/:id/messages", h.GetMessages)

	admin := r.Group("/admin/disputes")
	admin.Use(authMW)
	admin.Use(requireAdmin)
	admin.GET("", h.ListAll)
	admin.POST("/:id/resolve", h.AdminResolve)
}

type createDisputeRequest struct {
	OrderNo        string   `json:"order_no" binding:"required"`
	Reason         string   `json:"reason" binding:"required"`
	Description    string   `json:"description" binding:"required"`
	EvidenceImages []string `json:"evidence_images"`
	RefundAmount   float64  `json:"refund_amount"`
}

func (h *DisputeHandler) Create(c *gin.Context) {
	var req createDisputeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request: "+err.Error())
		return
	}

	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))

	dispute, err := h.svc.Create(c.Request.Context(), service.CreateDisputeInput{
		UserID:         userID,
		OrderNo:        req.OrderNo,
		Reason:         req.Reason,
		Description:    req.Description,
		EvidenceImages: req.EvidenceImages,
		RefundAmount:   req.RefundAmount,
	})
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	Success(c, dispute)
}

func (h *DisputeHandler) MyDisputes(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID := userIDStr.(string)

	var input service.DisputeListInput
	input.UserID = userID
	input.Status = c.Query("status")
	if c.Query("page") != "" {
		fmt.Sscanf(c.Query("page"), "%d", &input.Page)
	}
	if c.Query("page_size") != "" {
		fmt.Sscanf(c.Query("page_size"), "%d", &input.PageSize)
	}

	result, err := h.svc.List(input)
	if err != nil {
		InternalError(c, "failed to list disputes")
		return
	}
	Success(c, result)
}

func (h *DisputeHandler) GetByOrder(c *gin.Context) {
	orderNo := c.Param("order_no")

	dispute, err := h.svc.GetByOrderNo(orderNo)
	if err != nil {
		NotFound(c, "no dispute found for this order")
		return
	}

	// Only involved parties can view
	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))
	if dispute.InitiatorID != userID && dispute.SellerID != userID {
		Forbidden(c, "not authorized to view this dispute")
		return
	}

	Success(c, dispute)
}

type respondDisputeRequest struct {
	Description    string   `json:"description" binding:"required"`
	EvidenceImages []string `json:"evidence_images"`
}

func (h *DisputeHandler) Respond(c *gin.Context) {
	disputeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		BadRequest(c, "invalid dispute ID")
		return
	}

	var req respondDisputeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request: "+err.Error())
		return
	}

	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))

	dispute, err := h.svc.Respond(disputeID, service.RespondDisputeInput{
		UserID:         userID,
		Description:    req.Description,
		EvidenceImages: req.EvidenceImages,
	})
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	Success(c, dispute)
}

type sendMessageRequest struct {
	Content string   `json:"content" binding:"required"`
	Images  []string `json:"images"`
}

func (h *DisputeHandler) SendMessage(c *gin.Context) {
	disputeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		BadRequest(c, "invalid dispute ID")
		return
	}

	var req sendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request: "+err.Error())
		return
	}

	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))

	msg, err := h.svc.SendMessage(disputeID, userID, req.Content, req.Images)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	Success(c, msg)
}

func (h *DisputeHandler) GetMessages(c *gin.Context) {
	disputeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		BadRequest(c, "invalid dispute ID")
		return
	}

	messages, err := h.svc.GetMessages(disputeID)
	if err != nil {
		InternalError(c, "failed to get messages")
		return
	}
	Success(c, messages)
}

type adminResolveDisputeRequest struct {
	Decision  string  `json:"decision" binding:"required"` // full_refund | partial_refund | reject
	AdminNote *string `json:"admin_note"`
}

func (h *DisputeHandler) ListAll(c *gin.Context) {
	var input service.DisputeListInput
	input.Status = c.Query("status")
	if c.Query("page") != "" {
		fmt.Sscanf(c.Query("page"), "%d", &input.Page)
	}
	if c.Query("page_size") != "" {
		fmt.Sscanf(c.Query("page_size"), "%d", &input.PageSize)
	}

	result, err := h.svc.List(input)
	if err != nil {
		InternalError(c, "failed to list disputes")
		return
	}
	Success(c, result)
}

func (h *DisputeHandler) AdminResolve(c *gin.Context) {
	disputeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		BadRequest(c, "invalid dispute ID")
		return
	}

	var req adminResolveDisputeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request: "+err.Error())
		return
	}

	adminIDStr, _ := c.Get("user_id")
	adminID, _ := uuid.Parse(adminIDStr.(string))

	dispute, err := h.svc.AdminResolve(c.Request.Context(), disputeID, service.AdminResolveDisputeInput{
		AdminID:   adminID,
		Decision:  req.Decision,
		AdminNote: req.AdminNote,
	})
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	Success(c, dispute)
}
