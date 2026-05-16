package handler

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/planforever/nexusacg/internal/service"
)

type ProductHandler struct {
	svc      *service.ProductService
	category *service.CategoryService
}

func NewProductHandler(r *gin.RouterGroup, svc *service.ProductService, category *service.CategoryService, authMW gin.HandlerFunc) {
	h := &ProductHandler{svc: svc, category: category}

	public := r.Group("/products")
	public.GET("", h.List)
	public.GET("/:id", h.Get)
	public.GET("/categories", h.Categories)

	private := r.Group("/products")
	private.Use(authMW)
	private.POST("", h.Create)
	private.POST("/categories", h.CreateCategory)
	private.PUT("/categories/:id", h.UpdateCategory)
	private.DELETE("/categories/:id", h.DeleteCategory)
}

// ListProducts godoc
// @Summary List products
// @Description List products with pagination, filtering by zone/source_type/category
// @Tags products
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Param zone query string false "Filter by zone (cosplay/peripheral)"
// @Param source_type query string false "Filter by source type (official/agent/self_made)"
// @Param category_id query string false "Filter by category UUID"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Router /products [get]
func (h *ProductHandler) List(c *gin.Context) {
	var input service.ProductListInput
	if err := c.ShouldBindQuery(&input); err != nil {
		BadRequest(c, "invalid query parameters: "+err.Error())
		return
	}

	result, err := h.svc.List(c.Request.Context(), input)
	if err != nil {
		InternalError(c, err.Error())
		return
	}

	Success(c, result)
}

// GetProduct godoc
// @Summary Get product detail
// @Description Get a single product by ID
// @Tags products
// @Produce json
// @Param id path string true "Product UUID"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Router /products/{id} [get]
func (h *ProductHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		BadRequest(c, "invalid product id")
		return
	}

	product, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		NotFound(c, err.Error())
		return
	}

	Success(c, product)
}

// ListCategories godoc
// @Summary List categories
// @Description List product categories
// @Tags products
// @Produce json
// @Success 200 {object} Response
// @Router /products/categories [get]
func (h *ProductHandler) Categories(c *gin.Context) {
	var input service.CategoryListInput
	if err := c.ShouldBindQuery(&input); err != nil {
		BadRequest(c, "invalid query parameters")
		return
	}

	categories, err := h.category.List(c.Request.Context(), input)
	if err != nil {
		InternalError(c, err.Error())
		return
	}

	Success(c, categories)
}

// CreateProduct godoc
// @Summary Create product
// @Description Create a new product listing (requires auth)
// @Tags products
// @Accept json
// @Produce json
// @Param request body CreateProductRequest true "Product info"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Security BearerAuth
// @Router /products [post]
func (h *ProductHandler) Create(c *gin.Context) {
	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request: "+err.Error())
		return
	}

	sellerID, _ := c.Get("user_id")
	sid, _ := uuid.Parse(sellerID.(string))

	product, err := h.svc.Create(c.Request.Context(), service.CreateProductInput{
		SellerID:      sid,
		CategoryID:    req.CategoryID,
		Name:          req.Name,
		Description:   req.Description,
		Price:         req.Price,
		OriginalPrice: req.OriginalPrice,
		Zone:          req.Zone,
		SourceType:    req.SourceType,
		SellerType:    req.SellerType,
		Images:        req.Images,
		Stock:         req.Stock,
		Tags:          req.Tags,
		CharacterName: req.CharacterName,
		AnimeName:     req.AnimeName,
	})
	if err != nil {
		InternalError(c, err.Error())
		return
	}

	Success(c, product)
}

// CreateCategory godoc
// @Summary Create category
// @Description Create a new product category (requires auth)
// @Tags products
// @Accept json
// @Produce json
// @Param request body service.CreateCategoryInput true "Category info"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Security BearerAuth
// @Router /products/categories [post]
func (h *ProductHandler) CreateCategory(c *gin.Context) {
	var req service.CreateCategoryInput
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request: "+err.Error())
		return
	}

	category, err := h.category.Create(c.Request.Context(), req)
	if err != nil {
		InternalError(c, err.Error())
		return
	}

	Success(c, category)
}

