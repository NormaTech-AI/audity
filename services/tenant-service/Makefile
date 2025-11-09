.PHONY: help migrate-up migrate-down migrate-create sqlc-generate dev build

# Database connection
DB_URL ?= postgres://root:password@localhost:5432/tenant_db?sslmode=disable

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

migrate-up: ## Run database migrations up
	@echo "Running migrations..."
	migrate -path db/migrations -database "$(DB_URL)" -verbose up

migrate-down: ## Rollback database migrations
	@echo "Rolling back migrations..."
	migrate -path db/migrations -database "$(DB_URL)" -verbose down

migrate-create: ## Create a new migration file (usage: make migrate-create NAME=migration_name)
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME is required. Usage: make migrate-create NAME=migration_name"; \
		exit 1; \
	fi
	migrate create -ext sql -dir db/migrations -seq $(NAME)

sqlc-generate: ## Generate Go code from SQL queries
	@echo "Generating sqlc code..."
	sqlc generate

dev: ## Run development server with hot reload
	@echo "Starting development server..."
	air

build: ## Build the application
	@echo "Building application..."
	go build -o bin/tenant-service main.go

test: ## Run tests
	@echo "Running tests..."
	go test -v ./...

install-tools: ## Install development tools
	@echo "Installing development tools..."
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/air-verse/air@latest

tidy: ## Tidy go modules
	@echo "Tidying go modules..."
	go mod tidy

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf bin/ tmp/
