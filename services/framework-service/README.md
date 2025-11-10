# Framework Service

The Framework Service manages compliance frameworks (NSE, BSE, NCDEX, etc.) for the TPRM Audit Platform.

## Features

- Create, read, update, and delete compliance frameworks
- Store framework checklists as JSON
- Version management for frameworks
- RESTful API with JWT authentication

## API Endpoints

### Frameworks

- `GET /api/v1/frameworks` - List all frameworks
- `GET /api/v1/frameworks/:id` - Get a specific framework
- `GET /api/v1/frameworks/:id/checklist` - Get framework checklist
- `POST /api/v1/frameworks` - Create a new framework
- `PUT /api/v1/frameworks/:id` - Update a framework
- `DELETE /api/v1/frameworks/:id` - Delete a framework

## Configuration

Configuration is managed through `config.yaml`:

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
  jwt_secret: "your-secret-key"
```

## Development

### Prerequisites

- Go 1.23.2+
- PostgreSQL 15+
- sqlc
- golang-migrate

### Setup

1. Install dependencies:
```bash
go mod download
```

2. Generate database code:
```bash
make sqlc
```

3. Run migrations:
```bash
make migrate-up
```

4. Run the service:
```bash
make run
```

Or use hot reload with air:
```bash
make dev
```

### Database Migrations

Create a new migration:
```bash
migrate create -ext sql -dir db/migrations -seq migration_name
```

Run migrations:
```bash
make migrate-up
```

Rollback migrations:
```bash
make migrate-down
```

## Docker

Build the Docker image:
```bash
docker build -t framework-service .
```

Run the container:
```bash
docker run -p 8084:8084 framework-service
```

## API Documentation

Swagger documentation is available at `/swagger/index.html` when the service is running.

Generate swagger docs:
```bash
make swagger
```
