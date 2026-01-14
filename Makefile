.PHONY: help build run stop clean restart logs test docker-build docker-up docker-down docker-restart docker-logs docker-clean dev prod

# Default target
help:
	@echo "CentralHub Makefile Commands:"
	@echo ""
	@echo "Development:"
	@echo "  make dev              - Start development environment with Docker Compose"
	@echo "  make logs             - View application logs"
	@echo "  make restart          - Restart the application"
	@echo "  make stop             - Stop all containers"
	@echo ""
	@echo "Build:"
	@echo "  make build            - Build the Go application locally"
	@echo "  make docker-build     - Build Docker image"
	@echo ""
	@echo "Docker Compose:"
	@echo "  make docker-up        - Start all services"
	@echo "  make docker-down      - Stop and remove all services"
	@echo "  make docker-restart   - Restart all services"
	@echo "  make docker-logs      - View all service logs"
	@echo "  make docker-clean     - Clean up containers, volumes, and images"
	@echo ""
	@echo "Testing:"
	@echo "  make test             - Run tests"
	@echo ""
	@echo "Cleanup:"
	@echo "  make clean            - Clean build artifacts"

# Local build
build:
	@echo "Building CentralHub..."
	go build -o centralhub .
	@echo "Build complete: ./centralhub"

# Run locally (without Docker)
run:
	@echo "Running CentralHub locally..."
	go run main.go

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f centralhub
	rm -rf logs/*.log
	@echo "Clean complete"

# Docker build
docker-build:
	@echo "Building Docker image..."
	docker build -t centralhub:latest .
	@echo "Docker image built: centralhub:latest"

# Start development environment
dev: docker-up

# Start production environment
prod:
	@echo "Starting production environment..."
	docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
	@echo "Production environment started"

# Docker Compose: Start all services
docker-up:
	@echo "Starting all services..."
	docker-compose up -d
	@echo "Services started. Access app at http://localhost:8080"

# Docker Compose: Stop all services
docker-down: stop

stop:
	@echo "Stopping all services..."
	docker-compose down
	@echo "Services stopped"

# Docker Compose: Restart all services
docker-restart: restart

restart:
	@echo "Restarting all services..."
	docker-compose restart
	@echo "Services restarted"

# Docker Compose: View logs
docker-logs: logs

logs:
	docker-compose logs -f

# Docker Compose: View app logs only
logs-app:
	docker-compose logs -f app

# Docker Compose: View MongoDB logs only
logs-db:
	docker-compose logs -f mongodb

# Docker Compose: Clean up everything
docker-clean:
	@echo "Cleaning up Docker resources..."
	docker-compose down -v --rmi local
	docker system prune -f
	@echo "Docker cleanup complete"

# Setup: Create config file from example
setup-config:
	@if [ ! -f config.yaml ]; then \
		cp config/config.dev.yaml config.yaml; \
		echo "Created config.yaml from config.dev.yaml"; \
	else \
		echo "config.yaml already exists"; \
	fi

# Check service health
health:
	@echo "Checking service health..."
	@curl -s http://localhost:8080/health || echo "Service not responding"

# Shell into app container
shell-app:
	docker-compose exec app sh

# Shell into MongoDB container
shell-db:
	docker-compose exec mongodb mongosh -u admin -p admin123

# View running containers
ps:
	docker-compose ps
