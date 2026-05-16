package config

import (
	"fmt"
	"log"
	"os"
	"strings"
)

type Config struct {
	Env        string
	Port       string
	BaseURL    string // Public URL for payment callback endpoints
	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string
	RedisHost  string
	RedisPort  string
	JWTSecret  string
	// WeChat OAuth (Open Platform web/app login)
	WechatOAuthAppID     string // AppID from open.weixin.qq.com
	WechatOAuthAppSecret string // AppSecret from open.weixin.qq.com
	// WeChat Pay (API v3)
	WechatPayAppID      string // WeChat AppID (from open.weixin.qq.com)
	WechatPayMchID      string // Merchant ID
	WechatPayAPIv3Key   string // API v3 key for callback decryption
	WechatPayCertSerial string // Merchant certificate serial number
	WechatPayPrivateKey string // Path to merchant RSA private key PEM file
	// Alipay (smartwalle/alipay SDK)
	AlipayAppID          string
	AlipayAppPrivateKey  string // Path to RSA private key PEM file
	AlipayPublicKey      string // Path to Alipay RSA public key PEM file
	AlipaySandbox        bool   // Use sandbox gateway URL for testing
	// QQ OAuth (QQ Connect)
	QQOAuthAppID  string // AppID from connect.qq.com
	QQOAuthAppKey string // AppKey from connect.qq.com (not AppSecret)
	// AI Content Moderation
	DeepSeekAPIKey  string // DeepSeek API key for text moderation (V4 Flash)
	QwenAPIKey      string // Qwen API key for image/video moderation (Qwen3-VL-flash, DashScope)
	// SMTP for email verification
	SMTPHost     string // SMTP server host (e.g. smtp.qiye.aliyun.com)
	SMTPPort     int    // SMTP server port (e.g. 465 for SSL)
	SMTPUser     string // SMTP auth username
	SMTPPassword string // SMTP auth password
	SMTPFromName string // Display name in From header (e.g. "次元链")
	SMTPFromEmail string // From email address (e.g. noreply@example.com)
	// SMS (Aliyun)
	SMSAccessKeyID     string // Aliyun AccessKey ID for SMS
	SMSAccessKeySecret string // Aliyun AccessKey Secret for SMS
	SMSSignName        string // SMS sender name (e.g. "次元链") — traditional SMS only
	SMSTemplateCode    string // SMS template code (e.g. "SMS_123456789") — traditional SMS only
	// SMS Authentication (短信认证, for individual devs, uses SendSmsVerifyCode API)
	SMSAuthSchemeName string // Scheme name for SMS Auth API (optional, defaults to "默认方案")
	// Order timeout
	OrderTimeoutMinutes int           // Minutes before pending orders are auto-cancelled
	// Profit sharing
	PlatformFeePercent float64 // Platform fee percentage (default 0.05)
	AutoReleaseDays    int     // Days before shipped orders are auto-completed (default 7)
}

