#!/usr/bin/env bash
# EC2 bootstrap script — runs once on first start as root.
# Templated by Terraform: variables injected via templatefile().
set -euo pipefail

export DEBIAN_FRONTEND=noninteractive
LOG=/var/log/userdata.log
exec > >(tee -a "$LOG") 2>&1

echo "=== Bootstrap started $(date) ==="

# ─────────────────────────────────────────────────
# 1. System packages
# ─────────────────────────────────────────────────
apt-get update -y
apt-get install -y \
  ca-certificates curl gnupg git \
  nginx apache2-utils \
  awscli python3 python3-pip

# ─────────────────────────────────────────────────
# 2. Docker Engine
# ─────────────────────────────────────────────────
install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg \
  | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
chmod a+r /etc/apt/keyrings/docker.gpg

echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] \
  https://download.docker.com/linux/ubuntu \
  $(. /etc/os-release && echo "$VERSION_CODENAME") stable" \
  | tee /etc/apt/sources.list.d/docker.list > /dev/null

apt-get update -y
apt-get install -y \
  docker-ce docker-ce-cli containerd.io \
  docker-buildx-plugin docker-compose-plugin

systemctl enable docker
systemctl start docker
usermod -aG docker ubuntu

echo "Docker $(docker --version) installed."

# ─────────────────────────────────────────────────
# 3. Clone the repository
# ─────────────────────────────────────────────────
APP_DIR=/opt/prometheus-stack
git clone https://github.com/YOUR_USERNAME/prometheus-monitoring-grafana.git "$APP_DIR" \
  || (cd "$APP_DIR" && git pull)

# ─────────────────────────────────────────────────
# 4. Write .env from Terraform-injected values
#    In production, prefer fetching from SSM (see below)
# ─────────────────────────────────────────────────
DB_HOST="${use_rds == "true" ? db_host : "db"}"

cat > "$APP_DIR/.env" <<EOF
POSTGRES_DB=${postgres_db}
POSTGRES_USER=${postgres_user}
POSTGRES_PASSWORD=${postgres_password}
GRAFANA_ADMIN_USER=admin
GRAFANA_ADMIN_PASSWORD=${grafana_password}
DATABASE_URL=postgresql://${postgres_user}:${postgres_password}@$DB_HOST:5432/${postgres_db}?sslmode=disable
EOF

chmod 600 "$APP_DIR/.env"

# Optional: fetch secrets from SSM Parameter Store instead
# POSTGRES_PASSWORD=$(aws ssm get-parameter \
#   --name "/${project}/postgres_password" \
#   --with-decryption --query Parameter.Value --output text --region ${aws_region})

# ─────────────────────────────────────────────────
# 5. Start the stack
# ─────────────────────────────────────────────────
cd "$APP_DIR"

COMPOSE_CMD="docker compose -f docker-compose.yml -f deploy/aws/docker-compose.prod.yml"

# Pull images first so startup is faster
$COMPOSE_CMD pull --quiet

# Start all services (excluding containerised db if using RDS)
if [ "${use_rds}" = "true" ]; then
  $COMPOSE_CMD up -d --scale db=0
else
  $COMPOSE_CMD up -d
fi

echo "Stack started. Waiting 30s for services to initialise..."
sleep 30

# Seed the database
bash "$APP_DIR/scripts/seed-data.sh" 2>/dev/null || true

# ─────────────────────────────────────────────────
# 6. nginx reverse proxy — exposes monitoring UIs
#    on paths under port 80/443 with basic auth
# ─────────────────────────────────────────────────
HTPASSWD_FILE=/etc/nginx/.htpasswd

# Create admin user for monitoring paths
htpasswd -cb "$HTPASSWD_FILE" admin "${grafana_password}"

cp "$APP_DIR/deploy/aws/nginx/monitoring-proxy.conf" \
   /etc/nginx/sites-available/monitoring

ln -sf /etc/nginx/sites-available/monitoring \
       /etc/nginx/sites-enabled/monitoring

rm -f /etc/nginx/sites-enabled/default

nginx -t && systemctl reload nginx

echo "=== Bootstrap complete $(date) ==="
echo "App:      http://$(curl -s ifconfig.me)"
echo "Grafana:  http://$(curl -s ifconfig.me)/grafana"
