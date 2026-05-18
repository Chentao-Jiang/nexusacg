package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/planforever/nexusacg/internal/model"

	"gorm.io/gorm"
)

type CertificationService struct {
	db *gorm.DB
}

func NewCertificationService(db *gorm.DB) *CertificationService {
	return &CertificationService{db: db}
}

type ApplyMerchantInput struct {
	UserID             uuid.UUID `json:"-"`
	BusinessLicenseURL string    `json:"business_license_url"`
	ProductCategory    string    `json:"product_category"`
	StoreName          string    `json:"store_name"`
}

type ApplyServiceProviderInput struct {
	UserID          uuid.UUID `json:"-"`
	ProviderType    string    `json:"provider_type"` // makeup_artist | wig_stylist | photographer | post_editor | props_maker
	PortfolioImages []string  `json:"portfolio_images"`
}

type ApplicationListInput struct {
	Status   string `form:"status"`
	Type     string `form:"type"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

type ApplicationListResult struct {
	Items    []model.CertificationApplication `json:"items"`
	Total    int64                           `json:"total"`
	Page     int                             `json:"page"`
	PageSize int                             `json:"page_size"`
}

type ReviewApplicationInput struct {
	AdminID         uuid.UUID `json:"-"`
	Approved        bool      `json:"approved"`
	RejectionReason *string   `json:"rejection_reason,omitempty"`
}

func (s *CertificationService) ApplyMerchant(ctx context.Context, input ApplyMerchantInput) (*model.CertificationApplication, error) {
	var existing model.CertificationApplication
	err := s.db.Where("user_id = ? AND status IN ?", input.UserID, []string{"pending", "approved"}).
		Order("created_at DESC").First(&existing).Error
	if err == nil {
		return nil, fmt.Errorf("already has a %s application", existing.Status)
	}

	app := model.CertificationApplication{
		ID:                 uuid.New(),
		UserID:             input.UserID,
		Type:               "merchant",
		BusinessLicenseURL: &input.BusinessLicenseURL,
		ProductCategory:    &input.ProductCategory,
		StoreName:          &input.StoreName,
		Status:             "pending",
	}
	if err := s.db.Create(&app).Error; err != nil {
		return nil, fmt.Errorf("failed to create application: %w", err)
	}
	return &app, nil
}

func (s *CertificationService) ApplyServiceProvider(ctx context.Context, input ApplyServiceProviderInput) (*model.CertificationApplication, error) {
	validTypes := map[string]bool{
		"makeup_artist": true,
		"wig_stylist":   true,
		"photographer":  true,
		"post_editor":   true,
		"props_maker":   true,
	}
	if !validTypes[input.ProviderType] {
		return nil, fmt.Errorf("invalid provider type: %s", input.ProviderType)
	}
	if len(input.PortfolioImages) == 0 {
		return nil, fmt.Errorf("portfolio images required")
	}

	var existing model.CertificationApplication
	err := s.db.Where("user_id = ? AND status IN ?", input.UserID, []string{"pending", "approved"}).
		Order("created_at DESC").First(&existing).Error
	if err == nil {
		return nil, fmt.Errorf("already has a %s application", existing.Status)
	}

	app := model.CertificationApplication{
		ID:              uuid.New(),
		UserID:          input.UserID,
		Type:            "service_provider",
		ProviderType:    &input.ProviderType,
		PortfolioImages: input.PortfolioImages,
		Status:          "pending",
	}
	if err := s.db.Create(&app).Error; err != nil {
		return nil, fmt.Errorf("failed to create application: %w", err)
	}
	return &app, nil
}

func (s *CertificationService) GetUserApplication(userID uuid.UUID) (*model.CertificationApplication, error) {
	var app model.CertificationApplication
	err := s.db.Where("user_id = ?", userID).Order("created_at DESC").First(&app).Error
	if err != nil {
		return nil, fmt.Errorf("no application found")
	}
	return &app, nil
}

func (s *CertificationService) ListApplications(input ApplicationListInput) (*ApplicationListResult, error) {
	if input.Page <= 0 {
		input.Page = 1
	}
	if input.PageSize <= 0 {
		input.PageSize = 20
	}

	query := s.db.Model(&model.CertificationApplication{})
	if input.Status != "" {
		query = query.Where("status = ?", input.Status)
	}
	if input.Type != "" {
		query = query.Where("type = ?", input.Type)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count applications: %w", err)
	}

	var items []model.CertificationApplication
	offset := (input.Page - 1) * input.PageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(input.PageSize).Find(&items).Error; err != nil {
		return nil, fmt.Errorf("failed to list applications: %w", err)
	}

	return &ApplicationListResult{
		Items:    items,
		Total:    total,
		Page:     input.Page,
		PageSize: input.PageSize,
	}, nil
}

func (s *CertificationService) ReviewApplication(ctx context.Context, appID uuid.UUID, input ReviewApplicationInput) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var app model.CertificationApplication
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
			updates["status"] = "approved"
			if err := tx.Model(&app).Updates(updates).Error; err != nil {
				return err
			}

			// Find or create ServiceProvider record
			var sp model.ServiceProvider
			err := tx.Where("user_id = ?", app.UserID).First(&sp).Error
			if err == gorm.ErrRecordNotFound {
				providerType := ""
				if app.ProviderType != nil {
					providerType = *app.ProviderType
				}
				sp = model.ServiceProvider{
					ID:         uuid.New(),
					UserID:     app.UserID,
					Type:       providerType,
					IsVerified: true,
					Status:     "active",
				}
				if err := tx.Create(&sp).Error; err != nil {
					return err
				}
			} else if err != nil {
				return err
			} else {
				if app.ProviderType != nil {
					sp.Type = *app.ProviderType
				}
				sp.IsVerified = true
				sp.Status = "active"
				if err := tx.Save(&sp).Error; err != nil {
					return err
				}
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
