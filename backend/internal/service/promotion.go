package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/planforever/nexusacg/internal/model"
	"gorm.io/gorm"
)

type PromotionService struct {
	db *gorm.DB
}

func NewPromotionService(db *gorm.DB) *PromotionService {
	return &PromotionService{db: db}
}

type CreatePromotionApplicationInput struct {
	UserID     uuid.UUID `json:"-"`
	TargetType string    `json:"target_type"` // product | service_product
	TargetID   uuid.UUID `json:"target_id"`
	Budget     float64   `json:"budget"`
	Duration   int       `json:"duration"` // days
	Reason     string    `json:"reason"`
}

type ReviewPromotionInput struct {
	AdminID         uuid.UUID `json:"-"`
	Approved        bool      `json:"approved"`
	RejectionReason *string   `json:"rejection_reason,omitempty"`
}

type PromotionListInput struct {
	Status     string `form:"status"`
	TargetType string `form:"target_type"`
	Page       int    `form:"page"`
	PageSize   int    `form:"page_size"`
}

type PromotionListResult struct {
	Items    []model.PromotionApplication `json:"items"`
	Total    int64                        `json:"total"`
	Page     int                          `json:"page"`
	PageSize int                          `json:"page_size"`
}

type PromotionWithTarget struct {
	model.PromotionApplication
	TargetName string `json:"target_name"`
	TargetZone string `json:"target_zone"`
}

func (s *PromotionService) CreateApplication(ctx context.Context, input CreatePromotionApplicationInput) (*model.PromotionApplication, error) {
	if input.TargetType != "product" && input.TargetType != "service_product" {
		return nil, fmt.Errorf("invalid target type: must be product or service_product")
	}

	if err := s.verifyTargetOwnership(input.TargetType, input.TargetID, input.UserID); err != nil {
		return nil, err
	}

	var existing model.PromotionApplication
	err := s.db.Where("target_type = ? AND target_id = ? AND status IN ?",
		input.TargetType, input.TargetID, []string{"pending", "active", "approved"}).
		First(&existing).Error
	if err == nil {
		return nil, fmt.Errorf("target already has an active or pending promotion")
	}

	app := model.PromotionApplication{
		ID:         uuid.New(),
		UserID:     input.UserID,
		TargetType: input.TargetType,
		TargetID:   input.TargetID,
		Budget:     input.Budget,
		Duration:   input.Duration,
		Reason:     input.Reason,
		Status:     "pending",
	}

	if err := s.db.Create(&app).Error; err != nil {
		return nil, fmt.Errorf("failed to create promotion application: %w", err)
	}
	return &app, nil
}

func (s *PromotionService) verifyTargetOwnership(targetType string, targetID, userID uuid.UUID) error {
	if targetType == "product" {
		var p model.Product
		if err := s.db.Where("id = ? AND seller_id = ? AND status = ?", targetID, userID, "active").First(&p).Error; err != nil {
			return fmt.Errorf("product not found or does not belong to user")
		}
	} else {
		var sp model.ServiceProduct
		if err := s.db.Where("id = ? AND user_id = ? AND status = ?", targetID, userID, "active").First(&sp).Error; err != nil {
			return fmt.Errorf("service product not found or does not belong to user")
		}
	}
	return nil
}

func (s *PromotionService) GetUserApplications(userID uuid.UUID, page, pageSize int) (*PromotionListResult, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	query := s.db.Model(&model.PromotionApplication{}).Where("user_id = ?", userID)
	var total int64
	query.Count(&total)

	var items []model.PromotionApplication
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, fmt.Errorf("failed to list promotions: %w", err)
	}

	return &PromotionListResult{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *PromotionService) ReviewApplication(ctx context.Context, appID uuid.UUID, input ReviewPromotionInput) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var app model.PromotionApplication
		if err := tx.Where("id = ? AND status = ?", appID, "pending").First(&app).Error; err != nil {
			return fmt.Errorf("application not found or already reviewed")
		}

		now := time.Now()
		updates := map[string]interface{}{
			"reviewed_by": input.AdminID,
			"reviewed_at": now,
			"updated_at":  now,
		}

		if input.Approved {
			updates["status"] = "active"
			updates["activated_at"] = now
			updates["expires_at"] = now.AddDate(0, 0, app.Duration)
			if err := tx.Model(&app).Updates(updates).Error; err != nil {
				return err
			}
		} else {
			updates["status"] = "rejected"
			updates["rejection_reason"] = input.RejectionReason
			if err := tx.Model(&app).Updates(updates).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *PromotionService) ListApplications(input PromotionListInput) (*PromotionListResult, error) {
	if input.Page <= 0 {
		input.Page = 1
	}
	if input.PageSize <= 0 {
		input.PageSize = 20
	}

	query := s.db.Model(&model.PromotionApplication{})
	if input.Status != "" {
		query = query.Where("status = ?", input.Status)
	}
	if input.TargetType != "" {
		query = query.Where("target_type = ?", input.TargetType)
	}

	var total int64
	query.Count(&total)

	var items []model.PromotionApplication
	offset := (input.Page - 1) * input.PageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(input.PageSize).Find(&items).Error; err != nil {
		return nil, fmt.Errorf("failed to list promotions: %w", err)
	}

	return &PromotionListResult{
		Items:    items,
		Total:    total,
		Page:     input.Page,
		PageSize: input.PageSize,
	}, nil
}

func (s *PromotionService) GetPromotionsWithTarget(input PromotionListInput) ([]PromotionWithTarget, error) {
	var apps []model.PromotionApplication
	query := s.db.Model(&model.PromotionApplication{})
	if input.Status != "" {
		query = query.Where("status = ?", input.Status)
	}
	if input.TargetType != "" {
		query = query.Where("target_type = ?", input.TargetType)
	}
	if err := query.Order("created_at DESC").Limit(50).Find(&apps).Error; err != nil {
		return nil, fmt.Errorf("failed to list promotions: %w", err)
	}

	var results []PromotionWithTarget
	for _, app := range apps {
		var name, zone string
		if app.TargetType == "product" {
			var p model.Product
			if err := s.db.Select("name, zone").Where("id = ?", app.TargetID).First(&p).Error; err == nil {
				name = p.Name
				zone = p.Zone
			}
		} else {
			var sp model.ServiceProduct
			if err := s.db.Select("name").Where("id = ?", app.TargetID).First(&sp).Error; err == nil {
				name = sp.Name
				zone = "service"
			}
		}
		results = append(results, PromotionWithTarget{
			PromotionApplication: app,
			TargetName:           name,
			TargetZone:           zone,
		})
	}
	return results, nil
}

func (s *PromotionService) ExpirePromotions(ctx context.Context) (int64, error) {
	now := time.Now()
	result := s.db.Model(&model.PromotionApplication{}).
		Where("status = ? AND expires_at < ?", "active", now).
		Update("status", "expired")
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}
