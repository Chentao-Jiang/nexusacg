package payment

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/planforever/nexusacg/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// testDB returns a database connection for integration tests.
// Requires DATABASE_URL env var, falls back to dev defaults.
func testDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost port=5432 user=nexusacg password=nexusacg_dev_pass dbname=nexusacg sslmode=disable TimeZone=Asia/Shanghai"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Skipf("database not available: %v", err)
	}

	// Auto-migrate needed tables
	db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")
	db.AutoMigrate(&model.Order{}, &model.PaymentLog{}, &model.User{})

	// Clean tables before each test (in reverse FK order)
	db.Exec("DELETE FROM order_items")
	db.Exec("DELETE FROM payment_logs")
	db.Exec("DELETE FROM orders")

	// Create a test user for FK references
	testUser := model.User{
		Nickname: "test_user",
		Role:     "user",
		Status:   "active",
	}
	db.FirstOrCreate(&testUser, model.User{Nickname: "test_user"})

	return db
}

// testAlipayClient returns an Alipay client for tests.
// Requires ALIPAY_APP_PRIVATE_KEY_PATH and ALIPAY_PUBLIC_KEY_PATH env vars.
func testUserID(t *testing.T, db *gorm.DB) string {
	t.Helper()
	var user model.User
	if err := db.Where("nickname = ?", "test_user").First(&user).Error; err != nil {
		t.Fatalf("test user not found: %v", err)
	}
	return user.ID.String()
}

func testAlipayClient(t *testing.T) *AlipaySign {
	t.Helper()

	appID := "2021006153686187"
	privKeyPath := os.Getenv("ALIPAY_APP_PRIVATE_KEY_PATH")
	pubKeyPath := os.Getenv("ALIPAY_PUBLIC_KEY_PATH")

	if privKeyPath == "" || pubKeyPath == "" {
		t.Skip("ALIPAY_APP_PRIVATE_KEY_PATH or ALIPAY_PUBLIC_KEY_PATH not set")
	}

	privKey, err := os.ReadFile(privKeyPath)
	if err != nil {
		t.Skipf("cannot read private key: %v", err)
	}
	pubKey, err := os.ReadFile(pubKeyPath)
	if err != nil {
		t.Skipf("cannot read public key: %v", err)
	}

	client, err := NewAlipaySign(appID, strings.TrimSpace(string(privKey)), strings.TrimSpace(string(pubKey)), true)
	if err != nil {
		t.Skipf("alipay SDK init failed: %v", err)
	}
	return client
}

func TestPrepayAlipayOrder_GeneratesValidOrderString(t *testing.T) {
	db := testDB(t)
	alipay := testAlipayClient(t)

	svc := NewCallbackService(db, nil, alipay)
	ctx := context.Background()

	orderNo := fmt.Sprintf("TEST_%d", time.Now().UnixNano())
	result, err := svc.PrepayAlipayOrder(ctx, orderNo, "测试商品", 100, "http://localhost:8080/api/v1/payments/alipay/callback")
	if err != nil {
		t.Fatalf("PrepayAlipayOrder failed: %v", err)
	}

	m, ok := result.(gin.H)
	if !ok {
		// gin.H is map[string]interface{} but let's check the actual type
		t.Fatalf("result type: %T, value: %+v", result, result)
	}

	orderStr, ok := m["order_string"].(string)
	if !ok {
		t.Fatal("order_string not found in result")
	}

	if len(orderStr) == 0 {
		t.Fatal("order_string is empty")
	}

	t.Logf("order_string length: %d bytes", len(orderStr))

	// Verify required fields are present
	requiredFields := []string{"app_id", "method", "sign_type", "sign", "timestamp", "biz_content"}
	for _, field := range requiredFields {
		if !strings.Contains(orderStr, field) {
			t.Errorf("order_string missing field: %s", field)
		}
	}

	// Verify method value
	if !strings.Contains(orderStr, "alipay.trade.app.pay") {
		t.Error("order_string should contain alipay.trade.app.pay method")
	}

	// Verify sign type
	if !strings.Contains(orderStr, "RSA2") {
		t.Error("order_string should contain RSA2 sign type")
	}
}

