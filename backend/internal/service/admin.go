package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/planforever/nexusacg/internal/model"
	"gorm.io/gorm"
)

// AdminService handles admin panel operations: product/post/event audit, order management, stats.
type AdminService struct {
	db *gorm.DB
}

func NewAdminService(db *gorm.DB) *AdminService {
	return &AdminService{db: db}
}

// --- Product Audit ---

func (s *AdminService) ListPendingProducts(ctx context.Context, page, pageSize int) ([]model.Product, int64, error) {
	query := s.db.Model(&model.Product{}).Where("status = ?", "pending_review")
	var total int64
	query.Count(&total)
	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}
	var products []model.Product
	if err := query.Offset(offset).Limit(pageSize).Order("created_at ASC").Find(&products).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list pending products: %w", err)
	}
	return products, total, nil
}

func (s *AdminService) ApproveProduct(ctx context.Context, productID uuid.UUID) error {
	result := s.db.Model(&model.Product{}).Where("id = ?", productID).Updates(map[string]interface{}{
		"status":     "active",
		"updated_at": time.Now(),
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("product not found")
	}
	return nil
}

func (s *AdminService) RejectProduct(ctx context.Context, productID uuid.UUID) error {
	result := s.db.Model(&model.Product{}).Where("id = ?", productID).Updates(map[string]interface{}{
		"status":     "rejected",
		"updated_at": time.Now(),
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("product not found")
	}
	return nil
}

// --- Post Audit ---

func (s *AdminService) ListPendingPosts(ctx context.Context, page, pageSize int) ([]model.Post, int64, error) {
	query := s.db.Model(&model.Post{}).Where("status = ?", "pending_review")
	var total int64
	query.Count(&total)
	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}
	var posts []model.Post
	if err := query.Offset(offset).Limit(pageSize).Order("created_at ASC").Find(&posts).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list pending posts: %w", err)
	}
	return posts, total, nil
}

func (s *AdminService) ApprovePost(ctx context.Context, postID uuid.UUID) error {
	result := s.db.Model(&model.Post{}).Where("id = ?", postID).Updates(map[string]interface{}{
		"status":     "approved",
		"updated_at": time.Now(),
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("post not found")
	}
	return nil
}

func (s *AdminService) RejectPost(ctx context.Context, postID uuid.UUID) error {
	result := s.db.Model(&model.Post{}).Where("id = ?", postID).Updates(map[string]interface{}{
		"status":     "rejected",
		"updated_at": time.Now(),
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("post not found")
	}
	return nil
}

// --- Order Management ---

func (s *AdminService) ListOrders(ctx context.Context, status string, page, pageSize int) ([]model.Order, int64, error) {
	query := s.db.Model(&model.Order{}).Preload("Items")
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
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&orders).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list orders: %w", err)
	}
	return orders, total, nil
}

func (s *AdminService) ProcessRefund(ctx context.Context, orderNo string) error {
	order := &model.Order{}
	if err := s.db.Where("order_no = ?", orderNo).First(order).Error; err != nil {
		return fmt.Errorf("order not found")
	}
	if order.PaymentStatus != "paid" {
		return fmt.Errorf("only paid orders can be refunded")
	}

	tx := s.db.Begin()
	if err := tx.Model(&model.Order{}).Where("id = ?", order.ID).Updates(map[string]interface{}{
		"order_status":   "refunded",
		"payment_status": "refunded",
		"updated_at":     time.Now(),
	}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to refund: %w", err)
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

// --- Stats Dashboard ---

type DashboardStats struct {
	TotalUsers     int64 `json:"total_users"`
	TotalProducts  int64 `json:"total_products"`
	TotalOrders    int64 `json:"total_orders"`
	TotalPosts     int64 `json:"total_posts"`
	PendingProducts int64 `json:"pending_products"`
	PendingPosts   int64 `json:"pending_posts"`
	PendingOrders  int64 `json:"pending_orders"`
	PaidOrders     int64 `json:"paid_orders"`
	TotalRevenue   float64 `json:"total_revenue"`
}

func (s *AdminService) GetDashboardStats(ctx context.Context) (*DashboardStats, error) {
	stats := &DashboardStats{}
	s.db.Model(&model.User{}).Where("status = ?", "active").Count(&stats.TotalUsers)
	s.db.Model(&model.Product{}).Where("status = ?", "active").Count(&stats.TotalProducts)
	s.db.Model(&model.Product{}).Where("status = ?", "pending_review").Count(&stats.PendingProducts)
	s.db.Model(&model.Post{}).Where("status = ?", "pending_review").Count(&stats.PendingPosts)
	s.db.Model(&model.Post{}).Where("status = ?", "approved").Count(&stats.TotalPosts)
	s.db.Model(&model.Order{}).Where("order_status = ?", "pending").Count(&stats.PendingOrders)
	s.db.Model(&model.Order{}).Where("order_status = ?", "paid").Count(&stats.PaidOrders)
	s.db.Model(&model.Order{}).Count(&stats.TotalOrders)
	s.db.Model(&model.Order{}).Where("payment_status = ?", "paid").Select("COALESCE(SUM(total_amount), 0)").Scan(&stats.TotalRevenue)
	return stats, nil
}

// --- Admin User Management ---

func (s *AdminService) ListUsers(ctx context.Context, page, pageSize int) ([]model.User, int64, error) {
	query := s.db.Model(&model.User{})
	var total int64
	query.Count(&total)
	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}
	var users []model.User
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}
	return users, total, nil
}

func (s *AdminService) BanUser(ctx context.Context, userID uuid.UUID) error {
	result := s.db.Model(&model.User{}).Where("id = ?", userID).Update("status", "banned")
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}
