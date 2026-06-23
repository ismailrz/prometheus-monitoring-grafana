#!/usr/bin/env bash
# Seed the database — creates products (if none exist) then creates sample orders.
set -euo pipefail

API="${API:-http://localhost:8080/api/v1}"
BASE="${API/\/api\/v1/}"

# Check backend is reachable
if ! curl -sf --max-time 5 "$BASE/health" > /dev/null 2>&1; then
  echo "Error: backend not reachable at $BASE/health"
  echo "Make sure the stack is running: make up"
  exit 1
fi

# ─── Seed products if the table is empty ─────────────────────────────────────
PRODUCT_COUNT=$(curl -sf "$API/products" | python3 -c "
import json, sys
data = json.loads(sys.stdin.read())
print(data.get('count', 0))
")

if [[ "$PRODUCT_COUNT" -eq 0 ]]; then
  echo "No products found — seeding 15 products..."

  seed_product() {
    curl -sf -X POST "$API/products" \
      -H "Content-Type: application/json" \
      -d "$1" > /dev/null
  }

  seed_product '{"name":"MacBook Pro 14\"",  "description":"Apple M3 Pro, 18GB RAM, 512GB SSD", "price":1999.99,"stock":45,"category":"laptops"}'
  seed_product '{"name":"MacBook Air 15\"",  "description":"Apple M2, 8GB RAM, 256GB SSD",      "price":1299.99,"stock":80,"category":"laptops"}'
  seed_product '{"name":"Dell XPS 15",       "description":"Intel i7, 32GB RAM, 1TB NVMe",      "price":1799.99,"stock":30,"category":"laptops"}'
  seed_product '{"name":"iPhone 15 Pro",     "description":"256GB, Titanium, A17 Pro chip",     "price":999.99, "stock":120,"category":"phones"}'
  seed_product '{"name":"Samsung Galaxy S24","description":"256GB, Snapdragon 8 Gen 3",         "price":799.99, "stock":90,"category":"phones"}'
  seed_product '{"name":"Google Pixel 8 Pro","description":"256GB, Tensor G3",                  "price":899.99, "stock":60,"category":"phones"}'
  seed_product '{"name":"AirPods Pro 2",     "description":"Active Noise Cancellation, USB-C",  "price":249.99, "stock":200,"category":"audio"}'
  seed_product '{"name":"Sony WH-1000XM5",   "description":"Over-ear ANC, 30hr battery",        "price":349.99, "stock":75,"category":"audio"}'
  seed_product '{"name":"Bose QC45",         "description":"Over-ear noise cancelling",          "price":279.99, "stock":50,"category":"audio"}'
  seed_product '{"name":"iPad Pro 12.9\"",   "description":"M2 chip, 256GB WiFi",               "price":999.99, "stock":40,"category":"tablets"}'
  seed_product '{"name":"iPad Air 5",        "description":"M1 chip, 64GB WiFi",                "price":599.99, "stock":65,"category":"tablets"}'
  seed_product '{"name":"Apple Watch S9",    "description":"GPS + Cellular, 45mm Aluminum",     "price":429.99, "stock":150,"category":"wearables"}'
  seed_product '{"name":"Garmin Fenix 7",    "description":"Solar GPS multisport watch",         "price":699.99, "stock":30,"category":"wearables"}'
  seed_product '{"name":"LG 27\" 4K Monitor","description":"IPS, 144Hz, USB-C",                 "price":599.99, "stock":25,"category":"electronics"}'
  seed_product '{"name":"Logitech MX Master 3","description":"Advanced wireless mouse",          "price":99.99,  "stock":300,"category":"electronics"}'

  echo "Products seeded."
else
  echo "Found $PRODUCT_COUNT existing products — skipping product seed."
fi

# ─── Seed orders ─────────────────────────────────────────────────────────────
echo "Fetching product IDs..."
PRODUCTS=$(curl -sf "$API/products" | python3 -c "
import json, sys
data = json.loads(sys.stdin.read())
ids = [p['id'] for p in data.get('data', [])]
print('\n'.join(ids))
")

if [[ -z "$PRODUCTS" ]]; then
  echo "Error: still no products after seeding. Check backend logs: docker compose logs backend"
  exit 1
fi

IDS=($PRODUCTS)
COUNT=${#IDS[@]}
echo "Creating 10 sample orders from $COUNT products..."

EMAILS=("alice@example.com" "bob@example.com" "carol@example.com" "dave@example.com" "eve@example.com")

for i in $(seq 0 9); do
  EMAIL="${EMAILS[$((i % ${#EMAILS[@]}))]}"
  P1="${IDS[$((RANDOM % COUNT))]}"
  P2="${IDS[$((RANDOM % COUNT))]}"
  QTY1=$((RANDOM % 3 + 1))
  QTY2=$((RANDOM % 2 + 1))

  curl -sf -X POST "$API/orders" \
    -H "Content-Type: application/json" \
    -d "{\"customer_email\":\"$EMAIL\",\"items\":[{\"product_id\":\"$P1\",\"quantity\":$QTY1},{\"product_id\":\"$P2\",\"quantity\":$QTY2}]}" \
    > /dev/null

  echo "  Created order $((i+1))/10 for $EMAIL"
done

echo "Done. Run 'make load-test' to generate continuous traffic."
