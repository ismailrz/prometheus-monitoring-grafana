# Deploying to AWS

This guide covers deploying the full Prometheus + Grafana monitoring stack to AWS. Two options are described — pick the one that matches your scale.

---

## Option Comparison

| | Option A — Single EC2 | Option B — ECS Fargate |
|---|---|---|
| **Complexity** | Low | High |
| **Cost (est.)** | ~$30–60/mo | ~$80–150/mo |
| **Scaling** | Manual | Automatic |
| **Ops overhead** | Low | Medium |
| **Best for** | Learning, small teams, staging | Production, multi-team |
| **Setup time** | ~20 min | ~2–3 hours |

**This guide focuses on Option A (EC2 + Docker Compose)** — a direct lift-and-shift of the local stack. Option B (ECS) is outlined at the end.

---

## Architecture on AWS

```
                        Internet
                           │
                    ┌──────▼──────┐
                    │  Elastic IP │
                    └──────┬──────┘
                           │ :80 / :443
                    ┌──────▼──────────────────────────┐
                    │  EC2  t3.large                   │
                    │                                  │
                    │  ┌─────────────────────────┐     │
                    │  │  Host nginx (port 80)   │     │
                    │  │  /           → frontend │     │
                    │  │  /api/       → backend  │     │
                    │  │  /grafana/   → grafana  │ ←── basic auth
                    │  │  /prometheus/→ prom     │ ←── basic auth
                    │  │  /alertmgr/  → alertmgr │ ←── basic auth
                    │  └─────────────────────────┘     │
                    │                                  │
                    │  Docker containers (bridge net): │
                    │  frontend  backend  prometheus    │
                    │  grafana   loki     promtail      │
                    │  alertmgr  exporters              │
                    │                                  │
                    └──────────────┬───────────────────┘
                    VPC 10.0.0.0/16│
                                   │ :5432 (private subnet)
                    ┌──────────────▼──────────────┐
                    │  RDS PostgreSQL  (optional)  │
                    │  db.t3.micro                 │
                    └─────────────────────────────┘

  Security Groups:
  - EC2:  22 (your IP only), 80, 443 (public)
  - RDS:  5432 (EC2 SG only — never public)
```

---

## Prerequisites

