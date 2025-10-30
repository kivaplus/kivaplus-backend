# Hexagonal Architecture Makefile

# Variables
BINARY_NAME=main
AWS_REGION=us-east-1
DATABASE_URL ?= $(shell echo $DATABASE_URL)
MIGRATIONS_PATH = internal/database/migrations

# Build targets
.ONESHELL:
.PHONY: build clean test deploy build-all help

# Build all Lambda functions
build-all: build-auth build-users build-authorizer

# Build auth Lambda (public routes: register, login, refresh)
build-auth:
	@echo "🔨 Building auth Lambda..."
	cd cmd/auth && GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME) .
	cd cmd/auth && zip auth.zip $(BINARY_NAME) && mv auth.zip ../../dist/
	@echo "✅ Auth Lambda built successfully"

# Build users Lambda (private routes: profile, usuarios)
build-users:
	@echo "🔨 Building users Lambda..."
	cd cmd/users && GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME) .
	cd cmd/users && zip users.zip $(BINARY_NAME) && mv users.zip ../../dist/
	@echo "✅ Users Lambda built successfully"

# Build authorizer Lambda
build-authorizer:
	@echo "🔨 Building authorizer Lambda..."
	cd cmd/authorizer && GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME) .
	cd cmd/authorizer && zip authorizer.zip $(BINARY_NAME) && mv authorizer.zip ../../dist/
	@echo "✅ Authorizer Lambda built successfully"

# Legacy build (for backward compatibility) - DEPRECATED
# Use build-users instead
build:
	@echo "⚠️  Legacy build is deprecated. Use 'make build-users' instead."
	@echo "🔄 Running build-users..."
	@make build-users

build-migration:
	@echo "🔨 Building migration Lambda..."
	cd cmd/run-migrations && \
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -tags lambda.norpc -o bootstrap main.go && \
	zip migrations.zip bootstrap && \
	rm bootstrap && \
	mv migrations.zip ../../dist
	@echo "✅ Migration Lambda built successfully"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -f $(BINARY_NAME) function.zip
	rm -f cmd/users/$(BINARY_NAME) cmd/users/users.zip
	rm -f cmd/authorizer/$(BINARY_NAME) cmd/authorizer/authorizer.zip
	@cd dist && rm -f migrations.zip
	@echo "✅ Clean completed"

# Run tests
test:
	@echo "🧪 Running all tests..."
	go test ./...
	@echo "✅ Tests completed"

# Run unit tests
test-unit:
	@echo "🧪 Running unit tests..."
	go test ./internal/... -short
	@echo "✅ Unit tests completed"

# Run integration tests
test-integration:
	@echo "🧪 Running integration tests..."
	@if [ -z "$(TEST_DATABASE_URL)" ]; then \
		echo "⚠️  TEST_DATABASE_URL not set, using default test database"; \
		export TEST_DATABASE_URL="postgres://postgres:password@localhost:5432/kivaplus_test?sslmode=disable"; \
	fi
	@export TEST_DATABASE_URL=$(TEST_DATABASE_URL) && \
	export JWT_SECRET="test-secret-key-for-integration-tests" && \
	go test -v ./test/... -tags=integration
	@echo "✅ Integration tests completed"

# Run integration tests with script
test-integration-script:
	@echo "🧪 Running integration tests with setup script..."
	@bash scripts/run-integration-tests.sh

# Deploy functions
deploy-users: build-users
	@echo "🚀 Deploying users Lambda..."
	aws lambda update-function-code \
		--region $(AWS_REGION) \
		--function-name KivaPlusBackendStack-dev-UsersLambda \
		--zip-file fileb://cmd/users/users.zip
	@echo "✅ Users Lambda deployed"

deploy-authorizer: build-authorizer
	@echo "🚀 Deploying authorizer Lambda..."
	aws lambda update-function-code \
		--region $(AWS_REGION) \
		--function-name KivaPlusBackendStack-dev-AuthorizerLambda \
		--zip-file fileb://cmd/authorizer/authorizer.zip
	@echo "✅ Authorizer Lambda deployed"

# Deploy legacy function - DEPRECATED
deploy-legacy:
	@echo "⚠️  Legacy deploy is deprecated. Use 'make deploy-users' instead."
	@echo "🔄 Running deploy-users..."
	@make deploy-users

