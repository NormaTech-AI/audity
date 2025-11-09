# Client Service

Microservice responsible for managing client onboarding, provisioning, and lifecycle.

## Features

- **Client Onboarding**: Create and manage client organizations
- **Database Provisioning**: Automatically provision isolated PostgreSQL databases for each client
- **Storage Provisioning**: Automatically create MinIO buckets for client file storage
- **RBAC Integration**: Role-based access control for client operations
- **RESTful API**: Clean API endpoints for client management

## Architecture

The client-service is part of a microservices architecture:
- **Port**: 8081
- **Database**: Shares `tenant_db` with tenant-service for client metadata
- **Dependencies**: PostgreSQL, MinIO, auth-service

## API Endpoints

### Public Endpoints
- `GET /` - Service info
- `GET /health` - Health check

### Protected Endpoints (Require Authentication)
- `GET /api/clients` - List all clients
- `POST /api/clients` - Create a new client
- `GET /api/clients/:id` - Get client by ID

## Development

### Prerequisites
- Go 1.24.4+
- PostgreSQL 15+
- MinIO
- sqlc (for code generation)

### Setup

1. Install dependencies:
```bash
go mod download
```

2. Generate sqlc code:
```bash
sqlc generate
```

3. Run the service:
```bash
go run main.go
```

### Configuration

Edit `config.yaml` to configure:
- Server port and host
- Database connection
- MinIO connection
- JWT secret
- Logging level

## Docker

Build:
```bash
docker build -t client-service .
```

Run:
```bash
docker run -p 8081:8081 client-service
```

## Database Schema

The service manages:
- `clients` - Client organizations
- `client_databases` - Database connection info for each client
- `client_buckets` - MinIO bucket info for each client

## License

Proprietary - NormaTech AI
