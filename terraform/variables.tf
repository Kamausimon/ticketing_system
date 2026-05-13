variable "aws_region" {
  description = "AWS region to deploy into"
  type        = string
  default     = "us-east-1"
}

variable "project_name" {
  description = "Used as a prefix on every resource name"
  type        = string
  default     = "ticketing"
}

variable "environment" {
  description = "Deployment environment (dev, staging, production)"
  type        = string
  default     = "dev"
}

# ── VPC ───────────────────────────────────────────────────────────────────────

variable "vpc_cidr" {
  description = "CIDR block for the VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "availability_zones" {
  description = "List of AZs to deploy subnets into (at least 2 for RDS)"
  type        = list(string)
  default     = ["us-east-1a", "us-east-1b"]
}

# ── RDS Aurora ────────────────────────────────────────────────────────────────

variable "db_name" {
  description = "Name of the initial PostgreSQL database"
  type        = string
  default     = "ticketing_system"
}

variable "db_username" {
  description = "Master username for the Aurora cluster"
  type        = string
  default     = "ticketing_admin"
}

variable "db_password" {
  description = "Master password for the Aurora cluster"
  type        = string
  sensitive   = true
}

variable "aurora_instance_class" {
  description = "Aurora instance class"
  type        = string
  default     = "db.t3.medium"
}

# ── ElastiCache ───────────────────────────────────────────────────────────────

variable "redis_node_type" {
  description = "ElastiCache node type"
  type        = string
  default     = "cache.t3.micro"
}

variable "redis_num_cache_nodes" {
  description = "Number of cache nodes (1 for dev, 2+ for prod)"
  type        = number
  default     = 1
}

# ── SQS ───────────────────────────────────────────────────────────────────────

variable "sqs_message_retention_seconds" {
  description = "How long SQS retains unprocessed messages"
  type        = number
  default     = 86400 # 1 day
}

variable "sqs_visibility_timeout_seconds" {
  description = "How long a received message is hidden from other consumers"
  type        = number
  default     = 300 # 5 minutes — should exceed max worker processing time
}
