.PHONY: help build clean deps proto up down logs restart migrate test lint format

# Variables
DOCKER_COMPOSE = docker-compose
GO = go
PROTOC = protoc

# Default target
help: ## Show this help message
	@echo "OtelLab - OpenTelemetry & Jaeger Learning Project"
	@echo ""
	@echo "Available commands:"
	@awk 'BEGIN {FS = ":.*##"; printf "\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Setup Commands
deps: ## Install Go dependencies
	@echo "Installing Go dependencies..."
	$(GO) mod download
	$(GO) mod tidy

proto: ## Generate Go code from protobuf files
	@echo "Generating protobuf files..."
	mkdir -p proto/task proto/user
	$(PROTOC) --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/task.proto proto/user.proto

##@ Development Commands
build: ## Build all services
	@echo "Building services..."
	$(GO) build -o bin/api-gateway ./api-gateway
	$(GO) build -o bin/task-service ./task-service
	$(GO) build -o bin/user-service ./user-service

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -rf tmp/
	$(DOCKER_COMPOSE) down --remove-orphans --volumes

##@ Docker Commands
up: ## Start all services with Docker Compose
	@echo "Starting all services..."
	$(DOCKER_COMPOSE) up -d

down: ## Stop all services
	@echo "Stopping all services..."
	$(DOCKER_COMPOSE) down

logs: ## View logs from all services
	$(DOCKER_COMPOSE) logs -f

restart: ## Restart all services
	@echo "Restarting services..."
	$(DOCKER_COMPOSE) restart

##@ Database Commands
migrate: ## Run database migrations
	@echo "Running database migrations..."
	$(DOCKER_COMPOSE) exec task-service echo "Migrations run automatically on startup"

##@ Development Tools
test: ## Run all tests
	@echo "Running tests..."
	$(GO) test -v -race -cover ./...

lint: ## Run linter
	@echo "Running linter..."
	golangci-lint run ./...

format: ## Format Go code
	@echo "Formatting code..."
	$(GO) fmt ./...
	$(GO) mod tidy

##@ Utility Commands
jaeger: ## Open Jaeger UI in browser
	@echo "Opening Jaeger UI..."
	open http://localhost:16686

health: ## Check health of all services
	@echo "Checking service health..."
	@curl -s http://localhost:8080/health | jq . || echo "API Gateway not responding"

demo: ## Run a quick demo of the system
	@echo "Creating a demo task..."
	@curl -X POST http://localhost:8080/api/users \
		-H "Content-Type: application/json" \
		-d '{"name":"Demo User","email":"demo@example.com"}' | jq .
	@echo ""
	@curl -X POST http://localhost:8080/api/tasks \
		-H "Content-Type: application/json" \
		-d '{"title":"Demo Task","description":"This is a demo task","assignee_id":"user-001"}' | jq .

install-tools: ## Install development tools
	@echo "Installing development tools..."
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GO) install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	$(GO) install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

##@ Quick Start
quick-start: deps proto up ## Quick start: install deps, generate proto, and start services
	@echo "🚀 OtelLab is starting up!"
	@echo "📊 Jaeger UI: http://localhost:16686"
	@echo "🌐 API Gateway: http://localhost:8080"
	@echo "❤️ Health Check: http://localhost:8080/health"
	@echo ""
	@echo "Waiting for services to be ready..."
	@sleep 10
	@make health
	@echo ""
	@echo "✅ Ready! Try running 'make demo' to create some test data."