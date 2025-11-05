# KivaPlus Backend - Lambda Functions
# ===================================

.PHONY: help build-all build-auth build-users build-condominiums clean test

# Variables
BINARY_NAME=main

help: ## Show available commands
	@echo "KivaPlus Backend - Lambda Functions"
	@echo "=================================="
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build-all: build-auth build-users build-condominiums ## Build all Lambda functions

build-auth: ## Build auth Lambda (register, login, refresh)
	@echo "🔨 Building auth Lambda..."
	cd cmd/auth && GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME) .
	cd cmd/auth && zip auth.zip $(BINARY_NAME) && mv auth.zip ../../dist/
	@echo "✅ Auth Lambda built"

build-users: ## Build users Lambda (profile, users management)
	@echo "🔨 Building users Lambda..."
	cd cmd/users && GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME) .
	cd cmd/users && zip users.zip $(BINARY_NAME) && mv users.zip ../../dist/
	@echo "✅ Users Lambda built"

build-condominiums: ## Build condominiums Lambda (condominiums CRUD)
	@echo "🔨 Building condominiums Lambda..."
	cd cmd/condominiums && GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME) .
	cd cmd/condominiums && zip condominiums.zip $(BINARY_NAME) && mv condominiums.zip ../../dist/
	@echo "✅ Condominiums Lambda built"

clean: ## Clean build artifacts
	@echo "🧹 Cleaning build artifacts..."
	rm -f $(BINARY_NAME) function.zip
	rm -f cmd/*/$(BINARY_NAME) cmd/*/*.zip
	rm -f dist/*.zip
	@echo "✅ Clean completed"

test: ## Run unit tests
	@echo "🧪 Running unit tests..."
	go test ./internal/... -short
	@echo "✅ Unit tests completed"

test-integration: ## Run integration tests
	@echo "🧪 Running integration tests..."
	go test -v ./tests/integration -timeout 60s
	@echo "✅ Integration tests completed"

test-all: test test-integration ## Run all tests

# Docker commands for backend repository
setup: ## Complete setup for backend repository
	@echo "🏗️  Setting up KivaPlus Backend environment..."
	@echo "1️⃣  Starting core services..."
	docker compose up -d postgres redis
	@echo "⏳ Waiting for services to start..."
	@sleep 15
	@echo "2️⃣  Building Lambda functions..."
	docker compose --profile build run --rm lambda-builder
	@echo "3️⃣  Starting LocalStack (optional)..."
	docker compose --profile localstack up -d localstack
	@echo "⏳ Waiting for LocalStack..."
	@sleep 20
	@echo "4️⃣  Creating AWS resources (if LocalStack is running)..."
	-docker compose --profile setup run --rm localstack-setup
	@echo "✅ Backend setup completed!"

docker-up: ## Start all services
	@echo "🐳 Starting all services..."
	docker compose up -d postgres redis
	@echo "✅ Core services started"

docker-up-full: ## Start all services including LocalStack
	@echo "🐳 Starting all services including LocalStack..."
	docker compose --profile localstack up -d
	@echo "✅ All services started"

docker-down: ## Stop all services
	@echo "🛑 Stopping all services..."
	docker compose down
	@echo "✅ All services stopped"

docker-logs: ## View logs
	docker compose logs -f

docker-status: ## Check service status
	@echo "📊 Service Status:"
	@docker compose ps
	@echo ""
	@echo "🔍 Health Checks:"
	@echo -n "PostgreSQL: "
	@docker compose exec -T postgres pg_isready -U kivaplus -d kivaplus_local > /dev/null 2>&1 && echo "✅ Ready" || echo "❌ Not Ready"
	@echo -n "Redis: "
	@docker compose exec -T redis redis-cli -a kivaplus123 ping > /dev/null 2>&1 && echo "✅ Ready" || echo "❌ Not Ready"

docker-clean: ## Clean all Docker resources
	@echo "🧹 Cleaning Docker resources..."
	docker compose down -v --remove-orphans
	docker system prune -f
	@echo "✅ Cleanup completed"

format: ## Format code
	@echo "🎨 Formatting code..."
	go fmt ./...
	@echo "✅ Code formatted"

.DEFAULT_GOAL := help
