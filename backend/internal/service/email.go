package service

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"log"
	"net/smtp"
	"time"

	"github.com/google/uuid"
	"github.com/planforever/nexusacg/internal/config"
	"github.com/planforever/nexusacg/internal/model"
	"gorm.io/gorm"
)

// EmailService handles sending verification emails and managing tokens.
type EmailService struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewEmailService(db *gorm.DB, cfg *config.Config) *EmailService {
	return &EmailService{db: db, cfg: cfg}
}

// IsConfigured reports whether SMTP is set up.
func (s *EmailService) IsConfigured() bool {
	return s.cfg.SMTPHost != "" && s.cfg.SMTPFromEmail != ""
}

// GenerateToken creates a new verification token for the given user.
// Deletes any existing token for this user first (one active token per user).
func (s *EmailService) GenerateToken(userID uuid.UUID) (string, error) {
	s.db.Where("user_id = ?", userID).Delete(&model.EmailVerificationToken{})

	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate random: %w", err)
	}
	token := hex.EncodeToString(bytes)

	record := model.EmailVerificationToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	if err := s.db.Create(&record).Error; err != nil {
		return "", fmt.Errorf("save token: %w", err)
	}
	return token, nil
}

// VerifyToken checks if the token is valid and not expired, then activates the user.
func (s *EmailService) VerifyToken(token string) (*model.User, error) {
	var record model.EmailVerificationToken
	if err := s.db.Where("token = ? AND expires_at > ?", token, time.Now()).First(&record).Error; err != nil {
		return nil, fmt.Errorf("invalid or expired verification link")
	}

	var user model.User
	if err := s.db.First(&user, "id = ?", record.UserID).Error; err != nil {
		return nil, fmt.Errorf("user not found")
	}

	if err := s.db.Model(&user).Updates(map[string]interface{}{
		"status": "active",
	}).Error; err != nil {
		return nil, fmt.Errorf("activate user: %w", err)
	}

	s.db.Delete(&record)
	return &user, nil
}

// ResendToken generates a fresh token for a pending user by email.
func (s *EmailService) ResendToken(email string) (uuid.UUID, error) {
	var user model.User
	if err := s.db.Where("email = ? AND status = ?", email, "pending_email").First(&user).Error; err != nil {
		return uuid.Nil, fmt.Errorf("no pending registration for this email")
	}
	_, err := s.GenerateToken(user.ID)
	return user.ID, err
}

// SendVerificationEmail sends the verification email via SMTP.
// Falls back to console output in development if SMTP is not configured.
func (s *EmailService) SendVerificationEmail(email, token string) error {
	verifyURL := fmt.Sprintf("nexusacg://verify?token=%s", token)

	subject := "【次元链】请验证您的邮箱"
	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="font-family:sans-serif;max-width:600px;margin:0 auto;padding:20px">
<h2>欢迎加入次元链！</h2>
<p>请点击以下链接验证您的邮箱，验证后账号即可激活使用：</p>
<p><a href="%s" style="display:inline-block;padding:12px 24px;background:#6366f1;color:#fff;text-decoration:none;border-radius:6px">验证邮箱</a></p>
<p style="color:#666;font-size:12px">链接将在 24 小时后失效。如果不是您本人操作，请忽略此邮件。</p>
</body>
</html>`, verifyURL)

	if !s.IsConfigured() {
		log.Printf("[EMAIL DEV MODE] Verification link for %s: %s", email, verifyURL)
		return nil
	}

	return s.sendSMTP(email, subject, body)
}

func (s *EmailService) sendSMTP(to, subject, body string) error {
	from := fmt.Sprintf("%s <%s>", s.cfg.SMTPFromName, s.cfg.SMTPFromEmail)
	addr := fmt.Sprintf("%s:%d", s.cfg.SMTPHost, s.cfg.SMTPPort)

	msg := "From: " + from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n\r\n" +
		body

	// Use TLS dial (required for port 465 / implicit SSL)
	conn, err := tls.Dial("tcp", addr, &tls.Config{
		ServerName: s.cfg.SMTPHost,
	})
	if err != nil {
		return fmt.Errorf("tls dial: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.cfg.SMTPHost)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer client.Quit()

	auth := smtp.PlainAuth("", s.cfg.SMTPUser, s.cfg.SMTPPassword, s.cfg.SMTPHost)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}

	if err := client.Mail(s.cfg.SMTPFromEmail); err != nil {
		return fmt.Errorf("smtp mail: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp rcpt: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	_, err = w.Write([]byte(msg))
	if err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}
	return w.Close()
}
