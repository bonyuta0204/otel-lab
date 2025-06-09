FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

# Install air for hot reload
RUN go install github.com/cosmtrek/air@latest

FROM golang:1.21-alpine

WORKDIR /app

# Install air
RUN go install github.com/cosmtrek/air@latest

# Copy source code
COPY . .

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]