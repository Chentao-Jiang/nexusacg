package service

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/planforever/nexusacg/internal/model"
	"gorm.io/gorm"
)

type SellerRatingService struct{ db *gorm.DB }

func NewSellerRatingService(db *gorm.DB) *SellerRatingService { return &SellerRatingService{db: db} }

func (s *SellerRatingService) Create(sr *model.SellerRating) error {
	// Verify order exists and belongs to buyer
	var order model.Order
	if err := s.db.Where("id = ? AND user_id = ? AND status = ?", sr.OrderID, sr.BuyerID, "completed").First(&order).Error; err != nil {
		return fmt.Errorf("order not found or not completed")
	}
	if err := s.db.Create(sr).Error; err != nil {
		return fmt.Errorf("already rated or invalid")
	}
	// Update seller rating aggregate
	var avg float64
	s.db.Model(&model.SellerRating{}).Where("seller_id = ?", sr.SellerID).Select("COALESCE(AVG(rating), 0)").Scan(&avg)
	s.db.Model(&model.User{}).Where("id = ?", sr.SellerID).Updates(map[string]interface{}{
		"seller_rating": avg,
		"seller_review_count": gorm.Expr("(SELECT COUNT(*) FROM seller_ratings WHERE seller_id = ?)", sr.SellerID),
	})
	return nil
}

func (s *SellerRatingService) GetSellerRatings(sellerID uuid.UUID, page, pageSize int) ([]model.SellerRating, int64, error) {
	if pageSize <= 0 { pageSize = 20 }
	if page <= 0 { page = 1 }
	var total int64
	s.db.Model(&model.SellerRating{}).Where("seller_id = ?", sellerID).Count(&total)
	var ratings []model.SellerRating
	err := s.db.Where("seller_id = ?", sellerID).Order("created_at DESC").
		Offset((page-1)*pageSize).Limit(pageSize).Find(&ratings).Error
	return ratings, total, err
}

func (s *SellerRatingService) DB() *gorm.DB { return s.db }