# Deploy all functions
deploy-all: deploy-usuarios deploy-authorizer
	@echo "✅ All Lambda functions deployed"

# Database operations (keeping existing functionality)
migrate-credentials:
	@echo "🔐 Loading AWS Secrets Manager credentials..."
	@if [ -z "$AWS_SECRET_NAME" ]; then \
		echo "❌ AWS_SECRET_NAME variable not defined!"; \
		echo "Run: export AWS_SECRET_NAME=kivaplus/db-credentials"; \
		exit 1; \
	fi
	@bash scripts/get-db-credentials.sh

migrate-up:
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "❌ DATABASE_URL not defined!"; \
		echo "Run first: source ./scripts/get-db-credentials.sh"; \
		exit 1; \
	fi
	@echo "🚀 Applying migrations..."
	migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" up

migrate-down:
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "❌ DATABASE_URL not defined!"; \
		echo "Run first: source ./scripts/get-db-credentials.sh"; \
		false; \
	fi
	@echo "⬇️  Rolling back migrations..."
	migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" down

migrate-down-1:
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "❌ DATABASE_URL not defined!"; \
		echo "Run first: source ./scripts/get-db-credentials.sh"; \
		exit 1; \
	fi
	@echo "⬇️  Rolling back last migration..."
	migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" down 1

migrate-version:
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "❌ DATABASE_URL not defined!"; \
		echo "Run first: source ./scripts/get-db-credentials.sh"; \
		exit 1; \
	fi
	@echo "📊 Current database version:"
	@if command -v migrate >/dev/null 2>&1; then \
		echo migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" version; \
	else \
		echo "⚠️  'migrate' CLI not found. Use: make install-migrate"; \
		echo "Or run: go run cmd/migrate/main.go version"; \
	fi

migrate-create:
	@read -p "Migration name: " name; \
	migrate create -ext sql -dir $(MIGRATIONS_PATH) -seq $$name

# Go-based migrations (hexagonal architecture compatible)
migrate-go-version:
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "❌ DATABASE_URL not defined!"; \
		echo "Run first: source ./scripts/get-db-credentials.sh"; \
		exit 1; \
	fi
	@echo "📊 Current database version (via Go):"
	go run cmd/migrate/main.go -action=version

migrate-go-up:
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "❌ DATABASE_URL not defined!"; \
		exit 1; \
	fi
	@echo "🚀 Applying migrations (via Go)..."
	go run cmd/migrate/main.go -action=up

# Run address column rename migration
migrate-address:
	@echo "🔄 Running address column rename migration..."
	@bash scripts/run-address-migration.sh

# Session Manager operations
connect-ssm:
	@echo "🔗 Connecting via Session Manager..."
	@bash scripts/connect-session-manager.sh

migrate-ssm:
	@echo "🚀 Running migrations via Session Manager..."
	@bash scripts/run-migrations-ssm.sh

# Development helpers
dev-setup:
	@echo "🛠️  Setting up development environment..."
	go mod tidy
	go mod download
	@echo "✅ Development environment ready"

install-migrate:
	@echo "🔧 Installing golang-migrate..."
	@bash scripts/install-migrate.sh

# Linting and formatting
lint:
	@echo "🔍 Running linter..."
	golangci-lint run
	@echo "✅ Linting completed"

format:
	@echo "🎨 Formatting code..."
	go fmt ./...
	@echo "✅ Code formatted"

# API testing
test-api:
	@echo "🧪 Testing API endpoints..."
	./scripts/test-jwt.sh
	@echo "✅ API tests completed"

test-endpoints:
	@echo "🧪 Testing API endpoints with curl..."
	./scripts/test-endpoints.sh
	@echo "✅ Endpoint tests completed"

test-handlers:
	@echo "🧪 Running handler tests..."
	go test -v -race ./internal/users/handlers/...
	@echo "✅ Handler tests completed"

test-coverage:
	@echo "🧪 Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./internal/...
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out
	@echo "✅ Coverage report generated: coverage.html"

test-jwt-roles:
	@echo "🧪 Testing JWT with roles..."
	./scripts/test-jwt-roles.sh
	@echo "✅ JWT roles tests completed"

test-complete-system:
	@echo "🧪 Testing complete system (roles, profile, cache)..."
	./scripts/test-complete-system.sh
	@echo "✅ Complete system tests completed"

