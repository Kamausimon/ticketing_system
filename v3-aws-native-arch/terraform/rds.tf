resource "aws_db_subnet_group" "main" {
  name       = "${local.name_prefix}-db-subnet-group"
  subnet_ids = aws_subnet.private[*].id

  tags = merge(local.common_tags, { Name = "${local.name_prefix}-db-subnet-group" })
}

resource "aws_rds_cluster" "postgres" {
  cluster_identifier      = "${local.name_prefix}-aurora-cluster"
  engine                  = "aurora-postgresql"
  engine_version          = "15.4"
  database_name           = var.db_name
  master_username         = var.db_username
  master_password         = var.db_password
  db_subnet_group_name    = aws_db_subnet_group.main.name
  vpc_security_group_ids  = [aws_security_group.rds.id]

  # Automated backups retained for 7 days.
  backup_retention_period = 7
  preferred_backup_window = "02:00-03:00"

  # Encrypt at rest using the default AWS-managed key.
  storage_encrypted = true

  skip_final_snapshot = var.environment != "production"

  tags = merge(local.common_tags, { Name = "${local.name_prefix}-aurora" })
}

resource "aws_rds_cluster_instance" "writer" {
  identifier         = "${local.name_prefix}-aurora-writer"
  cluster_identifier = aws_rds_cluster.postgres.id
  instance_class     = var.aurora_instance_class
  engine             = aws_rds_cluster.postgres.engine
  engine_version     = aws_rds_cluster.postgres.engine_version

  tags = merge(local.common_tags, { Role = "writer" })
}
