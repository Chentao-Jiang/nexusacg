package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/planforever/nexusacg/internal/config"
	"github.com/planforever/nexusacg/internal/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewAuthService(db *gorm.DB, cfg *config.Config) *AuthService {
	return &AuthService{db: db, cfg: cfg}
}

type RegisterInput struct {
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*model.User, error) {
	// Check if phone or email already exists
	var count int64
	if input.Phone != "" {
		s.db.Model(&model.User{}).Where("phone = ?", input.Phone).Count(&count)
		if count > 0 {
			return nil, fmt.Errorf("phone already registered")
		}
	}
	if input.Email != "" {
		s.db.Model(&model.User{}).Where("email = ?", input.Email).Count(&count)
		if count > 0 {
			return nil, fmt.Errorf("email already registered")
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), 12)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	hashStr := string(hash)

	user := model.User{
		ID:           uuid.New(),
		Nickname:     input.Nickname,
		PasswordHash: &hashStr,
		Role:         "user",
		Status:       "active",
	}

	if input.Phone != "" {
		user.Phone = &input.Phone
	}
	if input.Email != "" {
		user.Email = &input.Email
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

type LoginInput struct {
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

// WeChatOAuthLogin exchanges a WeChat authorization code for tokens,
// finds or creates a user by WechatOpenID, and returns JWT tokens.
func (s *AuthService) WeChatOAuthLogin(ctx context.Context, wechat *WeChatOAuthService, code string) (*model.User, *TokenPair, error) {
	if wechat == nil {
		return nil, nil, fmt.Errorf("WeChat OAuth not configured")
	}

	// Step 1: exchange code for access_token + openid
	tokenResp, err := wechat.GetAccessToken(code)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Step 2: get user info
	userInfo, err := wechat.GetUserInfo(tokenResp.AccessToken, tokenResp.OpenID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get wechat userinfo: %w", err)
	}

	// Step 3: find or create user by WechatOpenID
	var user model.User
	wechatOpenID := userInfo.OpenID
	err = s.db.Where("wechat_open_id = ?", wechatOpenID).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		// Auto-register: create new user with WeChat openid
		user = model.User{
			ID:           uuid.New(),
			WechatOpenID: &wechatOpenID,
			Nickname:     userInfo.Nickname,
			AvatarURL:    &userInfo.HeadimgURL,
			Role:         "user",
			Status:       "active",
		}
		if createErr := s.db.Create(&user).Error; createErr != nil {
			return nil, nil, fmt.Errorf("failed to create user: %w", createErr)
		}
	} else if err != nil {
		return nil, nil, fmt.Errorf("failed to query user: %w", err)
	}

	if user.Status != "active" {
		return nil, nil, fmt.Errorf("user account is not active")
	}

	// Step 4: issue JWT tokens
	tokens, err := s.generateTokens(&user)
	if err != nil {
		return nil, nil, err
	}

	return &user, tokens, nil
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (*TokenPair, error) {
	var user model.User
	query := s.db.Where("status = ?", "active")
	if input.Phone != "" {
		query = query.Where("phone = ?", input.Phone)
	} else if input.Email != "" {
		query = query.Where("email = ?", input.Email)
	} else {
		return nil, fmt.Errorf("phone or email required")
	}

	if err := query.First(&user).Error; err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if user.PasswordHash == nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	tokens, err := s.generateTokens(&user)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

func (s *AuthService) generateTokens(user *model.User) (*TokenPair, error) {
	// Access token - 2 hour expiry
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID.String(),
		"role":    user.Role,
		"exp":     time.Now().Add(2 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	})

	accessStr, err := accessToken.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Refresh token - 7 day expiry
	refreshBytes := make([]byte, 32)
	if _, err := rand.Read(refreshBytes); err != nil {
		return nil, err
	}
	refreshToken := hex.EncodeToString(refreshBytes)

	// Store refresh token in DB
	rt := model.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	s.db.Create(&rt)

	return &TokenPair{
		AccessToken:  accessStr,
		RefreshToken: refreshToken,
		ExpiresIn:    7200,
	}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, tokenStr string) (*TokenPair, error) {
	var rt model.RefreshToken
	if err := s.db.Where("token = ? AND expires_at > ?", tokenStr, time.Now()).First(&rt).Error; err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	var user model.User
	if err := s.db.Where("id = ?", rt.UserID).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user not found")
	}
	if user.Status != "active" {
		return nil, fmt.Errorf("user account is not active")
	}

	tokens, err := s.generateTokens(&user)
	if err != nil {
		return nil, err
	}

	// Delete old refresh token only after new tokens are generated
	s.db.Delete(&rt)

	return tokens, nil
}

// QQOAuthLogin exchanges a QQ authorization code for tokens,
// finds or creates a user by QQOpenID, and returns JWT tokens.
func (s *AuthService) QQOAuthLogin(ctx context.Context, qq *QQOAuthService, code, redirectURI string) (*model.User, *TokenPair, error) {
	if qq == nil {
		return nil, nil, fmt.Errorf("QQ OAuth not configured")
	}

	// Step 1: exchange code for access_token
	tokenResp, err := qq.GetAccessToken(code, redirectURI)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Step 2: get openid
	openid, err := qq.GetOpenID(tokenResp.AccessToken)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get qq openid: %w", err)
	}

	// Step 3: get user info
	userInfo, err := qq.GetUserInfo(tokenResp.AccessToken, openid)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get qq userinfo: %w", err)
	}

	// Step 4: find or create user by QQOpenID
	var user model.User
	err = s.db.Where("qq_open_id = ?", openid).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		// Auto-register: create new user with QQ openid
		user = model.User{
			ID:       uuid.New(),
			QQOpenID: &openid,
			Nickname: userInfo.Nickname,
			Role:     "user",
			Status:   "active",
		}
		if userInfo.Figureurl2 != "" {
			avatarURL := userInfo.Figureurl2
			user.AvatarURL = &avatarURL
		} else if userInfo.Figureurl1 != "" {
			avatarURL := userInfo.Figureurl1
			user.AvatarURL = &avatarURL
		}
		if createErr := s.db.Create(&user).Error; createErr != nil {
			return nil, nil, fmt.Errorf("failed to create user: %w", createErr)
		}
	} else if err != nil {
		return nil, nil, fmt.Errorf("failed to query user: %w", err)
	}

	if user.Status != "active" {
		return nil, nil, fmt.Errorf("user account is not active")
	}

	// Step 5: issue JWT tokens
	tokens, err := s.generateTokens(&user)
	if err != nil {
		return nil, nil, err
	}

	return &user, tokens, nil
}

// SMSLogin finds or creates a user by phone number and returns JWT tokens.
func (s *AuthService) SMSLogin(ctx context.Context, phone string) (*TokenPair, error) {
	var user model.User
	err := s.db.Where("phone = ? AND status = ?", phone, "active").First(&user).Error
	if err == gorm.ErrRecordNotFound {
		// Auto-register: create new user with phone
		phoneRef := phone
		user = model.User{
			ID:       uuid.New(),
			Phone:    &phoneRef,
			Nickname: phone[:3] + "****" + phone[len(phone)-4:],
			Role:     "user",
			Status:   "active",
		}
		if createErr := s.db.Create(&user).Error; createErr != nil {
			return nil, fmt.Errorf("failed to create user: %w", createErr)
		}
	} else if err != nil {
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	if user.Status != "active" {
		return nil, fmt.Errorf("user account is not active")
	}

	return s.generateTokens(&user)
}
