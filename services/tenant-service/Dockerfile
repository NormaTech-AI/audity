FROM golang:1.25-alpine
WORKDIR /app

# Install air
RUN go install github.com/air-verse/air@latest
# Install swag
RUN go install github.com/swaggo/swag/cmd/swag@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Run Swag once to generate initial docs
RUN swag init

# Port this service listens on *internally*
EXPOSE 8081

# Run air
CMD ["air"]