// UpdateCategory godoc
// @Summary Update category
// @Description Update a product category (requires auth)
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Category UUID"
// @Param request body service.UpdateCategoryInput true "Updated category info"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Security BearerAuth
// @Router /products/categories/{id} [put]
func (h *ProductHandler) UpdateCategory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		BadRequest(c, "invalid category id")
		return
	}

	var req service.UpdateCategoryInput
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request: "+err.Error())
		return
	}

	category, err := h.category.Update(c.Request.Context(), id, req)
	if err != nil {
		NotFound(c, err.Error())
		return
	}

	Success(c, category)
}

// DeleteCategory godoc
// @Summary Delete category
// @Description Delete a product category (requires auth)
// @Tags products
// @Produce json
// @Param id path string true "Category UUID"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Security BearerAuth
// @Router /products/categories/{id} [delete]
func (h *ProductHandler) DeleteCategory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		BadRequest(c, "invalid category id")
		return
	}

	if err := h.category.Delete(c.Request.Context(), id); err != nil {
		NotFound(c, err.Error())
		return
	}

	Success(c, nil)
}

type CreateProductRequest struct {
	Name          string     `json:"name" binding:"required,min=1,max=200"`
	Description   string     `json:"description" binding:"max=5000"`
	Price         float64    `json:"price" binding:"required,gt=0,lt=1000000"`
	OriginalPrice *float64   `json:"original_price"`
	Zone          string     `json:"zone" binding:"required,oneof=peripheral costume_makeup_props"`
	SellerType    string     `json:"seller_type" binding:"omitempty,oneof=certified_merchant certified_service uncertified"`
	SourceType    string     `json:"source_type" binding:"required,oneof=official agent self_made"`
	Images        []string   `json:"images" binding:"max=20"`
	Stock         int        `json:"stock" binding:"gte=0,lte=999999"`
	Tags          []string   `json:"tags" binding:"max=10"`
	CharacterName *string    `json:"character_name"`
	AnimeName     *string    `json:"anime_name"`
	CategoryID    *uuid.UUID `json:"category_id"`
}

type PostHandler struct {
	svc  *service.PostService
	mod  *service.ContentModerationService
}

func NewPostHandler(r *gin.RouterGroup, svc *service.PostService, authMW gin.HandlerFunc, mod *service.ContentModerationService) {
	h := &PostHandler{svc: svc, mod: mod}

	public := r.Group("/posts")
	public.GET("", h.List)
	public.GET("/:id", h.Get)
	public.GET("/:id/comments", h.Comments)

	private := r.Group("/posts")
	private.Use(authMW)
	private.POST("", h.Create)
	private.POST("/:id/like", h.Like)
	private.DELETE("/:id/like", h.Unlike)
	private.POST("/:id/comments", h.CreateComment)
}

// ListPosts godoc
// @Summary List posts
// @Description List posts with pagination
// @Tags posts
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Router /posts [get]
func (h *PostHandler) List(c *gin.Context) {
	var input service.PostListInput
	if err := c.ShouldBindQuery(&input); err != nil {
		BadRequest(c, "invalid query parameters")
		return
	}

	result, err := h.svc.List(c.Request.Context(), input)
	if err != nil {
		InternalError(c, err.Error())
		return
	}

	Success(c, result)
}

// GetPost godoc
// @Summary Get post detail
// @Description Get a single post by ID
// @Tags posts
// @Produce json
// @Param id path string true "Post UUID"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Router /posts/{id} [get]
func (h *PostHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		BadRequest(c, "invalid post id")
		return
	}

	post, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		NotFound(c, err.Error())
		return
	}

	Success(c, post)
}

// CreatePost godoc
// @Summary Create post
// @Description Create a new post with content, images, and tags (requires auth)
// @Tags posts
// @Accept json
// @Produce json
// @Param request body CreatePostRequest true "Post info"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Security BearerAuth
// @Router /posts [post]
func (h *PostHandler) Create(c *gin.Context) {
	var req CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request: "+err.Error())
		return
	}

	// Auto-moderate content before creation
	if h.mod != nil {
		videoURL := ""
		if req.VideoURL != nil {
			videoURL = *req.VideoURL
		}
		result, err := h.mod.AutoModeratePost(c.Request.Context(), req.Title, req.Content, req.Images, videoURL)
		if err != nil {
			log.Printf("content moderation error: %v", err)
		} else if !result.Pass {
			BadRequest(c, fmt.Sprintf("内容审核未通过: %s", result.Reason))
			return
		}
	}

	userID, _ := c.Get("user_id")
	uid, _ := uuid.Parse(userID.(string))

	post, err := h.svc.Create(c.Request.Context(), service.CreatePostInput{
		UserID:   uid,
		Title:    req.Title,
		Content:  req.Content,
		Images:   req.Images,
		VideoURL: req.VideoURL,
		Tags:     req.Tags,
	})
	if err != nil {
		InternalError(c, err.Error())
		return
	}

	Success(c, post)
}

