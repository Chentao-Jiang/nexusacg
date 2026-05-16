package handler

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/planforever/nexusacg/internal/middleware"
	"github.com/planforever/nexusacg/internal/service"

	"gorm.io/gorm"
)

type CertificationHandler struct {
	svc *service.CertificationService
	db  *gorm.DB
}

func NewCertificationHandler(r *gin.RouterGroup, svc *service.CertificationService, authMW gin.HandlerFunc, db *gorm.DB) {
	h := &CertificationHandler{svc: svc, db: db}

	user := r.Group("/certifications")
	user.Use(authMW)
	user.POST("/merchant", h.ApplyMerchant)
	user.POST("/service-provider", h.ApplyServiceProvider)
	user.GET("/my", h.GetMyApplication)

	admin := r.Group("/admin/certifications")
	admin.Use(authMW)
	admin.Use(middleware.RequireAdmin())
	admin.GET("", h.ListApplications)
	admin.POST("/:id/review", h.ReviewApplication)
}

type applyMerchantRequest struct {
	BusinessLicenseURL string `json:"business_license_url" binding:"required"`
	ProductCategory    string `json:"product_category" binding:"required"`
	StoreName          string `json:"store_name" binding:"required"`
}

func (h *CertificationHandler) ApplyMerchant(c *gin.Context) {
	var req applyMerchantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request: "+err.Error())
		return
	}

	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))

	app, err := h.svc.ApplyMerchant(c.Request.Context(), service.ApplyMerchantInput{
		UserID:             userID,
		BusinessLicenseURL: req.BusinessLicenseURL,
		ProductCategory:    req.ProductCategory,
		StoreName:          req.StoreName,
	})
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	Success(c, app)
}

type applyServiceProviderRequest struct {
	ProviderType    string   `json:"provider_type" binding:"required"`
	PortfolioImages []string `json:"portfolio_images" binding:"required"`
}

func (h *CertificationHandler) ApplyServiceProvider(c *gin.Context) {
	var req applyServiceProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request: "+err.Error())
		return
	}

	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))

	app, err := h.svc.ApplyServiceProvider(c.Request.Context(), service.ApplyServiceProviderInput{
		UserID:          userID,
		ProviderType:    req.ProviderType,
		PortfolioImages: req.PortfolioImages,
	})
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	Success(c, app)
}

func (h *CertificationHandler) GetMyApplication(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))

	app, err := h.svc.GetUserApplication(userID)
	if err != nil {
		NotFound(c, "no application found")
		return
	}
	Success(c, app)
}

func (h *CertificationHandler) ListApplications(c *gin.Context) {
	var input service.ApplicationListInput
	input.Status = c.Query("status")
	input.Type = c.Query("type")
	if c.Query("page") != "" {
		fmt.Sscanf(c.Query("page"), "%d", &input.Page)
	}
	if c.Query("page_size") != "" {
		fmt.Sscanf(c.Query("page_size"), "%d", &input.PageSize)
	}

	result, err := h.svc.ListApplications(input)
	if err != nil {
		InternalError(c, "failed to list applications")
		return
	}
	Success(c, result)
}

type reviewApplicationRequest struct {
	Approved        bool    `json:"approved"`
	RejectionReason *string `json:"rejection_reason,omitempty"`
}

func (h *CertificationHandler) ReviewApplication(c *gin.Context) {
	var req reviewApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request: "+err.Error())
		return
	}

	appID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		BadRequest(c, "invalid application ID")
		return
	}

	adminIDStr, _ := c.Get("user_id")
	adminID, _ := uuid.Parse(adminIDStr.(string))

	if err := h.svc.ReviewApplication(c.Request.Context(), appID, service.ReviewApplicationInput{
		AdminID:         adminID,
		Approved:        req.Approved,
		RejectionReason: req.RejectionReason,
	}); err != nil {
		BadRequest(c, err.Error())
		return
	}
	Success(c, gin.H{"reviewed": true})
}
