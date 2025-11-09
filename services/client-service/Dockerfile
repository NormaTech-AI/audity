# # Development stage
# FROM golang:1.25-alpine AS development

# WORKDIR /app

# # Install air for hot reload
# RUN go install github.com/air-verse/air@latest

# # Expose port
# EXPOSE 8081

# # Run with air for hot reload
# CMD ["air", "-c", "services/client-service/.air.toml"]

# # Build stage
# FROM golang:1.25-alpine AS builder

# WORKDIR /app

# # Copy go mod files
# COPY go.mod go.sum ./
# RUN go mod download

# # Copy source code
# COPY . .

# # Build the application
# RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# # Production stage
# FROM alpine:latest

# RUN apk --no-cache add ca-certificates

# WORKDIR /root/

# # Copy the binary from builder
# COPY --from=builder /app/main .
# COPY --from=builder /app/config.yaml .

# # Expose port
# EXPOSE 8081

# # Run the binary
# CMD ["./main"]

# --- Stage 1: Build ---
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy workspace files (needed for shared packages)
COPY go.work* ./
COPY packages/ ./packages/

# Copy service files
COPY services/client-service/go.mod services/client-service/go.sum* ./services/client-service/
WORKDIR /app/services/client-service
RUN go mod download

# Copy service source
COPY services/client-service/ ./

# Install swag for docs generation
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN swag init

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o client-service ./main.go

# --- Stage 2: Development (with hot reload) ---
FROM golang:1.25-alpine AS development

WORKDIR /app

# Install air and swag
RUN go install github.com/air-verse/air@latest
RUN go install github.com/swaggo/swag/cmd/swag@latest

# This stage expects volumes to be mounted
# Air will run from the service directory
WORKDIR /app/services/client-service
CMD ["air", "-c", ".air.toml"]

# --- Stage 3: Production ---
FROM alpine:latest AS production

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/services/client-service/client-service .

# Copy config and docs
COPY services/client-service/config.yaml .
COPY --from=builder /app/services/client-service/docs ./docs

# Expose port
EXPOSE 8081

# Run
CMD ["./client-service"]
