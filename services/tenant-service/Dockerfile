# --- Stage 1: Build ---
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy workspace files (needed for shared packages)
COPY go.work* ./
COPY packages/ ./packages/

# Copy all service go.mod files (required by go.work)
COPY services/auth-service/go.mod services/auth-service/go.sum* ./services/auth-service/
COPY services/client-service/go.mod services/client-service/go.sum* ./services/client-service/
COPY services/tenant-service/go.mod services/tenant-service/go.sum* ./services/tenant-service/
COPY services/framework-service/go.mod services/framework-service/go.sum* ./services/framework-service/

WORKDIR /app/services/tenant-service
RUN go mod download

# Copy service source
COPY services/tenant-service/ ./

# Install swag for docs generation
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN swag init

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o tenant-service ./main.go

# --- Stage 2: Development (with hot reload) ---
FROM golang:1.25-alpine AS development

WORKDIR /app

# Install air and swag
RUN go install github.com/air-verse/air@latest
RUN go install github.com/swaggo/swag/cmd/swag@latest

# This stage expects volumes to be mounted
CMD ["air"]

# --- Stage 3: Production ---
FROM alpine:latest AS production

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/services/tenant-service/tenant-service .

# Copy config and docs
COPY services/tenant-service/config.yaml .
COPY --from=builder /app/services/tenant-service/docs ./docs

# Expose port
EXPOSE 8081

# Run
CMD ["./tenant-service"]