func Load() *Config {
	cfg := &Config{
		Env:                  getEnv("ENV", "development"),
		Port:                 getEnv("PORT", "8080"),
		BaseURL:              getRailwayURL(),
		DBHost:               getEnv("DB_HOST", "localhost"),
		DBPort:               getEnv("DB_PORT", "5432"),
		DBName:               getEnv("DB_NAME", "nexusacg"),
		DBUser:               getEnv("DB_USER", "nexusacg"),
		DBPassword:           getEnv("DB_PASSWORD", "nexusacg_dev_pass"),
		RedisHost:            getEnv("REDIS_HOST", "localhost"),
		RedisPort:            getEnv("REDIS_PORT", "6379"),
		JWTSecret:            getEnv("JWT_SECRET", ""),
		// WeChat OAuth (Open Platform)
		WechatOAuthAppID:     getEnv("WECHAT_OAUTH_APP_ID", ""),
		WechatOAuthAppSecret: getEnv("WECHAT_OAUTH_APP_SECRET", ""),
		// WeChat Pay (API v3)
		WechatPayAppID:      getEnv("WECHAT_PAY_APP_ID", ""),
		WechatPayMchID:      getEnv("WECHAT_PAY_MCH_ID", ""),
		WechatPayAPIv3Key:   getEnv("WECHAT_PAY_APIV3_KEY", ""),
		WechatPayCertSerial: getEnv("WECHAT_PAY_CERT_SERIAL", ""),
		WechatPayPrivateKey: getEnv("WECHAT_PAY_PRIVATE_KEY_PATH", ""),
		AlipayAppID:          getEnv("ALIPAY_APP_ID", ""),
		AlipayAppPrivateKey:  getEnv("ALIPAY_APP_PRIVATE_KEY_PATH", ""),
		AlipayPublicKey:      getEnv("ALIPAY_PUBLIC_KEY_PATH", ""),
		AlipaySandbox:        getEnv("ALIPAY_SANDBOX", "false") == "true",
		QQOAuthAppID:         getEnv("QQ_OAUTH_APP_ID", ""),
		QQOAuthAppKey:        getEnv("QQ_OAUTH_APP_KEY", ""),
		DeepSeekAPIKey:       getEnv("DEEPSEEK_API_KEY", ""),
		QwenAPIKey:           getEnv("QWEN_API_KEY", ""),
		SMSAccessKeyID:       getEnv("SMS_ACCESS_KEY_ID", ""),
		SMSAccessKeySecret:   getEnv("SMS_ACCESS_KEY_SECRET", ""),
		SMSSignName:          getEnv("SMS_SIGN_NAME", ""),
		SMSTemplateCode:      getEnv("SMS_TEMPLATE_CODE", ""),
		SMSAuthSchemeName:    getEnv("SMS_AUTH_SCHEME_NAME", ""),
		SMTPHost:             getEnv("SMTP_HOST", ""),
		SMTPPort:             func() int { v := 465; fmt.Sscanf(getEnv("SMTP_PORT", "465"), "%d", &v); return v }(),
		SMTPUser:             getEnv("SMTP_USER", ""),
		SMTPPassword:         getEnv("SMTP_PASSWORD", ""),
		SMTPFromName:         getEnv("SMTP_FROM_NAME", "次元链"),
		SMTPFromEmail:        getEnv("SMTP_FROM_EMAIL", ""),
		OrderTimeoutMinutes:  func() int { v := 0; fmt.Sscanf(getEnv("ORDER_TIMEOUT_MINUTES", "30"), "%d", &v); return v }(),
		PlatformFeePercent:   func() float64 { v := 0.05; fmt.Sscanf(getEnv("PLATFORM_FEE_PERCENT", "0.05"), "%f", &v); return v }(),
		AutoReleaseDays:      func() int { v := 7; fmt.Sscanf(getEnv("AUTO_RELEASE_DAYS", "7"), "%d", &v); return v }(),
	}

	// Validate JWT secret is set in production
	if cfg.Env != "development" && (cfg.JWTSecret == "" || strings.Contains(cfg.JWTSecret, "change-me")) {
		log.Fatal("JWT_SECRET must be set to a strong random value in production")
	}
	if cfg.JWTSecret == "" {
		cfg.JWTSecret = "dev-secret-change-me"
	}

	return cfg
}

func (c *Config) DSN() string {
	sslMode := getEnv("DB_SSLMODE", "require")
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Shanghai",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, sslMode)
}

func (c *Config) RedisAddr() string {
	return fmt.Sprintf("%s:%s", c.RedisHost, c.RedisPort)
}

// ReadAlipayPrivateKey reads the private key from file.
func (c *Config) ReadAlipayPrivateKey() (string, error) {
	if c.AlipayAppPrivateKey == "" {
		return "", fmt.Errorf("ALIPAY_APP_PRIVATE_KEY_PATH not set")
	}
	data, err := os.ReadFile(c.AlipayAppPrivateKey)
	if err != nil {
		return "", fmt.Errorf("read private key file: %w", err)
	}
	return strings.TrimSpace(string(data)), nil
}

// ReadAlipayPublicKey reads the public key from file.
func (c *Config) ReadAlipayPublicKey() (string, error) {
	if c.AlipayPublicKey == "" {
		return "", fmt.Errorf("ALIPAY_PUBLIC_KEY_PATH not set")
	}
	data, err := os.ReadFile(c.AlipayPublicKey)
	if err != nil {
		return "", fmt.Errorf("read public key file: %w", err)
	}
	return strings.TrimSpace(string(data)), nil
}

// ReadWechatPayPrivateKey reads the WeChat Pay merchant RSA private key from file.
func (c *Config) ReadWechatPayPrivateKey() (string, error) {
	if c.WechatPayPrivateKey == "" {
		return "", fmt.Errorf("WECHAT_PAY_PRIVATE_KEY_PATH not set")
	}
	data, err := os.ReadFile(c.WechatPayPrivateKey)
	if err != nil {
		return "", fmt.Errorf("read wechat private key file: %w", err)
	}
	return strings.TrimSpace(string(data)), nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// getRailwayURL returns the public URL from Railway env vars, falling back to localhost.
// Railway injects RAILWAY_PUBLIC_DOMAIN automatically; we prepend https:// for the full URL.
func getRailwayURL() string {
	if v := os.Getenv("BASE_URL"); v != "" {
		return v
	}
	if domain := os.Getenv("RAILWAY_PUBLIC_DOMAIN"); domain != "" {
		return "https://" + domain
	}
	return "http://localhost:8080"
}
