output "instance_public_ip" {
  description = "Elastic IP of the EC2 instance"
  value       = aws_eip.app.public_ip
}

output "instance_id" {
  description = "EC2 instance ID"
  value       = aws_instance.app.id
}

output "ssh_command" {
  description = "SSH command to connect to the instance (use your downloaded .pem file)"
  value       = "ssh -i ~/Downloads/${var.key_pair_name}.pem ubuntu@${aws_eip.app.public_ip}"
}

output "app_url" {
  description = "Frontend URL (configure DNS to point here)"
  value       = "http://${aws_eip.app.public_ip}"
}

output "grafana_url" {
  description = "Grafana dashboard URL"
  value       = "http://${aws_eip.app.public_ip}/grafana"
}

output "prometheus_url" {
  description = "Prometheus UI (restricted to operator_cidr via nginx)"
  value       = "http://${aws_eip.app.public_ip}/prometheus"
}

output "rds_endpoint" {
  description = "RDS PostgreSQL endpoint (only set when use_rds = true)"
  value       = var.use_rds ? aws_db_instance.postgres[0].address : "n/a (using containerised postgres)"
}

output "vpc_id" {
  value = aws_vpc.main.id
}