setup-test-roles:
	@echo "🔧 Setting up test roles..."
	@echo "Connect to Session Manager and run:"
	@echo "  psql \$$DATABASE_URL -f /path/to/add-test-roles.sql"
	@echo "Or copy the SQL content from scripts/add-test-roles.sql"

test-postgres:
	@echo "🧪 Running PostgreSQL tests..."
	@bash scripts/test-postgres.sh

# Infrastructure
infra-diff:
	@echo "🏗️  Diff infrastructure..."
	cd ../infra && cdk diff
	@echo "✅ Infrastructure deployed"

infra-deploy:
	@echo "🏗️  Deploying infrastructure..."
	cd ../infra && cdk deploy KivaPlusBackendStack-dev --require-approval never
	@echo "✅ Infrastructure deployed"

infra-destroy:
	@echo "💥 Destroying infrastructure..."
	cd ../infra && cdk destroy KivaPlusBackendStack-dev --force
	@echo "✅ Infrastructure destroyed"

# Connection helpers
connect-db:
	@echo "🔗 Connecting to RDS via bastion host..."
	@bash scripts/connect-db.sh

port-forward:
	@echo "🔗 Starting port forwarding to RDS..."
	@bash scripts/port-forward.sh

# Debug helpers
debug-rds:
	@echo "🔍 Debugging RDS endpoints..."
	@bash scripts/debug-rds.sh

debug-sg:
	@echo "🔍 Debugging Security Groups..."
	@bash scripts/debug-security-groups.sh

# Docker-based Local Development
docker-setup:
	@echo "🐳 Setting up Docker development environment..."
	@if [ ! -f .env ]; then \
		echo "📝 Creating .env from .env.docker..."; \
		cp .env.docker .env; \
	fi
	@docker-compose up -d postgres localstack
	@echo "⏳ Waiting for services to be ready..."
	@sleep 10
	@make docker-migrate
	@echo "✅ Docker development environment ready"

docker-up:
	@echo "🐳 Starting all Docker services..."
	@docker-compose up -d

docker-down:
	@echo "🐳 Stopping all Docker services..."
	@docker-compose down

docker-logs:
	@echo "📋 Showing Docker logs..."
	@docker-compose logs -f

docker-migrate:
	@echo "🚀 Running migrations in Docker..."
	@docker-compose exec postgres psql -U postgres -d kivaplus_local -c "SELECT 'Database ready' as status;"
	@if command -v migrate >/dev/null 2>&1; then \
		DATABASE_URL="postgres://postgres:password@localhost:5432/kivaplus_local?sslmode=disable" \
		migrate -path internal/database/migrations -database "$$DATABASE_URL" up; \
	else \
		echo "⚠️  migrate CLI not found. Install with: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"; \
	fi

run-local:
	@echo "🚀 Starting local development server..."
	@if [ ! -f .env ]; then \
		echo "📝 Creating .env from .env.docker..."; \
		cp .env.docker .env; \
	fi
	@echo "Make sure Docker services are running: make docker-up"
	go run cmd/local-server/main.go

run-docker:
	@echo "🐳 Starting application in Docker..."
	@docker-compose up app

# Legacy local setup (without Docker)
setup-local-native:
	@echo "🛠️  Setting up native local development environment..."
	@bash scripts/setup-local-db.sh
	@echo "✅ Native local development environment ready"

test-local:
	@echo "🧪 Testing local API endpoints..."
	@bash scripts/test-local-api.sh

debug-local:
	@echo "🔍 Starting local server with debug logging..."
	@export LOG_LEVEL=debug && go run cmd/local-server/main.go

# Database helpers for local development
local-db-reset:
	@echo "🔄 Resetting local database..."
	@dropdb kivaplus_local 2>/dev/null || true
	@make setup-local

local-db-shell:
	@echo "🐘 Opening PostgreSQL shell..."
	@psql postgres://postgres:password@localhost:5432/kivaplus_local

# Testing targets for local development
test-local-unit:
	@echo "🧪 Running unit tests..."
	go test -short ./internal/...

test-local-integration:
	@echo "🧪 Running local integration tests..."
	@echo "Make sure local server is running (make run-local)"
	go test -v ./test/local_integration_test.go

test-local-repo:
	@echo "🧪 Running repository tests against local database..."
	go test -v ./internal/users/repository/...

