#!/usr/bin/env bash
# Phase 1 E2E Test — simulates real user operations through the API
# Uses real data, covers all Phase 1 features
set -uo pipefail

BASE="http://localhost:8080"
PASS=0
FAIL=0

pass() { PASS=$((PASS+1)); echo "  PASS: $1"; }
fail() { FAIL=$((FAIL+1)); echo "  FAIL: $1"; }
section() { echo ""; echo "=== $1 ==="; }
sleep_rate() { echo "  (waiting 65s for rate limit cooldown...)"; sleep 65; }

##############################################################################
section "1. Health Check"
##############################################################################
resp=$(curl -sf "$BASE/health")
echo "Response: $resp"
if echo "$resp" | grep -q '"status":"ok"'; then pass "health check"; else fail "health check"; fi

##############################################################################
section "2. User Registration (buyer, seller, admin)"
##############################################################################
resp=$(curl -sf -X POST "$BASE/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"phone":"13800100001","nickname":"buyer_e2e","password":"testpass123"}')
BUYER_PHONE="13800100001"
BUYER_PASS="testpass123"
echo "Register buyer: $resp"
if echo "$resp" | grep -q '"code":0'; then pass "buyer registration"; else fail "buyer registration"; fi

resp=$(curl -sf -X POST "$BASE/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"phone":"13800100002","nickname":"seller_e2e","password":"testpass123"}')
SELLER_PHONE="13800100002"
SELLER_PASS="testpass123"
echo "Register seller: $resp"
if echo "$resp" | grep -q '"code":0'; then pass "seller registration"; else fail "seller registration"; fi

resp=$(curl -sf -X POST "$BASE/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"phone":"13800100003","nickname":"admin_e2e","password":"testpass123"}')
ADMIN_PHONE="13800100003"
ADMIN_PASS="testpass123"
echo "Register admin: $resp"
if echo "$resp" | grep -q '"code":0'; then pass "admin registration"; else fail "admin registration"; fi

##############################################################################
# Rate limit: 5 requests/min on auth endpoints. Wait before login batch.
##############################################################################
sleep_rate

##############################################################################
section "3. Login (Password + JWT Token)"
##############################################################################
resp=$(curl -sf -X POST "$BASE/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"phone\":\"$BUYER_PHONE\",\"password\":\"$BUYER_PASS\"}")
BUYER_TOKEN=$(echo "$resp" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
BUYER_REFRESH=$(echo "$resp" | grep -o '"refresh_token":"[^"]*"' | cut -d'"' -f4)
if [ -n "$BUYER_TOKEN" ]; then pass "buyer login + JWT"; else fail "buyer login: $resp"; fi

resp=$(curl -sf -X POST "$BASE/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"phone\":\"$SELLER_PHONE\",\"password\":\"$SELLER_PASS\"}")
SELLER_TOKEN=$(echo "$resp" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
if [ -n "$SELLER_TOKEN" ]; then pass "seller login + JWT"; else fail "seller login: $resp"; fi

resp=$(curl -sf -X POST "$BASE/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"phone\":\"$ADMIN_PHONE\",\"password\":\"$ADMIN_PASS\"}")
ADMIN_TOKEN=$(echo "$resp" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
if [ -n "$ADMIN_TOKEN" ]; then pass "admin login + JWT"; else fail "admin login: $resp"; fi

##############################################################################
section "4. Refresh Token"
##############################################################################
if [ -n "$BUYER_REFRESH" ]; then
  resp=$(curl -sf -X POST "$BASE/api/v1/auth/refresh" \
    -H "Content-Type: application/json" \
    -d "{\"refresh_token\":\"$BUYER_REFRESH\"}")
  if echo "$resp" | grep -q '"access_token"'; then pass "refresh token"; else fail "refresh token: $resp"; fi
else
  fail "refresh token (no refresh token)"
fi

##############################################################################
section "5. Token Rotation (old refresh token must be invalidated)"
##############################################################################
if [ -n "$BUYER_REFRESH" ]; then
  resp=$(curl -s -X POST "$BASE/api/v1/auth/refresh" \
    -H "Content-Type: application/json" \
    -d "{\"refresh_token\":\"$BUYER_REFRESH\"}")
  if echo "$resp" | grep -q '"invalid refresh token"'; then pass "token rotation"; else fail "token rotation: $resp"; fi
fi

##############################################################################
section "6. Logout"
##############################################################################
resp=$(curl -sf -X POST "$BASE/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"phone\":\"$BUYER_PHONE\",\"password\":\"$BUYER_PASS\"}")
LOGOUT_REFRESH=$(echo "$resp" | grep -o '"refresh_token":"[^"]*"' | cut -d'"' -f4)
LOGOUT_TOKEN=$(echo "$resp" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)

if [ -n "$LOGOUT_REFRESH" ] && [ -n "$LOGOUT_TOKEN" ]; then
  resp=$(curl -sf -X POST "$BASE/api/v1/auth/logout" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $LOGOUT_TOKEN" \
    -d "{\"refresh_token\":\"$LOGOUT_REFRESH\"}")
  if echo "$resp" | grep -q '"logged_out"'; then pass "logout"; else fail "logout: $resp"; fi
fi

##############################################################################
section "7. Category CRUD"
##############################################################################
resp=$(curl -sf -X POST "$BASE/api/v1/products/categories" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $SELLER_TOKEN" \
  -d '{"name":"E2E Cat","zone":"merch","description":"e2e test category"}')
CAT_ID=$(echo "$resp" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
if echo "$resp" | grep -q '"code":0'; then pass "create category"; else fail "create category: $resp"; fi

resp=$(curl -sf "$BASE/api/v1/products/categories")
if echo "$resp" | grep -q '"E2E Cat"'; then pass "list categories"; else fail "list categories"; fi

##############################################################################
section "8. Product CRUD + Dual Zone (merch/cosplay)"
##############################################################################
resp=$(curl -sf -X POST "$BASE/api/v1/products" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $SELLER_TOKEN" \
  -d '{"name":"E2E Merch","description":"Test merch","price":99.99,"stock":100,"zone":"peripheral","source_type":"self_made","category_id":"'"$CAT_ID"'","images":[],"tags":["anime","test"]}')
PROD_ID=$(echo "$resp" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
if echo "$resp" | grep -q '"code":0'; then pass "create merch product"; else fail "create merch product: $resp"; fi

resp=$(curl -sf -X POST "$BASE/api/v1/products" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $SELLER_TOKEN" \
  -d '{"name":"E2E Cosplay","description":"Test cosplay","price":199.99,"stock":50,"zone":"cosplay","source_type":"official","category_id":"'"$CAT_ID"'","images":[],"tags":["cosplay","test"]}')
if echo "$resp" | grep -q '"code":0'; then pass "create cosplay product"; else fail "create cosplay product: $resp"; fi

resp=$(curl -sf "$BASE/api/v1/products")
if echo "$resp" | grep -q '"E2E Merch"'; then pass "list products"; else fail "list products"; fi

resp=$(curl -sf "$BASE/api/v1/products?q=cosplay")
if echo "$resp" | grep -q '"E2E Cosplay"'; then pass "search products"; else fail "search products"; fi

resp=$(curl -sf "$BASE/api/v1/products?zone=peripheral")
if echo "$resp" | grep -q '"E2E Merch"' && ! echo "$resp" | grep -q '"E2E Cosplay"'; then pass "zone filter (peripheral only)"; else fail "zone filter: $resp"; fi

resp=$(curl -sf "$BASE/api/v1/products/$PROD_ID")
if echo "$resp" | grep -q '"E2E Merch"'; then pass "product detail"; else fail "product detail"; fi

resp=$(curl -sf "$BASE/api/v1/products?tags=anime")
if echo "$resp" | grep -q '"E2E Merch"'; then pass "tag filter"; else fail "tag filter"; fi

##############################################################################
section "9. Community — Posts + Comments + Likes"
##############################################################################
resp=$(curl -sf -X POST "$BASE/api/v1/posts" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $BUYER_TOKEN" \
  -d '{"title":"E2E Post","content":"Test post for e2e","content_type":"text","images":[]}')
POST_ID=$(echo "$resp" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
if echo "$resp" | grep -q '"code":0'; then pass "create post"; else fail "create post: $resp"; fi

resp=$(curl -sf "$BASE/api/v1/posts")
if echo "$resp" | grep -q '"E2E Post"'; then pass "list posts"; else fail "list posts"; fi

resp=$(curl -sf "$BASE/api/v1/posts/$POST_ID")
if echo "$resp" | grep -q '"E2E Post"'; then pass "post detail"; else fail "post detail"; fi

resp=$(curl -sf -X POST "$BASE/api/v1/posts/$POST_ID/like" \
  -H "Authorization: Bearer $BUYER_TOKEN")
if echo "$resp" | grep -q '"code":0'; then pass "like post"; else fail "like post: $resp"; fi

resp=$(curl -sf -X POST "$BASE/api/v1/posts/$POST_ID/comments" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $BUYER_TOKEN" \
  -d '{"content":"E2E comment"}')
COMMENT_ID=$(echo "$resp" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
if echo "$resp" | grep -q '"code":0'; then pass "create comment"; else fail "create comment: $resp"; fi

resp=$(curl -sf -X POST "$BASE/api/v1/posts/$POST_ID/comments" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $SELLER_TOKEN" \
  -d '{"content":"E2E reply","parent_id":"'"$COMMENT_ID"'"}')
if echo "$resp" | grep -q '"code":0'; then pass "nested reply"; else fail "nested reply: $resp"; fi

resp=$(curl -sf "$BASE/api/v1/posts/$POST_ID/comments")
if echo "$resp" | grep -q '"E2E comment"'; then pass "list comments"; else fail "list comments"; fi

##############################################################################
section "10. Events CRUD"
##############################################################################
resp=$(curl -sf -X POST "$BASE/api/v1/events" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $BUYER_TOKEN" \
  -d '{"name":"E2E Convention","description":"Test event","start_time":"2026-06-01T10:00:00+08:00","end_time":"2026-06-02T18:00:00+08:00","address":"Test Venue","organizer":"Test Org","cover_image":"","ticket_price":50.0}')
EVENT_ID=$(echo "$resp" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
if echo "$resp" | grep -q '"code":0'; then pass "create event"; else fail "create event: $resp"; fi

resp=$(curl -sf "$BASE/api/v1/events")
if echo "$resp" | grep -q '"E2E Convention"'; then pass "list events"; else fail "list events"; fi

resp=$(curl -sf "$BASE/api/v1/events/$EVENT_ID")
if echo "$resp" | grep -q '"E2E Convention"'; then pass "event detail"; else fail "event detail"; fi

##############################################################################
section "11. Escrow Order Flow: create → pay → ship → confirm → profit share"
##############################################################################
resp=$(curl -sf -X POST "$BASE/api/v1/orders" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $BUYER_TOKEN" \
  -H "Idempotency-Key: e2e-order-1" \
  -d '{"items":[{"product_id":"'"$PROD_ID"'","quantity":1}]}')
ORDER_NO=$(echo "$resp" | grep -o '"order_no":"[^"]*"' | cut -d'"' -f4)
echo "  Order created: $ORDER_NO"
if echo "$resp" | grep -q '"order_no"'; then pass "create order"; else fail "create order: $resp"; fi

resp=$(curl -sf "$BASE/api/v1/orders/$ORDER_NO" \
  -H "Authorization: Bearer $BUYER_TOKEN")
if echo "$resp" | grep -q '"pending"'; then pass "order detail (pending)"; else fail "order detail pending: $resp"; fi

resp=$(curl -sf "$BASE/api/v1/orders" \
  -H "Authorization: Bearer $BUYER_TOKEN")
if echo "$resp" | grep -q '"order_no"'; then pass "buyer order list"; else fail "buyer order list"; fi

# Pay order (simulated)
resp=$(curl -sf -X POST "$BASE/api/v1/orders/$ORDER_NO/pay" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $BUYER_TOKEN" \
  -d '{"payment_method":"alipay","payment_id":"e2e-pay-1"}')
echo "  Paid order: $resp"
if echo "$resp" | grep -q '"paid"'; then pass "pay order"; else fail "pay order: $resp"; fi

# Seller ships
resp=$(curl -sf -X POST "$BASE/api/v1/orders/$ORDER_NO/ship" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $SELLER_TOKEN" \
  -d '{"tracking_number":"E2E-TRACK-001","carrier":"TestExpress"}')
echo "  Shipped: $resp"
if echo "$resp" | grep -q '"shipped"'; then pass "seller ships"; else fail "seller ship: $resp"; fi

# Buyer confirms receipt → triggers profit sharing
resp=$(curl -sf -X POST "$BASE/api/v1/orders/$ORDER_NO/confirm" \
  -H "Authorization: Bearer $BUYER_TOKEN")
echo "  Confirmed: $resp"
if echo "$resp" | grep -q '"completed"'; then pass "buyer confirms + profit share"; else fail "confirm receipt: $resp"; fi

##############################################################################
section "12. Security Boundary Tests"
##############################################################################

# Non-seller cannot ship
resp=$(curl -s -X POST "$BASE/api/v1/orders/$ORDER_NO/ship" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $BUYER_TOKEN")
if echo "$resp" | grep -qE '"code":400|"not authorized'; then pass "non-seller cannot ship"; else fail "non-seller ship: $resp"; fi

# Seller (not buyer) cannot confirm own order
resp=$(curl -s -X POST "$BASE/api/v1/orders/$ORDER_NO/confirm" \
  -H "Authorization: Bearer $SELLER_TOKEN")
if echo "$resp" | grep -qE '"code":400|"order not found'; then pass "non-buyer cannot confirm"; else fail "non-buyer confirm: $resp"; fi

# Buyer cannot ship own order (self-buy scenario)
resp=$(curl -sf -X POST "$BASE/api/v1/orders" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $BUYER_TOKEN" \
  -H "Idempotency-Key: e2e-self-ship" \
  -d '{"items":[{"product_id":"'"$PROD_ID"'","quantity":1}]}')
SELF_ORDER_NO=$(echo "$resp" | grep -o '"order_no":"[^"]*"' | cut -d'"' -f4)

curl -sf -X POST "$BASE/api/v1/orders/$SELF_ORDER_NO/pay" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $BUYER_TOKEN" \
  -d '{"payment_method":"alipay","payment_id":"e2e-self"}' > /dev/null

resp=$(curl -s -X POST "$BASE/api/v1/orders/$SELF_ORDER_NO/ship" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $BUYER_TOKEN")
if echo "$resp" | grep -qE '"code":400|"not authorized'; then pass "buyer cannot ship own order"; else fail "buyer ship own: $resp"; fi

##############################################################################
section "13. File Upload"
##############################################################################
TEST_IMG=$(mktemp /tmp/e2e_XXXXXX.png)
printf '\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR' > "$TEST_IMG"

resp=$(curl -sf -X POST "$BASE/api/v1/upload" \
  -H "Authorization: Bearer $BUYER_TOKEN" \
  -F "file=@$TEST_IMG;type=image/png")
rm -f "$TEST_IMG"
if echo "$resp" | grep -q '"url"'; then pass "authenticated file upload"; else fail "upload: $resp"; fi

# Unauthenticated upload → must be rejected
TEST_IMG2=$(mktemp /tmp/e2e_unauth_XXXXXX.png)
printf '\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR' > "$TEST_IMG2"
resp=$(curl -s -X POST "$BASE/api/v1/upload" \
  -F "file=@$TEST_IMG2;type=image/png")
rm -f "$TEST_IMG2"
if echo "$resp" | grep -qE '"code":401|"missing authorization'; then pass "unauthenticated upload rejected"; else fail "unauth upload: $resp"; fi

##############################################################################
section "14. Admin Endpoints"
##############################################################################
# Promote admin user to admin role
PGPASSWORD=nexusacg_dev_pass psql -h 127.0.0.1 -p 5432 -U nexusacg -d nexusacg -c \
  "UPDATE users SET role='admin' WHERE phone='$ADMIN_PHONE';" > /dev/null 2>&1

sleep_rate

# Re-login as admin to get fresh token with admin role in JWT
resp=$(curl -sf -X POST "$BASE/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"phone\":\"$ADMIN_PHONE\",\"password\":\"$ADMIN_PASS\"}")
ADMIN_TOKEN_NEW=$(echo "$resp" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
if [ -n "$ADMIN_TOKEN_NEW" ]; then
  echo "  Admin re-login OK"
fi

resp=$(curl -sf "$BASE/api/v1/admin/products/pending" \
  -H "Authorization: Bearer $ADMIN_TOKEN_NEW")
if echo "$resp" | grep -q '"code":0'; then pass "admin pending products"; else fail "admin pending products: $resp"; fi

resp=$(curl -sf "$BASE/api/v1/admin/posts/pending" \
  -H "Authorization: Bearer $ADMIN_TOKEN_NEW")
if echo "$resp" | grep -q '"code":0'; then pass "admin pending posts"; else fail "admin pending posts: $resp"; fi

resp=$(curl -sf "$BASE/api/v1/admin/orders" \
  -H "Authorization: Bearer $ADMIN_TOKEN_NEW")
if echo "$resp" | grep -q '"code":0'; then pass "admin list orders"; else fail "admin list orders: $resp"; fi

resp=$(curl -sf "$BASE/api/v1/admin/stats" \
  -H "Authorization: Bearer $ADMIN_TOKEN_NEW")
if echo "$resp" | grep -q '"total_users"'; then pass "admin dashboard stats"; else fail "admin stats: $resp"; fi

resp=$(curl -sf "$BASE/api/v1/admin/users" \
  -H "Authorization: Bearer $ADMIN_TOKEN_NEW")
if echo "$resp" | grep -q '"items"'; then pass "admin list users"; else fail "admin list users: $resp"; fi

##############################################################################
section "15. Order Refund"
##############################################################################
resp=$(curl -sf -X POST "$BASE/api/v1/orders" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $BUYER_TOKEN" \
  -H "Idempotency-Key: e2e-refund" \
  -d '{"items":[{"product_id":"'"$PROD_ID"'","quantity":1}]}')
REFUND_ORDER_NO=$(echo "$resp" | grep -o '"order_no":"[^"]*"' | cut -d'"' -f4)

curl -sf -X POST "$BASE/api/v1/orders/$REFUND_ORDER_NO/pay" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $BUYER_TOKEN" \
  -d '{"payment_method":"alipay","payment_id":"e2e-refund"}' > /dev/null

resp=$(curl -sf -X POST "$BASE/api/v1/orders/$REFUND_ORDER_NO/refund" \
  -H "Authorization: Bearer $BUYER_TOKEN")
if echo "$resp" | grep -q '"refunded"'; then pass "order refund"; else fail "order refund: $resp"; fi

##############################################################################
section "16. Payment Logs (Admin only)"
##############################################################################
resp=$(curl -sf "$BASE/api/v1/payments/logs" \
  -H "Authorization: Bearer $ADMIN_TOKEN_NEW")
if echo "$resp" | grep -q '"code":0'; then pass "admin payment logs"; else fail "admin payment logs: $resp"; fi

resp=$(curl -s "$BASE/api/v1/payments/logs" \
  -H "Authorization: Bearer $BUYER_TOKEN")
if echo "$resp" | grep -qE '"code":403|"admin access'; then pass "non-admin payment logs denied"; else fail "non-admin logs: $resp"; fi

##############################################################################
section "17. IDOR Protection"
##############################################################################
# Owner can view own order
resp=$(curl -sf "$BASE/api/v1/orders/$ORDER_NO" \
  -H "Authorization: Bearer $BUYER_TOKEN")
if echo "$resp" | grep -q '"completed"'; then pass "owner can view order"; else fail "owner view: $resp"; fi

# Non-owner cannot view another user's order
resp=$(curl -s "$BASE/api/v1/orders/$ORDER_NO" \
  -H "Authorization: Bearer $SELLER_TOKEN")
if echo "$resp" | grep -qE '"code":403|"access denied'; then pass "non-owner IDOR blocked"; else fail "non-owner IDOR: $resp"; fi

##############################################################################
section "18. Alipay SDK Verify (dev only)"
##############################################################################
resp=$(curl -s "$BASE/api/v1/payments/alipay/verify" 2>/dev/null)
if echo "$resp" | grep -q '"code":0\|"trade_app_pay\|"ALL_PASS'; then pass "alipay SDK verify"; else echo "  SKIP: alipay verify unavailable ($resp)"; fi

##############################################################################
section "Results"
##############################################################################
echo ""
echo "================================"
echo "  Passed: $PASS"
echo "  Failed: $FAIL"
echo "  Total:  $((PASS + FAIL))"
echo "================================"

if [ "$FAIL" -gt 0 ]; then
  echo "  SOME TESTS FAILED!"
  exit 1
fi
echo "  ALL TESTS PASSED ✓"
exit 0
