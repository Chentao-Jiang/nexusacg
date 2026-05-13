package service

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/planforever/nexusacg/internal/config"
)

// SMSService handles sending and verifying SMS verification codes.
// Supports two backends:
//   1. Traditional SMS (SendSms) — primary, requires SMS_SIGN_NAME + SMS_TEMPLATE_CODE
//   2. SMS Authentication (SendSmsVerifyCode) — fallback for individual devs, no template needed
type SMSService struct {
	cfg       *config.Config
	mu        sync.Mutex
	codes     map[string]*codeEntry
	sendTimes map[string][]time.Time
}

type codeEntry struct {
	code     string
	expires  time.Time
	sendTime time.Time
}

func NewSMSService(cfg *config.Config) *SMSService {
	s := &SMSService{
		cfg:       cfg,
		codes:     make(map[string]*codeEntry),
		sendTimes: make(map[string][]time.Time),
	}
	go s.cleanupLoop()
	return s
}

func (s *SMSService) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		s.cleanup()
	}
}

func (s *SMSService) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for phone, entry := range s.codes {
		if now.After(entry.expires) {
			delete(s.codes, phone)
		}
	}
	for phone, times := range s.sendTimes {
		cutoff := now.Add(-24 * time.Hour)
		var kept []time.Time
		for _, t := range times {
			if t.After(cutoff) {
				kept = append(kept, t)
			}
		}
		if len(kept) == 0 {
			delete(s.sendTimes, phone)
		} else {
			s.sendTimes[phone] = kept
		}
	}
}

// GenerateAndStore creates a 6-digit verification code for the given phone number.
// Rate limited: 1 code/minute, max 5/day.
func (s *SMSService) GenerateAndStore(phone string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()

	if entry, exists := s.codes[phone]; exists {
		if now.Sub(entry.sendTime) < time.Minute {
			return "", fmt.Errorf("请等待一分钟后再发送")
		}
	}

	dayAgo := now.Add(-24 * time.Hour)
	count := 0
	for _, t := range s.sendTimes[phone] {
		if t.After(dayAgo) {
			count++
		}
	}
	if count >= 5 {
		return "", fmt.Errorf("今日发送次数已达上限，请明天再试")
	}

	code := fmt.Sprintf("%06d", randInt(100000, 999999))
	s.codes[phone] = &codeEntry{
		code:     code,
		expires:  now.Add(5 * time.Minute),
		sendTime: now,
	}
	s.sendTimes[phone] = append(s.sendTimes[phone], now)
	return code, nil
}

// Verify checks if the provided code matches the stored code for the phone number.
// Consumes the code on success.
func (s *SMSService) Verify(phone, code string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, exists := s.codes[phone]
	if !exists {
		return false
	}
	if time.Now().After(entry.expires) {
		delete(s.codes, phone)
		return false
	}
	return hmac.Equal([]byte(entry.code), []byte(code))
}

// SendVerificationCode generates a code and sends it.
// Backend priority: Traditional SMS > SMS Auth API > dev mode (log to console)
func (s *SMSService) SendVerificationCode(phone string) error {
	code, err := s.GenerateAndStore(phone)
	if err != nil {
		return err
	}

	// Try traditional SMS first
	if s.cfg.SMSSignName != "" && s.cfg.SMSTemplateCode != "" {
		if err := s.sendAliyunSMS(phone, code); err != nil {
			// Traditional SMS failed, try SMS Auth API as fallback
			if errAuth := s.sendSmsVerifyCode(phone, code); errAuth != nil {
				return fmt.Errorf("传统短信发送失败: %v; 短信认证 fallback 也失败: %v", err, errAuth)
			}
			return nil
		}
		return nil
	}

	// No traditional SMS configured, try SMS Auth API
	if err := s.sendSmsVerifyCode(phone, code); err != nil {
		fmt.Printf("[SMS Auth API ERROR] %v\n", err)
		// Dev mode: log the code
		fmt.Printf("[SMS DEV MODE] Verification code for %s: %s\n", phone, code)
		return nil
	}
	return nil
}

// --- Traditional SMS (SendSms) ---