# Development workflow
dev-workflow:
	@echo "🔄 Complete development workflow..."
	@make setup-local
	@echo "✅ Database setup complete"
	@make run-local &
	@echo "⏳ Waiting for server to start..."
	@sleep 3
	@make test-local
	@echo "✅ Development workflow complete"

# Help
help:
	@echo "📚 Available commands:"
	@echo ""
	@echo "🏗️  Build Commands:"
	@echo "  build-all        - Build all Lambda functions (hexagonal)"
	@echo "  build-users      - Build users Lambda function"
	@echo "  build-authorizer - Build authorizer Lambda function"
	@echo "  build            - Build legacy Lambda function"
	@echo "  build-migration  - Build migration Lambda function"
	@echo "  clean            - Clean build artifacts"
	@echo ""
	@echo "🚀 Deploy Commands:"
	@echo "  deploy-all       - Deploy all Lambda functions"
	@echo "  deploy-users     - Deploy users Lambda function"
	@echo "  deploy-authorizer- Deploy authorizer Lambda function"
	@echo "  deploy-legacy    - Deploy legacy Lambda function"
	@echo ""
	@echo "🗄️  Database Commands:"
	@echo "  migrate-up      - Run database migrations"
	@echo "  migrate-down    - Rollback database migrations"
	@echo "  migrate-down-1  - Rollback last migration"
	@echo "  migrate-version - Check migration status"
	@echo "  migrate-create  - Create new migration"
	@echo "  migrate-go-up   - Run migrations via Go"
	@echo "  migrate-go-version - Check version via Go"
	@echo "  migrate-ssm     - Run migrations via Session Manager"
	@echo ""
	@echo "🧪 Test Commands:"
	@echo "  test            - Run all tests"
	@echo "  test-unit       - Run unit tests only"
	@echo "  test-integration- Run integration tests only"
	@echo "  test-api        - Test API endpoints"
	@echo "  test-postgres   - Run PostgreSQL tests"
	@echo ""
	@echo "🛠️  Development Commands:"
	@echo "  dev-setup       - Setup development environment"
	@echo "  lint            - Run linter"
	@echo "  format          - Format code"
	@echo "  install-migrate - Install golang-migrate CLI"
	@echo ""
	@echo "🔗 Connection Commands:"
	@echo "  connect-ssm     - Connect via Session Manager (recommended)"
	@echo "  connect-db      - Connect via bastion host"
	@echo "  port-forward    - Start port forwarding"
	@echo ""
	@echo "🏗️  Infrastructure Commands:"
	@echo "  infra-diff      - Diff infrastructure"
	@echo "  infra-deploy    - Deploy infrastructure"
	@echo "  infra-destroy   - Destroy infrastructure"
	@echo ""
	@echo "🔍 Debug Commands:"
	@echo "  debug-rds       - Debug RDS endpoints"
	@echo "  debug-sg        - Debug Security Groups"

.DEFAULT_GOAL := help

# Docker-based Local Development
docker-setup:
	@echo "🐳 Setting up Docker-based local development environment..."
	@bash scripts/docker-setup.sh

docker-up:
	@echo "🐳 Starting Docker services..."
	@docker compose up -d postgres
	@echo "✅ PostgreSQL container started"

docker-up-all:
	@echo "🐳 Starting all Docker services (including LocalStack)..."
	@docker compose --profile localstack --profile cache up -d
	@echo "✅ All containers started"

docker-up-with-app:
	@echo "🐳 Starting Docker services with Go application..."
	@docker compose --profile app up -d
	@echo "✅ All containers started including Go app on port 8080"

docker-app-logs:
	@echo "📋 Showing Go application logs..."
	@docker compose logs -f app

docker-down:
	@echo "🛑 Stopping Docker services..."
	@docker compose down
	@echo "✅ All containers stopped"

docker-logs:
	@echo "📋 Showing Docker container logs..."
	@docker compose logs -f

docker-db-shell:
	@echo "🐘 Opening PostgreSQL shell in Docker container..."
	@docker compose exec postgres psql -U kivaplus -d kivaplus_local

docker-reset:
	@echo "🔄 Resetting Docker environment..."
	@docker compose down -v
	@docker compose up -d postgres
	@echo "⏳ Waiting for PostgreSQL to be ready..."
	@sleep 5
	@make docker-migrate
	@echo "✅ Docker environment reset complete"

