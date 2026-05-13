package payment

import (
	"context"
	"fmt"
	"log"
	"net/url"

	"github.com/smartwalle/alipay/v3"
)

// AlipaySign verifies Alipay RSA2 signatures using the official smartwalle/alipay SDK.
// Configuration requires: AppID, AppPrivateKey (your RSA private key),
// and AlipayPublicKey (from Alipay Open Platform console).
type AlipaySign struct {
	Client *alipay.Client
}

// NewAlipaySign initializes the Alipay SDK client for signature verification.
// Uses RSA2 (SHA256WithRSA) signing. appPrivateKey is your app's RSA private key (PEM);
// alipayPublicKey is the Alipay public key string from the Open Platform console.
// When sandbox is true, uses the Alipay sandbox gateway URL.
func NewAlipaySign(appID, appPrivateKey, alipayPublicKey string, sandbox bool) (*AlipaySign, error) {
	client, err := alipay.New(appID, appPrivateKey, sandbox)
	if err != nil {
		return nil, fmt.Errorf("init alipay client: %w", err)
	}
	if err := client.LoadAliPayPublicKey(alipayPublicKey); err != nil {
		return nil, fmt.Errorf("load alipay public key: %w", err)
	}
	return &AlipaySign{Client: client}, nil
}

// Verify checks Alipay notification parameters using the SDK's built-in verifier.
// The SDK handles: removing sign/sign_type, sorting, building query string,
// and RSA-SHA256 (RSA2) verification. Returns error if signature is invalid.
func (a *AlipaySign) Verify(params map[string]string, expectedSign string) bool {
	if a.Client == nil {
		log.Printf("alipay client not initialized, skipping verification")
		return false
	}
	// Convert to url.Values for SDK
	v := make(url.Values)
	for k, val := range params {
		v[k] = []string{val}
	}
	err := a.Client.VerifySign(context.Background(), v)
	if err != nil {
		log.Printf("alipay signature verification failed: %v", err)
		return false
	}
	return true
}
