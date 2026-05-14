package handler

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/planforever/nexusacg/internal/service"
)

type AdminHandler struct {
	admin *service.AdminService
}

func NewAdminHandler(r *gin.RouterGroup, admin *service.AdminService, middlewares ...gin.HandlerFunc) {
	h := &AdminHandler{admin: admin}

	adminGroup := r.Group("/admin")
	for _, mw := range middlewares {
		adminGroup.Use(mw)
	}

	// Product audit
	adminGroup.GET("/products/pending", h.PendingProducts)
	adminGroup.POST("/products/:id/approve", h.ApproveProduct)
	adminGroup.POST("/products/:id/reject", h.RejectProduct)

	// Post audit
	adminGroup.GET("/posts/pending", h.PendingPosts)
	adminGroup.POST("/posts/:id/approve", h.ApprovePost)
	adminGroup.POST("/posts/:id/reject", h.RejectPost)

	// Order management
	adminGroup.GET("/orders", h.ListOrders)
	adminGroup.POST("/orders/:order_no/refund", h.RefundOrder)

	// Dashboard
	adminGroup.GET("/stats", h.Stats)

	// User management
	adminGroup.GET("/users", h.ListUsers)
	adminGroup.POST("/users/:id/ban", h.BanUser)
}

// PendingProducts godoc
// @Summary List pending products
// @Description List products awaiting admin approval
// @Tags admin
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} Response
// @Failure 500 {object} Response
// @Security BearerAuth
// @Router /admin/products/pending [get]
func (h *AdminHandler) PendingProducts(c *gin.Context) {
	page := 1
	if p := c.Query("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}
	pageSize := 20
	if ps := c.Query("page_size"); ps != "" {
		fmt.Sscanf(ps, "%d", &pageSize)
	}

	products, total, err := h.admin.ListPendingProducts(c.Request.Context(), page, pageSize)
	if err != nil {
		InternalError(c, err.Error())
		return
	}

	Success(c, gin.H{"items": products, "total": total, "page": page, "size": pageSize})
}

// ApproveProduct godoc
// @Summary Approve product
// @Description Approve a pending product (admin only)
// @Tags admin
// @Produce json
// @Param id path string true "Product UUID"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Security BearerAuth
// @Router /admin/products/{id}/approve [post]
func (h *AdminHandler) ApproveProduct(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		BadRequest(c, "invalid product id")
		return
	}

	if err := h.admin.ApproveProduct(c.Request.Context(), id); err != nil {
		NotFound(c, err.Error())
		return
	}

	Success(c, nil)
}

// RejectProduct godoc
// @Summary Reject product
// @Description Reject a pending product (admin only)
// @Tags admin
// @Produce json
// @Param id path string true "Product UUID"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Security BearerAuth
// @Router /admin/products/{id}/reject [post]
func (h *AdminHandler) RejectProduct(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		BadRequest(c, "invalid product id")
		return
	}

	if err := h.admin.RejectProduct(c.Request.Context(), id); err != nil {
		NotFound(c, err.Error())
		return
	}

	Success(c, nil)
}

// PendingPosts godoc
// @Summary List pending posts
// @Description List posts awaiting admin approval
// @Tags admin
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} Response
// @Failure 500 {object} Response
// @Security BearerAuth
// @Router /admin/posts/pending [get]
func (h *AdminHandler) PendingPosts(c *gin.Context) {
	page := 1
	if p := c.Query("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}
	pageSize := 20
	if ps := c.Query("page_size"); ps != "" {
		fmt.Sscanf(ps, "%d", &pageSize)
	}

	posts, total, err := h.admin.ListPendingPosts(c.Request.Context(), page, pageSize)
	if err != nil {
		InternalError(c, err.Error())
		return
	}

	Success(c, gin.H{"items": posts, "total": total, "page": page, "size": pageSize})
}

// ApprovePost godoc
// @Summary Approve post
// @Description Approve a pending post (admin only)
// @Tags admin
// @Produce json
// @Param id path string true "Post UUID"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Security BearerAuth
// @Router /admin/posts/{id}/approve [post]
func (h *AdminHandler) ApprovePost(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		BadRequest(c, "invalid post id")
		return
	}

	if err := h.admin.ApprovePost(c.Request.Context(), id); err != nil {
		NotFound(c, err.Error())
		return
	}

	Success(c, nil)
}

