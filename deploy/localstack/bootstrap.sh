#!/usr/bin/env bash
set -euo pipefail

DYNAMO_URL="${DYNAMODB_ENDPOINT:-http://dynamodb-local:8000}"
LS_URL="${AWS_ENDPOINT_URL:-http://localstack:4566}"
REGION="${AWS_REGION:-us-east-1}"
TABLE="${DYNAMODB_TABLE:-orders}"

AWS_DYNAMO="aws --endpoint-url=$DYNAMO_URL --region=$REGION"
AWS_LS="aws --endpoint-url=$LS_URL --region=$REGION"

# ---------------------------------------------------------------------------
# Wait for services
# ---------------------------------------------------------------------------
echo "Waiting for DynamoDB Local..."
until $AWS_DYNAMO dynamodb list-tables > /dev/null 2>&1; do
  echo "  DynamoDB not ready, retrying in 2s..."
  sleep 2
done

echo "Waiting for LocalStack (SNS/SQS)..."
until $AWS_LS sns list-topics > /dev/null 2>&1; do
  echo "  LocalStack not ready, retrying in 2s..."
  sleep 2
done

# ---------------------------------------------------------------------------
# DynamoDB table
# ---------------------------------------------------------------------------
echo "Creating DynamoDB table '$TABLE'..."
$AWS_DYNAMO dynamodb create-table \
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
$AWS_DYNAMO dynamodb update-time-to-live \
  --table-name "$TABLE" \
  --time-to-live-specification "Enabled=true,AttributeName=ExpiresAt" \
  > /dev/null 2>&1 || true

# ---------------------------------------------------------------------------
# SNS topics
# ---------------------------------------------------------------------------
echo "Creating SNS topics..."
ORDERS_TOPIC_ARN=$($AWS_LS sns create-topic --name orders-events-topic --output text --query TopicArn)
STOCK_TOPIC_ARN=$($AWS_LS sns create-topic --name stock-events-topic --output text --query TopicArn)
PAYMENTS_TOPIC_ARN=$($AWS_LS sns create-topic --name payments-events-topic --output text --query TopicArn)

echo "  orders-events-topic:   $ORDERS_TOPIC_ARN"
echo "  stock-events-topic:    $STOCK_TOPIC_ARN"
echo "  payments-events-topic: $PAYMENTS_TOPIC_ARN"

# ---------------------------------------------------------------------------
# SQS queues helper
# Creates queue + DLQ, applies RedrivePolicy, subscribes to SNS with FilterPolicy
# Usage: create_queue_and_subscribe <queue-name> <topic-arn> <event_type>
# ---------------------------------------------------------------------------
create_queue_and_subscribe() {
  local name="$1"
  local topic_arn="$2"
  local event_type="$3"

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

  # SNS subscription with FilterPolicy on MessageAttribute event_type
  $AWS_LS sns subscribe \
    --topic-arn "$topic_arn" \
    --protocol sqs \
    --notification-endpoint "$q_arn" \
    --attributes "FilterPolicy={\"event_type\":[\"$event_type\"]}" \
    > /dev/null

  echo "  ✓ $name  ← $event_type"
}

echo "Creating SQS queues and SNS subscriptions..."

# Orders consumes from stock and payments topics
create_queue_and_subscribe "orders-stock-available-queue"          "$STOCK_TOPIC_ARN"    "stock.available"
create_queue_and_subscribe "orders-stock-unavailable-queue"        "$STOCK_TOPIC_ARN"    "stock.unavailable"
create_queue_and_subscribe "orders-stock-updated-queue"            "$STOCK_TOPIC_ARN"    "stock.updated"
create_queue_and_subscribe "orders-payment-generated-queue"        "$PAYMENTS_TOPIC_ARN" "payment.generated"
create_queue_and_subscribe "orders-payment-approved-queue"         "$PAYMENTS_TOPIC_ARN" "payment.approved"
create_queue_and_subscribe "orders-payment-failed-queue"           "$PAYMENTS_TOPIC_ARN" "payment.failed"

# Lambda Notification consumes from orders topic (declared here for local testing)
create_queue_and_subscribe "notification-approval-requested-queue" "$ORDERS_TOPIC_ARN"   "order.approval-requested"

echo ""
echo "Bootstrap complete."
echo "  DynamoDB table : $TABLE"
echo "  Orders topic   : $ORDERS_TOPIC_ARN"
