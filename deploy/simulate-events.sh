#!/usr/bin/env bash
set -euo pipefail

# ---------------------------------------------------------------------------
# LocalStack config
# ---------------------------------------------------------------------------
AWS_ENDPOINT_URL="${AWS_ENDPOINT_URL:-http://localhost:4566}"
REGION="${AWS_REGION:-us-east-1}"

SNS_STOCK_RESERVED_ARN="${SNS_STOCK_RESERVED_ARN:-arn:aws:sns:us-east-1:000000000000:catalog-stock-reserved-topic}"
SNS_STOCK_INSUFFICIENT_ARN="${SNS_STOCK_INSUFFICIENT_ARN:-arn:aws:sns:us-east-1:000000000000:catalog-stock-insufficient-topic}"
SNS_BACKORDER_CREATED_ARN="${SNS_BACKORDER_CREATED_ARN:-arn:aws:sns:us-east-1:000000000000:catalog-backorder-created-topic}"
SNS_PAYMENTS_ARN="${SNS_PAYMENTS_ARN:-arn:aws:sns:us-east-1:000000000000:payments-events-topic}"

AWS="aws --endpoint-url=$AWS_ENDPOINT_URL --region=$REGION"

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------
random_uuid() {
  python3 -c "import uuid; print(uuid.uuid4())"
}

sns_publish() {
  local topic_arn="$1"
  local message="$2"
  local attrs="${3:-}"

  if [[ -n "$attrs" ]]; then
    $AWS sns publish --topic-arn "$topic_arn" --message "$message" --message-attributes "$attrs" > /dev/null
  else
    $AWS sns publish --topic-arn "$topic_arn" --message "$message" > /dev/null
  fi
  echo "  Published to: $topic_arn"
}

# ---------------------------------------------------------------------------
# Order ID (argument or prompt)
# ---------------------------------------------------------------------------
if [[ $# -ge 1 ]]; then
  ORDER_ID="$1"
else
  read -rp "Order ID: " ORDER_ID
fi

if [[ -z "$ORDER_ID" ]]; then
  echo "Error: order ID is required." >&2
  exit 1
fi

PRODUCT_ID=$(random_uuid)
EVENT_ID=$(random_uuid)

# ---------------------------------------------------------------------------
# Menu
# ---------------------------------------------------------------------------
echo ""
echo "Order: $ORDER_ID"
echo ""
echo "Choose the event to simulate:"
echo "  1) stock.reserved           — AWAITING_STOCK_CONSULT → IN_PROGRESS"
echo "  2) stock.insufficient       — AWAITING_STOCK_CONSULT → AWAITING_STOCK_ORDER"
echo "  3) backorder.created        — AWAITING_STOCK_ORDER   → IN_PROGRESS"
echo "  4) payment.checkout_created — FINISHED               → AWAITING_PAYMENT"
echo "  5) payment.approved         — AWAITING_PAYMENT       → PAYMENT_APPROVED"
echo "  6) payment.failed           — AWAITING_PAYMENT       → PAYMENT_FAILED"
echo ""
read -rp "Option [1-6]: " OPT

echo ""
case "$OPT" in
  1)
    echo "Publishing stock.reserved..."
    sns_publish "$SNS_STOCK_RESERVED_ARN" \
      "{\"product_id\": \"$PRODUCT_ID\", \"order_id\": \"$ORDER_ID\", \"quantity\": 1, \"type\": \"STOCK_RESERVED\"}"
    ;;
  2)
    echo "Publishing stock.insufficient..."
    sns_publish "$SNS_STOCK_INSUFFICIENT_ARN" \
      "{\"product_id\": \"$PRODUCT_ID\", \"order_id\": \"$ORDER_ID\", \"quantity\": 1, \"type\": \"STOCK_INSUFFICIENT\"}"
    ;;
  3)
    echo "Publishing backorder.created..."
    sns_publish "$SNS_BACKORDER_CREATED_ARN" \
      "{\"product_id\": \"$PRODUCT_ID\", \"order_id\": \"$ORDER_ID\", \"quantity\": 1, \"type\": \"BACKORDER_CREATED\"}"
    ;;
  4)
    echo "Publishing payment.checkout_created..."
    sns_publish "$SNS_PAYMENTS_ARN" \
      "{\"event_id\": \"$EVENT_ID\", \"event_type\": \"payment.checkout_created\", \"data\": {\"order_id\": \"$ORDER_ID\"}}" \
      '{"event_type": {"DataType": "String", "StringValue": "payment.checkout_created"}}'
    ;;
  5)
    echo "Publishing payment.approved..."
    sns_publish "$SNS_PAYMENTS_ARN" \
      "{\"event_id\": \"$EVENT_ID\", \"event_type\": \"payment.approved\", \"data\": {\"order_id\": \"$ORDER_ID\"}}" \
      '{"event_type": {"DataType": "String", "StringValue": "payment.approved"}}'
    ;;
  6)
    read -rp "Failure reason: " REASON
    echo "Publishing payment.failed..."
    sns_publish "$SNS_PAYMENTS_ARN" \
      "{\"event_id\": \"$EVENT_ID\", \"event_type\": \"payment.failed\", \"data\": {\"order_id\": \"$ORDER_ID\", \"reason\": \"$REASON\"}}" \
      '{"event_type": {"DataType": "String", "StringValue": "payment.failed"}}'
    ;;
  *)
    echo "Invalid option." >&2
    exit 1
    ;;
esac

echo "Done."
