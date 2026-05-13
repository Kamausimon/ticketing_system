#!/usr/bin/env bash
# Creates all SNS topics and SQS queues in LocalStack to mirror what
# Terraform provisions in a real AWS environment.
set -euo pipefail

ENDPOINT="http://localhost:4566"
REGION="us-east-1"
ACCOUNT="000000000000"   # LocalStack's fake account ID
PREFIX="ticketing-dev"

AWS="aws --endpoint-url=$ENDPOINT --region=$REGION"

SNS_TOPICS=(
  order-created
  order-payment-confirmed
  order-confirmed
  order-cancelled
  order-completed
  reservation-expired
  inventory-released
  event-cancelled
  payment-completed
  tickets-generated
  notification-email-requested
  refund-approved
  refund-completed
)

WORKER_QUEUES=(
  "payment-worker:order-created"
  "ticket-worker:payment-completed"
  "ticket-email-worker:tickets-generated"
  "email-worker:notification-email-requested"
  "notification-worker:order-confirmed"
  "event-cancellation-worker:event-cancelled"
  "refund-processor-worker:refund-approved"
  "waitlist-worker:inventory-released"
  "analytics-order-created:order-created"
  "analytics-order-confirmed:order-confirmed"
  "analytics-payment-completed:payment-completed"
  "analytics-order-cancelled:order-cancelled"
  "analytics-event-cancelled:event-cancelled"
  "analytics-tickets-generated:tickets-generated"
)

echo "Creating SNS topics..."
declare -A TOPIC_ARNS
for topic in "${SNS_TOPICS[@]}"; do
  ARN=$($AWS sns create-topic --name "${PREFIX}-${topic}" --query TopicArn --output text)
  TOPIC_ARNS[$topic]=$ARN
  echo "  ✓ $topic → $ARN"
done

echo ""
echo "Creating SQS queues and SNS subscriptions..."
for entry in "${WORKER_QUEUES[@]}"; do
  QUEUE="${entry%%:*}"
  TOPIC="${entry##*:}"

  QUEUE_URL=$($AWS sqs create-queue --queue-name "${PREFIX}-${QUEUE}" --query QueueUrl --output text)
  QUEUE_ARN="arn:aws:sqs:${REGION}:${ACCOUNT}:${PREFIX}-${QUEUE}"

  $AWS sns subscribe \
    --topic-arn "${TOPIC_ARNS[$TOPIC]}" \
    --protocol sqs \
    --notification-endpoint "$QUEUE_ARN" > /dev/null

  echo "  ✓ ${QUEUE} subscribed to ${TOPIC}"
done

echo ""
echo "Done. Copy the following into your .env:"
echo ""
for topic in "${SNS_TOPICS[@]}"; do
  VAR=$(echo "$topic" | tr '[:lower:]' '[:upper:]' | tr '-' '_' | sed 's/\./_/g')
  echo "SNS_ARN_${VAR}=${TOPIC_ARNS[$topic]}"
done
echo ""
for entry in "${WORKER_QUEUES[@]}"; do
  QUEUE="${entry%%:*}"
  VAR=$(echo "$QUEUE" | tr '[:lower:]' '[:upper:]' | tr '-' '_')
  URL="http://localhost:4566/000000000000/${PREFIX}-${QUEUE}"
  echo "SQS_URL_${VAR}=${URL}"
done
