FROM golang:1.21-alpine

WORKDIR /app

# Install air
RUN go install github.com/cosmtrek/air@latest

# Copy source code
COPY . .

EXPOSE 8081

CMD ["air", "-c", ".air.toml"]