docker-migrate:
	@echo "🚀 Running migrations on Docker database..."
	@echo "🔄 Running migrations inside Docker container (safer approach)..."
	@bash scripts/run-migrations-docker.sh

# Local Development Workflow
setup-local: docker-setup
	@echo "✅ Local development environment ready with Docker"

run-local:
	@echo "🚀 Starting local development server..."
	@if [ ! -f .env ]; then \
		echo "📝 Creating .env file from template..."; \
		cp .env.local .env; \
	fi
	@echo "🐳 Ensuring PostgreSQL container is running..."
	@docker compose up -d postgres
	@echo "⏳ Waiting for database to be ready..."
	@sleep 3
	@echo "🚀 Starting Go server..."
	@go run cmd/local-server/main.go

test-local: docker-up
	@echo "🧪 Testing local API endpoints..."
	@sleep 2  # Wait for database to be ready
	@bash scripts/test-local-api.sh

debug-local: docker-up
	@echo "🔍 Starting local server with debug logging..."
	@export LOG_LEVEL=debug && go run cmd/local-server/main.go

# Development helpers
dev-full-reset:
	@echo "🔄 Complete development environment reset..."
	@make docker-reset
	@make run-local &
	@sleep 5
	@make test-local
	@echo "✅ Full reset and test completed"

# ===== ROUTES MANAGEMENT =====
routes-list:
	@echo "📋 Listing all routes..."
	@go run cmd/routes-cli/main.go -cmd=list

routes-validate:
	@echo "✅ Validating routes configuration..."
	@go run cmd/routes-cli/main.go -cmd=validate

routes-add:
	@echo "➕ Adding new route..."
	@echo "Usage: make routes-add GROUP=users NAME=create_user METHOD=POST PATH=/usuarios LAMBDA=users"
	@if [ -z "$(GROUP)" ] || [ -z "$(NAME)" ] || [ -z "$(METHOD)" ] || [ -z "$(PATH)" ]; then \
		echo "❌ Missing required parameters: GROUP, NAME, METHOD, PATH"; \
		exit 1; \
	fi
	@go run cmd/routes-cli/main.go -cmd=add -group=$(GROUP) -name=$(NAME) -method=$(METHOD) -path=$(PATH) -lambda=$(LAMBDA) -public=$(PUBLIC) -roles=$(ROLES) -resource=$(RESOURCE) -action=$(ACTION) -scope=$(SCOPE)

routes-remove:
	@echo "➖ Removing route..."
	@echo "Usage: make routes-remove NAME=create_user"
	@if [ -z "$(NAME)" ]; then \
		echo "❌ Missing required parameter: NAME"; \
		exit 1; \
	fi
	@go run cmd/routes-cli/main.go -cmd=remove -name=$(NAME)

routes-export:
	@echo "📤 Exporting routes configuration..."
	@go run cmd/routes-cli/main.go -cmd=export -format=json

routes-sync:
	@echo "🔄 Syncing routes configuration to infrastructure..."
	@if [ -f "scripts/sync-routes-api.sh" ]; then \
		bash scripts/sync-routes-api.sh sync; \
	else \
		echo "💡 Using GitHub Actions for sync (recommended)"; \
		echo "Push your changes to trigger automatic sync"; \
	fi

routes-publish:
	@echo "📦 Publishing routes configuration as package..."
	@bash scripts/publish-routes-config.sh

routes-setup-shared:
	@echo "🔗 Setting up shared configuration repository..."
	@bash scripts/setup-shared-config.sh

routes-help:
	@echo "🔧 Routes Management Commands:"
	@echo "  make routes-list                    # List all routes"
	@echo "  make routes-validate                # Validate configuration"
	@echo "  make routes-add GROUP=... NAME=...  # Add new route"
	@echo "  make routes-remove NAME=...         # Remove route"
	@echo "  make routes-export                  # Export configuration"
	@echo ""
	@echo "🔄 Cross-Repository Sync Commands:"
	@echo "  make routes-sync                    # Sync to infrastructure repo"
	@echo "  make routes-publish                 # Publish as NPM package"
	@echo "  make routes-setup-shared            # Setup shared config repo"
	@echo ""
	@echo "Example: Add a new route"
	@echo "  make routes-add GROUP=users NAME=delete_user METHOD=DELETE PATH=/usuarios/{id} LAMBDA=users ROLES=1,2"