// LikePost godoc
// @Summary Like a post
// @Description Like a post (requires auth)
// @Tags posts
// @Produce json
// @Param id path string true "Post UUID"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Security BearerAuth
// @Router /posts/{id}/like [post]
func (h *PostHandler) Like(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		BadRequest(c, "invalid post id")
		return
	}

	userID, _ := c.Get("user_id")
	uid, _ := uuid.Parse(userID.(string))

	if err := h.svc.Like(c.Request.Context(), uid, id); err != nil {
		BadRequest(c, err.Error())
		return
	}

	Success(c, nil)
}

// UnlikePost godoc
// @Summary Unlike a post
// @Description Remove like from a post (requires auth)
// @Tags posts
// @Produce json
// @Param id path string true "Post UUID"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Security BearerAuth
// @Router /posts/{id}/like [delete]
func (h *PostHandler) Unlike(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		BadRequest(c, "invalid post id")
		return
	}

	userID, _ := c.Get("user_id")
	uid, _ := uuid.Parse(userID.(string))

	if err := h.svc.Unlike(c.Request.Context(), uid, id); err != nil {
		BadRequest(c, err.Error())
		return
	}

	Success(c, nil)
}

// ListComments godoc
// @Summary List post comments
// @Description Get comments for a post
// @Tags posts
// @Produce json
// @Param id path string true "Post UUID"
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Router /posts/{id}/comments [get]
func (h *PostHandler) Comments(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		BadRequest(c, "invalid post id")
		return
	}

	page := 1
	if p := c.Query("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}
	pageSize := 20
	if ps := c.Query("page_size"); ps != "" {
		fmt.Sscanf(ps, "%d", &pageSize)
	}

	result, err := h.svc.ListComments(c.Request.Context(), id, page, pageSize)
	if err != nil {
		InternalError(c, err.Error())
		return
	}

	Success(c, result)
}

type CreateCommentRequest struct {
	Content  string     `json:"content" binding:"required,max=2000"`
	ParentID *uuid.UUID `json:"parent_id"`
}

// CreateComment godoc
// @Summary Create comment
// @Description Add a comment to a post (requires auth)
// @Tags posts
// @Accept json
// @Produce json
// @Param id path string true "Post UUID"
// @Param request body CreateCommentRequest true "Comment content and optional parent_id"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Security BearerAuth
// @Router /posts/{id}/comments [post]
func (h *PostHandler) CreateComment(c *gin.Context) {
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		BadRequest(c, "invalid post id")
		return
	}

	var req CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request: "+err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	uid, _ := uuid.Parse(userID.(string))

	comment, err := h.svc.CreateComment(c.Request.Context(), service.CommentInput{
		PostID:   postID,
		UserID:   uid,
		Content:  req.Content,
		ParentID: req.ParentID,
	})
	if err != nil {
		InternalError(c, err.Error())
		return
	}

	Success(c, comment)
}

type CreatePostRequest struct {
	Title    string   `json:"title" binding:"max=200"`
	Content  string   `json:"content" binding:"required,max=10000"`
	Images   []string `json:"images" binding:"max=20"`
	VideoURL *string  `json:"video_url"`
	Tags     []string `json:"tags" binding:"max=10"`
}

type EventHandler struct {
	svc *service.EventService
}

func NewEventHandler(r *gin.RouterGroup, svc *service.EventService, authMW gin.HandlerFunc) {
	h := &EventHandler{svc: svc}

	public := r.Group("/events")
	public.GET("", h.List)
	public.GET("/:id", h.Get)

	private := r.Group("/events")
	private.Use(authMW)
	private.POST("", h.Create)
}

