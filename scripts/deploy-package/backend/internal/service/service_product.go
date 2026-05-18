package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/planforever/nexusacg/internal/model"
	"gorm.io/gorm"
)

type ServiceProductService struct {
	db *gorm.DB
}

func NewServiceProductService(db *gorm.DB) *ServiceProductService {
	return &ServiceProductService{db: db}
}

type ServiceProductListInput struct {
	ServiceType string  `form:"service_type"`
	CategoryID  string  `form:"category_id"`
	Keyword     string  `form:"keyword"`
	MinPrice    float64 `form:"min_price"`
	MaxPrice    float64 `form:"max_price"`
	Page        int     `form:"page,default=1"`
	PageSize    int     `form:"page_size,default=20"`
	Sort        string  `form:"sort,default=created_at"`
}

type ServiceProductWithProvider struct {
	model.ServiceProduct
	ProviderType   string  `json:"provider_type"`
	IsVerified     bool    `json:"is_verified"`
	ProviderName   string  `json:"provider_name"`
	ProviderAvatar *string `json:"provider_avatar"`
	Rating         float64 `json:"rating"`
}

type CreateServiceProductInput struct {
	ServiceProviderID uuid.UUID  `json:"service_provider_id"`
	UserID            uuid.UUID  `json:"-"`
	CategoryID        *uuid.UUID `json:"category_id"`
	Name              string     `json:"name"`
	Description       string     `json:"description"`
	Price             float64    `json:"price"`
	OriginalPrice     *float64   `json:"original_price"`
	ServiceType       string     `json:"service_type"`
	Images            []string   `json:"images"`
	PortfolioImages   []string   `json:"portfolio_images"`
	Tags              []string   `json:"tags"`
}

func (s *ServiceProductService) List(ctx context.Context, input ServiceProductListInput) ([]ServiceProductWithProvider, int64, error) {
	query := s.db.Table("service_products").
		Select("service_products.*, sp.type AS provider_type, sp.is_verified, sp.rating, u.nickname AS provider_name, u.avatar_url AS provider_avatar").
		Joins("JOIN service_providers sp ON service_products.service_provider_id = sp.id").
		Joins("JOIN users u ON service_products.user_id = u.id").
		Where("service_products.status = ?", "active")

	if input.ServiceType != "" {
		query = query.Where("service_products.service_type = ?", input.ServiceType)
	}
	if input.CategoryID != "" {
		if catID, err := uuid.Parse(input.CategoryID); err == nil {
			query = query.Where("service_products.category_id = ?", catID)
		}
	}
	if input.Keyword != "" {
		words := []string{}
		for _, word := range strings.Fields(input.Keyword) {
			words = append(words, "%"+word+"%")
		}
		conditions := make([]string, len(words))
		args := make([]interface{}, 0, len(words)*2)
		for i, w := range words {
			conditions[i] = "(service_products.name ILIKE ? OR service_products.description ILIKE ?)"
			args = append(args, w, w)
		}
		query = query.Where(strings.Join(conditions, " AND "), args...)
	}
	if input.MinPrice > 0 {
		query = query.Where("service_products.price >= ?", input.MinPrice)
	}
	if input.MaxPrice > 0 {
		query = query.Where("service_products.price <= ?", input.MaxPrice)
	}

	var total int64
	countQuery := s.db.Table("service_products").Where("status = ?", "active")
	if input.ServiceType != "" {
		countQuery = countQuery.Where("service_type = ?", input.ServiceType)
	}
	if input.CategoryID != "" {
		if catID, err := uuid.Parse(input.CategoryID); err == nil {
			countQuery = countQuery.Where("category_id = ?", catID)
		}
	}
	if input.Keyword != "" {
		words := []string{}
		for _, word := range strings.Fields(input.Keyword) {
			words = append(words, "%"+word+"%")
		}
		conditions := make([]string, len(words))
		args := make([]interface{}, 0, len(words)*2)
		for i, w := range words {
			conditions[i] = "(name ILIKE ? OR description ILIKE ?)"
			args = append(args, w, w)
		}
		countQuery = countQuery.Where(strings.Join(conditions, " AND "), args...)
	}
	if input.MinPrice > 0 {
		countQuery = countQuery.Where("price >= ?", input.MinPrice)
	}
	if input.MaxPrice > 0 {
		countQuery = countQuery.Where("price <= ?", input.MaxPrice)
	}
	countQuery.Count(&total)

	offset := (input.Page - 1) * input.PageSize
	if offset < 0 {
		offset = 0
	}

	switch input.Sort {
	case "price_asc":
		query = query.Order("service_products.price ASC")
	case "price_desc":
		query = query.Order("service_products.price DESC")
	case "rating":
		query = query.Order("sp.rating DESC")
	case "verified_first":
		query = query.Order("sp.is_verified DESC, service_products.created_at DESC")
	default:
		query = query.Order("service_products.created_at DESC")
	}

	var results []ServiceProductWithProvider
	if err := query.Offset(offset).Limit(input.PageSize).Find(&results).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list service products: %w", err)
	}
	return results, total, nil
}

