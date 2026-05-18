package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/planforever/nexusacg/internal/model"
	"gorm.io/gorm"
)

type CategoryService struct {
	db *gorm.DB
}

func NewCategoryService(db *gorm.DB) *CategoryService {
	return &CategoryService{db: db}
}

type CategoryListInput struct {
	Zone string `form:"zone"`
}

func (s *CategoryService) List(ctx context.Context, input CategoryListInput) ([]model.Category, error) {
	query := s.db.Model(&model.Category{})
	if input.Zone != "" {
		query = query.Where("zone = ?", input.Zone)
	}
	var categories []model.Category
	if err := query.Order("sort_order ASC, created_at ASC").Find(&categories).Error; err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}
	return categories, nil
}

func (s *CategoryService) Get(ctx context.Context, id uuid.UUID) (*model.Category, error) {
	var category model.Category
	if err := s.db.Where("id = ?", id).First(&category).Error; err != nil {
		return nil, fmt.Errorf("category not found")
	}
	return &category, nil
}

type CreateCategoryInput struct {
	Name      string     `json:"name" binding:"required"`
	Zone      string     `json:"zone" binding:"required"`
	ParentID  *uuid.UUID `json:"parent_id"`
	IconURL   *string    `json:"icon_url"`
	SortOrder int        `json:"sort_order"`
}

func (s *CategoryService) Create(ctx context.Context, input CreateCategoryInput) (*model.Category, error) {
	category := model.Category{
		ID:        uuid.New(),
		Name:      input.Name,
		Zone:      input.Zone,
		ParentID:  input.ParentID,
		IconURL:   input.IconURL,
		SortOrder: input.SortOrder,
	}
	if err := s.db.Create(&category).Error; err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}
	return &category, nil
}

type UpdateCategoryInput struct {
	Name      *string    `json:"name"`
	Zone      *string    `json:"zone"`
	ParentID  *uuid.UUID `json:"parent_id"`
	IconURL   *string    `json:"icon_url"`
	SortOrder *int       `json:"sort_order"`
}

func (s *CategoryService) Update(ctx context.Context, id uuid.UUID, input UpdateCategoryInput) (*model.Category, error) {
	updates := map[string]interface{}{}
	if input.Name != nil {
		updates["name"] = *input.Name
	}
	if input.Zone != nil {
		updates["zone"] = *input.Zone
	}
	if input.ParentID != nil {
		updates["parent_id"] = *input.ParentID
	}
	if input.IconURL != nil {
		updates["icon_url"] = *input.IconURL
	}
	if input.SortOrder != nil {
		updates["sort_order"] = *input.SortOrder
	}
	if len(updates) == 0 {
		return s.Get(ctx, id)
	}

	result := s.db.Model(&model.Category{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to update category: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("category not found")
	}
	return s.Get(ctx, id)
}

func (s *CategoryService) Delete(ctx context.Context, id uuid.UUID) error {
	result := s.db.Where("id = ?", id).Delete(&model.Category{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete category: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("category not found")
	}
	return nil
}
