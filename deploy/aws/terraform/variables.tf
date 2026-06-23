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
  description = "Your public IP in CIDR notation for SSH access. Use 0.0.0.0/0 to allow all (practice only). Get your IP: curl -s ifconfig.me"
  type        = string
  default     = "0.0.0.0/0"
}

variable "ssh_public_key_path" {
  description = "Local path to your SSH public key file. Generate a dedicated key: ssh-keygen -t rsa -b 2048 -f ~/.ssh/prom-stack -N \"\""
  type        = string
  default     = "~/.ssh/prom-stack.pub"
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

variable "create_iam_role" {
  description = "Set to true to create an IAM role for the EC2 instance (SSM + CloudWatch access). Requires iam:CreateRole permission. Set false for restricted AWS accounts."
  type        = bool
  default     = false
}

variable "store_secrets_in_ssm" {
  description = "Set to true to store passwords in SSM Parameter Store. Requires ssm:PutParameter permission. Set false for restricted AWS accounts."
  type        = bool
  default     = false
}
