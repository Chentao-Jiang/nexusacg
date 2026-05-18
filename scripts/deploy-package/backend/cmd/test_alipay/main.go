package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/smartwalle/alipay/v3"
)

func main() {
	privateKeyPath := "/home/jct/nexusacg/alipay_app_private_key.pem"
	publicKeyPath := "/home/jct/nexusacg/alipay_public_key.pem"
	appID := "2021006153686187"

	fmt.Println("=== Alipay SDK Verification Test ===")
	fmt.Println()

	// 1. Load keys
	privKey, err := readFile(privateKeyPath)
	if err != nil {
		fmt.Printf("FAIL: read private key: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("[1] Private key loaded: %d bytes\n", len(privKey))

	pubKey, err := readFile(publicKeyPath)
	if err != nil {
		fmt.Printf("FAIL: read public key: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("[2] Public key loaded: %d bytes\n", len(pubKey))
	fmt.Println()

	// 2. Initialize SDK (sandbox mode)
	client, err := alipay.New(appID, privKey, true)
	if err != nil {
		fmt.Printf("FAIL: init alipay client: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("[3] Alipay client created: AppID=%s, sandbox=true\n", appID)

	// 3. Load Alipay public key
	if err := client.LoadAliPayPublicKey(pubKey); err != nil {
		fmt.Printf("FAIL: load alipay public key: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("[4] Alipay public key loaded successfully")
	fmt.Println()

	// 4. Generate a test order string (TradeAppPay)
	testOrderNo := fmt.Sprintf("VERIFY_%d", time.Now().Unix())
	param := alipay.TradeAppPay{
		Trade: alipay.Trade{
			Subject:     "沙箱验证订单",
			OutTradeNo:  testOrderNo,
			TotalAmount: "0.01",
			ProductCode: "QUICK_MSECURITY_PAY",
			Body:        "Alipay SDK verification test",
		},
	}

	orderStr, err := client.TradeAppPay(param)
	if err != nil {
		fmt.Printf("FAIL: TradeAppPay: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("[5] TradeAppPay PASS — order string generated: %d bytes\n", len(orderStr))
	fmt.Println()

	// 5. Parse the order string and check structure
	parsed, err := url.ParseQuery(orderStr)
	if err != nil {
		fmt.Printf("FAIL: parse order string: %v\n", err)
		os.Exit(1)
	}

	checks := map[string]bool{
		"app_id":         parsed.Get("app_id") == appID,
		"method":         parsed.Get("method") == "alipay.trade.app.pay",
		"sign_type":      parsed.Get("sign_type") == "RSA2",
		"sign":           len(parsed.Get("sign")) > 0,
		"timestamp":      len(parsed.Get("timestamp")) > 0,
		"biz_content":    len(parsed.Get("biz_content")) > 0,
	}

	allPass := true
	for name, passed := range checks {
		status := "PASS"
		if !passed {
			status = "FAIL"
			allPass = false
		}
		fmt.Printf("    %s: %s\n", name, status)
	}
	fmt.Println()

	if !allPass {
		fmt.Println("FAIL: some checks did not pass")
		os.Exit(1)
	}

	// Summary
	fmt.Println("=== ALL TESTS PASSED ===")
	fmt.Println()
	fmt.Println("Alipay SDK is correctly configured and signing works.")
	fmt.Println()
	fmt.Println("What was verified:")
	fmt.Println("  - SDK initialization with sandbox mode")
	fmt.Println("  - TradeAppPay order string generation")
	fmt.Println("  - Order string contains all required fields with correct values")
	fmt.Println("  - RSA2 signature is present")
	fmt.Println()
	fmt.Println("For full sandbox flow testing, you need:")
	fmt.Println("  1. Alipay sandbox app at opendocs.alipay.com/open/200/105311")
	fmt.Println("  2. Sandbox buyer account (phone number)")
	fmt.Println("  3. Sandbox Alipay public key (different from production)")
	fmt.Println("  4. A reachable callback URL for async notifications")
	fmt.Println("  5. Use Alipay sandbox App to complete a test payment")
}

func readFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}
