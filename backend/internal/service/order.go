package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/planforever/nexusacg/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OrderService struct {
	db *gorm.DB
}

func NewOrderService(db *gorm.DB) *OrderService {
	return &OrderService{db: db}
}

type CreateOrderInput struct {
	UserID         uuid.UUID `json:"user_id"`
	IdempotencyKey *string   `json:"idempotency_key"`
	Items          []struct {
		ProductID uuid.UUID `json:"product_id"`
		Quantity  int       `json:"quantity"`
	} `json:"items"`
}

func (s *OrderService) Create(ctx context.Context, input CreateOrderInput) (*model.Order, error) {
	// All operations in a single transaction to prevent race conditions
	// between idempotency check and order creation.
	tx := s.db.Begin()

	// Idempotency check within the transaction
	if input.IdempotencyKey != nil && *input.IdempotencyKey != "" {
		var existing model.Order
		if err := tx.Where("idempotency_key = ?", input.IdempotencyKey).First(&existing).Error; err == nil {
			tx.Rollback()
			return &existing, nil
		}
	}

	orderNo := fmt.Sprintf("NAC%s%d", time.Now().Format("20060102150405"), uuid.New().ID()&0xFFFF)

	var totalAmount float64
	var orderItems []model.OrderItem

	for _, item := range input.Items {
		var product model.Product
		if err := tx.Where("id = ? AND status = ?", item.ProductID, "active").First(&product).Error; err != nil {
			return nil, fmt.Errorf("product %s not available", item.ProductID)
		}
		if product.Stock < item.Quantity {
			return nil, fmt.Errorf("product %s insufficient stock", item.ProductID)
		}
		totalAmount += product.Price * float64(item.Quantity)
		orderItems = append(orderItems, model.OrderItem{
			ID:        uuid.New(),
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     product.Price,
		})
	}

	order := model.Order{
		ID:             uuid.New(),
		UserID:         input.UserID,
		OrderNo:        orderNo,
		TotalAmount:    totalAmount,
		PaymentStatus:  "pending",
		OrderStatus:    "pending",
		IdempotencyKey: input.IdempotencyKey,
		Items:          orderItems,
	}

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Deduct stock atomically with WHERE stock >= N to prevent overselling
	for _, item := range orderItems {
		result := tx.Model(&model.Product{}).Where("id = ? AND stock >= ?", item.ProductID, item.Quantity).
			Update("stock", gorm.Expr("stock - ?", item.Quantity))
		if result.Error != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to deduct stock for product %s: %w", item.ProductID, result.Error)
		}
		if result.RowsAffected == 0 {
			tx.Rollback()
			return nil, fmt.Errorf("product %s insufficient stock (race condition)", item.ProductID)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit order: %w", err)
	}
	return &order, nil
}

func (s *OrderService) GetByUser(ctx context.Context, userID uuid.UUID, page, pageSize int, status string) ([]model.Order, int64, error) {
	query := s.db.Model(&model.Order{}).Where("user_id = ?", userID)
	if status != "" {
		query = query.Where("order_status = ?", status)
	}
	var total int64
	query.Count(&total)
	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}
	var orders []model.Order
	if err := query.Preload("Items").Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&orders).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get orders: %w", err)
	}
	return orders, total, nil
}

func (s *OrderService) GetByOrderNo(ctx context.Context, orderNo string) (*model.Order, error) {
	var order model.Order
	if err := s.db.Preload("Items").Where("order_no = ?", orderNo).First(&order).Error; err != nil {
		return nil, fmt.Errorf("order not found")
	}
	return &order, nil
}

func (s *OrderService) Cancel(ctx context.Context, userID uuid.UUID, orderNo string) error {
	var order model.Order
	if err := s.db.Where("order_no = ? AND user_id = ?", orderNo, userID).First(&order).Error; err != nil {
		return fmt.Errorf("order not found")
	}
	if order.OrderStatus != "pending" {
		return fmt.Errorf("only pending orders can be cancelled")
	}

	tx := s.db.Begin()
	if err := tx.Model(&model.Order{}).Where("id = ?", order.ID).Updates(map[string]interface{}{
		"order_status": "cancelled",
		"updated_at":   time.Now(),
	}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to cancel order: %w", err)
	}
	// Restore stock within same transaction
	var items []model.OrderItem
	tx.Where("order_id = ?", order.ID).Find(&items)
	for _, item := range items {
		tx.Model(&model.Product{}).Where("id = ?", item.ProductID).
			Update("stock", gorm.Expr("stock + ?", item.Quantity))
	}
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit cancel: %w", err)
	}
	return nil
}

func (s *OrderService) Refund(ctx context.Context, orderNo string) error {
	var order model.Order
	if err := s.db.Where("order_no = ?", orderNo).First(&order).Error; err != nil {
		return fmt.Errorf("order not found")
	}
	if order.PaymentStatus != "paid" {
		return fmt.Errorf("only paid orders can be refunded")
	}
	if order.OrderStatus == "refunded" {
		return fmt.Errorf("order already refunded")
	}

	tx := s.db.Begin()
	if err := tx.Model(&model.Order{}).Where("id = ?", order.ID).Updates(map[string]interface{}{
		"order_status":   "refunded",
		"payment_status": "refunded",
		"updated_at":     time.Now(),
	}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to refund order: %w", err)
	}
	// Restore stock within same transaction
	var items []model.OrderItem
	tx.Where("order_id = ?", order.ID).Find(&items)
	for _, item := range items {
		tx.Model(&model.Product{}).Where("id = ?", item.ProductID).
			Update("stock", gorm.Expr("stock + ?", item.Quantity))
	}
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit refund: %w", err)
	}
	return nil
}

func (s *OrderService) Pay(ctx context.Context, orderID uuid.UUID, paymentMethod, paymentID string) error {
	// Use a database transaction with row-level locking to prevent race conditions
	// on concurrent payment callbacks.
	tx := s.db.Begin()

	var order model.Order
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", orderID).First(&order).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("order not found")
	}

	// Idempotent: if already paid with the same payment ID, return success
	if order.PaymentStatus == "paid" && order.PaymentID != nil && *order.PaymentID == paymentID {
		tx.Rollback()
		return nil
	}

	if order.PaymentStatus != "pending" {
		tx.Rollback()
		return fmt.Errorf("order already processed")
	}

	now := time.Now()
	if err := tx.Model(&model.Order{}).Where("id = ?", orderID).Updates(map[string]interface{}{
		"payment_status":  "paid",
		"order_status":    "paid",
		"payment_method":  paymentMethod,
		"payment_id":      paymentID,
		"paid_at":         now,
		"updated_at":      now,
	}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to process payment: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit payment: %w", err)
	}
	return nil
}
