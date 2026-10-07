# FuzeFlow

FuzeFlow is a developer workflow and job execution platform designed around durable execution, distributed workers, integrations, and observable automation.

## Phase 1

Phase 1 establishes the project foundation:

- Go services for API, worker, and CLI entry points
- strict configuration from environment variables
- structured JSON logging
- HTTP server timeouts and graceful shutdown
- baseline security headers
- health and readiness endpoints
- PostgreSQL and Redis development services
- initial multi-tenant identity schema
- Docker build
- automated API foundation test

The execution engine, authentication, queue protocol, integrations, and dashboard are intentionally not implemented yet.

## Requirements

- Go 1.25+
- Docker Desktop

## Local development

```bash
go test ./...
go vet ./...
go run ./cmd/server
```

The API starts on `http://localhost:8080` by default.

```text
GET /healthz
GET /readyz
GET /api/v1
```

Start infrastructure with:

```bash
docker compose up -d
```

## Engineering standard

FuzeFlow favors small packages, explicit boundaries, standard-library primitives where they are sufficient, server-side validation, minimal dependencies, deterministic behavior, and tests around critical behavior.
