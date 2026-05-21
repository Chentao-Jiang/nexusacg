package handler

import (
	"strconv"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/planforever/nexusacg/internal/model"
	"github.com/planforever/nexusacg/internal/service"
)

type SPHandler struct{ svc *service.ServiceProviderSvc }

func NewSPHandler(r *gin.RouterGroup, svc *service.ServiceProviderSvc, authMW gin.HandlerFunc) {
	h := &SPHandler{svc: svc}
	public := r.Group("/service-providers")
	private := r.Group("/service-providers")
	private.Use(authMW)

	public.GET("", h.List)
	public.GET("/:id", h.Get)
	public.GET("/:id/reviews", h.GetReviews)
	private.POST("/:id/review", h.AddReview)
	private.GET("/me", h.MyProfile)
	private.PUT("/me", h.UpdateProfile)
	private.POST("/:id/book", h.Book)
	private.GET("/my-bookings", h.MyBookings)
}

func (h *SPHandler) List(c *gin.Context) {
	var input service.SPListInput
	c.ShouldBindQuery(&input)
	result, _ := h.svc.List(input)
	Success(c, result)
}

func (h *SPHandler) Get(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	sp, err := h.svc.Get(id)
	if err != nil { NotFound(c, err.Error()); return }
	Success(c, sp)
}

func (h *SPHandler) MyProfile(c *gin.Context) {
	uid, _ := c.Get("user_id")
	userID, _ := uuid.Parse(uid.(string))
	sp, err := h.svc.GetByUserID(userID)
	if err != nil { NotFound(c, "profile not found"); return }
	Success(c, sp)
}

func (h *SPHandler) UpdateProfile(c *gin.Context) {
	uid, _ := c.Get("user_id")
	userID, _ := uuid.Parse(uid.(string))
	var sp struct {
		Type            string   `json:"type"`
		Description     string   `json:"description"`
		PortfolioImages []string `json:"portfolio_images"`
		PriceList       []string `json:"price_list"`
	}
	c.ShouldBindJSON(&sp)
	provider := &model.ServiceProvider{ID: uuid.New(), UserID: userID, Type: sp.Type, Description: sp.Description,
		PortfolioImages: sp.PortfolioImages, PriceList: sp.PriceList}
	if err := h.svc.CreateOrUpdate(provider); err != nil {
		BadRequest(c, err.Error())
		return
	}
	Success(c, provider)
}

func (h *SPHandler) Book(c *gin.Context) {
	uid, _ := c.Get("user_id")
	userID, _ := uuid.Parse(uid.(string))
	providerID, _ := uuid.Parse(c.Param("id"))
	var input struct {
		ScheduleID  string `json:"schedule_id"`
		ServiceType string `json:"service_type"`
		Notes       string `json:"notes"`
	}
	c.ShouldBindJSON(&input)
	var sid *uuid.UUID
	if input.ScheduleID != "" {
		id, _ := uuid.Parse(input.ScheduleID)
		sid = &id
	}
	booking := &service.Booking{
		ServiceProviderID: providerID, UserID: userID,
		ScheduleID: sid, ServiceType: input.ServiceType, Notes: input.Notes,
	}
	if err := h.svc.CreateBooking(booking); err != nil {
		BadRequest(c, err.Error()); return
	}
	Success(c, gin.H{"booked": true})
}

func (h *SPHandler) MyBookings(c *gin.Context) {
	uid, _ := c.Get("user_id")
	userID, _ := uuid.Parse(uid.(string))
	bookings, _ := h.svc.GetMyBookings(userID)
	if bookings == nil { bookings = []map[string]interface{}{} }
	Success(c, gin.H{"items": bookings})
}

func (h *SPHandler) AddReview(c *gin.Context) {
	uid, _ := c.Get("user_id")
	userID, _ := uuid.Parse(uid.(string))
	providerID, _ := uuid.Parse(c.Param("id"))
	var input struct {
		Rating  int    `json:"rating"`
		Comment string `json:"comment"`
	}
	c.ShouldBindJSON(&input)
	if input.Rating < 1 || input.Rating > 5 {
		BadRequest(c, "rating must be 1-5"); return
	}
	if err := h.svc.AddReview(userID, providerID, input.Rating, input.Comment); err != nil {
		BadRequest(c, err.Error()); return
	}
	Success(c, gin.H{"reviewed": true})
}

func (h *SPHandler) GetReviews(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	ps, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	reviews, total, _ := h.svc.GetReviews(id, page, ps)
	Success(c, gin.H{"items": reviews, "total": total})
}
