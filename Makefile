.PHONY: migrate-up migrate-down migrate-create migrate-force migrate-version

# Variáveis
DATABASE_URL ?= $(shell echo $$DATABASE_URL)
MIGRATIONS_PATH = internal/database/migrations

build:
	@cd lambda && \
		GOOS=linux GOARCH=amd64 go build -o bootstrap && \
		zip auth.zip bootstrap && \
		mv auth.zip ../dist && \
		rm bootstrap
clean:
	@cd dist && \
		rm auth.zip
build-migration:
	@echo "🔨 Building Lambda function..."
	cd lambda/run-migrations && \
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -tags lambda.norpc -o bootstrap main.go && \
	zip migrations.zip bootstrap && \
	rm bootstrap && \
	mv migrations.zip ../../dist


# Carregar credenciais do AWS Secrets Manager
migrate-credentials:
	@echo "🔐 Carregando credenciais do AWS Secrets Manager..."
	@if [ -z "$$AWS_SECRET_NAME" ]; then \
		echo "❌ Variável AWS_SECRET_NAME não definida!"; \
		echo "Execute: export AWS_SECRET_NAME=kivaplus/db-credentials"; \
		exit 1; \
	fi
	@bash scripts/get-db-credentials.sh

# Subir todas as migrations
migrate-up:
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "❌ DATABASE_URL não definida!"; \
		echo "Execute primeiro: source ./scripts/get-db-credentials.sh"; \
		exit 1; \
	fi
	@echo "🚀 Aplicando migrations..."
	migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" up

# Reverter todas as migrations
migrate-down:
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "❌ DATABASE_URL não definida!"; \
		echo "Execute primeiro: source ./scripts/get-db-credentials.sh"; \
		exit 1; \
	fi
	@echo "⬇️  Revertendo migrations..."
	migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" down

# Reverter apenas a última migration
migrate-down-1:
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "❌ DATABASE_URL não definida!"; \
		echo "Execute primeiro: source ./scripts/get-db-credentials.sh"; \
		exit 1; \
	fi
	@echo "⬇️  Revertendo última migration..."
	migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" down 1

# Forçar versão específica (em caso de erro)
migrate-force:
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "❌ DATABASE_URL não definida!"; \
		echo "Execute primeiro: source ./scripts/get-db-credentials.sh"; \
		exit 1; \
	fi
	@echo "⚠️  Forçando versão..."
	@read -p "Digite a versão: " version; \
	migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" force $$version

# Ver versão atual
migrate-version:
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "❌ DATABASE_URL não definida!"; \
		echo "Execute primeiro: source ./scripts/get-db-credentials.sh"; \
		exit 1; \
	fi
	@echo "📊 Versão atual do banco:"
	@if command -v migrate >/dev/null 2>&1; then \
		migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" version; \
	else \
		echo "⚠️  CLI 'migrate' não encontrado. Use: make install-migrate"; \
		echo "Ou execute: go run cmd/migrate/main.go version"; \
	fi

# Criar nova migration
migrate-create:
	@read -p "Nome da migration: " name; \
	migrate create -ext sql -dir $(MIGRATIONS_PATH) -seq $$name

# Instalar golang-migrate CLI
install-migrate:
	@echo "🔧 Instalando golang-migrate..."
	@bash scripts/install-migrate.sh

# Configurar chave SSH no bastion
setup-ssh:
	@echo "🔑 Configurando chave SSH no bastion..."
	@bash scripts/setup-ssh-key.sh

# Conexão com banco via bastion host
connect-db:
	@echo "🔗 Conectando ao RDS via bastion host..."
	@bash scripts/connect-db.sh

# Debug RDS endpoints
debug-rds:
	@echo "🔍 Debugando RDS endpoints..."
	@bash scripts/debug-rds.sh

# Debug security groups
debug-sg:
	@echo "🔍 Debugando Security Groups..."
	@bash scripts/debug-security-groups.sh

# Testar se o túnel está funcionando
test-tunnel:
	@echo "🔍 Testando túnel SSH..."
	@bash scripts/test-tunnel.sh

# Port forwarding via Systems Manager (alternativo)
port-forward:
	@echo "🔗 Iniciando port forwarding para RDS..."
	@bash scripts/port-forward.sh

