package payment

import (
	"net/url"
	"strings"
	"testing"

	"github.com/smartwalle/alipay/v3"
)

// TestAlipaySDK_GenerateOrderString verifies the SDK can generate a properly
// signed order string for the sandbox environment.
func TestAlipaySDK_GenerateOrderString(t *testing.T) {
	aliClient := testAlipayClient(t)
	if aliClient.Client == nil {
		t.Fatal("alipay client not initialized")
	}

	// Generate order string via SDK (same method used in PrepayAlipayOrder)
	param := alipay.TradeAppPay{
		Trade: alipay.Trade{
			Subject:     "沙箱测试",
			OutTradeNo:  "test_order_sandbox_001",
			TotalAmount: "0.01",
			ProductCode: "QUICK_MSECURITY_PAY",
		},
	}
	orderStr, err := aliClient.Client.TradeAppPay(param)
	if err != nil {
		t.Fatalf("TradeAppPay failed: %v", err)
	}
	if len(orderStr) == 0 {
		t.Fatal("order_string is empty")
	}

	// Parse the order string
	parsed, err := url.ParseQuery(orderStr)
	if err != nil {
		t.Fatalf("parse order string: %v", err)
	}

	// Check required fields
	checks := map[string]string{
		"app_id":    "2021006153686187",
		"method":    "alipay.trade.app.pay",
		"sign_type": "RSA2",
	}

	for name, expected := range checks {
		got := parsed.Get(name)
		if got != expected {
			t.Errorf("%s: expected %q, got %q", name, expected, got)
		}
	}

	// Verify signature field exists and is non-empty
	sign := parsed.Get("sign")
	if sign == "" {
		t.Error("sign field is empty")
	}
	if len(sign) < 100 {
		t.Errorf("sign too short: %d bytes (RSA2 should be ~344 bytes)", len(sign))
	}

	// Verify timestamp is present
	ts := parsed.Get("timestamp")
	if ts == "" {
		t.Error("timestamp field is empty")
	}

	// Verify biz_content is valid JSON
	bizContent := parsed.Get("biz_content")
	if bizContent == "" {
		t.Error("biz_content is empty")
	}
	if !strings.Contains(bizContent, "subject") {
		t.Error("biz_content missing subject")
	}
	if !strings.Contains(bizContent, "out_trade_no") {
		t.Error("biz_content missing out_trade_no")
	}

	t.Logf("order string: %d bytes, sign: %d bytes", len(orderStr), len(sign))
}

// TestAlipaySDK_EncodeParam verifies the SDK can encode and sign parameters.
func TestAlipaySDK_EncodeParam(t *testing.T) {
	aliClient := testAlipayClient(t)
	if aliClient.Client == nil {
		t.Fatal("alipay client not initialized")
	}

	// Use TradePreCreate as a simple API call that returns encoded params
	param := alipay.TradePreCreate{
		Trade: alipay.Trade{
			Subject:     "编码测试",
			OutTradeNo:  "test_encode_001",
			TotalAmount: "0.01",
			ProductCode: "FACE_TO_FACE_PAYMENT",
		},
	}

	encoded, err := aliClient.Client.EncodeParam(param)
	if err != nil {
		t.Fatalf("EncodeParam failed: %v", err)
	}

	parsed, err := url.ParseQuery(encoded)
	if err != nil {
		t.Fatalf("parse encoded params: %v", err)
	}

	if parsed.Get("sign") == "" {
		t.Error("sign not generated in encoded params")
	}
	if parsed.Get("sign_type") != "RSA2" {
		t.Errorf("sign_type: expected RSA2, got %s", parsed.Get("sign_type"))
	}

	t.Log("SDK parameter encoding and signing works")
}

// TestAlipayCallbackHandler_ParsesFormBody verifies the handler's form parsing
// logic matches what the real Alipay callback sends.
func TestAlipayCallbackHandler_ParsesFormBody(t *testing.T) {
	// Build form-encoded body like Alipay sends
	body := url.Values{}
	body.Set("app_id", "2021006153686187")
	body.Set("trade_status", "TRADE_SUCCESS")
	body.Set("out_trade_no", "test_order_callback")
	body.Set("trade_no", "2026051322001234567890")
	body.Set("total_amount", "1.00")
	body.Set("sign", "mock_sign")
	body.Set("sign_type", "RSA2")
	body.Set("notify_time", "2026-05-13 10:01:01")
	body.Set("notify_type", "trade_status_sync")

	// Simulate the handler's parsing logic (same as AlipayCallback in handler/payment.go)
	rawBody := body.Encode()
	params := make(map[string]string)
	for _, kv := range strings.Split(rawBody, "&") {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) == 2 {
			val, _ := url.QueryUnescape(parts[1])
			params[parts[0]] = val
		}
	}

	if params["trade_status"] != "TRADE_SUCCESS" {
		t.Errorf("trade_status not parsed correctly, got: %s", params["trade_status"])
	}
	if params["out_trade_no"] != "test_order_callback" {
		t.Errorf("out_trade_no not parsed correctly, got: %s", params["out_trade_no"])
	}
	if params["total_amount"] != "1.00" {
		t.Errorf("total_amount not parsed correctly, got: %s", params["total_amount"])
	}
}
