# Copy these values into your .env file after running `terraform apply`.

# ── SNS ARNs ─────────────────────────────────────────────────────────────────

output "sns_arns" {
  description = "Map of topic-name → SNS ARN (use for SNS_ARN_* env vars)"
  value       = { for k, v in aws_sns_topic.events : k => v.arn }
}

# ── SQS Queue URLs ────────────────────────────────────────────────────────────

output "sqs_queue_urls" {
  description = "Map of worker-name → SQS queue URL (use for SQS_URL_* env vars)"
  value       = { for k, v in aws_sqs_queue.worker : k => v.id }
}

# ── RDS ───────────────────────────────────────────────────────────────────────

output "db_endpoint" {
  description = "Aurora cluster writer endpoint — use as DB_HOST"
  value       = aws_rds_cluster.postgres.endpoint
}

output "db_reader_endpoint" {
  description = "Aurora cluster reader endpoint — point read-only queries here"
  value       = aws_rds_cluster.postgres.reader_endpoint
}

output "db_port" {
  value = aws_rds_cluster.postgres.port
}

# ── ElastiCache ───────────────────────────────────────────────────────────────

output "redis_endpoint" {
  description = "ElastiCache primary endpoint — use as REDIS_ADDR"
  value       = "${aws_elasticache_cluster.redis.cache_nodes[0].address}:${aws_elasticache_cluster.redis.port}"
}
