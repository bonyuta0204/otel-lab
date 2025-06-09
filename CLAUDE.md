# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

OtelLab is an educational OpenTelemetry microservices project demonstrating distributed tracing with Jaeger. It consists of three Go services:
- **API Gateway** (port 8080): HTTP REST entry point using Gorilla Mux
- **Task Service** (port 8081): gRPC service with PostgreSQL storage
- **User Service** (port 8082): gRPC service with in-memory storage

Services communicate via gRPC and send traces to Jaeger (UI: http://localhost:16686).

## Essential Development Commands

### Quick Setup
```bash
make quick-start  # Install deps, generate proto, start all services
make demo         # Create demo user and task for testing
```

### Development Workflow
```bash
make build        # Build all services
make test         # Run tests with race detection and coverage
make lint         # Run golangci-lint
make format       # Format code and tidy modules
```

### Service Management
```bash
make up           # Start all services with Docker Compose
make down         # Stop all services
make restart      # Restart all services
make logs         # View logs from all services
make health       # Check service health
```

### Code Generation
```bash
make proto        # Generate Go code from protobuf files (required after .proto changes)
```

## Architecture

```
Client → API Gateway → Task Service → PostgreSQL
               ↓         User Service → Memory
           Jaeger ← ← ← ← ← ← ← ← ← ← ←
```

### Key Files by Service

**API Gateway** (`/api-gateway/`):
- `handlers/` - HTTP request handlers for REST endpoints
- `middleware/middleware.go` - Request ID, logging, CORS
- `tracing/tracer.go` - OpenTelemetry initialization

**Task Service** (`/task-service/`):
- `server/server.go` - gRPC service implementation
- `storage/postgres.go` - PostgreSQL repository layer

**User Service** (`/user-service/`):
- `server/server.go` - gRPC service implementation  
- `storage/memory.go` - In-memory storage with seeded data

**Shared**:
- `/proto/` - Protocol buffer definitions for gRPC contracts
- `/migrations/` - Database migration files

## OpenTelemetry Integration

All services use OpenTelemetry for distributed tracing:
- Automatic HTTP/gRPC instrumentation via interceptors
- Trace context propagation across service boundaries
- Jaeger exporter configuration in each service's `tracing/tracer.go`
- Environment-based configuration (JAEGER_ENDPOINT, etc.)

## Development Notes

- PostgreSQL migrations run automatically on task-service startup
- Services use graceful shutdown patterns
- Hot reload enabled via Air during development
- gRPC health checks implemented for service dependencies
- Consistent error handling and logging across services