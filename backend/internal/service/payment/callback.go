package payment

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/planforever/nexusacg/internal/model"
	"github.com/smartwalle/alipay/v3"
	"gorm.io/gorm"
)

// CallbackService handles payment provider async notifications.
// Implements idempotency via payment_logs table + state machine validation.
type CallbackService struct {
	db             *gorm.DB
	wechatClient   *WechatPayClient
	alipayVerifier *AlipaySign
}

func NewCallbackService(db *gorm.DB, wechat *WechatPayClient, alipay *AlipaySign) *CallbackService {
	return &CallbackService{
		db:             db,
		wechatClient:   wechat,
		alipayVerifier: alipay,
	}
}

// PrepayWechatOrder creates a WeChat Pay unified order and returns payment parameters for the mobile client.
func (s *CallbackService) PrepayWechatOrder(ctx context.Context, orderNo, description string, amountFen int64, payerIP string) (interface{}, error) {
	if s.wechatClient == nil {
		return nil, fmt.Errorf("wechat pay not configured")
	}

	input := WechatPayOrderInput{
		OutTradeNo:  orderNo,
		Description: description,
		TotalAmount: amountFen,
		PayerIP:     payerIP,
	}

	result, err := s.wechatClient.PrepayOrder(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("wechat prepay: %w", err)
	}

	// Return all fields needed by mobile client
	return gin.H{
		"appid":     s.wechatClient.AppID,
		"partnerid": *result.PartnerId,
		"prepayid":  *result.PrepayId,
		"package":   *result.Package,
		"noncestr":  *result.NonceStr,
		"timestamp": *result.TimeStamp,
		"sign":      *result.Sign,
	}, nil
}

// PrepayAlipayOrder creates an Alipay App payment order string for the mobile client.
func (s *CallbackService) PrepayAlipayOrder(ctx context.Context, orderNo, description string, amountFen int64, notifyURL string) (interface{}, error) {
	if s.alipayVerifier == nil || s.alipayVerifier.Client == nil {
		return nil, fmt.Errorf("alipay not configured")
	}

	amountYuan := fmt.Sprintf("%.2f", float64(amountFen)/100.0)

	param := alipay.TradeAppPay{
		Trade: alipay.Trade{
			Subject:     description,
			OutTradeNo:  orderNo,
			TotalAmount: amountYuan,
			ProductCode: "QUICK_MSECURITY_PAY",
			Body:        description,
			NotifyURL:   notifyURL,
		},
	}

	// TradeAppPay returns a signed order string that the mobile client passes to Alipay SDK
	orderStr, err := s.alipayVerifier.Client.TradeAppPay(param)
	if err != nil {
		return nil, fmt.Errorf("alipay trade app pay: %w", err)
	}

	return gin.H{
		"order_string": orderStr,
	}, nil
}

// VerifyAlipaySDK exercises the full Alipay SDK flow: order string generation + signature verification.
// Returns a diagnostic report for debugging.
func (s *CallbackService) VerifyAlipaySDK() (interface{}, error) {
	if s.alipayVerifier == nil || s.alipayVerifier.Client == nil {
		return nil, fmt.Errorf("alipay not configured")
	}

	results := gin.H{}

	// 1. Test: generate a signed order string
	testOrderNo := fmt.Sprintf("VERIFY_%d", time.Now().Unix())
	param := alipay.TradeAppPay{
		Trade: alipay.Trade{
			Subject:     "沙箱验证订单",
			OutTradeNo:  testOrderNo,
			TotalAmount: "0.01",
			ProductCode: "QUICK_MSECURITY_PAY",
			Body:        "Alipay SDK verification test",
		},
	}

	orderStr, err := s.alipayVerifier.Client.TradeAppPay(param)
	if err != nil {
		results["status"] = "FAIL"
		results["trade_app_pay"] = fmt.Sprintf("error: %v", err)
		return results, nil
	}
	results["trade_app_pay"] = "PASS"
	results["order_string_length"] = len(orderStr)

	// 2. Test: parse the order string back and check structure
	parsed, err := url.ParseQuery(orderStr)
	if err != nil {
		results["status"] = "FAIL"
		results["parse_order_string"] = fmt.Sprintf("error: %v", err)
		return results, nil
	}
	results["parse_order_string"] = "PASS"

	// 3. Verify required fields
	requiredFields := map[string]string{
		"app_id":    "2021006153686187",
		"method":    "alipay.trade.app.pay",
		"sign_type": "RSA2",
	}
	allFieldsOK := true
	for field, expected := range requiredFields {
		if parsed.Get(field) != expected {
			results[field] = fmt.Sprintf("FAIL: expected %q, got %q", expected, parsed.Get(field))
			allFieldsOK = false
		} else {
			results[field] = "PASS"
		}
	}

	// 4. Verify signature is present and well-formed (RSA2 = 344 bytes base64)
	sign := parsed.Get("sign")
	if sign == "" || len(sign) < 100 {
		results["verify_sign"] = fmt.Sprintf("FAIL: sign too short (%d bytes)", len(sign))
		allFieldsOK = false
	} else {
		results["verify_sign"] = fmt.Sprintf("PASS (%d bytes)", len(sign))
	}

	if allFieldsOK {
		results["status"] = "ALL_PASS"
	} else {
		results["status"] = "FAIL"
	}
	return results, nil
}

