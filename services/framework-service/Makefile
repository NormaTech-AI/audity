.PHONY: help build run test clean sqlc migrate-up migrate-down swagger

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build the application
	go build -o bin/framework-service main.go

run: ## Run the application
	go run main.go

test: ## Run tests
	go test -v ./...

clean: ## Clean build artifacts
	rm -rf bin/
	rm -rf tmp/

sqlc: ## Generate sqlc code
	sqlc generate

migrate-up: ## Run database migrations up
	migrate -path db/migrations -database "postgres://postgres:postgres@localhost:5432/framework_db?sslmode=disable" up

migrate-down: ## Run database migrations down
	migrate -path db/migrations -database "postgres://postgres:postgres@localhost:5432/framework_db?sslmode=disable" down

swagger: ## Generate swagger documentation
	swag init -g main.go --output docs

dev: ## Run with air for hot reload
	air