// ListEvents godoc
// @Summary List events
// @Description List events with pagination
// @Tags events
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Router /events [get]
func (h *EventHandler) List(c *gin.Context) {
	var input service.EventListInput
	if err := c.ShouldBindQuery(&input); err != nil {
		BadRequest(c, "invalid query parameters")
		return
	}

	result, err := h.svc.List(c.Request.Context(), input)
	if err != nil {
		InternalError(c, err.Error())
		return
	}

	Success(c, result)
}

// GetEvent godoc
// @Summary Get event detail
// @Description Get a single event by ID
// @Tags events
// @Produce json
// @Param id path string true "Event UUID"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Router /events/{id} [get]
func (h *EventHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		BadRequest(c, "invalid event id")
		return
	}

	event, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		NotFound(c, err.Error())
		return
	}

	Success(c, event)
}

// CreateEvent godoc
// @Summary Create event
// @Description Create a new event (requires auth)
// @Tags events
// @Accept json
// @Produce json
// @Param request body CreateEventRequest true "Event info"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Security BearerAuth
// @Router /events [post]
func (h *EventHandler) Create(c *gin.Context) {
	var req CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request: "+err.Error())
		return
	}

	event, err := h.svc.Create(c.Request.Context(), service.CreateEventInput{
		Name:        req.Name,
		Description: req.Description,
		CoverURL:    req.CoverURL,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Address:     req.Address,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
	})
	if err != nil {
		InternalError(c, err.Error())
		return
	}

	Success(c, event)
}

type CreateEventRequest struct {
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description"`
	CoverURL    *string  `json:"cover_url"`
	StartTime   string   `json:"start_time" binding:"required"`
	EndTime     string   `json:"end_time" binding:"required"`
	Address     string   `json:"address" binding:"required"`
	Latitude    *float64 `json:"latitude"`
	Longitude   *float64 `json:"longitude"`
}

type OrderHandler struct {
	svc        *service.OrderService
	profitSvc  *service.ProfitShareService
}

func NewOrderHandler(r *gin.RouterGroup, svc *service.OrderService, profitSvc *service.ProfitShareService, authMW gin.HandlerFunc) {
	h := &OrderHandler{svc: svc, profitSvc: profitSvc}

	private := r.Group("/orders")
	private.Use(authMW)
	private.POST("", h.Create)
	private.GET("", h.List)
	private.GET("/:order_no", h.Detail)
	private.POST("/:order_no/pay", h.Pay)
	private.POST("/:order_no/cancel", h.Cancel)
	private.POST("/:order_no/refund", h.Refund)
	private.POST("/:order_no/ship", h.Ship)
	private.POST("/:order_no/confirm", h.Confirm)
}

// CreateOrder godoc
// @Summary Create order
// @Description Create a new order with product items (requires auth)
// @Tags orders
// @Accept json
// @Produce json
// @Param request body CreateOrderRequest true "Order items"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Security BearerAuth
// @Router /orders [post]
func (h *OrderHandler) Create(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request: "+err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	uid, _ := uuid.Parse(userID.(string))

	items := make([]struct {
		ProductID uuid.UUID `json:"product_id"`
		Quantity  int       `json:"quantity"`
	}, len(req.Items))
	for i, item := range req.Items {
		items[i].ProductID = item.ProductID
		items[i].Quantity = item.Quantity
	}

	order, err := h.svc.Create(c.Request.Context(), service.CreateOrderInput{
		UserID:         uid,
		IdempotencyKey: req.IdempotencyKey,
		Items:          items,
	})
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	Success(c, order)
}

type CreateOrderRequest struct {
	IdempotencyKey *string `json:"idempotency_key"`
	Items          []struct {
		ProductID uuid.UUID `json:"product_id" binding:"required"`
		Quantity  int       `json:"quantity" binding:"required,gt=0"`
	} `json:"items" binding:"required,min=1"`
}