func (s *ServiceProductService) Get(ctx context.Context, id uuid.UUID) (*ServiceProductWithProvider, error) {
	var result ServiceProductWithProvider
	err := s.db.Table("service_products").
		Select("service_products.*, sp.type AS provider_type, sp.is_verified, sp.rating, u.nickname AS provider_name, u.avatar_url AS provider_avatar").
		Joins("JOIN service_providers sp ON service_products.service_provider_id = sp.id").
		Joins("JOIN users u ON service_products.user_id = u.id").
		Where("service_products.id = ? AND service_products.status = ?", id, "active").
		First(&result).Error
	if err != nil {
		return nil, fmt.Errorf("service product not found")
	}
	return &result, nil
}

func (s *ServiceProductService) GetByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.ServiceProduct, int64, error) {
	query := s.db.Model(&model.ServiceProduct{}).Where("user_id = ?", userID)
	var total int64
	query.Count(&total)
	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}
	var products []model.ServiceProduct
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&products).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list service products: %w", err)
	}
	return products, total, nil
}

func (s *ServiceProductService) Create(ctx context.Context, input CreateServiceProductInput) (*model.ServiceProduct, error) {
	var sp model.ServiceProvider
	if err := s.db.Where("id = ? AND user_id = ?", input.ServiceProviderID, input.UserID).First(&sp).Error; err != nil {
		return nil, fmt.Errorf("service provider not found or does not belong to user")
	}

	if input.ServiceType != sp.Type {
		return nil, fmt.Errorf("service type must match provider type")
	}

	if len(input.PortfolioImages) == 0 {
		return nil, fmt.Errorf("portfolio images required for service listings")
	}

	if input.Images == nil {
		input.Images = []string{}
	}
	if input.Tags == nil {
		input.Tags = []string{}
	}

	product := model.ServiceProduct{
		ID:                uuid.New(),
		ServiceProviderID: input.ServiceProviderID,
		UserID:            input.UserID,
		CategoryID:        input.CategoryID,
		Name:              input.Name,
		Description:       input.Description,
		Price:             input.Price,
		OriginalPrice:     input.OriginalPrice,
		ServiceType:       input.ServiceType,
		Images:            input.Images,
		PortfolioImages:   input.PortfolioImages,
		Status:            "active",
		Tags:              input.Tags,
	}

	if err := s.db.Create(&product).Error; err != nil {
		return nil, fmt.Errorf("failed to create service product: %w", err)
	}
	return &product, nil
}

func (s *ServiceProductService) Update(ctx context.Context, id uuid.UUID, userID uuid.UUID, input CreateServiceProductInput) (*model.ServiceProduct, error) {
	var product model.ServiceProduct
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&product).Error; err != nil {
		return nil, fmt.Errorf("service product not found")
	}

	updates := map[string]interface{}{}
	if input.Name != "" {
		updates["name"] = input.Name
	}
	if input.Description != "" {
		updates["description"] = input.Description
	}
	if input.Price > 0 {
		updates["price"] = input.Price
	}
	if input.OriginalPrice != nil {
		updates["original_price"] = *input.OriginalPrice
	}
	if len(input.Images) > 0 {
		updates["images"] = input.Images
	}
	if len(input.PortfolioImages) > 0 {
		updates["portfolio_images"] = input.PortfolioImages
	}
	if len(input.Tags) > 0 {
		updates["tags"] = input.Tags
	}

	if len(updates) == 0 {
		return &product, nil
	}

	if err := s.db.Model(&product).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update service product: %w", err)
	}

	var updated model.ServiceProduct
	if err := s.db.Where("id = ?", id).First(&updated).Error; err != nil {
		return nil, fmt.Errorf("service product not found after update")
	}
	return &updated, nil
}

func (s *ServiceProductService) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	result := s.db.Model(&model.ServiceProduct{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("status", "deleted")
	if result.Error != nil {
		return fmt.Errorf("failed to delete service product: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("service product not found")
	}
	return nil
}

func (s *ServiceProductService) GetSchedules(ctx context.Context, serviceProductID uuid.UUID, eventID *uuid.UUID) ([]model.ServiceSchedule, error) {
	var sp model.ServiceProduct
	if err := s.db.Where("id = ?", serviceProductID).First(&sp).Error; err != nil {
		return nil, fmt.Errorf("service product not found")
	}

	query := s.db.Where("service_provider_id = ?", sp.ServiceProviderID)
	if eventID != nil {
		query = query.Where("event_id = ?", eventID)
	}
	var schedules []model.ServiceSchedule
	if err := query.Order("date ASC").Find(&schedules).Error; err != nil {
		return nil, fmt.Errorf("failed to list schedules: %w", err)
	}
	return schedules, nil
}
