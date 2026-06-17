.PHONY: up down build logs ps clean seed load-test reload-prometheus fmt

up:
	docker compose up -d --build

up-detach:
	docker compose up -d

down:
	docker compose down

build:
	docker compose build

logs:
	docker compose logs -f

logs-backend:
	docker compose logs -f backend

logs-prometheus:
	docker compose logs -f prometheus

ps:
	docker compose ps

clean:
	docker compose down -v --remove-orphans

seed:
	@echo "Seeding database..."
	@bash scripts/seed-data.sh

load-test:
	@echo "Running load test..."
	@bash scripts/load-test.sh

reload-prometheus:
	curl -s -X POST http://localhost:9090/-/reload && echo "Prometheus config reloaded"

fmt:
	cd backend && gofmt -w ./...

urls:
	@echo ""
	@echo "  Frontend:    http://localhost:3000"
	@echo "  Backend API: http://localhost:8080/api/v1"
	@echo "  Prometheus:  http://localhost:9090"
	@echo "  Grafana:     http://localhost:3001  (admin/admin123)"
	@echo "  AlertMgr:    http://localhost:9093"
	@echo "  Loki:        http://localhost:3100"
	@echo "  Promtail:    http://localhost:9080"
	@echo "  cAdvisor:    http://localhost:8081"
	@echo ""
