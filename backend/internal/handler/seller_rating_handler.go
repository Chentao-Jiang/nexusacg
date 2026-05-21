package handler

import (
	"strconv"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/planforever/nexusacg/internal/model"
	"github.com/planforever/nexusacg/internal/service"
)

type SellerRatingHandler struct{ svc *service.SellerRatingService }

func NewSellerRatingHandler(r *gin.RouterGroup, svc *service.SellerRatingService, authMW gin.HandlerFunc) {
	h := &SellerRatingHandler{svc: svc}
	r.POST("/rate/order/:order_id", authMW, h.Rate)
	r.GET("/ratings/seller/:seller_id", h.GetRatings)
}

func (h *SellerRatingHandler) Rate(c *gin.Context) {
	uid, _ := c.Get("user_id")
	buyerID, _ := uuid.Parse(uid.(string))
	orderID, _ := uuid.Parse(c.Param("order_id"))
	var input struct {
		Rating  int    `json:"rating"`
		Comment string `json:"comment"`
	}
	c.ShouldBindJSON(&input)
	if input.Rating < 1 || input.Rating > 5 { BadRequest(c, "rating must be 1-5"); return }
	var order model.Order
	if err := h.svc.DB().Preload("Items").Where("id = ? AND user_id = ?", orderID, buyerID).First(&order).Error; err != nil {
		BadRequest(c, "order not found"); return
	}
	if len(order.Items) == 0 { BadRequest(c, "order has no items"); return }
	var product model.Product
	if err := h.svc.DB().Where("id = ?", order.Items[0].ProductID).First(&product).Error; err != nil {
		BadRequest(c, "product not found"); return
	}
	if err := h.svc.Create(&model.SellerRating{
		ID: uuid.New(), SellerID: product.SellerID, BuyerID: buyerID,
		OrderID: orderID, Rating: input.Rating, Comment: input.Comment,
	}); err != nil { BadRequest(c, err.Error()); return }
	Success(c, gin.H{"rated": true})
}

func (h *SellerRatingHandler) GetRatings(c *gin.Context) {
	sellerID, _ := uuid.Parse(c.Param("seller_id"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	ps, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	ratings, total, _ := h.svc.GetSellerRatings(sellerID, page, ps)
	Success(c, gin.H{"items": ratings, "total": total})
}
