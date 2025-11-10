# Framework Service - Quick Start

## Setup (First Time)

```bash
# 1. Navigate to service directory
cd services/framework-service

# 2. Install dependencies
go mod tidy

# 3. Generate database code
make sqlc

# 4. Create database
psql -U postgres -h localhost -c "CREATE DATABASE framework_db;"

# 5. Run migrations
make migrate-up
```

## Running the Service

### Option 1: Docker Compose (Recommended)
```bash
# From project root
docker-compose up framework-service
```

### Option 2: Local Development with Hot Reload
```bash
cd services/framework-service
make dev
```

### Option 3: Direct Run
```bash
cd services/framework-service
make run
```

## Testing

### Health Check
```bash
curl http://localhost:8084/health
```

### List Frameworks (requires JWT)
```bash
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
     http://localhost:8080/api/frameworks
```

### Create Framework
```bash
curl -X POST \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "NSE",
    "description": "National Stock Exchange",
    "version": "1.0",
    "checklist_json": {
      "sections": [{
        "name": "Section 1",
        "questions": [{
          "number": "1.1",
          "text": "Question text",
          "type": "yes_no",
          "help_text": "Help",
          "is_mandatory": true
        }]
      }]
    }
  }' \
  http://localhost:8080/api/frameworks
```

## Common Commands

```bash
# Build the service
make build

# Run tests
make test

# Generate swagger docs
make swagger

# Clean build artifacts
make clean

# Run migrations up
make migrate-up

# Run migrations down
make migrate-down
```

## Ports

- **8084**: Framework service (internal)
- **8080**: Gateway (public access)

## Configuration

Edit `config.yaml` or set environment variables:

```bash
export SERVER_PORT=8084
export DATABASE_HOST=localhost
export DATABASE_DBNAME=framework_db
export AUTH_JWT_SECRET=your-secret
```

## Troubleshooting

### Service won't start
```bash
# Check logs
docker-compose logs framework-service

# Verify database
psql -U postgres -h localhost -d framework_db

# Check port
lsof -i :8084
```

### Import errors
```bash
cd services/framework-service
go mod tidy
make sqlc
```

### Database errors
```bash
# Recreate database
psql -U postgres -h localhost -c "DROP DATABASE IF EXISTS framework_db;"
psql -U postgres -h localhost -c "CREATE DATABASE framework_db;"
make migrate-up
```

## Documentation

- Full README: `services/framework-service/README.md`
- Migration Guide: `FRAMEWORK-SERVICE-MIGRATION.md`
- API Docs: `http://localhost:8084/swagger/index.html`