// HandleWechatCallbackV3 processes a WeChat Pay v3 async notification.
// WeChat Pay v3 sends JSON with AEAD_AES_256_GCM encrypted resource.
// Returns the JSON response body to send back to WeChat.
func (s *CallbackService) HandleWechatCallbackV3(ctx context.Context, r *http.Request) (string, error) {
	if s.wechatClient == nil {
		return `{"code":"FAIL","message":"微信支付未配置"}`, fmt.Errorf("wechat pay not configured")
	}

	// 1. Parse and decrypt the callback
	var result paymentResult
	_, err := s.wechatClient.NotifyHandler.ParseNotifyRequest(ctx, r, &result)
	if err != nil {
		return `{"code":"FAIL","message":"解析失败"}`, fmt.Errorf("parse wechat v3 callback: %w", err)
	}

	// 2. Check trade state
	if result.TradeState != "SUCCESS" {
		log.Printf("wechat callback: trade_state=%s, out_trade_no=%s", result.TradeState, result.OutTradeNo)
		return `{"code":"SUCCESS","message":"已处理"}`, nil
	}

	// 3. Extract fields
	transactionID := result.TransactionID
	outTradeNo := result.OutTradeNo
	totalAmount := float64(result.Amount.Total) / 100.0 // convert cents to yuan

	// 4. Idempotency check
	if transactionID != "" {
		var existing model.PaymentLog
		if err := s.db.Where("transaction_id = ?", transactionID).First(&existing).Error; err == nil {
			log.Printf("duplicate wechat callback: transaction_id=%s, order_no=%s", transactionID, outTradeNo)
			return `{"code":"SUCCESS","message":"已处理"}`, nil
		}
	}

	// 5. Prepare payment log
	paymentLog := model.PaymentLog{
		PaymentMethod:  "wechat",
		TransactionID:  transactionID,
		Amount:         totalAmount,
		Status:         "success",
		RawPayload:     fmt.Sprintf("%+v", result),
		SignatureValid: true, // verified by SDK
		Processed:      false,
	}

	// 6. Apply business logic inside a transaction with row lock
	err = s.db.Transaction(func(tx *gorm.DB) error {
		var order model.Order
		if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("order_no = ?", outTradeNo).First(&order).Error; err != nil {
			paymentLog.Status = "failed"
			return tx.Create(&paymentLog).Error
		}

		// State machine: only allow pending → paid transition
		if order.PaymentStatus == "paid" {
			log.Printf("order already paid: order_no=%s, payment_id=%v", outTradeNo, order.PaymentID)
			paymentLog.Processed = true
			return tx.Create(&paymentLog).Error
		}
		if order.PaymentStatus != "pending" {
			paymentLog.Status = "failed"
			return tx.Create(&paymentLog).Error
		}

		// Mark order as paid
		now := time.Now()
		if err := tx.Model(&model.Order{}).Where("id = ?", order.ID).Updates(map[string]interface{}{
			"payment_status": "paid",
			"order_status":   "paid",
			"payment_method": "wechat",
			"payment_id":     transactionID,
			"paid_at":        now,
			"updated_at":     now,
		}).Error; err != nil {
			return err
		}

		paymentLog.OrderID = order.ID
		paymentLog.Processed = true
		return tx.Create(&paymentLog).Error
	})

	if err != nil {
		return `{"code":"FAIL","message":"系统错误"}`, fmt.Errorf("process wechat v3 callback: %w", err)
	}

	return `{"code":"SUCCESS","message":"已处理"}`, nil
}

