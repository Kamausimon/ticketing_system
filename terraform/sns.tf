# One SNS topic per event type — mirrors the former Kafka topic constants.

locals {
  sns_topics = [
    "order-created",
    "order-payment-confirmed",
    "order-confirmed",
    "order-cancelled",
    "order-completed",
    "reservation-expired",
    "inventory-released",
    "event-cancelled",
    "payment-completed",
    "tickets-generated",
    "notification-email-requested",
    "refund-approved",
    "refund-completed",
  ]
}

resource "aws_sns_topic" "events" {
  for_each = toset(local.sns_topics)

  name = "${local.name_prefix}-${each.key}"
  tags = merge(local.common_tags, { TopicName = each.key })
}