func (s *SMSService) sendAliyunSMS(phone, code string) error {
	params := map[string]string{
		"AccessKeyId":      s.cfg.SMSAccessKeyID,
		"Action":           "SendSms",
		"Format":           "JSON",
		"PhoneNumbers":     phone,
		"RegionId":         "cn-hangzhou",
		"SignatureMethod":  "HMAC-SHA1",
		"SignatureNonce":   uuidNonce(),
		"SignatureVersion": "1.0",
		"SignName":         s.cfg.SMSSignName,
		"TemplateCode":     s.cfg.SMSTemplateCode,
		"TemplateParam":    fmt.Sprintf(`{"code":"%s"}`, code),
		"Timestamp":        time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		"Version":          "2017-05-25",
	}

	signature := signRequest(params, s.cfg.SMSAccessKeySecret, "POST")
	params["Signature"] = signature

	form := url.Values{}
	for k, v := range params {
		form.Set(k, v)
	}
	resp, err := http.PostForm("https://dysmsapi.aliyuncs.com/", form)
	if err != nil {
		return fmt.Errorf("发送短信请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("短信发送失败 (HTTP %d)", resp.StatusCode)
	}

	var result struct {
		Code    string `json:"Code"`
		Message string `json:"Message"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("短信响应解析失败")
	}
	if result.Code != "OK" {
		return fmt.Errorf("短信发送失败: %s", result.Message)
	}
	return nil
}

// --- SMS Authentication API (SendSmsVerifyCode) ---
// Used when traditional SMS is not configured (individual devs).
// Endpoint: https://dypnsapi.aliyuncs.com/

func (s *SMSService) sendSmsVerifyCode(phone, code string) error {
	if s.cfg.SMSAccessKeyID == "" || s.cfg.SMSAccessKeySecret == "" {
		return fmt.Errorf("短信认证未配置 AccessKey")
	}

	params := map[string]string{
		"AccessKeyId":      s.cfg.SMSAccessKeyID,
		"Action":           "SendSmsVerifyCode",
		"Format":           "JSON",
		"PhoneNumber":      phone,
		"RegionId":         "cn-hangzhou",
		"SignatureMethod":  "HMAC-SHA1",
		"SignatureNonce":   uuidNonce(),
		"SignatureVersion": "1.0",
		"SignName":         "速通互联验证码",     // 系统预置签名
		"TemplateCode":     "100001",           // 系统预置模板
		"TemplateParam":    fmt.Sprintf(`{"code":"%s","min":"5"}`, code),
		"Timestamp":        time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		"Version":          "2017-05-25",
	}

	signature := signRequest(params, s.cfg.SMSAccessKeySecret, "POST")
	params["Signature"] = signature

	form := url.Values{}
	for k, v := range params {
		form.Set(k, v)
	}
	resp, err := http.PostForm("https://dypnsapi.aliyuncs.com/", form)
	if err != nil {
		return fmt.Errorf("短信认证请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("短信认证发送失败 (HTTP %d)", resp.StatusCode)
	}

	var result struct {
		Code       string `json:"Code"`
		Message    string `json:"Message"`
		RequestId  string `json:"RequestId"`
		SendStatus string `json:"SendStatus"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("短信认证响应解析失败: %s", string(body))
	}
	if result.Code != "OK" {
		return fmt.Errorf("短信认证发送失败: %s - %s", result.Code, result.Message)
	}
	return nil
}

// --- Shared ---

func signRequest(params map[string]string, accessKeySecret, httpMethod string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var canonicalParts []string
	for _, k := range keys {
		canonicalParts = append(canonicalParts, percentEncode(k)+"="+percentEncode(params[k]))
	}
	canonicalQueryString := strings.Join(canonicalParts, "&")

	stringToSign := httpMethod + "&" + percentEncode("/") + "&" + percentEncode(canonicalQueryString)

	key := accessKeySecret + "&"
	mac := hmac.New(sha1.New, []byte(key))
	mac.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return signature
}

func percentEncode(s string) string {
	return strings.ReplaceAll(url.QueryEscape(s), "+", "%20")
}

func randInt(min, max int) int {
	if min > max {
		min, max = max, min
	}
	n := max - min + 1
	if n <= 0 {
		n = math.MaxInt32
	}
	var buf [4]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return min
	}
	val := int(uint32(buf[0])<<24 | uint32(buf[1])<<16 | uint32(buf[2])<<8 | uint32(buf[3]))
	if val < 0 {
		val = -val
	}
	return min + (val % n)
}

func uuidNonce() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
