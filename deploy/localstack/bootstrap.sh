#!/usr/bin/env bash
set -euo pipefail

LS_URL="${AWS_ENDPOINT_URL:-http://localstack:4566}"
REGION="${AWS_REGION:-us-east-1}"
TABLE="${DYNAMODB_TABLE:-orders}"

AWS_LS="aws --endpoint-url=$LS_URL --region=$REGION"

# ---------------------------------------------------------------------------
# Wait for services
# ---------------------------------------------------------------------------
echo "Waiting for LocalStack (DynamoDB/SNS/SQS)..."
until $AWS_LS dynamodb list-tables > /dev/null 2>&1; do
  echo "  LocalStack not ready, retrying in 2s..."
  sleep 2
done

# ---------------------------------------------------------------------------
# DynamoDB table
# ---------------------------------------------------------------------------
echo "Creating DynamoDB table '$TABLE'..."
$AWS_LS dynamodb create-table \
  --table-name "$TABLE" \
  --attribute-definitions \
    AttributeName=PK,AttributeType=S \
    AttributeName=SK,AttributeType=S \
    AttributeName=GSI1PK,AttributeType=S \
    AttributeName=GSI1SK,AttributeType=S \
    AttributeName=GSI2PK,AttributeType=S \
    AttributeName=GSI2SK,AttributeType=S \
  --key-schema \
    AttributeName=PK,KeyType=HASH \
    AttributeName=SK,KeyType=RANGE \
  --billing-mode PAY_PER_REQUEST \
  --global-secondary-indexes \
    '[
      {
        "IndexName":"GSI1",
        "KeySchema":[
          {"AttributeName":"GSI1PK","KeyType":"HASH"},
          {"AttributeName":"GSI1SK","KeyType":"RANGE"}
        ],
        "Projection":{"ProjectionType":"ALL"}
      },
      {
        "IndexName":"GSI2",
        "KeySchema":[
          {"AttributeName":"GSI2PK","KeyType":"HASH"},
          {"AttributeName":"GSI2SK","KeyType":"RANGE"}
        ],
        "Projection":{"ProjectionType":"ALL"}
      }
    ]' \
  2>&1 | grep -v "ResourceInUseException" || true

echo "Enabling TTL on attribute 'ExpiresAt'..."
$AWS_LS dynamodb update-time-to-live \
  --table-name "$TABLE" \
  --time-to-live-specification "Enabled=true,AttributeName=ExpiresAt" \
  > /dev/null 2>&1 || true

# ---------------------------------------------------------------------------
# SNS topics
# ---------------------------------------------------------------------------
echo "Creating SNS topics..."
ORDERS_TOPIC_ARN=$($AWS_LS sns create-topic --name orders-events-topic --output text --query TopicArn)
CATALOG_STOCK_RESERVED_ARN=$($AWS_LS sns create-topic --name catalog-stock-reserved-topic --output text --query TopicArn)
CATALOG_STOCK_INSUFFICIENT_ARN=$($AWS_LS sns create-topic --name catalog-stock-insufficient-topic --output text --query TopicArn)
CATALOG_BACKORDER_CREATED_ARN=$($AWS_LS sns create-topic --name catalog-backorder-created-topic --output text --query TopicArn)
PAYMENTS_TOPIC_ARN=$($AWS_LS sns create-topic --name payments-events-topic --output text --query TopicArn)

echo "  orders-events-topic:             $ORDERS_TOPIC_ARN"
echo "  catalog-stock-reserved-topic:    $CATALOG_STOCK_RESERVED_ARN"
echo "  catalog-stock-insufficient-topic:$CATALOG_STOCK_INSUFFICIENT_ARN"
echo "  catalog-backorder-created-topic: $CATALOG_BACKORDER_CREATED_ARN"
echo "  payments-events-topic:           $PAYMENTS_TOPIC_ARN"

# ---------------------------------------------------------------------------
# SQS queues helper
# Creates queue + DLQ, applies RedrivePolicy, subscribes to SNS topic.
# Usage: create_queue_and_subscribe <queue-name> <topic-arn> [filter-policy-json]
# Omit filter-policy-json to subscribe without filtering (receives all topic messages).
# filter-policy-json example: '{"event_type":["payment.approved","payment.failed"]}'
# ---------------------------------------------------------------------------
create_queue_and_subscribe() {
  local name="$1"
  local topic_arn="$2"
  local filter_policy="${3:-}"

  # DLQ
  local dlq_url
  dlq_url=$($AWS_LS sqs create-queue --queue-name "${name}-dlq" --output text --query QueueUrl)
  local dlq_arn
  dlq_arn=$($AWS_LS sqs get-queue-attributes \
    --queue-url "$dlq_url" --attribute-names QueueArn \
    --output text --query 'Attributes.QueueArn')

  # Main queue with RedrivePolicy
  local q_url
  q_url=$($AWS_LS sqs create-queue \
    --queue-name "$name" \
    --attributes "RedrivePolicy={\"deadLetterTargetArn\":\"$dlq_arn\",\"maxReceiveCount\":\"5\"}" \
    --output text --query QueueUrl)
  local q_arn
  q_arn=$($AWS_LS sqs get-queue-attributes \
    --queue-url "$q_url" --attribute-names QueueArn \
    --output text --query 'Attributes.QueueArn')

  local sub_args=(--topic-arn "$topic_arn" --protocol sqs --notification-endpoint "$q_arn")
  if [ -n "$filter_policy" ]; then
    sub_args+=(--attributes "FilterPolicy=$filter_policy")
  fi
  $AWS_LS sns subscribe "${sub_args[@]}" > /dev/null

  echo "  ✓ $name"
}

echo "Creating SQS queues and SNS subscriptions..."

# Orders consumes from catalog topics (one queue per topic — no filter needed)
create_queue_and_subscribe "orders-stock-reserved-queue"     "$CATALOG_STOCK_RESERVED_ARN"
create_queue_and_subscribe "orders-stock-insufficient-queue" "$CATALOG_STOCK_INSUFFICIENT_ARN"
create_queue_and_subscribe "orders-backorder-created-queue"  "$CATALOG_BACKORDER_CREATED_ARN"

# Orders consumes from payments topic (filter by event_type MessageAttribute)
create_queue_and_subscribe "orders-payment-events-queue" "$PAYMENTS_TOPIC_ARN" '{"event_type":["payment.checkout_created","payment.approved","payment.failed"]}'

# Lambda Notification consumes from orders topic (declared here for local testing)
create_queue_and_subscribe "notification-approval-requested-queue" "$ORDERS_TOPIC_ARN" '{"event_type":["order.approval-requested"]}'

echo ""
echo "Bootstrap complete."
echo "  DynamoDB table : $TABLE"
echo "  Orders topic   : $ORDERS_TOPIC_ARN"
