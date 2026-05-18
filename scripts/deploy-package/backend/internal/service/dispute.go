package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/planforever/nexusacg/internal/model"
	"gorm.io/gorm"
)

type DisputeService struct {
	db *gorm.DB
}

func NewDisputeService(db *gorm.DB) *DisputeService {
	return &DisputeService{db: db}
}

type CreateDisputeInput struct {
	UserID           uuid.UUID `json:"-"`
	OrderNo          string    `json:"order_no"`
	Reason           string    `json:"reason"`          // quality_issue | not_as_described | missing_item | counterfeit | service_not_delivered | other
	Description      string    `json:"description"`
	EvidenceImages   []string  `json:"evidence_images"`
	RefundAmount     float64   `json:"refund_amount"` // 0 = full refund
}

type RespondDisputeInput struct {
	UserID         uuid.UUID `json:"-"`
	Description    string    `json:"description"`
	EvidenceImages []string  `json:"evidence_images"`
}

type AdminResolveDisputeInput struct {
	AdminID       uuid.UUID `json:"-"`
	Decision      string    `json:"decision"` // full_refund | partial_refund | reject
	AdminNote     *string   `json:"admin_note,omitempty"`
}

type DisputeListInput struct {
	Status   string `form:"status"`
	UserID   string `form:"user_id"` // admin filter
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

type DisputeListResult struct {
	Items    []model.Dispute `json:"items"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
}

var validDisputeReasons = map[string]bool{
	"quality_issue":           true,
	"not_as_described":        true,
	"missing_item":            true,
	"counterfeit":             true,
	"service_not_delivered":   true,
	"other":                   true,
}

func (s *DisputeService) Create(ctx context.Context, input CreateDisputeInput) (*model.Dispute, error) {
	if !validDisputeReasons[input.Reason] {
		return nil, fmt.Errorf("invalid dispute reason: %s", input.Reason)
	}
	if input.Description == "" {
		return nil, fmt.Errorf("description is required")
	}

	var order model.Order
	if err := s.db.Where("order_no = ?", input.OrderNo).First(&order).Error; err != nil {
		return nil, fmt.Errorf("order not found")
	}

	if order.PaymentStatus != "paid" {
		return nil, fmt.Errorf("only paid orders can be disputed")
	}
	validStatuses := map[string]bool{"paid": true, "shipped": true, "completed": true}
	if !validStatuses[order.OrderStatus] {
		return nil, fmt.Errorf("order status %s cannot be disputed", order.OrderStatus)
	}

	if order.UserID != input.UserID {
		return nil, fmt.Errorf("only the order owner can open a dispute")
	}

	if order.DisputeStatus != "none" {
		return nil, fmt.Errorf("order already has an active dispute")
	}

	var dispute model.Dispute
	err := s.db.Transaction(func(tx *gorm.DB) error {
		dispute = model.Dispute{
			ID:               uuid.New(),
			OrderID:          order.ID,
			OrderNo:          input.OrderNo,
			InitiatorID:      input.UserID,
			SellerID:         order.UserID,
			Reason:           input.Reason,
			Description:      input.Description,
			EvidenceImages:   input.EvidenceImages,
			RefundAmount:     input.RefundAmount,
			Status:           "pending",
		}
		if err := tx.Create(&dispute).Error; err != nil {
			return err
		}

		return tx.Model(&model.Order{}).Where("id = ?", order.ID).Update("dispute_status", "disputed").Error
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create dispute: %w", err)
	}
	return &dispute, nil
}

func (s *DisputeService) GetByOrderNo(orderNo string) (*model.Dispute, error) {
	var order model.Order
	if err := s.db.Where("order_no = ?", orderNo).First(&order).Error; err != nil {
		return nil, fmt.Errorf("order not found")
	}

	var dispute model.Dispute
	if err := s.db.Where("order_id = ?", order.ID).Order("created_at DESC").First(&dispute).Error; err != nil {
		return nil, fmt.Errorf("dispute not found")
	}
	return &dispute, nil
}

func (s *DisputeService) List(input DisputeListInput) (*DisputeListResult, error) {
	if input.Page <= 0 {
		input.Page = 1
	}
	if input.PageSize <= 0 {
		input.PageSize = 20
	}

	query := s.db.Model(&model.Dispute{})
	if input.Status != "" {
		query = query.Where("status = ?", input.Status)
	}
	if input.UserID != "" {
		uid, err := uuid.Parse(input.UserID)
		if err == nil {
			query = query.Where("initiator_id = ? OR seller_id = ?", uid, uid)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count disputes: %w", err)
	}

	var items []model.Dispute
	offset := (input.Page - 1) * input.PageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(input.PageSize).Find(&items).Error; err != nil {
		return nil, fmt.Errorf("failed to list disputes: %w", err)
	}

	return &DisputeListResult{
		Items:    items,
		Total:    total,
		Page:     input.Page,
		PageSize: input.PageSize,
	}, nil
}

func (s *DisputeService) Respond(disputeID uuid.UUID, input RespondDisputeInput) (*model.Dispute, error) {
	var dispute model.Dispute
	if err := s.db.Where("id = ? AND status = ?", disputeID, "pending").First(&dispute).Error; err != nil {
		return nil, fmt.Errorf("dispute not found or already responded")
	}

	if dispute.SellerID != input.UserID {
		return nil, fmt.Errorf("only the seller can respond to this dispute")
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":       "seller_responded",
		"responded_by": input.UserID,
		"responded_at": now,
		"description":  dispute.Description + "\n[Seller Response]\n" + input.Description,
		"updated_at":   now,
	}

	if len(input.EvidenceImages) > 0 {
		existing := dispute.EvidenceImages
		if existing == nil {
			existing = []string{}
		}
		updates["evidence_images"] = append(existing, input.EvidenceImages...)
	}

	if err := s.db.Model(&dispute).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to respond to dispute: %w", err)
	}

	dispute.Status = "seller_responded"
	dispute.RespondedBy = &input.UserID
	dispute.RespondedAt = &now
	return &dispute, nil
}

func (s *DisputeService) SendMessage(disputeID uuid.UUID, senderID uuid.UUID, content string, images []string) (*model.DisputeMessage, error) {
	var dispute model.Dispute
	if err := s.db.Where("id = ?", disputeID).First(&dispute).Error; err != nil {
		return nil, fmt.Errorf("dispute not found")
	}

	// Determine role from sender
	role := "buyer"
	if dispute.SellerID == senderID {
		role = "seller"
	}

	msg := model.DisputeMessage{
		ID:         uuid.New(),
		DisputeID:  disputeID,
		SenderID:   senderID,
		SenderRole: role,
		Content:    content,
		Images:     images,
	}

	if err := s.db.Create(&msg).Error; err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}
	return &msg, nil
}

func (s *DisputeService) GetMessages(disputeID uuid.UUID) ([]model.DisputeMessage, error) {
	var messages []model.DisputeMessage
	if err := s.db.Where("dispute_id = ?", disputeID).Order("created_at ASC").Find(&messages).Error; err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	return messages, nil
}

func (s *DisputeService) AdminResolve(ctx context.Context, disputeID uuid.UUID, input AdminResolveDisputeInput) (*model.Dispute, error) {
	validDecisions := map[string]bool{
		"full_refund":    true,
		"partial_refund": true,
		"reject":         true,
	}
	if !validDecisions[input.Decision] {
		return nil, fmt.Errorf("invalid decision: %s", input.Decision)
	}

	var dispute model.Dispute
	if err := s.db.Where("id = ?", disputeID).First(&dispute).Error; err != nil {
		return nil, fmt.Errorf("dispute not found")
	}
	if dispute.Status == "resolved" || dispute.Status == "buyer_won" || dispute.Status == "seller_won" {
		return nil, fmt.Errorf("dispute already resolved")
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		updates := map[string]interface{}{
			"admin_decision": input.Decision,
			"admin_note":     input.AdminNote,
			"resolved_by":    input.AdminID,
			"resolved_at":    now,
			"updated_at":     now,
		}

		switch input.Decision {
		case "full_refund":
			updates["status"] = "buyer_won"

			var order model.Order
			if err := tx.Where("id = ?", dispute.OrderID).First(&order).Error; err != nil {
				return err
			}
			if err := tx.Model(&model.Order{}).Where("id = ?", order.ID).Updates(map[string]interface{}{
				"order_status":   "refunded",
				"payment_status": "refunded",
				"dispute_status": "resolved",
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

		case "partial_refund":
			updates["status"] = "buyer_won"
			if err := tx.Model(&model.Order{}).Where("id = ?", dispute.OrderID).Updates(map[string]interface{}{
				"dispute_status": "resolved",
				"updated_at":     now,
			}).Error; err != nil {
				return err
			}

		case "reject":
			updates["status"] = "seller_won"
			if err := tx.Model(&model.Order{}).Where("id = ?", dispute.OrderID).Updates(map[string]interface{}{
				"dispute_status": "none",
				"updated_at":     now,
			}).Error; err != nil {
				return err
			}
		}

		return tx.Model(&dispute).Updates(updates).Error
	})

	if err != nil {
		return nil, fmt.Errorf("failed to resolve dispute: %w", err)
	}

	dispute.Status = "resolved"
	dispute.AdminDecision = input.Decision
	dispute.AdminNote = input.AdminNote
	dispute.ResolvedBy = &input.AdminID
	resolvedAt := time.Now()
	dispute.ResolvedAt = &resolvedAt
	return &dispute, nil
}
