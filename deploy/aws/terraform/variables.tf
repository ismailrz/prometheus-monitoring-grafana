variable "project" {
  description = "Project name — used as a prefix for all resource names"
  type        = string
  default     = "prom-stack"
}

variable "environment" {
  description = "Deployment environment"
  type        = string
  default     = "production"
}

variable "aws_region" {
  description = "AWS region to deploy into"
  type        = string
  default     = "us-east-1"
}

variable "instance_type" {
  description = "EC2 instance type. t3.medium (4GB) is the minimum; t3.large (8GB) is recommended."
  type        = string
  default     = "t3.medium"
}

variable "operator_cidr" {
  description = "Your public IP in CIDR notation — SSH access is restricted to this address. Get it with: curl -s ifconfig.me"
  type        = string
  # example: "203.0.113.42/32"
}

variable "ssh_public_key_path" {
  description = "Local path to your SSH public key file"
  type        = string
  default     = "~/.ssh/id_rsa.pub"
}

variable "postgres_password" {
  description = "PostgreSQL superuser password"
  type        = string
  sensitive   = true
}

variable "grafana_password" {
  description = "Grafana admin password"
  type        = string
  sensitive   = true
}

variable "use_rds" {
  description = "Set to true to provision an RDS PostgreSQL instance instead of containerised Postgres. Recommended for production."
  type        = bool
  default     = false
}
