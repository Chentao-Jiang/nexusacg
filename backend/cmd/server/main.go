package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "github.com/planforever/nexusacg/docs"

	"github.com/planforever/nexusacg/internal/config"
	"github.com/planforever/nexusacg/internal/database"
	"github.com/planforever/nexusacg/internal/handler"
	"github.com/planforever/nexusacg/internal/middleware"
	"github.com/planforever/nexusacg/internal/service"
	"github.com/planforever/nexusacg/internal/service/payment"
	"github.com/planforever/nexusacg/internal/storage"
)

func main() {
	cfg := config.Load()
	db := database.Connect(cfg)
	authMW := middleware.JWTAuth(cfg)
	ctx := context.Background()

	// Services
	authSvc := service.NewAuthService(db, cfg)
	smsSvc := service.NewSMSService(cfg)
	productSvc := service.NewProductService(db)
	categorySvc := service.NewCategoryService(db)
	postSvc := service.NewPostService(db)
	eventSvc := service.NewEventService(db)
	orderSvc := service.NewOrderService(db)
	adminSvc := service.NewAdminService(db)
	moderationSvc := service.NewContentModerationService(db, cfg.ModerationAPIKey, cfg.ModerationAPISecret)

	// Storage (deferred to after router init)
	store := storage.NewLocalStorage("./uploads", cfg.BaseURL)

	var wechatOauth *service.WeChatOAuthService
	if cfg.WechatOAuthAppID != "" && cfg.WechatOAuthAppSecret != "" {
		wechatOauth = service.NewWeChatOAuthService(cfg.WechatOAuthAppID, cfg.WechatOAuthAppSecret)
		log.Println("wechat oauth service initialized")
	} else {
		log.Println("wechat oauth not configured, skipping")
	}

	var qqOauth *service.QQOAuthService
	if cfg.QQOAuthAppID != "" && cfg.QQOAuthAppKey != "" {
		qqOauth = service.NewQQOAuthService(cfg.QQOAuthAppID, cfg.QQOAuthAppKey)
		log.Println("qq oauth service initialized")
	} else {
		log.Println("qq oauth not configured, skipping")
	}

	// Payment services
	var wechatClient *payment.WechatPayClient
	if cfg.WechatPayAppID != "" && cfg.WechatPayMchID != "" {
		privKeyContent, err := cfg.ReadWechatPayPrivateKey()
		if err != nil {
			log.Printf("warning: wechat pay key file not readable: %v", err)
		} else {
			privKey, err := payment.LoadWechatPayPrivateKey(privKeyContent)
			if err != nil {
				log.Printf("warning: wechat pay private key invalid: %v", err)
			} else {
				wc, err := payment.NewWechatPayClient(ctx, payment.WechatPayClientConfig{
					AppID:      cfg.WechatPayAppID,
					MchID:      cfg.WechatPayMchID,
					CertSerial: cfg.WechatPayCertSerial,
					APIv3Key:   cfg.WechatPayAPIv3Key,
					PrivateKey: privKey,
					NotifyURL:  cfg.BaseURL + "/api/v1/payments/wechat/callback",
				})
				if err != nil {
					log.Printf("warning: wechat pay client init failed: %v", err)
				} else {
					wechatClient = wc
					log.Println("wechat pay client initialized successfully")
				}
			}
		}
	}

	var alipaySign *payment.AlipaySign
	if cfg.AlipayAppID != "" {
		privKey, err1 := cfg.ReadAlipayPrivateKey()
		pubKey, err2 := cfg.ReadAlipayPublicKey()
		if err1 != nil || err2 != nil {
			log.Printf("warning: alipay key files not readable: priv=%v, pub=%v", err1, err2)
		} else {
			as, err := payment.NewAlipaySign(cfg.AlipayAppID, privKey, pubKey, cfg.AlipaySandbox)
			if err != nil {
				log.Printf("warning: alipay SDK init failed: %v", err)
			} else {
				alipaySign = as
				log.Println("alipay SDK initialized successfully")
			}
		}
	}
	paymentSvc := payment.NewCallbackService(db, wechatClient, alipaySign)

	// Router
	r := gin.Default()
	r.Use(middleware.CORS())
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.RateLimit())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "nexusacg"})
	})

	// Serve uploaded files
	r.Static("/uploads", "./uploads")

	// Swagger API documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1
	v1 := r.Group("/api/v1")
	handler.NewAuthHandler(v1, authSvc, smsSvc, wechatOauth, qqOauth, cfg)
	handler.NewProductHandler(v1, productSvc, categorySvc, authMW)
	handler.NewPostHandler(v1, postSvc, authMW, moderationSvc)
	handler.NewEventHandler(v1, eventSvc, authMW)
	handler.NewOrderHandler(v1, orderSvc, authMW)
	handler.NewPaymentHandler(v1, paymentSvc, authMW, cfg.BaseURL+"/api/v1/payments/alipay/callback")
	handler.NewUploadHandler(v1, store, authMW)
	handler.NewAdminHandler(v1, adminSvc, authMW)

	// Order timeout cron: cancel pending orders after configured timeout
	if cfg.OrderTimeoutMinutes > 0 {
		timeout := time.Duration(cfg.OrderTimeoutMinutes) * time.Minute
		go func() {
			ticker := time.NewTicker(5 * time.Minute)
			defer ticker.Stop()
			log.Printf("order timeout cron started: cancelling pending orders after %v", timeout)
			// Run immediately on startup
			cancelPendingOrders(ctx, paymentSvc, timeout)
			for range ticker.C {
				cancelPendingOrders(ctx, paymentSvc, timeout)
			}
		}()
	}

	// Graceful shutdown
	log.Printf("server starting on :%s", cfg.Port)
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server shutdown failed: %v", err)
	}
	log.Println("server stopped")
}

func cancelPendingOrders(ctx context.Context, svc *payment.CallbackService, timeout time.Duration) {
	cancelled, err := svc.CancelTimeoutOrders(ctx, timeout)
	if err != nil {
		log.Printf("order timeout cron error: %v", err)
		return
	}
	if cancelled > 0 {
		log.Printf("order timeout cron: cancelled %d pending orders", cancelled)
	}
}
