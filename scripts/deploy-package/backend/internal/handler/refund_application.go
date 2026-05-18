package handler

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/planforever/nexusacg/internal/service"
)

type RefundApplicationHandler struct {
	svc *service.RefundApplicationService
}

func NewRefundApplicationHandler(r *gin.RouterGroup, svc *service.RefundApplicationService, authMW gin.HandlerFunc) {
	h := &RefundApplicationHandler{svc: svc}

	user := r.Group("/refund-applications")
	user.Use(authMW)
	user.POST("", h.Create)
	user.GET("/my", h.MyApplications)
	user.POST("/:id/review", h.SellerReview)

	seller := r.Group("/seller/refund-applications")
	seller.Use(authMW)
	seller.GET("", h.SellerApplications)

	admin := r.Group("/admin/refund-applications")
	admin.Use(authMW)
	admin.GET("", h.AdminApplications)
	admin.POST("/:id/execute", h.AdminExecute)
}

type createRefundApplicationRequest struct {
	OrderNo      string   `json:"order_no" binding:"required"`
	RefundType   string   `json:"refund_type" binding:"required"`
	Reason       string   `json:"reason" binding:"required"`
	EvidenceURLs []string `json:"evidence_urls"`
	Amount       float64  `json:"amount"`
}

func (h *RefundApplicationHandler) Create(c *gin.Context) {
	var req createRefundApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request: "+err.Error())
		return
	}

	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))

	app, err := h.svc.Create(c.Request.Context(), service.CreateRefundApplicationInput{
		UserID:       userID,
		OrderNo:      req.OrderNo,
		RefundType:   req.RefundType,
		Reason:       req.Reason,
		EvidenceURLs: req.EvidenceURLs,
		Amount:       req.Amount,
	})
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	Success(c, app)
}

func (h *RefundApplicationHandler) MyApplications(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")

	var input service.RefundApplicationListInput
	input.UserID = userIDStr.(string)
	input.Status = c.Query("status")
	if c.Query("page") != "" {
		fmt.Sscanf(c.Query("page"), "%d", &input.Page)
	}
	if c.Query("page_size") != "" {
		fmt.Sscanf(c.Query("page_size"), "%d", &input.PageSize)
	}

	result, err := h.svc.List(input)
	if err != nil {
		InternalError(c, "failed to list refund applications")
		return
	}
	Success(c, result)
}

type sellerReviewRefundApplicationRequest struct {
	Approved bool    `json:"approved"`
	Note     *string `json:"note,omitempty"`
}

func (h *RefundApplicationHandler) SellerReview(c *gin.Context) {
	appID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		BadRequest(c, "invalid application ID")
		return
	}

	var req sellerReviewRefundApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request: "+err.Error())
		return
	}

	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))

	app, err := h.svc.SellerReview(appID, service.ReviewRefundApplicationInput{
		ReviewerID: userID,
		Approved:   req.Approved,
		Note:       req.Note,
	})
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	Success(c, app)
}

func (h *RefundApplicationHandler) SellerApplications(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")

	var input service.RefundApplicationListInput
	input.SellerID = userIDStr.(string)
	input.Status = c.Query("status")
	if c.Query("page") != "" {
		fmt.Sscanf(c.Query("page"), "%d", &input.Page)
	}
	if c.Query("page_size") != "" {
		fmt.Sscanf(c.Query("page_size"), "%d", &input.PageSize)
	}

	result, err := h.svc.List(input)
	if err != nil {
		InternalError(c, "failed to list refund applications")
		return
	}
	Success(c, result)
}

type adminExecuteRefundApplicationRequest struct {
	Note *string `json:"note,omitempty"`
}

func (h *RefundApplicationHandler) AdminApplications(c *gin.Context) {
	var input service.RefundApplicationListInput
	input.Status = c.Query("status")
	if c.Query("page") != "" {
		fmt.Sscanf(c.Query("page"), "%d", &input.Page)
	}
	if c.Query("page_size") != "" {
		fmt.Sscanf(c.Query("page_size"), "%d", &input.PageSize)
	}

	result, err := h.svc.List(input)
	if err != nil {
		InternalError(c, "failed to list refund applications")
		return
	}
	Success(c, result)
}

func (h *RefundApplicationHandler) AdminExecute(c *gin.Context) {
	appID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		BadRequest(c, "invalid application ID")
		return
	}

	var req adminExecuteRefundApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request: "+err.Error())
		return
	}

	adminIDStr, _ := c.Get("user_id")
	adminID, _ := uuid.Parse(adminIDStr.(string))

	app, err := h.svc.AdminExecute(c.Request.Context(), appID, adminID, req.Note)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	Success(c, app)
}