// ListOrders godoc
// @Summary List my orders
// @Description List orders for the authenticated user (requires auth)
// @Tags orders
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Param status query string false "Filter by order status"
// @Success 200 {object} Response
// @Failure 401 {object} Response
// @Security BearerAuth
// @Router /orders [get]
func (h *OrderHandler) List(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, _ := uuid.Parse(userID.(string))

	page := 1
	if p := c.Query("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}
	pageSize := 20
	if ps := c.Query("page_size"); ps != "" {
		fmt.Sscanf(ps, "%d", &pageSize)
	}
	status := c.Query("status")

	orders, total, err := h.svc.GetByUser(c.Request.Context(), uid, page, pageSize, status)
	if err != nil {
		InternalError(c, err.Error())
		return
	}

	Success(c, gin.H{
		"items": orders,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

// GetOrder godoc
// @Summary Get order detail
// @Description Get order details by order number (requires auth)
// @Tags orders
// @Produce json
// @Param order_no path string true "Order number"
// @Success 200 {object} Response
// @Failure 404 {object} Response
// @Security BearerAuth
// @Router /orders/{order_no} [get]
func (h *OrderHandler) Detail(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, _ := uuid.Parse(userID.(string))

	order, err := h.svc.GetByOrderNo(c.Request.Context(), c.Param("order_no"))
	if err != nil {
		NotFound(c, err.Error())
		return
	}

	if order.UserID != uid {
		Forbidden(c, "access denied")
		return
	}

	Success(c, order)
}

// CancelOrder godoc
// @Summary Cancel an order
// @Description Cancel a pending order (requires auth)
// @Tags orders
// @Produce json
// @Param order_no path string true "Order number"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Security BearerAuth
// @Router /orders/{order_no}/cancel [post]
func (h *OrderHandler) Cancel(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, _ := uuid.Parse(userID.(string))

	if err := h.svc.Cancel(c.Request.Context(), uid, c.Param("order_no")); err != nil {
		BadRequest(c, err.Error())
		return
	}

	Success(c, gin.H{"status": "cancelled"})
}

// RefundOrder godoc
// @Summary Refund an order
// @Description Process a refund for an order (requires auth)
// @Tags orders
// @Produce json
// @Param order_no path string true "Order number"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Security BearerAuth
// @Router /orders/{order_no}/refund [post]
func (h *OrderHandler) Refund(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, _ := uuid.Parse(userID.(string))

	order, err := h.svc.GetByOrderNo(c.Request.Context(), c.Param("order_no"))
	if err != nil {
		NotFound(c, err.Error())
		return
	}

	if order.UserID != uid {
		Forbidden(c, "access denied")
		return
	}

	if err := h.svc.Refund(c.Request.Context(), c.Param("order_no")); err != nil {
		BadRequest(c, err.Error())
		return
	}

	Success(c, gin.H{"status": "refunded"})
}

type PayRequest struct {
	PaymentMethod string `json:"payment_method" binding:"required,oneof=wechat alipay"`
	PaymentID     string `json:"payment_id" binding:"required"`
}

// PayOrder godoc
// @Summary Pay for an order
// @Description Record payment for an order (requires auth)
// @Tags orders
// @Accept json
// @Produce json
// @Param order_no path string true "Order number"
// @Param request body PayRequest true "Payment method and ID"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Security BearerAuth
// @Router /orders/{order_no}/pay [post]
func (h *OrderHandler) Pay(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, _ := uuid.Parse(userID.(string))

	orderNo := c.Param("order_no")
	if orderNo == "" {
		BadRequest(c, "missing order_no")
		return
	}

	var req PayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request")
		return
	}

	// Look up order by order_no to get the ID
	order, err := h.svc.GetByOrderNo(c.Request.Context(), orderNo)
	if err != nil {
		NotFound(c, err.Error())
		return
	}

	// Verify the authenticated user owns the order
	if order.UserID != uid {
		Forbidden(c, "access denied")
		return
	}

	if err := h.svc.Pay(c.Request.Context(), order.ID, req.PaymentMethod, req.PaymentID); err != nil {
		BadRequest(c, err.Error())
		return
	}

	Success(c, gin.H{"status": "paid"})
}

// ShipOrder godoc
// @Summary Ship an order
// @Description Mark a paid order as shipped (requires auth, seller only)
// @Tags orders
// @Accept json
// @Produce json
// @Param order_no path string true "Order number"
// @Param request body ShipOrderRequest false "Shipping info"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Security BearerAuth
// @Router /orders/{order_no}/ship [post]
func (h *OrderHandler) Ship(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(string)

	if err := h.profitSvc.ShipOrder(c.Param("order_no"), uid); err != nil {
		BadRequest(c, err.Error())
		return
	}

	Success(c, gin.H{"status": "shipped"})
}

// ConfirmReceipt godoc
// @Summary Confirm receipt of an order
// @Description Mark a shipped order as completed and trigger profit sharing (requires auth, buyer only)
// @Tags orders
// @Produce json
// @Param order_no path string true "Order number"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Security BearerAuth
// @Router /orders/{order_no}/confirm [post]
func (h *OrderHandler) Confirm(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(string)

	if err := h.profitSvc.ConfirmReceipt(c.Param("order_no"), uid); err != nil {
		BadRequest(c, err.Error())
		return
	}

	Success(c, gin.H{"status": "completed"})
}

type ShipOrderRequest struct {
	TrackingNumber string `json:"tracking_number"`
	Carrier        string `json:"carrier"`
}

type EventServiceListingHandler struct {
	svc *service.EventServiceListingService
}

func NewEventServiceListingHandler(r *gin.RouterGroup, svc *service.EventServiceListingService, authMW gin.HandlerFunc) {
	h := &EventServiceListingHandler{svc: svc}

	// Public: view event services
	public := r.Group("/events")
	public.GET("/:id/service-listings", h.ListByEvent)

	// Private: manage event services and schedules
	private := r.Group("")
	private.Use(authMW)
	private.POST("/events/:id/service-listings", h.CreateListing)
	private.POST("/service-schedules", h.CreateSchedule)
	private.GET("/service-schedules/:provider_id", h.ListSchedules)
}

type createListingRequest struct {
	ServiceProviderID uuid.UUID `json:"service_provider_id" binding:"required"`
	Price             float64   `json:"price" binding:"required,gt=0"`
	Description       string    `json:"description" binding:"required"`
}

func (h *EventServiceListingHandler) CreateListing(c *gin.Context) {
	var req createListingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request: "+err.Error())
		return
	}

	eventID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		BadRequest(c, "invalid event ID")
		return
	}

	listing, err := h.svc.CreateListing(c.Request.Context(), service.CreateListingInput{
		EventID:           eventID,
		ServiceProviderID: req.ServiceProviderID,
		Price:             req.Price,
		Description:       req.Description,
	})
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	Success(c, listing)
}

func (h *EventServiceListingHandler) ListByEvent(c *gin.Context) {
	eventID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		BadRequest(c, "invalid event ID")
		return
	}

	listings, err := h.svc.ListByEvent(c.Request.Context(), eventID)
	if err != nil {
		InternalError(c, "failed to list event services")
		return
	}
	Success(c, listings)
}