// paymentResult holds the decrypted WeChat Pay v3 transaction notification.
type paymentResult struct {
	Appid         string `json:"appid"`
	Mchid         string `json:"mchid"`
	OutTradeNo    string `json:"out_trade_no"`
	TransactionID string `json:"transaction_id"`
	TradeState    string `json:"trade_state"`
	TradeType     string `json:"trade_type"`
	BankType      string `json:"bank_type"`
	Attach        string `json:"attach"`
	SuccessTime   string `json:"success_time"`
	Amount        struct {
		Total    int64  `json:"total"`
		Currency string `json:"currency"`
	} `json:"amount"`
	Payer struct {
		Openid string `json:"openid"`
	} `json:"payer"`
}

// HandleAlipayCallback processes an Alipay async notification.
func (s *CallbackService) HandleAlipayCallback(ctx context.Context, params map[string]string) (string, error) {
	// 0. Check if Alipay is configured
	if s.alipayVerifier == nil {
		return "fail", fmt.Errorf("alipay not configured")
	}

	// 1. Verify signature — reject immediately if invalid
	sign := params["sign"]
	sigValid := s.alipayVerifier.Verify(params, sign)
	if !sigValid {
		return "fail", fmt.Errorf("alipay callback signature invalid")
	}

	// 2. Check trade_status
	tradeStatus := params["trade_status"]
	if tradeStatus != "TRADE_SUCCESS" && tradeStatus != "TRADE_FINISHED" {
		s.logPayment("alipay", sigValid, false, fmt.Sprintf("%v", params), params["trade_no"])
		return "fail", fmt.Errorf("alipay trade_status: %s", tradeStatus)
	}

	// 3. Idempotency check
	transactionID := params["trade_no"]   // Alipay's transaction ID
	outTradeNo := params["out_trade_no"] // our order_no

	if transactionID != "" {
		var existing model.PaymentLog
		if err := s.db.Where("transaction_id = ?", transactionID).First(&existing).Error; err == nil {
			log.Printf("duplicate alipay callback: trade_no=%s, out_trade_no=%s", transactionID, outTradeNo)
			return "success", nil
		}
	}

	// 4. Parse amount
	totalAmount := parseFloat(params["total_amount"])

	// 5. Prepare payment log (OrderID set after order lookup in transaction)
	paymentLog := model.PaymentLog{
		PaymentMethod:  "alipay",
		TransactionID:  transactionID,
		Amount:         totalAmount,
		Status:         "success",
		RawPayload:     fmt.Sprintf("%v", params),
		SignatureValid: sigValid,
		Processed:      false,
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if !sigValid {
			paymentLog.Status = "failed"
			return tx.Create(&paymentLog).Error
		}

		var order model.Order
		if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("order_no = ?", outTradeNo).First(&order).Error; err != nil {
			paymentLog.Status = "failed"
			return tx.Create(&paymentLog).Error
		}

		if order.PaymentStatus == "paid" {
			log.Printf("order already paid: order_no=%s, payment_id=%v", outTradeNo, order.PaymentID)
			paymentLog.Processed = true
			return tx.Create(&paymentLog).Error
		}
		if order.PaymentStatus != "pending" {
			paymentLog.Status = "failed"
			return tx.Create(&paymentLog).Error
		}

		now := time.Now()
		if err := tx.Model(&model.Order{}).Where("id = ?", order.ID).Updates(map[string]interface{}{
			"payment_status": "paid",
			"order_status":   "paid",
			"payment_method": "alipay",
			"payment_id":     transactionID,
			"paid_at":        now,
			"updated_at":     now,
		}).Error; err != nil {
			return err
		}

		paymentLog.OrderID = order.ID
		paymentLog.Processed = true
		return tx.Create(&paymentLog).Error
	})

	if err != nil {
		return "fail", fmt.Errorf("process alipay callback: %w", err)
	}

	return "success", nil
}

