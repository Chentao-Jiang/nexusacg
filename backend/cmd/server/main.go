package main

import (
	"context"
	_ "embed"
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

//go:embed static/verify.html
var verifyHTML string

//go:embed static/login.html
var loginHTML string

func main() {
	cfg := config.Load()
	db := database.Connect(cfg)
	authMW := middleware.JWTAuth(cfg)
	ctx := context.Background()

	// Services
	authSvc := service.NewAuthService(db, cfg)
	emailSvc := service.NewEmailService(db, cfg)
	if emailSvc.IsConfigured() {
		log.Println("email service initialized with SMTP")
	} else {
		log.Println("email service in dev mode (no SMTP configured)")
	}
	smsSvc := service.NewSMSService(cfg)
	productSvc := service.NewProductService(db)
	categorySvc := service.NewCategoryService(db)
	postSvc := service.NewPostService(db)
	eventSvc := service.NewEventService(db)
	orderSvc := service.NewOrderService(db)
	adminSvc := service.NewAdminService(db)
	moderationSvc := service.NewContentModerationService(cfg.DeepSeekAPIKey, cfg.QwenAPIKey)

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
	profitShareSvc := service.NewProfitShareService(db, cfg.PlatformFeePercent, paymentSvc)
	certificationSvc := service.NewCertificationService(db)
	eventListingSvc := service.NewEventServiceListingService(db)
	serviceProductSvc := service.NewServiceProductService(db)
	promotionSvc := service.NewPromotionService(db)

	// Router
	r := gin.Default()
	r.MaxMultipartMemory = 50 << 20 // 50MB for video uploads
	r.Use(middleware.CORS())
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.RateLimit())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Serve uploaded files
	r.Static("/uploads", "./uploads")

	// Verify page (embedded in binary)
	r.GET("/verify", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(verifyHTML))
	})

	// Root redirect to login page
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/login")
	})

	// Login landing page (embedded in binary)
	r.GET("/login", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(loginHTML))
	})

	// robots.txt to discourage search engine indexing
	r.GET("/robots.txt", func(c *gin.Context) {
		c.String(http.StatusOK, "User-agent: *\nDisallow: /uploads/\n")
	})

	// Swagger API documentation (disabled in production)
	if cfg.Env == "development" {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// API v1
	v1 := r.Group("/api/v1")
	handler.NewAuthHandler(v1, authSvc, smsSvc, emailSvc, wechatOauth, qqOauth, cfg)
	handler.NewProductHandler(v1, productSvc, categorySvc, authMW)
	handler.NewPostHandler(v1, postSvc, authMW, moderationSvc)
	handler.NewEventHandler(v1, eventSvc, authMW)
	handler.NewOrderHandler(v1, orderSvc, profitShareSvc, authMW)
	handler.NewPaymentHandler(v1, paymentSvc, authMW, cfg.BaseURL+"/api/v1/payments/alipay/callback", cfg.Env)
	handler.NewUploadHandler(v1, store, authMW, middleware.RequireAdmin())
	handler.NewAdminHandler(v1, adminSvc, authMW, middleware.RequireAdmin())
	handler.NewCertificationHandler(v1, certificationSvc, authMW, db)
	handler.NewEventServiceListingHandler(v1, eventListingSvc, authMW)
	handler.NewServiceProductHandler(v1, serviceProductSvc, authMW)
	handler.NewPromotionHandler(v1, promotionSvc, authMW, middleware.RequireAdmin())

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

	// Auto-release cron: complete shipped orders after configured days
	if cfg.AutoReleaseDays > 0 {
		go func() {
			ticker := time.NewTicker(30 * time.Minute)
			defer ticker.Stop()
			log.Printf("auto-release cron started: completing shipped orders after %d days", cfg.AutoReleaseDays)
			// Run immediately on startup
			autoReleaseOrders(profitShareSvc, cfg.AutoReleaseDays)
			for range ticker.C {
				autoReleaseOrders(profitShareSvc, cfg.AutoReleaseDays)
			}
		}()
	}

	// Promotion expiry cron: expire promotions past their end date
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		log.Println("promotion expiry cron started")
		// Run immediately on startup
		expired, err := promotionSvc.ExpirePromotions(ctx)
		if err != nil {
			log.Printf("promotion expiry cron error: %v", err)
		} else if expired > 0 {
			log.Printf("promotion expiry cron: expired %d promotions", expired)
		}
		for range ticker.C {
			expired, err := promotionSvc.ExpirePromotions(ctx)
			if err != nil {
				log.Printf("promotion expiry cron error: %v", err)
			} else if expired > 0 {
				log.Printf("promotion expiry cron: expired %d promotions", expired)
			}
		}
	}()

	// Graceful shutdown
	log.Printf("server starting on :%s", cfg.Port)

	// Body size limit for JSON endpoints (1MB)
	bodyLimitMW := func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
		c.Next()
	}
	r.Use(bodyLimitMW)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
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

func autoReleaseOrders(svc *service.ProfitShareService, days int) {
	released, err := svc.AutoReleaseOrders(days)
	if err != nil {
		log.Printf("auto-release cron error: %v", err)
		return
	}
	if released > 0 {
		log.Printf("auto-release cron: completed %d shipped orders", released)
	}
}
