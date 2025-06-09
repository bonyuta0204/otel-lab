FROM golang:1.24-alpine

WORKDIR /app

# Install air for hot reload
RUN go install github.com/air-verse/air@latest

# Copy go.mod and go.sum from project root
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire project (since we need proto files)
COPY . .

# Set working directory to api-gateway
WORKDIR /app/api-gateway

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]