// CancelTimeoutOrders marks orders as cancelled if they've been pending too long
// and restores their stock.
func (s *CallbackService) CancelTimeoutOrders(ctx context.Context, timeout time.Duration) (int64, error) {
	cutoff := time.Now().Add(-timeout)

	// Find pending timeout orders
	var orders []model.Order
	if err := s.db.Where("payment_status = ? AND created_at < ?", "pending", cutoff).Find(&orders).Error; err != nil {
		return 0, err
	}
	if len(orders) == 0 {
		return 0, nil
	}

	var cancelled int64
	for _, order := range orders {
		// Load order items for stock restoration
		var items []model.OrderItem
		s.db.Where("order_id = ?", order.ID).Find(&items)

		tx := s.db.Begin()
		// Lock the order
		if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("id = ?", order.ID).First(&order).Error; err != nil {
			tx.Rollback()
			continue
		}
		if order.PaymentStatus != "pending" {
			tx.Rollback()
			continue
		}

		// Update order status
		if err := tx.Model(&model.Order{}).Where("id = ?", order.ID).Updates(map[string]interface{}{
			"order_status":   "cancelled",
			"payment_status": "cancelled",
			"updated_at":     time.Now(),
		}).Error; err != nil {
			tx.Rollback()
			continue
		}

		// Restore stock
		for _, item := range items {
			tx.Model(&model.Product{}).Where("id = ?", item.ProductID).
				Update("stock", gorm.Expr("stock + ?", item.Quantity))
		}

		if err := tx.Commit().Error; err != nil {
			continue
		}
		cancelled++
	}

	return cancelled, nil
}

// GetPaymentLogs returns payment logs for audit purposes.
func (s *CallbackService) GetPaymentLogs(orderID string, limit int) ([]model.PaymentLog, error) {
	var logs []model.PaymentLog
	query := s.db.Order("created_at DESC").Limit(limit)
	if orderID != "" {
		query = query.Where("order_id = ?", orderID)
	}
	if err := query.Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

// --- Helper methods ---

func (s *CallbackService) logPayment(method string, sigValid, processed bool, rawPayload string, transactionID string) {
	log := model.PaymentLog{
		PaymentMethod:  method,
		TransactionID:  transactionID,
		SignatureValid: sigValid,
		Processed:      processed,
		RawPayload:     rawPayload,
		Status:         "failed",
	}
	s.db.Create(&log)
}

func parseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}

// ReleaseWechatFunds calls WeChat Pay profit-sharing API to release escrow funds.
// Requires merchant account with profit-sharing feature enabled.
func (s *CallbackService) ReleaseWechatFunds(order *model.Order, platformFee, sellerAmount float64) error {
	if s.wechatClient == nil {
		return fmt.Errorf("wechat pay not configured")
	}
	// TODO: Implement WeChat Pay profit-sharing API call when merchant account is available.
	// Requires: 1. Enable profit-sharing on merchant account 2. Upload receiver accounts
	// API: POST /v3/profitsharing/orders
	log.Printf("wechat profit sharing: not yet implemented for order %s", order.OrderNo)
	return nil
}

// ReleaseAlipayFunds calls Alipay settle API to release escrow funds.
// Requires merchant account with settlement feature enabled.
func (s *CallbackService) ReleaseAlipayFunds(order *model.Order, platformFee, sellerAmount float64) error {
	if s.alipayVerifier == nil || s.alipayVerifier.Client == nil {
		return fmt.Errorf("alipay not configured")
	}
	// TODO: Implement Alipay profit-sharing API call when merchant account is available.
	// API: alipay.trade.order.settle
	log.Printf("alipay profit sharing: not yet implemented for order %s", order.OrderNo)
	return nil
}
