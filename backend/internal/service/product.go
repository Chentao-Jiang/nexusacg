package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/planforever/nexusacg/internal/model"
	"gorm.io/gorm"
)

type ProductService struct {
	db *gorm.DB
}

func NewProductService(db *gorm.DB) *ProductService {
	return &ProductService{db: db}
}

type ProductListInput struct {
	Zone       string  `form:"zone"`
	CategoryID string  `form:"category_id"`
	Keyword    string  `form:"keyword"`
	Tags       string  `form:"tags"`
	MinPrice   float64 `form:"min_price"`
	MaxPrice   float64 `form:"max_price"`
	SellerType string  `form:"seller_type"`
	Page       int     `form:"page,default=1"`
	PageSize   int     `form:"page_size,default=20"`
	Sort       string  `form:"sort,default=certified_first"`
}

type ProductListResult struct {
	Items      []ProductWithCategory `json:"items"`
	Total      int64                 `json:"total"`
	Page       int                   `json:"page"`
	Size       int                   `json:"page_size"`
}

type ProductWithCategory struct {
	model.Product
	CategoryName string `json:"category_name"`
}

func (s *ProductService) List(ctx context.Context, input ProductListInput) (*ProductListResult, error) {
	query := s.db.Table("products").
		Select("products.*, COALESCE(categories.name, '') AS category_name").
		Joins("LEFT JOIN categories ON products.category_id = categories.id").
		Where("products.status = ?", "active")

	if input.Zone != "" {
		query = query.Where("products.zone = ?", input.Zone)
	}
	if input.CategoryID != "" {
		if catID, err := uuid.Parse(input.CategoryID); err == nil {
			query = query.Where("products.category_id = ?", catID)
		}
	}
	if input.SellerType != "" {
		query = query.Where("products.seller_type = ?", input.SellerType)
	}
	if input.Keyword != "" {
		words := []string{}
		for _, word := range strings.Fields(input.Keyword) {
			words = append(words, "%"+word+"%")
		}
		conditions := make([]string, len(words))
		args := make([]interface{}, 0, len(words)*4)
		for i, w := range words {
			conditions[i] = "(products.name ILIKE ? OR products.anime_name ILIKE ? OR products.character_name ILIKE ? OR products.description ILIKE ?)"
			args = append(args, w, w, w, w)
		}
		query = query.Where(strings.Join(conditions, " AND "), args...)
	}
	if input.Tags != "" {
		tagList := strings.Split(input.Tags, ",")
		for _, tag := range tagList {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				tagJSON, _ := json.Marshal([]string{tag})
				query = query.Where("products.tags @> ?::jsonb", string(tagJSON))
			}
		}
	}
	if input.MinPrice > 0 {
		query = query.Where("products.price >= ?", input.MinPrice)
	}
	if input.MaxPrice > 0 {
		query = query.Where("products.price <= ?", input.MaxPrice)
	}

	var total int64
	countQuery := s.db.Table("products").Where("status = ?", "active")
	if input.Zone != "" {
		countQuery = countQuery.Where("zone = ?", input.Zone)
	}
	if input.CategoryID != "" {
		if catID, err := uuid.Parse(input.CategoryID); err == nil {
			countQuery = countQuery.Where("category_id = ?", catID)
		}
	}
	if input.SellerType != "" {
		countQuery = countQuery.Where("seller_type = ?", input.SellerType)
	}
	if input.Keyword != "" {
		words := []string{}
		for _, word := range strings.Fields(input.Keyword) {
			words = append(words, "%"+word+"%")
		}
		conditions := make([]string, len(words))
		args := make([]interface{}, 0, len(words)*4)
		for i, w := range words {
			conditions[i] = "(name ILIKE ? OR anime_name ILIKE ? OR character_name ILIKE ? OR description ILIKE ?)"
			args = append(args, w, w, w, w)
		}
		countQuery = countQuery.Where(strings.Join(conditions, " AND "), args...)
	}
	if input.Tags != "" {
		tagList := strings.Split(input.Tags, ",")
		for _, tag := range tagList {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				tagJSON, _ := json.Marshal([]string{tag})
				countQuery = countQuery.Where("tags @> ?::jsonb", string(tagJSON))
			}
		}
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
	case "certified_first":
		query = query.Order("CASE products.seller_type WHEN 'certified_merchant' THEN 0 WHEN 'certified_service' THEN 1 ELSE 2 END, products.created_at DESC")
	case "price_asc":
		query = query.Order("products.price ASC")
	case "price_desc":
		query = query.Order("products.price DESC")
	case "popular":
		query = query.Order("products.created_at DESC, products.price ASC")
	default:
		query = query.Order("products.created_at DESC")
	}

	var results []ProductWithCategory
	if err := query.Offset(offset).Limit(input.PageSize).Find(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}

	return &ProductListResult{
		Items: results,
		Total: total,
		Page:  input.Page,
		Size:  input.PageSize,
	}, nil
}

