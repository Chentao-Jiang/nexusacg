package payment

import (
	"context"
	"crypto/rsa"
	"fmt"
	"net/http"
	"time"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/downloader"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/app"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"
	"github.com/wechatpay-apiv3/wechatpay-go/core/auth/verifiers"
)

// WechatPayClient wraps the official wechatpay-go SDK for API v3 operations.
type WechatPayClient struct {
	Client    *core.Client
	AppSvc    *app.AppApiService
	AppID     string
	MchID     string
	NotifyURL string
	// NotifyHandler handles callback verification and decryption.
	NotifyHandler *notify.Handler
}

// WechatPayClientConfig holds the configuration for creating a WechatPayClient.
type WechatPayClientConfig struct {
	AppID      string // WeChat Open Platform AppID
	MchID      string // Merchant ID
	CertSerial string // Merchant certificate serial number
	APIv3Key   string // API v3 key for callback decryption
	PrivateKey *rsa.PrivateKey
	NotifyURL  string // Callback URL for WeChat Pay notifications
}

// NewWechatPayClient creates a WeChat Pay API v3 client with auto certificate management.
func NewWechatPayClient(ctx context.Context, cfg WechatPayClientConfig) (*WechatPayClient, error) {
	// Use WithWechatPayAutoAuthCipher for automatic platform certificate management
	// (downloads and refreshes platform certificates using APIv3 key)
	coreClient, err := core.NewClient(ctx,
		option.WithWechatPayAutoAuthCipher(cfg.MchID, cfg.CertSerial, cfg.PrivateKey, cfg.APIv3Key),
	)
	if err != nil {
		return nil, fmt.Errorf("init wechatpay client: %w", err)
	}

	// Create notify handler using the same certificate downloader
	mgr := downloader.MgrInstance()
	verifier := verifiers.NewSHA256WithRSAVerifier(mgr.GetCertificateVisitor(cfg.MchID))
	notifyHandler, err := notify.NewRSANotifyHandler(cfg.APIv3Key, verifier)
	if err != nil {
		return nil, fmt.Errorf("init notify handler: %w", err)
	}

	return &WechatPayClient{
		Client:        coreClient,
		AppSvc:        &app.AppApiService{Client: coreClient},
		AppID:         cfg.AppID,
		MchID:         cfg.MchID,
		NotifyURL:     cfg.NotifyURL,
		NotifyHandler: notifyHandler,
	}, nil
}

// WechatPayOrderInput contains the parameters for creating a WeChat Pay order.
type WechatPayOrderInput struct {
	OutTradeNo  string  // Merchant order number (order_no)
	Description string  // Product description
	TotalAmount int64   // Amount in cents (fen)
	OpenID      string  // Payer's OpenID (for JSAPI)
	PayerIP     string  // Payer's client IP
	NotifyURL   string  // Callback URL (overrides default if non-empty)
	TimeExpire  *time.Time // Order expiry time
}

// PrepayOrder creates a WeChat Pay unified order and returns payment parameters for the client.
func (w *WechatPayClient) PrepayOrder(ctx context.Context, input WechatPayOrderInput) (*app.PrepayWithRequestPaymentResponse, error) {
	notifyURL := input.NotifyURL
	if notifyURL == "" {
		notifyURL = w.NotifyURL
	}

	req := app.PrepayRequest{
		Appid:       &w.AppID,
		Mchid:       &w.MchID,
		Description: &input.Description,
		OutTradeNo:  &input.OutTradeNo,
		NotifyUrl:   &notifyURL,
		Amount: &app.Amount{
			Total:    &input.TotalAmount,
			Currency: strPtr("CNY"),
		},
		SceneInfo: &app.SceneInfo{
			PayerClientIp: &input.PayerIP,
		},
	}

	if input.TimeExpire != nil {
		req.TimeExpire = input.TimeExpire
	}

	// PrepayWithRequestPayment calls prepay AND builds the sign parameters for the mobile client
	resp, _, err := w.AppSvc.PrepayWithRequestPayment(ctx, req)
	return resp, err
}

// QueryOrderByOutTradeNo queries order status by merchant order number.
func (w *WechatPayClient) QueryOrderByOutTradeNo(ctx context.Context, outTradeNo string) (*payments.Transaction, error) {
	req := app.QueryOrderByOutTradeNoRequest{
		OutTradeNo: &outTradeNo,
		Mchid:      &w.MchID,
	}
	resp, _, err := w.AppSvc.QueryOrderByOutTradeNo(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("query wechat order: %w", err)
	}
	return resp, nil
}

// CloseOrder closes a WeChat Pay order by merchant order number.
func (w *WechatPayClient) CloseOrder(ctx context.Context, outTradeNo string) error {
	req := app.CloseOrderRequest{
		OutTradeNo: &outTradeNo,
		Mchid:      &w.MchID,
	}
	_, err := w.AppSvc.CloseOrder(ctx, req)
	if err != nil {
		return fmt.Errorf("close wechat order: %w", err)
	}
	return nil
}

// ParseCallback verifies and decrypts a WeChat Pay v3 callback notification.
// Returns the decrypted content as a map for flexible field access.
func (w *WechatPayClient) ParseCallback(ctx context.Context, request interface{}) (interface{}, error) {
	httpReq, ok := request.(*http.Request)
	if !ok {
		return nil, fmt.Errorf("request must be *http.Request")
	}

	var content notify.ContentMap
	_, err := w.NotifyHandler.ParseNotifyRequest(ctx, httpReq, &content)
	if err != nil {
		return nil, fmt.Errorf("parse wechat callback: %w", err)
	}
	return content, nil
}

func strPtr(s string) *string { return &s }

// LoadWechatPayPrivateKey loads a WeChat Pay merchant RSA private key from PEM string.
func LoadWechatPayPrivateKey(pemContent string) (*rsa.PrivateKey, error) {
	key, err := utils.LoadPrivateKey(pemContent)
	if err != nil {
		return nil, fmt.Errorf("load wechat pay private key: %w", err)
	}
	return key, nil
}
