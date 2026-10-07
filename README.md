# FuzeFlow

FuzeFlow is a developer workflow and job execution platform designed around durable execution, distributed workers, integrations, and observable automation.

## Phase 1

Phase 1 established the Go service foundation, structured logging, HTTP hardening, graceful shutdown, PostgreSQL/Redis development infrastructure, migrations, and baseline tests.

## Phase 2

Phase 2 establishes the identity and security boundary:

- email/password registration and login
- bcrypt password hashing
- opaque 256-bit session tokens with SHA-256 token storage
- HttpOnly, SameSite session cookies
- session revocation and expiry
- GitHub OAuth with state validation
- automatic organization creation for new accounts
- organization memberships with owner/admin/member/viewer roles
- server-side organization membership and role enforcement
- bounded request bodies
- request IDs
- authentication rate limiting
- security headers and no-store responses
- security audit records with request and network metadata
- PostgreSQL-backed identity persistence
- unit tests and GitHub Actions CI

The workflow engine, queue protocol, worker leasing, container execution, GitHub webhooks, and integrations remain outside this phase.

## Configuration

Copy .env.example to .env and configure PostgreSQL before starting the server.

For GitHub OAuth, create an OAuth application and set:

- FUZE_GITHUB_CLIENT_ID
- FUZE_GITHUB_CLIENT_SECRET
- FUZE_GITHUB_CALLBACK_URL

The callback must exactly match the URL configured in GitHub.

Set FUZE_COOKIE_SECURE=true when serving FuzeFlow over HTTPS.

## Local development

Start infrastructure:

    docker compose up -d

Apply migrations in order:

    migrations/000001_initial.sql
    migrations/000002_auth.sql

Then run:

    go test ./...
    go test -race ./...
    go vet ./...
    go run ./cmd/server

The API listens on http://localhost:8080 by default.

## Authentication API

    POST /api/v1/auth/signup
    POST /api/v1/auth/login
    POST /api/v1/auth/logout
    GET  /api/v1/auth/me
    GET  /api/v1/auth/github
    GET  /api/v1/auth/github/callback
    GET  /api/v1/organizations

Organization-scoped APIs must use the authenticated user's membership on the server. The client cannot grant itself organization access by choosing a different organization ID.

## Engineering standard

FuzeFlow favors small packages, explicit boundaries, standard-library primitives where they are sufficient, server-side authorization, minimal dependencies, deterministic behavior, and tests around security-critical behavior.
