package service

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/planforever/nexusacg/internal/model"
	"gorm.io/gorm"
)

// PaymentChannelRelease defines the interface for calling payment provider profit-sharing APIs.
type PaymentChannelRelease interface {
	ReleaseWechatFunds(order *model.Order, platformFee, sellerAmount float64) error
	ReleaseAlipayFunds(order *model.Order, platformFee, sellerAmount float64) error
}

type ProfitShareService struct {
	db                 *gorm.DB
	platformFeePercent float64
	releaser           PaymentChannelRelease
}

func NewProfitShareService(db *gorm.DB, platformFeePercent float64, releaser PaymentChannelRelease) *ProfitShareService {
	return &ProfitShareService{db: db, platformFeePercent: platformFeePercent, releaser: releaser}
}

// ShipOrder marks a paid order as shipped. Only the seller of items in the order can ship.
func (s *ProfitShareService) ShipOrder(orderNo, sellerID string) error {
	var order model.Order
	if err := s.db.Preload("Items").Where("order_no = ?", orderNo).First(&order).Error; err != nil {
		return fmt.Errorf("order not found")
	}
	if order.OrderStatus != "paid" {
		return fmt.Errorf("only paid orders can be shipped")
	}

	// Verify user is the seller of at least one item in the order
	productIDs := make([]uuid.UUID, len(order.Items))
	for i, item := range order.Items {
		productIDs[i] = item.ProductID
	}
	var productCount int64
	s.db.Model(&model.Product{}).Where("id IN ? AND seller_id = ?", productIDs, sellerID).Count(&productCount)
	if productCount == 0 {
		return fmt.Errorf("not authorized: you are not the seller of items in this order")
	}

	now := time.Now()
	if err := s.db.Model(&model.Order{}).Where("id = ?", order.ID).Updates(map[string]interface{}{
		"order_status": "shipped",
		"shipped_at":   now,
		"updated_at":   now,
	}).Error; err != nil {
		return fmt.Errorf("failed to ship order: %w", err)
	}
	return nil
}

// ConfirmReceipt marks a shipped order as completed and triggers profit sharing.
// Only the buyer (order owner) can confirm receipt.
func (s *ProfitShareService) ConfirmReceipt(orderNo, buyerID string) error {
	var order model.Order
	if err := s.db.Where("order_no = ? AND user_id = ?", orderNo, buyerID).First(&order).Error; err != nil {
		return fmt.Errorf("order not found")
	}
	if order.OrderStatus != "shipped" {
		return fmt.Errorf("only shipped orders can be confirmed")
	}

	return s.releaseFunds(&order)
}

// AutoReleaseOrders finds shipped orders past the timeout and auto-completes them.
func (s *ProfitShareService) AutoReleaseOrders(autoReleaseDays int) (int64, error) {
	if autoReleaseDays <= 0 {
		return 0, nil
	}
	cutoff := time.Now().AddDate(0, 0, -autoReleaseDays)

	var orders []model.Order
	if err := s.db.Where("order_status = ? AND shipped_at < ?", "shipped", cutoff).Find(&orders).Error; err != nil {
		return 0, err
	}
	if len(orders) == 0 {
		return 0, nil
	}

	var released int64
	for i := range orders {
		if err := s.releaseFunds(&orders[i]); err != nil {
			log.Printf("auto-release failed for order %s: %v", orders[i].OrderNo, err)
			continue
		}
		released++
	}
	return released, nil
}

// releaseFunds creates a profit share record, calls payment channel, and completes the order.
func (s *ProfitShareService) releaseFunds(order *model.Order) error {
	platformFee := order.TotalAmount * s.platformFeePercent
	sellerAmount := order.TotalAmount - platformFee

	record := model.ProfitShareRecord{
		OrderID:       order.ID,
		TotalAmount:   order.TotalAmount,
		PlatformFee:   platformFee,
		SellerAmount:  sellerAmount,
		PaymentMethod: "unknown",
		Status:        "pending",
	}
	if order.PaymentMethod != nil {
		record.PaymentMethod = *order.PaymentMethod
	}

	// Call payment channel profit-sharing (if configured)
	if err := s.callPaymentChannelRelease(&record, order); err != nil {
		log.Printf("payment channel release failed for order %s: %v (continuing with local record)", order.OrderNo, err)
		record.Status = "failed"
	} else {
		record.Status = "released"
	}

	now := time.Now()
	record.ReleasedAt = &now

	tx := s.db.Begin()
	if err := tx.Create(&record).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("create profit share record: %w", err)
	}

	if err := tx.Model(&model.Order{}).Where("id = ?", order.ID).Updates(map[string]interface{}{
		"order_status": "completed",
		"completed_at": now,
		"updated_at":   now,
	}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("complete order: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("commit profit share: %w", err)
	}

	log.Printf("profit share completed for order %s: platform_fee=%.2f, seller_amount=%.2f",
		order.OrderNo, platformFee, sellerAmount)
	return nil
}

func (s *ProfitShareService) callPaymentChannelRelease(record *model.ProfitShareRecord, order *model.Order) error {
	if s.releaser == nil {
		log.Printf("profit share: payment channel release skipped (not configured) for order %s", order.OrderNo)
		return nil
	}

	platformFee := record.PlatformFee
	sellerAmount := record.SellerAmount

	switch record.PaymentMethod {
	case "wechat":
		return s.releaser.ReleaseWechatFunds(order, platformFee, sellerAmount)
	case "alipay":
		return s.releaser.ReleaseAlipayFunds(order, platformFee, sellerAmount)
	default:
		log.Printf("profit share: unknown payment method %s for order %s", record.PaymentMethod, order.OrderNo)
		return nil
	}
}