type createScheduleRequest struct {
	ServiceProviderID uuid.UUID  `json:"service_provider_id" binding:"required"`
	EventID           *uuid.UUID `json:"event_id"`
	Date              string     `json:"date" binding:"required"`
	StartTime         *string    `json:"start_time"`
	EndTime           *string    `json:"end_time"`
	Status            string     `json:"status"`
	Notes             string     `json:"notes"`
}

func (h *EventServiceListingHandler) CreateSchedule(c *gin.Context) {
	var req createScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request: "+err.Error())
		return
	}

	schedule, err := h.svc.CreateSchedule(c.Request.Context(), service.CreateScheduleInput{
		ServiceProviderID: req.ServiceProviderID,
		EventID:           req.EventID,
		Date:              req.Date,
		StartTime:         req.StartTime,
		EndTime:           req.EndTime,
		Status:            req.Status,
		Notes:             req.Notes,
	})
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	Success(c, schedule)
}

func (h *EventServiceListingHandler) ListSchedules(c *gin.Context) {
	providerID, err := uuid.Parse(c.Param("provider_id"))
	if err != nil {
		BadRequest(c, "invalid provider ID")
		return
	}

	var eventID *uuid.UUID
	if eid := c.Query("event_id"); eid != "" {
		if parsed, err := uuid.Parse(eid); err == nil {
			eventID = &parsed
		}
	}

	schedules, err := h.svc.ListSchedules(c.Request.Context(), providerID, eventID)
	if err != nil {
		InternalError(c, "failed to list schedules")
		return
	}
	Success(c, schedules)
}

type ServiceProductHandler struct {
	svc *service.ServiceProductService
}