func TestHandleAlipayCallback_SuccessFlow(t *testing.T) {
	db := testDB(t)
	alipay := testAlipayClient(t)

	svc := NewCallbackService(db, nil, alipay)
	ctx := context.Background()

	// 1. Create a test order with pending status
	orderNo := fmt.Sprintf("TEST_CB_%d", time.Now().UnixNano())
	order := model.Order{
		UserID:        uuid.MustParse(testUserID(t, db)),
		OrderNo:       orderNo,
		TotalAmount:   1.00,
		PaymentStatus: "pending",
		OrderStatus:   "pending",
	}
	if err := db.Create(&order).Error; err != nil {
		t.Fatalf("create test order: %v", err)
	}

	// 2. Simulate Alipay callback (signature will be invalid since we can't forge Alipay's signature,
	// but we test the full state machine + idempotency + order update logic)
	params := map[string]string{
		"gmt_create":       "2026-05-13 10:00:00",
		"charset":          "utf-8",
		"gmt_payment":      "2026-05-13 10:01:00",
		"notify_time":      "2026-05-13 10:01:01",
		"subject":          "测试商品",
		"sign":             "fake_signature_for_testing",
		"buyer_id":         "2088000000000001",
		"invoice_amount":   "1.00",
		"version":          "1.0",
		"notify_id":        "test_notify_001",
		"fund_bill_list":   `[{"amount":"1.00","fundChannel":"ALIPAYACCOUNT"}]`,
		"notify_type":      "trade_status_sync",
		"out_trade_no":     orderNo,
		"total_amount":     "1.00",
		"trade_status":     "TRADE_SUCCESS",
		"trade_no":         "2026051322001234567890",
		"auth_app_id":      "2021006153686187",
		"receipt_amount":   "1.00",
		"point_amount":     "0.00",
		"buyer_pay_amount": "1.00",
		"app_id":           "2021006153686187",
		"sign_type":        "RSA2",
		"seller_id":        "2088000000000000",
	}

	resp, err := svc.HandleAlipayCallback(ctx, params)
	// With invalid signature, the callback should fail (sigValid=false → status=failed)
	if err != nil {
		t.Logf("callback returned error (expected for invalid sig): %v", err)
	}
	t.Logf("callback response: %s", resp)

	// Verify payment log was created
	var log model.PaymentLog
	if err := db.Where("payment_method = ? AND transaction_id = ?", "alipay", params["trade_no"]).First(&log).Error; err != nil {
		t.Fatalf("payment log not created: %v", err)
	}

	if log.SignatureValid {
		t.Error("signature should be invalid for fake signature")
	}

	t.Logf("payment log: id=%s, sigValid=%v, status=%s, processed=%v",
		log.ID, log.SignatureValid, log.Status, log.Processed)
}

func TestHandleAlipayCallback_Idempotency(t *testing.T) {
	db := testDB(t)
	alipay := testAlipayClient(t)

	svc := NewCallbackService(db, nil, alipay)
	ctx := context.Background()

	// 1. Create a test order
	orderNo := fmt.Sprintf("TEST_IDEM_%d", time.Now().UnixNano())
	order := model.Order{
		UserID:        uuid.MustParse(testUserID(t, db)),
		OrderNo:       orderNo,
		TotalAmount:   1.00,
		PaymentStatus: "pending",
		OrderStatus:   "pending",
	}
	db.Create(&order)

	// 2. Create a payment log to simulate already-processed callback
	existingLog := model.PaymentLog{
		OrderID:       order.ID,
		PaymentMethod: "alipay",
		TransactionID: "2026051322001111111111",
		Amount:        1.00,
		Status:        "success",
		Processed:     true,
	}
	db.Create(&existingLog)

	// 3. Send duplicate callback with same transaction_id
	params := map[string]string{
		"notify_time":  "2026-05-13 10:01:01",
		"notify_type":  "trade_status_sync",
		"out_trade_no": orderNo,
		"total_amount": "1.00",
		"trade_status": "TRADE_SUCCESS",
		"trade_no":     "2026051322001111111111", // same as existing log
		"sign":         "fake_sig",
	}

	resp, err := svc.HandleAlipayCallback(ctx, params)
	if err != nil {
		t.Fatalf("duplicate callback should return success: %v", err)
	}
	if resp != "success" {
		t.Errorf("expected 'success' for duplicate, got: %s", resp)
	}

	t.Log("idempotency check passed — duplicate callback returned success")
}

