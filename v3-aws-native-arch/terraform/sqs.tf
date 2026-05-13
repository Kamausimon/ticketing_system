# Each worker consumer group gets its own SQS queue subscribed to the
# relevant SNS topic(s). This gives independent scaling and dead-letter
# handling per worker type.

locals {
  # Map of queue-name → SNS topic name it subscribes to.
  # Analytics gets six separate queues (one per topic) for independent tracking.
  worker_queues = {
    "payment-worker"             = "order-created"
    "ticket-worker"              = "payment-completed"
    "ticket-email-worker"        = "tickets-generated"
    "email-worker"               = "notification-email-requested"
    "notification-worker"        = "order-confirmed"
    "event-cancellation-worker"  = "event-cancelled"
    "refund-processor-worker"    = "refund-approved"
    "waitlist-worker"            = "inventory-released"

    # Analytics fan-out queues
    "analytics-order-created"      = "order-created"
    "analytics-order-confirmed"    = "order-confirmed"
    "analytics-payment-completed"  = "payment-completed"
    "analytics-order-cancelled"    = "order-cancelled"
    "analytics-event-cancelled"    = "event-cancelled"
    "analytics-tickets-generated"  = "tickets-generated"
  }
}

# ── Dead-letter queues ────────────────────────────────────────────────────────

resource "aws_sqs_queue" "dlq" {
  for_each = local.worker_queues

  name                       = "${local.name_prefix}-${each.key}-dlq"
  message_retention_seconds  = 1209600 # 14 days — long enough for manual inspection
  tags                       = merge(local.common_tags, { Worker = each.key })
}

# ── Main queues ───────────────────────────────────────────────────────────────

resource "aws_sqs_queue" "worker" {
  for_each = local.worker_queues

  name                       = "${local.name_prefix}-${each.key}"
  visibility_timeout_seconds = var.sqs_visibility_timeout_seconds
  message_retention_seconds  = var.sqs_message_retention_seconds

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.dlq[each.key].arn
    maxReceiveCount     = 5
  })

  tags = merge(local.common_tags, { Worker = each.key })
}

# ── Queue policies (allow SNS to send) ───────────────────────────────────────

resource "aws_sqs_queue_policy" "worker" {
  for_each  = local.worker_queues
  queue_url = aws_sqs_queue.worker[each.key].id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Service = "sns.amazonaws.com" }
      Action    = "sqs:SendMessage"
      Resource  = aws_sqs_queue.worker[each.key].arn
      Condition = {
        ArnEquals = {
          "aws:SourceArn" = aws_sns_topic.events[each.value].arn
        }
      }
    }]
  })
}

# ── SNS → SQS subscriptions ──────────────────────────────────────────────────

resource "aws_sns_topic_subscription" "worker" {
  for_each = local.worker_queues

  topic_arn = aws_sns_topic.events[each.value].arn
  protocol  = "sqs"
  endpoint  = aws_sqs_queue.worker[each.key].arn

  # Deliver the raw event payload, not wrapped in the SNS envelope.
  # Set to false so the Go consumer's SNS envelope unwrapper handles both cases.
  raw_message_delivery = false
}
