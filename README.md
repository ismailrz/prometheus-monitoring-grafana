# Prometheus Monitoring & Grafana — 3-Tier App

A production-grade observability stack for a three-tier application built with **React**, **Go**, and **PostgreSQL**. The stack ships metrics, logs, and alerts out of the box with zero manual Grafana configuration.

---

## Table of Contents

1. [Architecture](#architecture)
2. [Stack at a Glance](#stack-at-a-glance)
3. [Prerequisites](#prerequisites)
4. [Quick Start — Local](#quick-start--local)
5. [Deploy to AWS EC2](#deploy-to-aws-ec2)
6. [Project Structure](#project-structure)
7. [Application Layer](#application-layer)
   - [Frontend](#frontend-react--nginx)
   - [Backend](#backend-go)
   - [Database](#database-postgresql)
8. [Monitoring Layer](#monitoring-layer)
   - [Prometheus](#prometheus)
   - [Grafana Dashboards](#grafana-dashboards)
   - [Loki & Promtail](#loki--promtail)
   - [AlertManager](#alertmanager)
9. [Exporters](#exporters)
10. [API Reference](#api-reference)
11. [Metrics Reference](#metrics-reference)
12. [Alert Rules](#alert-rules)
13. [Load Testing](#load-testing)
14. [Makefile Commands](#makefile-commands)
15. [Troubleshooting](#troubleshooting)

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                          Browser / Client                           │
└────────────────────────────┬────────────────────────────────────────┘
                             │ :3000
                ┌────────────▼────────────┐
                │   Frontend (nginx)      │  React SPA + nginx
                │   Port 3000             │  proxies /api → backend
                └────────────┬────────────┘
                             │ :8080
                ┌────────────▼────────────┐
                │   Backend (Go + Gin)    │  REST API
                │   Port 8080             │  /metrics  /health  /ready
                └────────────┬────────────┘
                             │ :5432
                ┌────────────▼────────────┐
                │   PostgreSQL 16         │  products / orders schema
                │   Port 5432             │
                └─────────────────────────┘

──────────────────────── Observability Plane ─────────────────────────

  ┌──────────────┐   scrape   ┌───────────────────────────────────┐
  │  Prometheus  │◄───────────│  backend       :8080/metrics       │
  │  :9090       │◄───────────│  node-exporter :9100               │
  │              │◄───────────│  postgres-exp  :9187               │
  │  Evaluates   │◄───────────│  cadvisor      :8081               │
  │  alert rules │◄───────────│  nginx-exp     :9113               │
  └──────┬───────┘            └───────────────────────────────────┘
         │ alerts
  ┌──────▼────────┐
  │ AlertManager  │  routes / silences / inhibits alerts
  │ :9093         │
  └───────────────┘

  ┌──────────────┐  query    ┌──────────────┐
  │   Grafana    │◄──────────│  Prometheus  │  metrics
  │   :3001      │◄──────────│  Loki :3100  │  logs
  └──────────────┘           └──────────────┘
                                    ▲
                             push   │
                        ┌───────────┴──────────┐
                        │  Promtail  :9080      │
                        │  reads Docker logs    │
                        │  via /var/run/docker  │
                        └───────────────────────┘
```

---

## Stack at a Glance

| Service | Image | Port | Role |
|---------|-------|------|------|
| **Frontend** | nginx:1.25-alpine | 3000 | React SPA + API proxy |
| **Backend** | golang:1.22-alpine | 8080 | REST API + `/metrics` |
| **PostgreSQL** | postgres:16-alpine | 5432 | Primary datastore |
| **Prometheus** | prom/prometheus:v2.51.0 | 9090 | Metrics store & query engine |
| **Grafana** | grafana/grafana:10.4.1 | 3001 | Dashboard UI |
| **Loki** | grafana/loki:2.9.5 | 3100 | Log aggregation |
| **Promtail** | grafana/promtail:2.9.5 | 9080 | Log shipper (Docker → Loki) |
| **AlertManager** | prom/alertmanager:v0.27.0 | 9093 | Alert routing & silencing |
| **Node Exporter** | prom/node-exporter:v1.7.0 | 9100 | Host OS/hardware metrics |
| **Postgres Exporter** | prometheuscommunity/postgres-exporter:v0.15.0 | 9187 | DB metrics |
| **cAdvisor** | gcr.io/cadvisor/cadvisor:v0.49.1 | 8081 | Per-container resource metrics |
| **Nginx Exporter** | nginx/nginx-prometheus-exporter:1.1.0 | 9113 | Nginx stub_status metrics |

---

## Prerequisites

| Tool | Minimum Version | Check |
|------|----------------|-------|
| Docker | 24.x | `docker --version` |
| Docker Compose | v2.x (plugin) | `docker compose version` |
| Make | any | `make --version` |
| curl | any | `curl --version` |
| Python 3 | 3.8+ | `python3 --version` *(used by seed/load-test scripts)* |

> **macOS note:** Docker Desktop must be running. The stack mounts `/var/run/docker.sock` for Promtail container log discovery — this works out of the box on Linux and macOS with Docker Desktop.

---

## Quick Start — Local

```bash
# 1. Clone / enter the directory
cd prometheus-monitoring-grafana

# 2. Start all 13 containers (first run builds images — ~3 min)
make up

# 3. Check everything is healthy
make ps

# 4. Seed the database with 15 products
make seed

# 5. Generate continuous traffic so dashboards show data
make load-test        # runs until Ctrl-C

# 6. Open Grafana
open http://localhost:3001   # admin / admin123
```

**All URLs after startup:**

| URL | Service |
|-----|---------|
| http://localhost:3000 | Frontend (React app) |
| http://localhost:8080/api/v1 | Backend REST API |
| http://localhost:8080/metrics | Raw Prometheus metrics |
| http://localhost:9090 | Prometheus UI |
| http://localhost:3001 | Grafana (admin / admin123) |
| http://localhost:3100 | Loki (HTTP API) |
| http://localhost:9080 | Promtail status page |
| http://localhost:9093 | AlertManager UI |
| http://localhost:8081 | cAdvisor container view |
| http://localhost:9090/targets | Prometheus scrape targets |
| http://localhost:9090/alerts | Prometheus active alerts |

---

## Deploy to AWS EC2

The full stack runs on a **single EC2 instance** using Docker Compose — a direct lift-and-shift of the local setup. All deployment files live in `deploy/aws/`.

### Architecture on AWS

```
                        Internet
                           │
                    ┌──────▼──────┐
                    │  Elastic IP │
                    └──────┬──────┘
                           │ :80 / :443
                    ┌──────▼──────────────────────────────┐
                    │  EC2  t3.large  (Ubuntu 22.04)       │
                    │                                      │
                    │  ┌──────────────────────────────┐    │
                    │  │  Host nginx  (port 80)        │    │
                    │  │  /           → frontend :3000 │    │
                    │  │  /api/       → backend  :8080 │    │
                    │  │  /grafana/   → grafana  :3001 │◄── basic auth
                    │  │  /prometheus/→ prom     :9090 │◄── basic auth
                    │  │  /alertmgr/  → alertmgr :9093 │◄── basic auth
                    │  └──────────────────────────────┘    │
                    │  Docker containers (all internal)     │
                    └───────────────────┬─────────────────-┘
                                        │ :5432
                    ┌───────────────────▼──────────────┐
                    │  RDS PostgreSQL  (optional)       │
                    │  private subnet — never public    │
                    └──────────────────────────────────┘
```

### Files

| File | Purpose |
|------|---------|
| `deploy/aws/terraform/main.tf` | VPC, EC2, RDS, IAM, SSM Parameters |
| `deploy/aws/terraform/variables.tf` | All input variables with descriptions |
| `deploy/aws/terraform/outputs.tf` | Public IP, SSH command, service URLs |
| `deploy/aws/terraform/terraform.tfvars.example` | Template — copy to `terraform.tfvars` |
| `deploy/aws/ec2-userdata.sh` | Bootstrap: installs Docker, clones repo, starts stack |
| `deploy/aws/docker-compose.prod.yml` | Production overrides (no host ports, resource limits) |
| `deploy/aws/nginx/monitoring-proxy.conf` | Host nginx reverse proxy with basic auth |

### Prerequisites

| Tool | Install |
|------|---------|
| [Terraform](https://developer.hashicorp.com/terraform/install) ≥ 1.6 | `brew install hashicorp/tap/terraform` |
| [AWS CLI](https://aws.amazon.com/cli/) v2 | `brew install awscli` |
| AWS Key Pair | EC2 → Key Pairs → **Create key pair** (download the `.pem` file) |

```bash
# Configure AWS credentials
aws configure
# AWS Access Key ID:     <your key>
# AWS Secret Access Key: <your secret>
# Default region:        us-east-1
# Default output format: json
```

### Step 1 — Configure Terraform variables

```bash
cd deploy/aws/terraform

cp terraform.tfvars.example terraform.tfvars
```

Edit `terraform.tfvars`:

```hcl
project       = "prom-stack"
aws_region    = "us-east-1"
aws_profile   = "default"           # AWS CLI profile from ~/.aws/credentials
instance_type = "t3.large"          # 2 vCPU / 8 GB RAM — recommended

operator_cidr = "0.0.0.0/0"        # restrict to your IP for production:
                                    # curl -s ifconfig.me → use <your-ip>/32

# Name of your existing AWS Key Pair (EC2 → Key Pairs in the AWS console)
key_pair_name = "your-key-pair-name"

postgres_password = "StrongPassword123!"
grafana_password  = "AdminPassword456!"

use_rds              = false  # true = RDS PostgreSQL (~+$15/mo, managed backups)
create_iam_role      = false  # true = IAM role for SSM + CloudWatch (requires iam:CreateRole)
store_secrets_in_ssm = false  # true = store passwords in SSM Parameter Store (requires ssm:PutParameter)
```

> `terraform.tfvars` is in `.gitignore` — never commit it.

### Step 2 — Deploy

```bash
terraform init
terraform plan      # review what will be created
terraform apply     # type "yes" — takes ~3 min
```

Terraform provisions: VPC · subnets · security groups · EC2 · Elastic IP · (optional IAM role, SSM parameters, RDS).

The EC2 instance runs `ec2-userdata.sh` automatically on first boot:
1. Installs Docker Engine + Docker Compose plugin
2. Clones `https://github.com/ismailrz/prometheus-monitoring-grafana`
3. Writes `.env` from the injected credentials
4. Starts all containers with the production compose override
5. Configures host nginx as a reverse proxy with basic auth

**First boot takes ~5–8 minutes.** Watch progress:

```bash
ssh -i ~/Downloads/your-key-pair-name.pem ubuntu@$(terraform output -raw instance_public_ip) "tail -f /var/log/userdata.log"
```

### Step 3 — Access the stack

```bash
terraform output instance_public_ip   # get the IP
```

| URL | Service | Auth |
|-----|---------|------|
| `http://<IP>` | React frontend | — |
| `http://<IP>/api/v1/products` | Backend API | — |
| `http://<IP>/grafana/` | Grafana dashboards | admin / *grafana_password* |
| `http://<IP>/prometheus/` | Prometheus UI | admin / *grafana_password* |
| `http://<IP>/alertmanager/` | AlertManager | admin / *grafana_password* |

### Deploy without Terraform (manual EC2)

If the EC2 instance is already running and you want to start the stack manually:

```bash
# SSH into your instance using the .pem key downloaded from the AWS console
ssh -i ~/Downloads/your-key-pair-name.pem ubuntu@<IP>

# Install Docker (if not already installed)
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg \
  | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
sudo chmod a+r /etc/apt/keyrings/docker.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] \
  https://download.docker.com/linux/ubuntu \
  $(. /etc/os-release && echo "$VERSION_CODENAME") stable" \
  | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt-get update -y
sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
sudo systemctl enable --now docker
sudo usermod -aG docker ubuntu

# Re-login so the docker group takes effect
newgrp docker

# Clone the repo
git clone https://github.com/ismailrz/prometheus-monitoring-grafana.git
cd prometheus-monitoring-grafana

# Create .env
cat > .env <<EOF
POSTGRES_DB=appdb
POSTGRES_USER=postgres
POSTGRES_PASSWORD=your_strong_password
GRAFANA_ADMIN_USER=admin
GRAFANA_ADMIN_PASSWORD=your_strong_password
EOF

# Start the stack
docker compose up -d

# Check all containers are healthy
docker compose ps
```

### Update after a code push

```bash
ssh -i ~/Downloads/your-key-pair-name.pem ubuntu@<IP>
cd ~/prometheus-monitoring-grafana

git pull
docker compose up -d --build
```

### Cost estimate

| Configuration | Monthly cost (us-east-1) |
|---------------|--------------------------|
| t3.medium (4 GB) + 40 GB EBS | ~$33 |
| t3.large (8 GB) + 40 GB EBS | ~$63 |
| t3.large + RDS db.t3.micro | ~$78 |

Use a **1-year Reserved Instance** to cut EC2 cost by ~40%.

### Add HTTPS (recommended)

Point a domain at your Elastic IP, then:

```bash
ssh ubuntu@<IP>
sudo apt install -y certbot python3-certbot-nginx
sudo certbot --nginx -d yourdomain.com
```

Certbot auto-configures nginx for TLS and sets up auto-renewal.

### Teardown

```bash
cd deploy/aws/terraform
terraform destroy   # removes all AWS resources
```

---

## Project Structure

```
prometheus-monitoring-grafana/
│
├── docker-compose.yml          # Orchestrates all 13 services
├── .env                        # Credentials (not committed in production)
├── Makefile                    # Convenience targets
│
├── backend/                    # Go REST API
│   ├── cmd/server/main.go      # Entry point, server setup, graceful shutdown
│   ├── internal/
│   │   ├── config/             # Env-based configuration
│   │   ├── database/           # Connection pool, schema migration, stats collector
│   │   ├── handlers/           # HTTP handlers: products, orders, health
│   │   ├── middleware/         # Prometheus metrics middleware, structured logger
│   │   └── models/             # Data models + repository pattern (products, orders)
│   ├── pkg/metrics/            # All Prometheus metric definitions (single source of truth)
│   ├── go.mod
│   └── Dockerfile              # Multi-stage build → distroless runtime image
│
├── frontend/                   # React SPA
│   ├── src/
│   │   ├── App.jsx             # Tab navigation (Products / Orders / Health)
│   │   ├── components/
│   │   │   ├── ProductList.jsx # CRUD product UI
│   │   │   ├── OrderList.jsx   # Create & view orders
│   │   │   └── HealthStatus.jsx# Service health + quick links
│   │   └── services/api.js     # Typed fetch wrapper for all backend endpoints
│   ├── nginx/nginx.conf        # Serves SPA, proxies /api, exposes /nginx_status
│   └── Dockerfile              # Multi-stage: Vite build → nginx:alpine
│
├── monitoring/
│   ├── prometheus/
│   │   ├── prometheus.yml      # Global config + all scrape targets
│   │   └── rules/
│   │       ├── app_alerts.yml  # Application-level alert rules
│   │       └── infra_alerts.yml# Infrastructure + Go runtime alerts
│   │
│   ├── grafana/
│   │   ├── provisioning/
│   │   │   ├── datasources/
│   │   │   │   ├── prometheus.yml  # Auto-provisions Prometheus datasource
│   │   │   │   └── loki.yml        # Auto-provisions Loki datasource
│   │   │   └── dashboards/
│   │   │       └── dashboards.yml  # Points Grafana at /var/lib/grafana/dashboards
│   │   └── dashboards/
│   │       ├── app-overview.json   # HTTP metrics dashboard
│   │       ├── go-runtime.json     # Go memory, GC, goroutines
│   │       ├── postgres.json       # PostgreSQL health & performance
│   │       ├── infrastructure.json # Host + container resources
│   │       └── logs.json           # Loki log explorer dashboard
│   │
│   ├── loki/
│   │   └── loki.yml            # Single-binary Loki, tsdb schema v13, 7-day retention
│   │
│   ├── promtail/
│   │   └── promtail.yml        # Docker socket discovery, level extraction, noise filter
│   │
│   ├── alertmanager/
│   │   └── alertmanager.yml    # Routes, receivers, inhibition rules
│   │
│   └── postgres-exporter/
│       └── queries.yml         # Custom SQL: table sizes, dead tuples, index usage, long queries
│
└── scripts/
    ├── init-db.sql             # Seeds 15 products on first DB start
    ├── seed-data.sh            # Creates 10 sample orders via API
    └── load-test.sh            # Continuous realistic traffic generator
```

---

## Application Layer

### Frontend (React + nginx)

A single-page app with three views:

| View | Description |
|------|-------------|
| **Products** | List, filter by category, create, delete products |
| **Orders** | Place orders (picks from live product list), view history |
| **Health** | Service health checks + quick links to all monitoring tools |

nginx is configured to:
- Serve the React build with proper SPA fallback (`try_files $uri /index.html`)
- Proxy `/api/` requests upstream to the Go backend with keepalive
- Expose `/nginx_status` for the nginx-prometheus-exporter
- Cache static assets with immutable headers (1 year)
- Add security headers (`X-Frame-Options`, `X-Content-Type-Options`)

### Backend (Go)

Built with [Gin](https://github.com/gin-gonic/gin). Production settings:

- **Graceful shutdown** — 30-second drain window on `SIGTERM`/`SIGINT`
- **Connection pool** — 25 max open, 10 max idle, 5-minute lifetime
- **Metrics middleware** — uses `c.FullPath()` (route template, not real URL) to prevent high-cardinality label explosion from UUIDs in paths
- **DB stats goroutine** — publishes pool stats (open/in-use/idle) to Prometheus every 15 seconds
- **Structured logger** — one-line `key=value` format, parsed by Promtail into log labels

Runtime image is **distroless/static** — no shell, no package manager, minimal attack surface.

### Database (PostgreSQL)

Schema (auto-migrated on backend startup):

```sql
products (id UUID PK, name, description, price, stock, category, created_at, updated_at)
orders   (id UUID PK, customer_email, status, total_amount, created_at, updated_at)
order_items (id UUID PK, order_id FK, product_id FK, quantity, unit_price, created_at)
```

Indexes: `products.category`, `orders.status`, `orders.customer_email`, `order_items.order_id`

`scripts/init-db.sql` seeds 15 products across 6 categories (laptops, phones, audio, tablets, wearables, electronics) on the first container start.

---

## Monitoring Layer

### Prometheus

**Config:** `monitoring/prometheus/prometheus.yml`

Scrape interval: **15s** globally, **10s** for the backend.

| Job | Target | Key metrics |
|-----|--------|-------------|
| `backend` | backend:8080 | `http_requests_total`, `http_request_duration_seconds`, `go_*`, `db_*` |
| `node-exporter` | node-exporter:9100 | `node_cpu_*`, `node_memory_*`, `node_disk_*`, `node_network_*` |
| `postgres-exporter` | postgres-exporter:9187 | `pg_stat_*`, `pg_locks_*`, `pg_table_size_*` |
| `cadvisor` | cadvisor:8080 | `container_cpu_*`, `container_memory_*` |
| `nginx-exporter` | nginx-exporter:9113 | `nginx_connections_*`, `nginx_http_requests_total` |
| `prometheus` | localhost:9090 | self-monitoring |
| `grafana` | grafana:3000 | self-monitoring |

**Useful Prometheus queries to explore:**

```promql
# Request rate by endpoint
sum by (method, path) (rate(http_requests_total[5m]))

# P99 latency
histogram_quantile(0.99, sum by (le) (rate(http_request_duration_seconds_bucket[5m])))

# Error rate percentage
100 * sum(rate(http_requests_total{status_code=~"5.."}[5m])) / sum(rate(http_requests_total[5m]))

# DB query p95 by operation
histogram_quantile(0.95, sum by (le, operation, table) (rate(db_query_duration_seconds_bucket[5m])))

# Container memory usage
container_memory_usage_bytes{name!=""}

# Host CPU usage
100 - (avg(rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)
```

### Grafana Dashboards

Grafana starts with **5 pre-provisioned dashboards** in the *Prometheus Stack* folder. No manual import needed.

#### 1. Application Overview
> Core RED metrics for the Go backend.

| Panel | Query |
|-------|-------|
| Request Rate | `sum(rate(http_requests_total[5m]))` |
| Error Rate (5xx) | `sum(rate(...{status_code=~"5.."}[5m])) / sum(rate(...[5m]))` |
| P95 Latency | `histogram_quantile(0.95, ...)` |
| In-Flight Requests | `http_requests_in_flight` |
| Request Rate by Status Code | Time series split by `status_code` label |
| Response Time Percentiles | p50 / p90 / p95 / p99 on one chart |
| Request Rate by Endpoint | Split by `method` + `path` |
| Connections Overview | In-flight HTTP + DB pool in-use/idle |
| DB Query Latency P95 | Per `operation` (select/insert/update) and `table` |
| Business Metrics | Products created, orders created, business errors (hourly rate) |

#### 2. Go Runtime
> Go process internals — useful for spotting memory leaks and GC pressure.

| Panel | What to look for |
|-------|-----------------|
| Goroutines | Steady growth = goroutine leak |
| GC Pause p90 | Spikes > 50ms = GC pressure, consider `GOGC` tuning |
| Heap Allocated | Saw-tooth pattern is normal GC behaviour |
| Memory Usage | Compare `heap_alloc` vs `heap_sys` — large gap = GC not keeping up |
| GC Duration / Cycles | High cycles/s with high pause = memory-intensive workload |
| Process CPU | Sustained 100% = CPU-bound, check goroutine count |

#### 3. PostgreSQL
> Database performance and health metrics from `postgres_exporter`.

| Panel | What to look for |
|-------|-----------------|
| Active Connections | Approaching `max_connections` (default 100)? Consider pgBouncer |
| Transactions per Second | Baseline + spikes |
| Cache Hit Ratio | Below 95% = consider raising `shared_buffers` |
| Long-Running Queries | Any value > 0 deserves investigation |
| Connections by State | `idle in transaction` means forgotten transactions |
| Dead Tuples | High values = autovacuum not keeping up |
| Locks by Mode | `ExclusiveLock` spikes = table contention |

#### 4. Infrastructure
> Host and container resource usage from node-exporter and cAdvisor.

| Panel | Notes |
|-------|-------|
| Host CPU / Memory / Disk | Aggregate host stats |
| Container CPU Usage | Per-container, overlaid. Helps identify noisy neighbours |
| Container Memory Usage | Spot containers approaching their limits |
| Host Network I/O | Bytes/s rx and tx per interface |
| Host Disk I/O | Read/write bytes/s per device |

#### 5. Logs
> Log explorer powered by Loki — correlated with metrics via shared time range.

| Panel | LogQL |
|-------|-------|
| Log Volume by Container | `sum by (container) (count_over_time(...))` |
| Log Volume by Level | Split by `level` label (info/warn/error) |
| stdout vs stderr | Split by `stream` label |
| Live Log Stream | `{service=~"$service"} \|= "$search"` — full-text search |
| Backend Logs | Filtered to `container="backend"` |
| PostgreSQL Logs | Filtered to `container="postgres"` |
| Errors & Panics | Regex filter `(?i)(error\|fatal\|panic\|critical)` |

The **Service** dropdown and **Search** text box let you narrow logs interactively.

### Loki & Promtail

**Loki** runs in single-binary/monolithic mode — appropriate for local and small-team environments.

- Schema: tsdb v13 (recommended for Loki 2.8+)
- Retention: 7 days
- Storage: local filesystem at `/loki`

**Promtail** uses Docker socket discovery (`docker_sd_configs`) — no static container list needed. It automatically picks up new containers. Labels attached to every log line:

| Label | Value example | Source |
|-------|--------------|--------|
| `container` | `backend` | `__meta_docker_container_name` |
| `service` | `backend` | `com.docker.compose.service` label |
| `stream` | `stdout` | Docker log stream |
| `image` | `prometheus-stack-backend` | Container image name |
| `level` | `info` / `warn` / `error` | Extracted from log content by pipeline |

Health-check log lines (`GET /health`, `GET /ready`, `GET /nginx_status`) are **dropped at the Promtail pipeline** stage to keep Loki storage clean.

**Useful LogQL queries:**

```logql
# All backend logs
{container="backend"}

# Error logs across all containers
{service=~".+"} |~ "(?i)(error|fatal|panic)"

# Backend request logs for a specific path
{container="backend"} |= "/api/v1/products"

# Log rate per container over time
sum by (container) (rate({service=~".+"}[5m]))

# Only error-level logs
{service=~".+"} | label_format level=level | level="error"
```

### AlertManager

**Config:** `monitoring/alertmanager/alertmanager.yml`

Alerts are visible in the AlertManager UI at http://localhost:9093. To wire up real notifications, edit the `receivers` section:

```yaml
# Slack example
receivers:
  - name: 'critical-receiver'
    slack_configs:
      - api_url: 'https://hooks.slack.com/services/YOUR/WEBHOOK/URL'
        channel: '#alerts'
        title: '{{ .GroupLabels.alertname }}'
        text: '{{ range .Alerts }}{{ .Annotations.description }}{{ end }}'
```

**Inhibition rules** prevent alert storms:
- If `BackendDown` fires → suppresses `HighErrorRate`, `HighP95Latency`, `NoIncomingTraffic`
- If `PostgresDown` fires → suppresses `PostgresHighConnections`

---

## Exporters

### Node Exporter (:9100)
Exposes OS-level metrics: CPU per core, memory breakdown, filesystem usage, network I/O, disk I/O, system load, process count. Mounts `/proc`, `/sys`, and `/rootfs` read-only.

### Postgres Exporter (:9187)
Uses the standard `pg_stat_*` views plus four **custom queries** defined in `monitoring/postgres-exporter/queries.yml`:

| Query | Metrics | Purpose |
|-------|---------|---------|
| `pg_table_size` | `total_bytes`, `heap_bytes`, `index_bytes` per table | Track table growth |
| `pg_long_running_queries` | count of queries > 30s | Alert on stuck transactions |
| `pg_dead_tuples` | `dead_tuples`, `live_tuples`, `dead_ratio` per table | VACUUM health |
| `pg_index_usage` | `scans`, `tuples_read`, `tuples_fetched` per index | Identify unused indexes |

### cAdvisor (:8081)
Reads Docker container cgroups. Provides per-container CPU, memory, network, and filesystem metrics. The `container_name` label maps to the Docker container name.

### Nginx Exporter (:9113)
Scrapes nginx's `stub_status` page (`/nginx_status`) and exports:
- `nginx_connections_active`
- `nginx_connections_reading` / `writing` / `waiting`
- `nginx_http_requests_total`

---

## API Reference

Base URL: `http://localhost:8080/api/v1`

### Products

| Method | Path | Description | Body |
|--------|------|-------------|------|
| `GET` | `/products` | List all products | — |
| `GET` | `/products?category=laptops` | Filter by category | — |
| `GET` | `/products/:id` | Get product by UUID | — |
| `POST` | `/products` | Create product | `{"name","price","stock","category","description"}` |
| `PUT` | `/products/:id` | Update product | same as POST |
| `DELETE` | `/products/:id` | Delete product | — |

**Create product example:**
```bash
curl -X POST http://localhost:8080/api/v1/products \
  -H "Content-Type: application/json" \
  -d '{"name":"ThinkPad X1","price":1499.99,"stock":20,"category":"laptops","description":"14-inch business laptop"}'
```

### Orders

| Method | Path | Description | Body |
|--------|------|-------------|------|
| `GET` | `/orders` | List orders (latest 100) | — |
| `GET` | `/orders?status=pending` | Filter by status | — |
| `GET` | `/orders/:id` | Get order with items | — |
| `POST` | `/orders` | Create order | `{"customer_email","items":[{"product_id","quantity"}]}` |

**Create order example:**
```bash
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{"customer_email":"user@example.com","items":[{"product_id":"<uuid>","quantity":2}]}'
```

Order status values: `pending` → `confirmed` → `shipped` → `delivered` / `cancelled`

### Health Endpoints

| Path | Returns | Used by |
|------|---------|---------|
| `GET /health` | `{"status":"ok"}` | Docker healthcheck, load balancer |
| `GET /ready` | `{"status":"ready"}` or 503 | Kubernetes readiness probe |
| `GET /metrics` | Prometheus text format | Prometheus scraper |

---

## Metrics Reference

All custom metrics are defined in `backend/pkg/metrics/metrics.go`.

### HTTP Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `http_requests_total` | Counter | `method`, `path`, `status_code` | Total requests processed |
| `http_request_duration_seconds` | Histogram | `method`, `path` | Request latency (11 buckets from 5ms to 10s) |
| `http_requests_in_flight` | Gauge | — | Requests currently being processed |
| `http_response_size_bytes` | Histogram | `method`, `path` | Response body size |

> `path` uses the **route template** (e.g. `/api/v1/products/:id`), not the real URL. This prevents label cardinality explosion from UUIDs.

### Database Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `db_query_duration_seconds` | Histogram | `operation`, `table` | Query execution time |
| `db_query_errors_total` | Counter | `operation`, `table` | Failed queries |
| `db_connections_open` | Gauge | — | Total open connections in pool |
| `db_connections_in_use` | Gauge | — | Connections executing a query |
| `db_connections_idle` | Gauge | — | Idle connections in pool |

### Business Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `products_created_total` | Counter | — | Lifetime product count |
| `products_deleted_total` | Counter | — | Lifetime delete count |
| `orders_created_total` | Counter | — | Lifetime order count |
| `order_value_total` | Counter | — | Cumulative order revenue |
| `business_errors_total` | Counter | `type` | Logic errors (e.g. `product_not_found`) |
| `app_info` | Gauge | `version`, `env` | Build metadata (always 1) |

### Go Runtime Metrics (auto-exposed)

| Metric | Description |
|--------|-------------|
| `go_goroutines` | Current goroutine count |
| `go_memstats_heap_alloc_bytes` | Heap bytes in use |
| `go_memstats_heap_sys_bytes` | Heap bytes obtained from OS |
| `go_gc_duration_seconds` | Stop-the-world GC pause histogram |
| `process_cpu_seconds_total` | Total CPU time |
| `process_open_fds` | Open file descriptors |
| `process_resident_memory_bytes` | RSS memory |

---

## Alert Rules

### Application Alerts (`monitoring/prometheus/rules/app_alerts.yml`)

| Alert | Condition | Severity | For |
|-------|-----------|----------|-----|
| `HighErrorRate` | 5xx rate > 5% | critical | 2m |
| `HighP95Latency` | p95 latency > 1s | warning | 5m |
| `BackendDown` | `up{job="backend"} == 0` | critical | 1m |
| `NoIncomingTraffic` | Zero requests for 10m | warning | 10m |
| `HighInFlightRequests` | In-flight > 100 | warning | 2m |
| `HighDatabaseErrorRate` | DB errors > 0.5/s | warning | 2m |
| `SlowDatabaseQueries` | DB p95 > 500ms | warning | 5m |
| `PostgresDown` | `pg_up == 0` | critical | 1m |
| `PostgresHighConnections` | Connections > 80% of max | warning | 5m |
| `PostgresLowCacheHitRatio` | Cache hit < 90% | warning | 10m |

### Infrastructure Alerts (`monitoring/prometheus/rules/infra_alerts.yml`)

| Alert | Condition | Severity |
|-------|-----------|----------|
| `HighCPUUsage` | CPU > 85% for 10m | warning |
| `HighMemoryUsage` | Memory > 90% for 5m | critical |
| `DiskSpaceLow` | Disk < 15% free | warning |
| `DiskSpaceCritical` | Disk < 5% free | critical |
| `ContainerHighCPU` | Container > 0.8 cores for 10m | warning |
| `ContainerHighMemory` | Container > 90% of memory limit | warning |
| `HighGoroutineCount` | Goroutines > 1000 for 5m | warning |
| `HighHeapMemory` | Heap > 500 MiB for 5m | warning |

---

## Load Testing

`scripts/load-test.sh` generates a realistic mix of API traffic:

| Traffic type | Weight | Description |
|--------------|--------|-------------|
| List products | 40% | `GET /api/v1/products` |
| Get single product | 20% | `GET /api/v1/products/:id` |
| List orders | 10% | `GET /api/v1/orders` |
| Create product | 10% | `POST /api/v1/products` |
| Create order | 10% | `POST /api/v1/orders` |
| 404 (unknown ID) | 10% | Generates visible error traffic |

```bash
make load-test                      # default 0.3s between requests (~3 req/s)
INTERVAL=0.05 make load-test        # 20 req/s
INTERVAL=0.01 make load-test        # ~100 req/s (stress test)
```

After ~2 minutes you will see populated charts in all Grafana dashboards.

---

## Makefile Commands

```bash
make up              # Build images and start all containers detached
make down            # Stop and remove containers (keeps volumes)
make build           # Rebuild images without starting
make logs            # Follow all container logs
make logs-backend    # Follow backend logs only
make logs-prometheus # Follow Prometheus logs only
make ps              # Show container status table
make clean           # Stop containers AND delete all volumes (full reset)
make seed            # Create 10 sample orders via the API
make load-test       # Run continuous traffic generator
make reload-prometheus  # Hot-reload Prometheus config without restart
make fmt             # Run gofmt over the Go backend
make urls            # Print all service URLs
```

---

## Troubleshooting

### Port already in use
```bash
# Find which process holds a port
lsof -i :3000 -sTCP:LISTEN

# Kill it
kill $(lsof -ti :3000)
```

### Backend can't connect to PostgreSQL at startup
Docker Compose waits for the `pg_isready` healthcheck before starting the backend. If the DB takes longer than expected:
```bash
docker compose logs db           # check for PostgreSQL startup errors
docker compose restart backend   # restart after DB is ready
```

### No data in Grafana dashboards
1. Confirm Prometheus is scraping: http://localhost:9090/targets — all targets should show **UP** in green.
2. Run `make load-test` to generate traffic.
3. Wait one full scrape cycle (15s) then refresh the dashboard.

### No logs in Loki / Logs dashboard
1. Check Promtail can reach the Docker socket:
   ```bash
   docker compose logs promtail
   ```
2. Check Loki is ready:
   ```bash
   curl http://localhost:3100/ready
   ```
3. Verify the Loki datasource in Grafana: **Connections → Data Sources → Loki → Test**.

### Grafana shows "No data"
- Ensure the time range includes the period when containers were running (default: last 1 hour).
- Check the datasource variable at the top of each dashboard matches your provisioned datasource.

### Full reset
```bash
make clean   # removes all containers + named volumes
make up      # fresh start
make seed
```

### Prometheus config validation
```bash
docker run --rm \
  -v $(pwd)/monitoring/prometheus:/etc/prometheus \
  prom/prometheus:v2.51.0 \
  promtool check config /etc/prometheus/prometheus.yml
```

### Go backend won't compile
The Dockerfile uses `-mod=mod` so `go.sum` is generated inside the build container. If you see a compile error locally:
```bash
cd backend
go mod tidy    # requires internet access via proxy.golang.org
go build ./...
```