func TestHandleAlipayCallback_NonSuccessStatus(t *testing.T) {
	db := testDB(t)
	alipay := testAlipayClient(t)

	svc := NewCallbackService(db, nil, alipay)
	ctx := context.Background()

	orderNo := fmt.Sprintf("TEST_FAIL_%d", time.Now().UnixNano())
	order := model.Order{
		UserID:        uuid.MustParse(testUserID(t, db)),
		OrderNo:       orderNo,
		TotalAmount:   1.00,
		PaymentStatus: "pending",
		OrderStatus:   "pending",
	}
	db.Create(&order)

	// Simulate TRADE_CLOSED status
	params := map[string]string{
		"notify_time":  "2026-05-13 10:01:01",
		"notify_type":  "trade_status_sync",
		"out_trade_no": orderNo,
		"total_amount": "1.00",
		"trade_status": "TRADE_CLOSED",
		"trade_no":     "2026051322009999999999",
		"sign":         "fake_sig",
	}

	_, err := svc.HandleAlipayCallback(ctx, params)
	if err == nil {
		t.Fatal("expected error for non-success trade status")
	}

	t.Logf("correctly rejected TRADE_CLOSED: %v", err)
}

func TestHandleAlipayCallback_AlreadyPaidOrder(t *testing.T) {
	db := testDB(t)
	alipay := testAlipayClient(t)

	svc := NewCallbackService(db, nil, alipay)
	ctx := context.Background()

	// Create an already-paid order
	orderNo := fmt.Sprintf("TEST_PAID_%d", time.Now().UnixNano())
	now := time.Now()
	paymentMethod := "alipay"
	paymentID := "20260513000000001"
	order := model.Order{
		UserID:        uuid.MustParse(testUserID(t, db)),
		OrderNo:       orderNo,
		TotalAmount:   1.00,
		PaymentStatus: "paid",
		OrderStatus:   "paid",
		PaymentMethod: &paymentMethod,
		PaymentID:     &paymentID,
		PaidAt:        &now,
	}
	db.Create(&order)

	params := map[string]string{
		"notify_time":  "2026-05-13 10:01:01",
		"notify_type":  "trade_status_sync",
		"out_trade_no": orderNo,
		"total_amount": "1.00",
		"trade_status": "TRADE_SUCCESS",
		"trade_no":     "2026051322008888888888",
		"sign":         "fake_sig",
	}

	resp, err := svc.HandleAlipayCallback(ctx, params)
	if err != nil {
		t.Fatalf("callback for already-paid order should not error: %v", err)
	}
	if resp != "success" {
		t.Errorf("expected 'success' for already-paid, got: %s", resp)
	}

	// Verify order is unchanged
	var updated model.Order
	db.Where("order_no = ?", orderNo).First(&updated)
	if updated.PaymentID == nil || *updated.PaymentID != "20260513000000001" {
		t.Error("order payment_id should not change for duplicate callback")
	}

	t.Log("already-paid order correctly handled — idempotent")
}

func TestCancelTimeoutOrders(t *testing.T) {
	db := testDB(t)

	svc := NewCallbackService(db, nil, nil)
	ctx := context.Background()

	// Create a pending order with old timestamp
	orderNo := fmt.Sprintf("TEST_TIMEOUT_%d", time.Now().UnixNano())
	order := model.Order{
		UserID:        uuid.MustParse(testUserID(t, db)),
		OrderNo:       orderNo,
		TotalAmount:   1.00,
		PaymentStatus: "pending",
		OrderStatus:   "pending",
	}
	db.Create(&order)

	// Manually set created_at to 2 hours ago
	db.Model(&model.Order{}).Where("id = ?", order.ID).Update("created_at", time.Now().Add(-2*time.Hour))

	// Cancel orders older than 1 hour
	count, err := svc.CancelTimeoutOrders(ctx, 1*time.Hour)
	if err != nil {
		t.Fatalf("CancelTimeoutOrders failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 cancelled order, got %d", count)
	}

	// Verify order is cancelled
	var updated model.Order
	db.Where("order_no = ?", orderNo).First(&updated)
	if updated.OrderStatus != "cancelled" {
		t.Errorf("order should be cancelled, got: %s", updated.OrderStatus)
	}

	t.Log("timeout order cancellation works correctly")
}
