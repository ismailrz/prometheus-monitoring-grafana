#!/usr/bin/env bash
# Generates realistic traffic to all API endpoints so Grafana dashboards
# have interesting data to show. Runs until Ctrl-C.
set -euo pipefail

API="http://localhost:8080/api/v1"
INTERVAL="${INTERVAL:-0.3}"   # seconds between requests (override with: INTERVAL=0.1 make load-test)

echo "Starting load test against $API (Ctrl-C to stop)"
echo "Set INTERVAL env var to change request rate (current: ${INTERVAL}s)"
echo ""

# Helper — fetch all product IDs
get_product_ids() {
  curl -sf "$API/products" | python3 -c "
import json, sys
data = json.load(sys.stdin)
ids = [p['id'] for p in data.get('data', [])]
if ids: print('\n'.join(ids))
"
}

EMAILS=("load-test-a@example.com" "load-test-b@example.com" "load-test-c@example.com")
REQ=0
ERR=0

while true; do
  IDS=($(get_product_ids 2>/dev/null || echo ""))
  COUNT=${#IDS[@]}

  if [[ $COUNT -eq 0 ]]; then
    echo "No products found — seeding first..."
    bash "$(dirname "$0")/seed-data.sh" 2>/dev/null || true
    sleep 2
    continue
  fi

  # Weighted mix of requests to simulate real traffic
  OP=$((RANDOM % 10))

  if [[ $OP -lt 4 ]]; then
    # 40% — list products (most common)
    STATUS=$(curl -so /dev/null -w "%{http_code}" "$API/products")
  elif [[ $OP -lt 6 ]]; then
    # 20% — get single product
    ID="${IDS[$((RANDOM % COUNT))]}"
    STATUS=$(curl -so /dev/null -w "%{http_code}" "$API/products/$ID")
  elif [[ $OP -lt 7 ]]; then
    # 10% — list orders
    STATUS=$(curl -so /dev/null -w "%{http_code}" "$API/orders")
  elif [[ $OP -lt 8 ]]; then
    # 10% — create product
    STATUS=$(curl -so /dev/null -w "%{http_code}" -X POST "$API/products" \
      -H "Content-Type: application/json" \
      -d "{\"name\":\"Load Test Item $RANDOM\",\"price\":$(echo "scale=2; $((RANDOM % 500 + 10)).99" | bc 2>/dev/null || echo "29.99"),\"stock\":$((RANDOM % 100)),\"category\":\"general\"}")
  elif [[ $OP -lt 9 ]]; then
    # 10% — create order
    P1="${IDS[$((RANDOM % COUNT))]}"
    P2="${IDS[$((RANDOM % COUNT))]}"
    EMAIL="${EMAILS[$((RANDOM % ${#EMAILS[@]}))]}"
    STATUS=$(curl -so /dev/null -w "%{http_code}" -X POST "$API/orders" \
      -H "Content-Type: application/json" \
      -d "{\"customer_email\":\"$EMAIL\",\"items\":[{\"product_id\":\"$P1\",\"quantity\":1},{\"product_id\":\"$P2\",\"quantity\":2}]}")
  else
    # 10% — 404 hit (generates error traffic for dashboards)
    STATUS=$(curl -so /dev/null -w "%{http_code}" "$API/products/00000000-0000-0000-0000-000000000000")
  fi

  REQ=$((REQ + 1))
  [[ "$STATUS" != "2"* && "$STATUS" != "404" ]] && ERR=$((ERR + 1))

  if [[ $((REQ % 50)) -eq 0 ]]; then
    printf "  requests: %d  errors: %d  error_rate: %.1f%%\n" \
      "$REQ" "$ERR" "$(echo "scale=1; $ERR * 100 / $REQ" | bc 2>/dev/null || echo "0")"
  fi

  sleep "$INTERVAL"
done