func NewServiceProductHandler(r *gin.RouterGroup, svc *service.ServiceProductService, authMW gin.HandlerFunc) {
	h := &ServiceProductHandler{svc: svc}

	public := r.Group("/service-products")
	public.GET("", h.List)
	public.GET("/:id", h.Get)
	public.GET("/:id/schedules", h.GetSchedules)

	private := r.Group("/service-products")
	private.Use(authMW)
	private.POST("", h.Create)
	private.PUT("/:id", h.Update)
	private.DELETE("/:id", h.Delete)
	private.GET("/my", h.MyListings)
}

type createServiceProductRequest struct {
	ServiceProviderID uuid.UUID  `json:"service_provider_id" binding:"required"`
	CategoryID        *uuid.UUID `json:"category_id"`
	Name              string     `json:"name" binding:"required,min=1,max=200"`
	Description       string     `json:"description" binding:"required,max=5000"`
	Price             float64    `json:"price" binding:"required,gt=0"`
	OriginalPrice     *float64   `json:"original_price"`
	ServiceType       string     `json:"service_type" binding:"required,oneof=makeup_artist wig_stylist photographer post_editor props_maker"`
	Images            []string   `json:"images" binding:"max=20"`
	PortfolioImages   []string   `json:"portfolio_images" binding:"required,min=1,max=30"`
	Tags              []string   `json:"tags" binding:"max=10"`
}

func (h *ServiceProductHandler) List(c *gin.Context) {
	var input service.ServiceProductListInput
	if err := c.ShouldBindQuery(&input); err != nil {
		BadRequest(c, "invalid query parameters: "+err.Error())
		return
	}
	results, total, err := h.svc.List(c.Request.Context(), input)
	if err != nil {
		InternalError(c, err.Error())
		return
	}
	Success(c, gin.H{
		"items": results,
		"total": total,
		"page":  input.Page,
		"size":  input.PageSize,
	})
}

func (h *ServiceProductHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		BadRequest(c, "invalid service product id")
		return
	}
	result, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		NotFound(c, err.Error())
		return
	}
	Success(c, result)
}

func (h *ServiceProductHandler) Create(c *gin.Context) {
	var req createServiceProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request: "+err.Error())
		return
	}
	userIDStr, _ := c.Get("user_id")
	uid, _ := uuid.Parse(userIDStr.(string))

	product, err := h.svc.Create(c.Request.Context(), service.CreateServiceProductInput{
		ServiceProviderID: req.ServiceProviderID,
		UserID:            uid,
		CategoryID:        req.CategoryID,
		Name:              req.Name,
		Description:       req.Description,
		Price:             req.Price,
		OriginalPrice:     req.OriginalPrice,
		ServiceType:       req.ServiceType,
		Images:            req.Images,
		PortfolioImages:   req.PortfolioImages,
		Tags:              req.Tags,
	})
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	Success(c, product)
}

func (h *ServiceProductHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		BadRequest(c, "invalid service product id")
		return
	}
	userIDStr, _ := c.Get("user_id")
	uid, _ := uuid.Parse(userIDStr.(string))

	var req createServiceProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request: "+err.Error())
		return
	}

	product, err := h.svc.Update(c.Request.Context(), id, uid, service.CreateServiceProductInput{
		Name:            req.Name,
		Description:     req.Description,
		Price:           req.Price,
		OriginalPrice:   req.OriginalPrice,
		Images:          req.Images,
		PortfolioImages: req.PortfolioImages,
		Tags:            req.Tags,
	})
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	Success(c, product)
}

func (h *ServiceProductHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		BadRequest(c, "invalid service product id")
		return
	}
	userIDStr, _ := c.Get("user_id")
	uid, _ := uuid.Parse(userIDStr.(string))

	if err := h.svc.Delete(c.Request.Context(), id, uid); err != nil {
		BadRequest(c, err.Error())
		return
	}
	Success(c, nil)
}

