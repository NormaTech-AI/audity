.PHONY: help dev build test tidy clean swagger

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

dev: ## Run development server with hot reload
	@echo "Starting development server..."
	air

build: ## Build the application
	@echo "Building application..."
	go build -o bin/auth-service main.go

test: ## Run tests
	@echo "Running tests..."
	go test -v ./...

tidy: ## Tidy go modules
	@echo "Tidying go modules..."
	go mod tidy

swagger: ## Generate Swagger documentation
	@echo "Generating Swagger docs..."
	swag init -g main.go --output docs

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf bin/ tmp/ docs/

install-tools: ## Install development tools
	@echo "Installing development tools..."
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/air-verse/air@latest
