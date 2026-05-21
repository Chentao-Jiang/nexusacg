package service

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/planforever/nexusacg/internal/model"
	"gorm.io/gorm"
)

type SPListInput struct {
	Type     string `form:"type"`
	Search   string `form:"search"`
	Sort     string `form:"sort"` // rating, newest
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=20"`
}

type SPListResult struct {
	Items []model.ServiceProvider `json:"items"`
	Total int64                   `json:"total"`
	Page  int                     `json:"page"`
	Size  int                     `json:"page_size"`
}

type ServiceProviderSvc struct{ db *gorm.DB }

func NewServiceProviderSvc(db *gorm.DB) *ServiceProviderSvc { return &ServiceProviderSvc{db: db} }

func (s *ServiceProviderSvc) List(input SPListInput) (*SPListResult, error) {
	if input.PageSize <= 0 { input.PageSize = 20 }
	if input.Page <= 0 { input.Page = 1 }
	q := s.db.Model(&model.ServiceProvider{}).Where("status = ?", "active")
	if input.Type != "" { q = q.Where("type = ?", input.Type) }
	if input.Search != "" { q = q.Where("description ILIKE ?", "%"+input.Search+"%") }
	var total int64
	q.Count(&total)
	order := "rating DESC"
	if input.Sort == "newest" { order = "created_at DESC" }
	var items []model.ServiceProvider
	err := q.Order(order).Offset((input.Page-1)*input.PageSize).Limit(input.PageSize).Find(&items).Error
	return &SPListResult{Items: items, Total: total, Page: input.Page, Size: input.PageSize}, err
}

func (s *ServiceProviderSvc) Get(id uuid.UUID) (*model.ServiceProvider, error) {
	var sp model.ServiceProvider
	if err := s.db.First(&sp, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("provider not found")
	}
	return &sp, nil
}

func (s *ServiceProviderSvc) CreateOrUpdate(sp *model.ServiceProvider) error {
	var existing model.ServiceProvider
	if err := s.db.Where("user_id = ?", sp.UserID).First(&existing).Error; err == nil {
		return s.db.Model(&existing).Updates(map[string]interface{}{
			"type": sp.Type, "description": sp.Description,
			"portfolio_images": sp.PortfolioImages, "price_list": sp.PriceList,
		}).Error
	}
	return s.db.Create(sp).Error
}

// Booking
type Booking struct {
	ID                uuid.UUID  `json:"id"`
	ServiceProviderID uuid.UUID  `json:"service_provider_id"`
	UserID            uuid.UUID  `json:"user_id"`
	ScheduleID        *uuid.UUID `json:"schedule_id"`
	ServiceType       string     `json:"service_type"`
	Status            string     `json:"status"` // pending, confirmed, completed, cancelled
	Notes             string     `json:"notes"`
	CreatedAt         string     `json:"created_at"`
}

func (s *ServiceProviderSvc) CreateBooking(booking *Booking) error {
	return s.db.Table("bookings").Create(map[string]interface{}{
		"id": uuid.New(),
		"service_provider_id": booking.ServiceProviderID,
		"user_id": booking.UserID,
		"schedule_id": booking.ScheduleID,
		"service_type": booking.ServiceType,
		"status": "pending",
		"notes": booking.Notes,
		"created_at": gorm.Expr("NOW()"),
	}).Error
}

func (s *ServiceProviderSvc) GetMyBookings(userID uuid.UUID) ([]map[string]interface{}, error) {
	var bookings []map[string]interface{}
	err := s.db.Table("bookings").Where("user_id = ?", userID).Order("created_at DESC").Find(&bookings).Error
	return bookings, err
}

// Review
func (s *ServiceProviderSvc) AddReview(userID, providerID uuid.UUID, rating int, comment string) error {
	exists := s.db.Table("reviews").Where("user_id = ? AND service_provider_id = ?", userID, providerID).First(&map[string]interface{}{})
	if exists.Error == nil { return fmt.Errorf("already reviewed") }

	tx := s.db.Begin()
	if err := tx.Table("reviews").Create(map[string]interface{}{
		"id": uuid.New(),
		"user_id": userID, "service_provider_id": providerID,
		"rating": rating, "comment": comment,
		"created_at": gorm.Expr("NOW()"),
	}).Error; err != nil { tx.Rollback(); return err }

	// Update avg rating
	if err := tx.Model(&model.ServiceProvider{}).Where("id = ?", providerID).Updates(map[string]interface{}{
		"rating": gorm.Expr("(SELECT COALESCE(AVG(rating), 0) FROM reviews WHERE service_provider_id = ?)", providerID),
		"review_count": gorm.Expr("(SELECT COUNT(*) FROM reviews WHERE service_provider_id = ?)", providerID),
	}).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (s *ServiceProviderSvc) GetReviews(providerID uuid.UUID, page, pageSize int) ([]map[string]interface{}, int64, error) {
	if pageSize <= 0 { pageSize = 20 }
	var total int64
	s.db.Table("reviews").Where("service_provider_id = ?", providerID).Count(&total)
	var reviews []map[string]interface{}
	err := s.db.Table("reviews").
		Select("reviews.*, users.nickname as user_name, users.avatar_url").
		Joins("JOIN users ON users.id = reviews.user_id").
		Where("service_provider_id = ?", providerID).
		Order("created_at DESC").Offset((page-1)*pageSize).Limit(pageSize).Find(&reviews).Error
	return reviews, total, err
}

func (s *ServiceProviderSvc) GetByUserID(userID uuid.UUID) (*model.ServiceProvider, error) {
	var sp model.ServiceProvider
	if err := s.db.Where("user_id = ?", userID).First(&sp).Error; err != nil {
		return nil, fmt.Errorf("profile not found")
	}
	return &sp, nil
}