func (h *ServiceProductHandler) MyListings(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	uid, _ := uuid.Parse(userIDStr.(string))
	page := 1
	if p := c.Query("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}
	pageSize := 20
	if ps := c.Query("page_size"); ps != "" {
		fmt.Sscanf(ps, "%d", &pageSize)
	}

	products, total, err := h.svc.GetByUser(c.Request.Context(), uid, page, pageSize)
	if err != nil {
		InternalError(c, err.Error())
		return
	}
	Success(c, gin.H{
		"items": products,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

func (h *ServiceProductHandler) GetSchedules(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		BadRequest(c, "invalid service product id")
		return
	}
	var eventID *uuid.UUID
	if eid := c.Query("event_id"); eid != "" {
		if parsed, err := uuid.Parse(eid); err == nil {
			eventID = &parsed
		}
	}
	schedules, err := h.svc.GetSchedules(c.Request.Context(), id, eventID)
	if err != nil {
		InternalError(c, err.Error())
		return
	}
	Success(c, schedules)
}

type PromotionHandler struct {
	svc *service.PromotionService
}

func NewPromotionHandler(r *gin.RouterGroup, svc *service.PromotionService, authMW gin.HandlerFunc, requireAdmin gin.HandlerFunc) {
	h := &PromotionHandler{svc: svc}

	user := r.Group("/promotions")
	user.Use(authMW)
	user.POST("", h.CreateApplication)
	user.GET("/my", h.MyApplications)

	admin := r.Group("/admin/promotions")
	admin.Use(authMW)
	admin.Use(requireAdmin)
	admin.GET("", h.ListApplications)
	admin.POST("/:id/review", h.ReviewApplication)
}

type createPromotionRequest struct {
	TargetType string  `json:"target_type" binding:"required,oneof=product service_product"`
	TargetID   string  `json:"target_id" binding:"required"`
	Budget     float64 `json:"budget" binding:"required,gt=0"`
	Duration   int     `json:"duration" binding:"required,gt=0,lte=365"`
	Reason     string  `json:"reason" binding:"required,max=1000"`
}

func (h *PromotionHandler) CreateApplication(c *gin.Context) {
	var req createPromotionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request: "+err.Error())
		return
	}
	targetID, err := uuid.Parse(req.TargetID)
	if err != nil {
		BadRequest(c, "invalid target_id")
		return
	}
	userIDStr, _ := c.Get("user_id")
	uid, _ := uuid.Parse(userIDStr.(string))

	app, err := h.svc.CreateApplication(c.Request.Context(), service.CreatePromotionApplicationInput{
		UserID:     uid,
		TargetType: req.TargetType,
		TargetID:   targetID,
		Budget:     req.Budget,
		Duration:   req.Duration,
		Reason:     req.Reason,
	})
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	Success(c, app)
}

func (h *PromotionHandler) MyApplications(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	uid, _ := uuid.Parse(userIDStr.(string))
	page := 1
	if p := c.Query("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}
	pageSize := 20
	if ps := c.Query("page_size"); ps != "" {
		fmt.Sscanf(ps, "%d", &pageSize)
	}

	result, err := h.svc.GetUserApplications(uid, page, pageSize)
	if err != nil {
		InternalError(c, err.Error())
		return
	}
	Success(c, result)
}

type reviewPromotionRequest struct {
	Approved        bool    `json:"approved"`
	RejectionReason *string `json:"rejection_reason,omitempty"`
}

func (h *PromotionHandler) ReviewApplication(c *gin.Context) {
	appID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		BadRequest(c, "invalid application ID")
		return
	}
	var req reviewPromotionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request: "+err.Error())
		return
	}
	adminIDStr, _ := c.Get("user_id")
	adminID, _ := uuid.Parse(adminIDStr.(string))

	if err := h.svc.ReviewApplication(c.Request.Context(), appID, service.ReviewPromotionInput{
		AdminID:         adminID,
		Approved:        req.Approved,
		RejectionReason: req.RejectionReason,
	}); err != nil {
		BadRequest(c, err.Error())
		return
	}
	Success(c, gin.H{"reviewed": true})
}

func (h *PromotionHandler) ListApplications(c *gin.Context) {
	var input service.PromotionListInput
	input.Status = c.Query("status")
	input.TargetType = c.Query("target_type")
	if p := c.Query("page"); p != "" {
		fmt.Sscanf(p, "%d", &input.Page)
	}
	if ps := c.Query("page_size"); ps != "" {
		fmt.Sscanf(ps, "%d", &input.PageSize)
	}

	results, err := h.svc.GetPromotionsWithTarget(input)
	if err != nil {
		InternalError(c, err.Error())
		return
	}
	Success(c, results)
}
