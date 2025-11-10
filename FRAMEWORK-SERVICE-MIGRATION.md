# Framework Service Migration Guide

## Overview

The framework functionality has been extracted from `tenant-service` into a new dedicated microservice called `framework-service`. This separation provides better scalability, maintainability, and follows microservices best practices.

## What Changed

### New Service: framework-service

A new microservice has been created at `services/framework-service/` with the following structure:

```
services/framework-service/
├── db/
│   ├── migrations/          # Database migrations
│   └── queries/             # SQL queries for sqlc
├── internal/
│   ├── config/              # Configuration management
│   ├── db/                  # Generated sqlc code (to be generated)
│   ├── handler/             # HTTP handlers
│   ├── router/              # Route definitions
│   ├── store/               # Database store wrapper
│   └── validator/           # Request validation
├── docs/                    # Swagger documentation (to be generated)
├── main.go                  # Service entry point
├── config.yaml              # Configuration file
├── sqlc.yaml                # sqlc configuration
├── Dockerfile               # Docker build file
├── Makefile                 # Build and development commands
└── README.md                # Service documentation
```

### Database

- **New Database**: `framework_db` - Dedicated database for framework service
- **Tables**: 
  - `compliance_frameworks` - Stores framework metadata and checklist JSON

### API Changes

Framework endpoints have been moved from tenant-service to framework-service:

**Old Endpoints (tenant-service):**
- `GET /api/tenant/frameworks`
- `GET /api/tenant/frameworks/:id`
- `GET /api/tenant/frameworks/:id/checklist`
- `POST /api/tenant/frameworks`
- `PUT /api/tenant/frameworks/:id`
- `DELETE /api/tenant/frameworks/:id`

**New Endpoints (framework-service):**
- `GET /api/frameworks`
- `GET /api/frameworks/:id`
- `GET /api/frameworks/:id/checklist`
- `POST /api/frameworks`
- `PUT /api/frameworks/:id`
- `DELETE /api/frameworks/:id`

The gateway (nginx) now routes `/api/frameworks` requests to the framework-service.

### Infrastructure Updates

1. **docker-compose.yml**: Added framework-service configuration
2. **nginx.conf**: Added routing for framework endpoints
3. **go.work**: Added framework-service to Go workspace

## Setup Instructions

### 1. Initialize Dependencies

```bash
cd services/framework-service
go mod tidy
```

### 2. Generate Database Code

```bash
cd services/framework-service
make sqlc
```

This generates the database access code in `internal/db/`.

### 3. Create Framework Database

```bash
# Connect to PostgreSQL
psql -U root -h localhost

# Create the database
CREATE DATABASE framework_db;
\q
```

### 4. Run Migrations

```bash
cd services/framework-service
make migrate-up
```

Or the migrations will run automatically when the service starts.

### 5. Generate Swagger Documentation (Optional)

```bash
cd services/framework-service
make swagger
```

### 6. Start the Service

**Using Docker Compose (Recommended):**
```bash
# From project root
docker-compose up framework-service
```

**Using Make:**
```bash
cd services/framework-service
make run
```

**Using Air (Hot Reload):**
```bash
cd services/framework-service
make dev
```

## Configuration

The service is configured via `config.yaml`:

```yaml
server:
  host: "0.0.0.0"
  port: "8084"

database:
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "postgres"
  dbname: "framework_db"
  sslmode: "disable"

auth:
  jwt_secret: "your-secret-key-change-in-production"
```

Environment variables can override these settings:
- `SERVER_HOST`
- `SERVER_PORT`
- `DATABASE_HOST`
- `DATABASE_PORT`
- `DATABASE_USER`
- `DATABASE_PASSWORD`
- `DATABASE_DBNAME`
- `AUTH_JWT_SECRET`

## Next Steps: Removing Framework Code from Tenant Service

After verifying the framework-service works correctly, you should:

1. **Remove framework handlers** from `services/tenant-service/internal/handler/framework.go`
2. **Remove framework routes** from `services/tenant-service/internal/router/router.go`
3. **Remove framework queries** from `services/tenant-service/db/queries/frameworks.sql`
4. **Update tenant-service** to call framework-service via HTTP when it needs framework data
5. **Migrate framework data** from tenant_db to framework_db:
   ```sql
   -- Export from tenant_db
   \copy (SELECT * FROM compliance_frameworks) TO '/tmp/frameworks.csv' CSV HEADER;
   
   -- Import to framework_db
   \copy compliance_frameworks FROM '/tmp/frameworks.csv' CSV HEADER;
   ```

## Testing

### Health Check

```bash
curl http://localhost:8084/health
```

### List Frameworks

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
    "description": "National Stock Exchange compliance framework",
    "version": "1.0",
    "checklist_json": {
      "sections": [
        {
          "name": "Section 1",
          "questions": [
            {
              "number": "1.1",
              "text": "Question text",
              "type": "yes_no",
              "help_text": "Help text",
              "is_mandatory": true
            }
          ]
        }
      ]
    }
  }' \
  http://localhost:8080/api/frameworks
```

## Troubleshooting

### Service won't start

1. Check database connection:
   ```bash
   psql -U postgres -h localhost -d framework_db
   ```

2. Check logs:
   ```bash
   docker-compose logs framework-service
   ```

3. Verify port 8084 is not in use:
   ```bash
   lsof -i :8084
   ```

### Database errors

1. Ensure migrations ran successfully:
   ```bash
   cd services/framework-service
   make migrate-up
   ```

2. Check database exists:
   ```bash
   psql -U postgres -h localhost -l | grep framework_db
   ```

### Import errors

Run `go mod tidy` in the framework-service directory:
```bash
cd services/framework-service
go mod tidy
```

## Architecture Benefits

1. **Separation of Concerns**: Framework management is isolated from tenant management
2. **Independent Scaling**: Framework service can be scaled independently
3. **Database Isolation**: Framework data is in its own database
4. **Easier Maintenance**: Changes to framework logic don't affect tenant service
5. **Better Testing**: Framework functionality can be tested in isolation
6. **Deployment Flexibility**: Services can be deployed and updated independently

## API Documentation

Once the service is running, Swagger documentation is available at:
```
http://localhost:8084/swagger/index.html
```

## Port Assignments

- **8082**: auth-service
- **8081**: tenant-service, client-service
- **8084**: framework-service (new)
- **8080**: nginx gateway

## Support

For issues or questions, refer to:
- Service README: `services/framework-service/README.md`
- Main project documentation
- Service logs: `docker-compose logs framework-service`
