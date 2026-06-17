#!/usr/bin/env bash
# Seed the database with sample orders using existing products.
set -euo pipefail

API="http://localhost:8080/api/v1"

echo "Fetching product IDs..."
PRODUCTS=$(curl -sf "$API/products" | python3 -c "
import json, sys
data = json.load(sys.stdin)
ids = [p['id'] for p in data.get('data', [])]
print('\n'.join(ids))
")

if [[ -z "$PRODUCTS" ]]; then
  echo "No products found. Make sure the stack is running: make up"
  exit 1
fi

IDS=($PRODUCTS)
COUNT=${#IDS[@]}
echo "Found $COUNT products. Creating sample orders..."

EMAILS=("alice@example.com" "bob@example.com" "carol@example.com" "dave@example.com" "eve@example.com")

for i in $(seq 0 9); do
  EMAIL="${EMAILS[$((i % ${#EMAILS[@]}))]}"
  P1="${IDS[$((RANDOM % COUNT))]}"
  P2="${IDS[$((RANDOM % COUNT))]}"
  QTY1=$((RANDOM % 3 + 1))
  QTY2=$((RANDOM % 2 + 1))

  curl -sf -X POST "$API/orders" \
    -H "Content-Type: application/json" \
    -d "{
      \"customer_email\": \"$EMAIL\",
      \"items\": [
        {\"product_id\": \"$P1\", \"quantity\": $QTY1},
        {\"product_id\": \"$P2\", \"quantity\": $QTY2}
      ]
    }" > /dev/null

  echo "  Created order $((i+1))/10 for $EMAIL"
done

echo "Done. Run 'make load-test' to generate continuous traffic."