| Tool | Install |
|------|---------|
| [Terraform](https://developer.hashicorp.com/terraform/install) ≥ 1.6 | `brew install terraform` |
| [AWS CLI](https://aws.amazon.com/cli/) v2 | `brew install awscli` |
| AWS account with IAM credentials | See below |
| SSH key pair | `ssh-keygen -t rsa -b 4096 -f ~/.ssh/id_rsa` |

### Configure AWS credentials

```bash
aws configure
# AWS Access Key ID:     <your key>
# AWS Secret Access Key: <your secret>
# Default region:        us-east-1
# Default output format: json
```

Minimum IAM permissions needed: EC2, VPC, RDS, SSM, IAM (for instance profile).

---

## Step-by-Step Deployment

### 1. Push your code to GitHub

The EC2 bootstrap script clones your repository. Update the URL in `deploy/aws/ec2-userdata.sh`:

```bash
# Edit line ~30 in ec2-userdata.sh:
git clone https://github.com/YOUR_USERNAME/prometheus-monitoring-grafana.git "$APP_DIR"
```

```bash
git init
git add .
git commit -m "initial commit"
git remote add origin https://github.com/YOUR_USERNAME/prometheus-monitoring-grafana.git
git push -u origin main
```

### 2. Configure Terraform variables

```bash
cd deploy/aws/terraform

cp terraform.tfvars.example terraform.tfvars
```

Edit `terraform.tfvars`:

```hcl
project       = "prom-stack"
aws_region    = "us-east-1"
instance_type = "t3.large"          # 2 vCPU / 8 GB RAM

# Your current public IP (restricts SSH access)
operator_cidr = "203.0.113.42/32"   # run: curl -s ifconfig.me

ssh_public_key_path = "~/.ssh/id_rsa.pub"

postgres_password = "SuperSecret123!"
grafana_password  = "AdminPass456!"

use_rds = false   # true = provision RDS (~+$15/mo)
```

> **Security:** `terraform.tfvars` contains passwords. It is already in `.gitignore` — never commit it.

### 3. Deploy infrastructure

```bash
terraform init
terraform plan    # review what will be created
terraform apply   # type "yes" to confirm
```

Terraform creates:
- VPC, subnets, internet gateway, route tables
- Security groups (EC2 + RDS)
- EC2 instance (Ubuntu 22.04)
- Elastic IP
- IAM role + instance profile (for SSM access)
- SSM Parameters (passwords stored as SecureString)
- *(Optional)* RDS PostgreSQL

The EC2 bootstrap script runs automatically on first start. It:
1. Installs Docker + Docker Compose
2. Clones your repository
3. Writes `.env` from Terraform-injected values
4. Starts all containers with the production compose override
5. Configures host nginx as a reverse proxy with basic auth

**First boot takes ~5–8 minutes.** Watch progress:

```bash
ssh ubuntu@$(terraform output -raw instance_public_ip)
tail -f /var/log/userdata.log
```

### 4. Verify deployment

```bash
# Get the public IP
terraform output instance_public_ip

# Check all containers are running
ssh ubuntu@<IP> "cd /opt/prometheus-stack && docker compose ps"

# Test the app
curl http://<IP>/health
curl http://<IP>/api/v1/products
```

### 5. Access the stack

| URL | Credentials |
|-----|-------------|
| `http://<IP>` | Frontend React app |
| `http://<IP>/grafana/` | admin / *your grafana_password* |
| `http://<IP>/prometheus/` | admin / *your grafana_password* (basic auth) |
| `http://<IP>/alertmanager/` | admin / *your grafana_password* (basic auth) |

---

## File Reference

```
deploy/aws/
├── terraform/
│   ├── main.tf                   # VPC, EC2, RDS, IAM, SSM
│   ├── variables.tf              # All input variables
│   ├── outputs.tf                # IP, URLs, SSH command
│   └── terraform.tfvars.example  # Template — copy to terraform.tfvars
├── ec2-userdata.sh               # Bootstrap: Docker install + stack start
├── docker-compose.prod.yml       # Production overrides (no host port bindings, resource limits)
└── nginx/
    └── monitoring-proxy.conf     # Host nginx: reverse proxy + basic auth for monitoring UIs
```

### docker-compose.prod.yml

The production override does three things:

1. **Removes all host port bindings** for internal services (Prometheus, Grafana, Loki, etc.) — they are only reachable through the nginx proxy on port 80.
2. **Adds resource limits** so a runaway container cannot starve the host.
3. **Configures Grafana** to serve from the `/grafana` sub-path.

Start with both files:
```bash
docker compose \
  -f docker-compose.yml \
  -f deploy/aws/docker-compose.prod.yml \
  up -d
```

### nginx/monitoring-proxy.conf

Runs on the **EC2 host** (not in Docker). Routes traffic:

| Path | Upstream | Auth |
|------|----------|------|
| `/` | frontend:3000 | none |
| `/api/` | backend:8080 | none |
| `/grafana/` | grafana:3001 | basic auth |
| `/prometheus/` | prometheus:9090 | basic auth |
| `/alertmanager/` | alertmanager:9093 | basic auth |
| `/loki/` | loki:3100 | basic auth |

---

## Adding HTTPS (strongly recommended)

Install Certbot and get a free TLS certificate. Requires a domain name pointed at your Elastic IP.

```bash
ssh ubuntu@<IP>

# Install certbot
sudo apt install -y certbot python3-certbot-nginx

# Get certificate (replace with your domain)
sudo certbot --nginx -d monitoring.yourdomain.com

# Auto-renewal is configured automatically
sudo certbot renew --dry-run
```

Certbot will modify `/etc/nginx/sites-available/monitoring` to add:
- `listen 443 ssl`
- `ssl_certificate` / `ssl_certificate_key` paths
- HTTP → HTTPS redirect

---

## Common Operations

### Update the application

```bash
ssh ubuntu@<IP>
cd /opt/prometheus-stack

git pull

docker compose \
  -f docker-compose.yml \
  -f deploy/aws/docker-compose.prod.yml \
  up -d --build
```

### Reload Prometheus config without restart

```bash
ssh ubuntu@<IP>
curl -s -X POST http://localhost:9090/-/reload
```

### View logs for a specific service

```bash
ssh ubuntu@<IP>
docker compose -f /opt/prometheus-stack/docker-compose.yml logs -f backend
```

### Backup Prometheus data

```bash
ssh ubuntu@<IP>

# Create a point-in-time snapshot
curl -s -XPOST http://localhost:9090/api/v1/admin/tsdb/snapshot
# Response includes snapshot name in /prometheus/snapshots/

# Copy to S3
aws s3 cp /var/lib/docker/volumes/prometheus-stack_prometheus-data \
  s3://your-backup-bucket/prometheus-$(date +%Y%m%d)/ \
  --recursive
```

### Teardown

```bash
cd deploy/aws/terraform
terraform destroy   # type "yes"
```

> If `use_rds = true` and `deletion_protection = true` on the RDS instance, you must disable it manually in the AWS console first.

---

## Cost Estimate

| Resource | Spec | Monthly cost (us-east-1) |
|----------|------|--------------------------|
| EC2 t3.medium | 2 vCPU / 4 GB | ~$30 |
| EC2 t3.large | 2 vCPU / 8 GB | ~$60 |
| EBS gp3 40 GB | root volume | ~$3 |
| Elastic IP | (free while attached) | $0 |
| RDS db.t3.micro | 2 vCPU / 1 GB | ~$15 |
| Data transfer out | first 100 GB free | ~$0 |
| **Total (t3.large + RDS)** | | **~$78/mo** |
| **Total (t3.large, no RDS)** | | **~$63/mo** |

All estimates use on-demand pricing. Use a **Reserved Instance (1 year)** to save ~40%.

---

## Option B — ECS Fargate (overview)

For a fully managed, auto-scaling deployment replace EC2 + Docker Compose with:

| Local component | AWS equivalent |
|-----------------|----------------|
| docker-compose.yml | ECS Task Definitions |
| Docker network | ECS Service Connect / Cloud Map |
| Volumes | EFS (Prometheus, Grafana, Loki) or S3 (Loki) |
| .env file | ECS Secrets (from Secrets Manager) |
| Host nginx | Application Load Balancer + target groups |
| PostgreSQL container | RDS PostgreSQL |
| node-exporter | CloudWatch Container Insights |

Key additional services needed:
- **ECR** — private registry for your built images
- **ALB** — load balancer with SSL termination
- **EFS** — persistent volumes for Prometheus TSDB and Loki chunks
- **Secrets Manager** — replaces `.env`
- **Route 53** — DNS
- **ACM** — free TLS certificate

Estimated additional setup: ~3–4 hours with Terraform or AWS CDK.

---

## Security Checklist

- [ ] `terraform.tfvars` is in `.gitignore` — never committed
- [ ] SSH access restricted to `operator_cidr` (your IP only)
- [ ] RDS is in a private subnet — not publicly accessible
- [ ] Monitoring UIs protected with basic auth (or VPN)
- [ ] HTTPS enabled via Certbot + ACM
- [ ] Passwords stored in SSM Parameter Store (SecureString)
- [ ] EC2 IAM role has least-privilege permissions
- [ ] EBS volume encrypted at rest
- [ ] `deletion_protection = true` on RDS
- [ ] Regular AMI snapshots or EBS snapshots scheduled