// RejectPost godoc
// @Summary Reject post
// @Description Reject a pending post (admin only)
// @Tags admin
// @Produce json
// @Param id path string true "Post UUID"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Security BearerAuth
// @Router /admin/posts/{id}/reject [post]
func (h *AdminHandler) RejectPost(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		BadRequest(c, "invalid post id")
		return
	}

	if err := h.admin.RejectPost(c.Request.Context(), id); err != nil {
		NotFound(c, err.Error())
		return
	}

	Success(c, nil)
}

// ListOrders godoc
// @Summary List all orders
// @Description List all orders with optional status filter (admin only)
// @Tags admin
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Param status query string false "Filter by order status"
// @Success 200 {object} Response
// @Failure 500 {object} Response
// @Security BearerAuth
// @Router /admin/orders [get]
func (h *AdminHandler) ListOrders(c *gin.Context) {
	page := 1
	if p := c.Query("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}
	pageSize := 20
	if ps := c.Query("page_size"); ps != "" {
		fmt.Sscanf(ps, "%d", &pageSize)
	}

	orders, total, err := h.admin.ListOrders(c.Request.Context(), c.Query("status"), page, pageSize)
	if err != nil {
		InternalError(c, err.Error())
		return
	}

	Success(c, gin.H{"items": orders, "total": total, "page": page, "size": pageSize})
}

// RefundOrder godoc
// @Summary Refund order
// @Description Process a refund for an order (admin only)
// @Tags admin
// @Produce json
// @Param order_no path string true "Order number"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Security BearerAuth
// @Router /admin/orders/{order_no}/refund [post]
func (h *AdminHandler) RefundOrder(c *gin.Context) {
	if err := h.admin.ProcessRefund(c.Request.Context(), c.Param("order_no")); err != nil {
		BadRequest(c, err.Error())
		return
	}

	Success(c, gin.H{"status": "refunded"})
}

// Stats godoc
// @Summary Dashboard stats
// @Description Get platform dashboard statistics (admin only)
// @Tags admin
// @Produce json
// @Success 200 {object} Response
// @Failure 500 {object} Response
// @Security BearerAuth
// @Router /admin/stats [get]
func (h *AdminHandler) Stats(c *gin.Context) {
	stats, err := h.admin.GetDashboardStats(c.Request.Context())
	if err != nil {
		InternalError(c, err.Error())
		return
	}

	Success(c, stats)
}

// ListUsers godoc
// @Summary List users
// @Description List all users (admin only)
// @Tags admin
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} Response
// @Failure 500 {object} Response
// @Security BearerAuth
// @Router /admin/users [get]
func (h *AdminHandler) ListUsers(c *gin.Context) {
	page := 1
	if p := c.Query("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}
	pageSize := 20
	if ps := c.Query("page_size"); ps != "" {
		fmt.Sscanf(ps, "%d", &pageSize)
	}

	users, total, err := h.admin.ListUsers(c.Request.Context(), page, pageSize)
	if err != nil {
		InternalError(c, err.Error())
		return
	}

	Success(c, gin.H{"items": users, "total": total, "page": page, "size": pageSize})
}

// BanUser godoc
// @Summary Ban user
// @Description Ban a user account (admin only)
// @Tags admin
// @Produce json
// @Param id path string true "User UUID"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Security BearerAuth
// @Router /admin/users/{id}/ban [post]
func (h *AdminHandler) BanUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		BadRequest(c, "invalid user id")
		return
	}

	if err := h.admin.BanUser(c.Request.Context(), id); err != nil {
		NotFound(c, err.Error())
		return
	}

	Success(c, nil)
}