func (s *ProductService) Get(ctx context.Context, id uuid.UUID) (*model.Product, error) {
	var product model.Product
	if err := s.db.Where("id = ? AND status = ?", id, "active").First(&product).Error; err != nil {
		return nil, fmt.Errorf("product not found")
	}
	return &product, nil
}

type CreateProductInput struct {
	SellerID   uuid.UUID  `json:"seller_id"`
	CategoryID *uuid.UUID `json:"category_id"`
	Name       string     `json:"name"`
	Description string    `json:"description"`
	Price      float64    `json:"price"`
	OriginalPrice *float64 `json:"original_price"`
	Zone       string     `json:"zone"`
	SourceType string     `json:"source_type"`
	SellerType string     `json:"seller_type"`
	Images     []string   `json:"images"`
	Stock      int        `json:"stock"`
	Tags       []string   `json:"tags"`
	CharacterName *string `json:"character_name"`
	AnimeName  *string    `json:"anime_name"`
}

func (s *ProductService) Create(ctx context.Context, input CreateProductInput) (*model.Product, error) {
	if input.Images == nil {
		input.Images = []string{}
	}
	if input.Tags == nil {
		input.Tags = []string{}
	}

	// Auto-determine seller_type if not provided
	if input.SellerType == "" {
		var app model.CertificationApplication
		err := s.db.Where("user_id = ? AND status = ?", input.SellerID, "approved").
			Order("updated_at DESC").First(&app).Error
		if err == nil {
			if app.Type == "merchant" {
				input.SellerType = "certified_merchant"
			} else if app.Type == "service_provider" {
				input.SellerType = "certified_service"
			}
		}
		if input.SellerType == "" {
			input.SellerType = "uncertified"
		}
	}

	product := model.Product{
		ID:            uuid.New(),
		SellerID:      input.SellerID,
		CategoryID:    input.CategoryID,
		Name:          input.Name,
		Description:   input.Description,
		Price:         input.Price,
		OriginalPrice: input.OriginalPrice,
		Zone:          input.Zone,
		SourceType:    input.SourceType,
		SellerType:    input.SellerType,
		Images:        input.Images,
		Stock:         input.Stock,
		Status:        "active",
		Tags:          input.Tags,
		CharacterName: input.CharacterName,
		AnimeName:     input.AnimeName,
	}

	if err := s.db.Create(&product).Error; err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return &product, nil
}

func (s *ProductService) GetMyProducts(userID uuid.UUID, page, pageSize int) (*ProductListResult, error) {
	if pageSize <= 0 { pageSize = 20 }
	if page <= 0 { page = 1 }
	var total int64
	s.db.Model(&model.Product{}).Where("user_id = ?", userID).Count(&total)
	var results []ProductWithCategory
	err := s.db.Table("products").
		Select("products.*, COALESCE(categories.name, '') AS category_name").
		Joins("LEFT JOIN categories ON products.category_id = categories.id").
		Where("products.user_id = ?", userID).
		Order("products.created_at DESC").
		Offset((page-1)*pageSize).Limit(pageSize).Find(&results).Error
	return &ProductListResult{Items: results, Total: total, Page: page, Size: pageSize}, err
}
