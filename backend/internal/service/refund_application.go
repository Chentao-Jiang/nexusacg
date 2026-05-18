package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/planforever/nexusacg/internal/model"
	"gorm.io/gorm"
)

type RefundApplicationService struct {
	db *gorm.DB
}

func NewRefundApplicationService(db *gorm.DB) *RefundApplicationService {
	return &RefundApplicationService{db: db}
}

type CreateRefundApplicationInput struct {
	UserID       uuid.UUID `json:"-"`
	OrderNo      string    `json:"order_no"`
	RefundType   string    `json:"refund_type"`   // refund_only | return_refund
	Reason       string    `json:"reason"`
	EvidenceURLs []string  `json:"evidence_urls"`
	Amount       float64   `json:"amount"` // 0 = full refund
}

type ReviewRefundApplicationInput struct {
	ReviewerID uuid.UUID `json:"-"`
	Approved   bool      `json:"approved"`
	Note       *string   `json:"note,omitempty"`
}

type RefundApplicationListInput struct {
	Status   string `form:"status"`
	UserID   string `form:"user_id"`
	SellerID string `form:"seller_id"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

type RefundApplicationListResult struct {
	Items    []model.RefundApplication `json:"items"`
	Total    int64                     `json:"total"`
	Page     int                       `json:"page"`
	PageSize int                       `json:"page_size"`
}

var validRefundTypes = map[string]bool{
	"refund_only":    true,
	"return_refund":  true,
}

func (s *RefundApplicationService) Create(ctx context.Context, input CreateRefundApplicationInput) (*model.RefundApplication, error) {
	if !validRefundTypes[input.RefundType] {
		return nil, fmt.Errorf("invalid refund type: %s", input.RefundType)
	}
	if input.Reason == "" {
		return nil, fmt.Errorf("reason is required")
	}

	var order model.Order
	if err := s.db.Where("order_no = ?", input.OrderNo).First(&order).Error; err != nil {
		return nil, fmt.Errorf("order not found")
	}

	if order.PaymentStatus != "paid" {
		return nil, fmt.Errorf("only paid orders can request a refund")
	}

	if order.UserID != input.UserID {
		return nil, fmt.Errorf("only the order owner can request a refund")
	}

	var existing model.RefundApplication
	err := s.db.Where("order_id = ? AND status NOT IN ?", order.ID, []string{"rejected", "completed"}).
		First(&existing).Error
	if err == nil {
		return nil, fmt.Errorf("order already has an active refund application (status: %s)", existing.Status)
	}

	if order.OrderStatus == "refunded" {
		return nil, fmt.Errorf("order already refunded")
	}

	var item model.OrderItem
	if err := s.db.Where("order_id = ?", order.ID).First(&item).Error; err != nil {
		return nil, fmt.Errorf("order has no items")
	}
	var product model.Product
	if err := s.db.Where("id = ?", item.ProductID).First(&product).Error; err != nil {
		return nil, fmt.Errorf("product not found")
	}

	if input.Amount < 0 {
		return nil, fmt.Errorf("refund amount cannot be negative")
	}
	if input.Amount > order.TotalAmount {
		return nil, fmt.Errorf("refund amount cannot exceed order total")
	}

	var app model.RefundApplication
	err = s.db.Transaction(func(tx *gorm.DB) error {
		app = model.RefundApplication{
			ID:           uuid.New(),
			OrderID:      order.ID,
			OrderNo:      input.OrderNo,
			UserID:       input.UserID,
			SellerID:     product.SellerID,
			RefundType:   input.RefundType,
			Reason:       input.Reason,
			EvidenceURLs: input.EvidenceURLs,
			Amount:       input.Amount,
			Status:       "pending",
		}
		return tx.Create(&app).Error
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create refund application: %w", err)
	}
	return &app, nil
}

func (s *RefundApplicationService) GetByOrder(orderID uuid.UUID) (*model.RefundApplication, error) {
	var app model.RefundApplication
	if err := s.db.Where("order_id = ?", orderID).Order("created_at DESC").First(&app).Error; err != nil {
		return nil, fmt.Errorf("refund application not found")
	}
	return &app, nil
}

func (s *RefundApplicationService) List(input RefundApplicationListInput) (*RefundApplicationListResult, error) {
	if input.Page <= 0 {
		input.Page = 1
	}
	if input.PageSize <= 0 {
		input.PageSize = 20
	}

	query := s.db.Model(&model.RefundApplication{})
	if input.Status != "" {
		query = query.Where("status = ?", input.Status)
	}
	if input.UserID != "" {
		uid, err := uuid.Parse(input.UserID)
		if err == nil {
			query = query.Where("user_id = ?", uid)
		}
	}
	if input.SellerID != "" {
		sid, err := uuid.Parse(input.SellerID)
		if err == nil {
			query = query.Where("seller_id = ?", sid)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count refund applications: %w", err)
	}

	var items []model.RefundApplication
	offset := (input.Page - 1) * input.PageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(input.PageSize).Find(&items).Error; err != nil {
		return nil, fmt.Errorf("failed to list refund applications: %w", err)
	}

	return &RefundApplicationListResult{
		Items:    items,
		Total:    total,
		Page:     input.Page,
		PageSize: input.PageSize,
	}, nil
}

func (s *RefundApplicationService) SellerReview(appID uuid.UUID, input ReviewRefundApplicationInput) (*model.RefundApplication, error) {
	var app model.RefundApplication
	if err := s.db.Where("id = ? AND status IN ?", appID, []string{"pending", "seller_review"}).First(&app).Error; err != nil {
		return nil, fmt.Errorf("refund application not found or already reviewed")
	}

	if app.SellerID != input.ReviewerID {
		return nil, fmt.Errorf("only the seller can review this application")
	}

	now := time.Now()
	updates := map[string]interface{}{
		"seller_note": input.Note,
		"reviewed_by": input.ReviewerID,
		"reviewed_at": now,
		"updated_at":  now,
	}

	if input.Approved {
		updates["status"] = "approved"
	} else {
		updates["status"] = "rejected"
	}

	if err := s.db.Model(&app).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to review refund application: %w", err)
	}

	app.Status = updates["status"].(string)
	app.ReviewedBy = &input.ReviewerID
	app.ReviewedAt = &now
	app.SellerNote = input.Note
	return &app, nil
}

func (s *RefundApplicationService) AdminExecute(ctx context.Context, appID uuid.UUID, adminID uuid.UUID, note *string) (*model.RefundApplication, error) {
	var app model.RefundApplication
	if err := s.db.Where("id = ? AND status = ?", appID, "approved").First(&app).Error; err != nil {
		return nil, fmt.Errorf("refund application not found or not approved")
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()

		var order model.Order
		if err := tx.Where("id = ?", app.OrderID).First(&order).Error; err != nil {
			return err
		}

		refundAmount := app.Amount
		if refundAmount <= 0 {
			refundAmount = order.TotalAmount
		}

		if err := tx.Model(&app).Updates(map[string]interface{}{
			"status":       "completed",
			"admin_note":   note,
			"reviewed_by":  adminID,
			"reviewed_at":  now,
			"completed_at": now,
			"updated_at":   now,
		}).Error; err != nil {
			return err
		}

		if order.OrderStatus != "refunded" {
			if err := tx.Model(&model.Order{}).Where("id = ?", order.ID).Updates(map[string]interface{}{
				"order_status":   "refunded",
				"payment_status": "refunded",
				"updated_at":     now,
			}).Error; err != nil {
				return err
			}

			var items []model.OrderItem
			tx.Where("order_id = ?", order.ID).Find(&items)
			for _, item := range items {
				tx.Model(&model.Product{}).Where("id = ?", item.ProductID).
					Update("stock", gorm.Expr("stock + ?", item.Quantity))
			}
		}

		paymentMethod := ""
		if order.PaymentMethod != nil {
			paymentMethod = *order.PaymentMethod
		}
		profitShare := model.ProfitShareRecord{
			ID:            uuid.New(),
			OrderID:       app.OrderID,
			TotalAmount:   refundAmount,
			PlatformFee:   0,
			SellerAmount:  -refundAmount,
			Status:        "refunded",
			PaymentMethod: paymentMethod,
		}
		return tx.Create(&profitShare).Error
	})

	if err != nil {
		return nil, fmt.Errorf("failed to execute refund: %w", err)
	}

	app.Status = "completed"
	resolvedAt := time.Now()
	app.CompletedAt = &resolvedAt
	app.ReviewedBy = &adminID
	app.ReviewedAt = &resolvedAt
	app.AdminNote = note
	return &app, nil
}
