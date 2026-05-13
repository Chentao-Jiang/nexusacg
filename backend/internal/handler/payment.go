package handler

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/planforever/nexusacg/internal/service/payment"
)

type PaymentHandler struct {
	svc         *payment.CallbackService
	alipayNotifyURL string
}

// NewPaymentHandler registers payment routes.
// Callbacks come from WeChat/Alipay servers (no auth), prepay requires auth.
func NewPaymentHandler(r *gin.RouterGroup, svc *payment.CallbackService, authMW gin.HandlerFunc, alipayNotifyURL string) {
	h := &PaymentHandler{svc: svc, alipayNotifyURL: alipayNotifyURL}

	// Payment callback endpoints — NO auth middleware
	pay := r.Group("/payments")
	pay.POST("/wechat/callback", h.WechatCallback)
	pay.POST("/alipay/callback", h.AlipayCallback)
	pay.GET("/logs", h.PaymentLogs)

	// Prepay endpoint — requires auth
	pay.POST("/prepay", authMW, h.Prepay)
	pay.GET("/alipay/verify", h.AlipayVerify)
}

// PrepayRequest is the request body for creating a payment order.
type PrepayRequest struct {
	OrderNo     string `json:"order_no" binding:"required"`
	PaymentType string `json:"payment_type" binding:"required,oneof=wechat alipay"`
	Description string `json:"description" binding:"required"`
	Amount      int64  `json:"amount" binding:"required,gt=0"` // amount in cents (fen)
}

// Prepay godoc
// @Summary Create payment order
// @Description Create a unified payment order (WeChat or Alipay) and return client parameters
// @Tags payments
// @Accept json
// @Produce json
// @Param request body PrepayRequest true "Payment order info"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Security BearerAuth
// @Router /payments/prepay [post]
func (h *PaymentHandler) Prepay(c *gin.Context) {
	var req PrepayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request: "+err.Error())
		return
	}

	payerIP := c.ClientIP()

	switch req.PaymentType {
	case "wechat":
		result, err := h.svc.PrepayWechatOrder(c.Request.Context(), req.OrderNo, req.Description, req.Amount, payerIP)
		if err != nil {
			InternalError(c, err.Error())
			return
		}
		Success(c, result)
	case "alipay":
		result, err := h.svc.PrepayAlipayOrder(c.Request.Context(), req.OrderNo, req.Description, req.Amount, h.alipayNotifyURL)
		if err != nil {
			InternalError(c, err.Error())
			return
		}
		Success(c, result)
	default:
		BadRequest(c, "unsupported payment type")
	}
}

// WechatCallback godoc
// @Summary WeChat Pay callback
// @Description Handle WeChat Pay payment callback (called by WeChat server)
// @Tags payments
// @Accept json
// @Produce plain
// @Success 200 {string} string "success response"
// @Failure 500 {string} string "error response"
// @Router /payments/wechat/callback [post]
func (h *PaymentHandler) WechatCallback(c *gin.Context) {
	log.Printf("wechat v3 callback received")

	resp, err := h.svc.HandleWechatCallbackV3(c.Request.Context(), c.Request)
	if err != nil {
		log.Printf("wechat callback error: %v", err)
		c.Header("Content-Type", "application/json; charset=UTF-8")
		c.String(http.StatusInternalServerError, `{"code":"FAIL","message":"处理失败"}`)
		return
	}

	c.Header("Content-Type", "application/json; charset=UTF-8")
	c.String(http.StatusOK, resp)
}

// AlipayCallback godoc
// @Summary Alipay callback
// @Description Handle Alipay payment callback (called by Alipay server)
// @Tags payments
// @Produce plain
// @Success 200 {string} string "success response"
// @Failure 400 {string} string "error response"
// @Router /payments/alipay/callback [post]
func (h *PaymentHandler) AlipayCallback(c *gin.Context) {
	// Read raw body to avoid Go's form parser converting '+' to space in base64 signatures
	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("read alipay callback body: %v", err)
		c.String(http.StatusBadRequest, "fail")
		return
	}
	defer c.Request.Body.Close()

	// Parse manually to preserve '+' in base64 values
	params := make(map[string]string)
	for _, kv := range strings.Split(string(rawBody), "&") {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) == 2 {
			key := parts[0]
			val, _ := url.QueryUnescape(parts[1])
			params[key] = val
		}
	}

	log.Printf("alipay callback received: trade_no=%s, out_trade_no=%s, trade_status=%s",
		params["trade_no"], params["out_trade_no"], params["trade_status"])

	resp, err := h.svc.HandleAlipayCallback(c.Request.Context(), params)
	if err != nil {
		log.Printf("alipay callback error: %v", err)
	}

	// Alipay expects plain text "success" or "fail"
	c.String(http.StatusOK, resp)
}

// PaymentLogs godoc
// @Summary Payment logs
// @Description Get payment logs for audit
// @Tags payments
// @Produce json
// @Param order_id query string false "Filter by order ID"
// @Success 200 {object} Response
// @Failure 500 {object} Response
// @Router /payments/logs [get]
func (h *PaymentHandler) PaymentLogs(c *gin.Context) {
	orderID := c.Query("order_id")
	limit := 50
	logs, err := h.svc.GetPaymentLogs(orderID, limit)
	if err != nil {
		InternalError(c, err.Error())
		return
	}
	Success(c, logs)
}

// AlipayVerify godoc
// @Summary Verify Alipay SDK
// @Description Verify Alipay SDK configuration for debugging
// @Tags payments
// @Produce json
// @Success 200 {object} Response
// @Failure 500 {object} Response
// @Router /payments/alipay/verify [get]
func (h *PaymentHandler) AlipayVerify(c *gin.Context) {
	results, err := h.svc.VerifyAlipaySDK()
	if err != nil {
		InternalError(c, err.Error())
		return
	}
	Success(c, results)
}