# Session Manager connection (recommended for development)
connect-ssm:
	@echo "🔗 Conectando via Session Manager..."
	@bash scripts/connect-session-manager.sh

# Run migrations via Session Manager
migrate-ssm:
	@echo "🚀 Executando migrations via Session Manager..."
	@bash scripts/run-migrations-ssm.sh

# Usar Go migration local (sem CLI migrate)
migrate-go-version:
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "❌ DATABASE_URL não definida!"; \
		echo "Execute primeiro: source ./scripts/get-db-credentials.sh"; \
		exit 1; \
	fi
	@echo "📊 Versão atual do banco (via Go):"
	go run cmd/migrate/main.go -action=version

migrate-go-up:
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "❌ DATABASE_URL não definida!"; \
		exit 1; \
	fi
	@echo "🚀 Aplicando migrations (via Go)..."
	go run cmd/migrate/main.go -action=up

# Comandos para usar com túnel SSH local
migrate-local-version:
	@if [ -z "$(DATABASE_URL_LOCAL)" ]; then \
		echo "❌ DATABASE_URL_LOCAL não definida!"; \
		echo "Execute primeiro: source ./scripts/get-local-db-url.sh"; \
		exit 1; \
	fi
	@echo "📊 Versão atual do banco (local via túnel):"
	DATABASE_URL="$(DATABASE_URL_LOCAL)" go run cmd/migrate/main.go -action=version

migrate-local-up:
	@if [ -z "$(DATABASE_URL_LOCAL)" ]; then \
		echo "❌ DATABASE_URL_LOCAL não definida!"; \
		echo "Execute primeiro: source ./scripts/get-local-db-url.sh"; \
		exit 1; \
	fi
	@echo "🚀 Aplicando migrations (local via túnel)..."
	DATABASE_URL="$(DATABASE_URL_LOCAL)" go run cmd/migrate/main.go -action=up

migrate-local-down:
	@if [ -z "$(DATABASE_URL_LOCAL)" ]; then \
		echo "❌ DATABASE_URL_LOCAL não definida!"; \
		echo "Execute primeiro: source ./scripts/get-local-db-url.sh"; \
		exit 1; \
	fi
	@echo "⬇️ Revertendo migrations (local via túnel)..."
	@read -p "Quantas migrations reverter? " steps; \
	DATABASE_URL="$(DATABASE_URL_LOCAL)" go run cmd/migrate/main.go -action=down -steps=$$steps

# Testes
test-postgres:
	@echo "🧪 Executando testes do PostgreSQL..."
	@bash scripts/test-postgres.sh

test-postgres-specific:
	@echo "🧪 Executando teste específico do PostgreSQL..."
	@read -p "Nome do teste: " test; \
	bash scripts/test-postgres.sh $$test

# Helper: mostrar ajuda
help:
	@echo "📚 Comandos disponíveis:"
	@echo ""
	@echo "  Para desenvolvimento local:"
	@echo "    make connect-ssm          - Conectar via Session Manager (recomendado)"
	@echo "    make migrate-ssm          - Executar migrations via Session Manager"
	@echo "    make setup-ssh            - Configurar chave SSH no bastion (legacy)"
	@echo "    make connect-db           - Conectar via bastion SSH (legacy)"
	@echo ""
	@echo "  Alternativo (Port Forwarding):"
	@echo "    make port-forward         - Port forwarding via Systems Manager"
	@echo "    source ./scripts/get-local-db-url.sh"
	@echo "    make migrate-local-version - Ver versão (local)"
	@echo "    make migrate-local-up     - Aplicar migrations (local)"
	@echo ""
	@echo "  Com migrate CLI (se instalado):"
	@echo "    make install-migrate    - Instalar CLI"
	@echo "    make migrate-version    - Ver versão atual"
	@echo "    make migrate-up         - Aplicar migrations"
	@echo "    make migrate-down-1     - Reverter última"
	@echo "    make migrate-create     - Criar nova"
	@echo ""
	@echo "  Testes:"
	@echo "    make test-postgres      - Executar testes do PostgreSQL"
	@echo "    make test-postgres-specific - Executar teste específico"
