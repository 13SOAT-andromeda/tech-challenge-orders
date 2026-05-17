# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
go build ./...            # build all packages
go test ./...             # run all tests
go test ./internal/... -run TestFoo  # run a single test
docker compose up -d      # start local dev (dynamodb-local + localstack + api + worker)
make up                   # deploy locally into an existing kind cluster (k8s)
make build                # docker build
```

Hot-reload in development uses `air` (`air.toml` at root).

## Architecture

Hexagonal/Clean architecture with three layers:

- **`internal/domain/`** — pure domain: `Order` aggregate, `Status` enum, `ErrOrderNotFound`. No framework imports.
- **`internal/application/`**
  - `ports/` — interfaces that the domain depends on (`OrderRepository`, `EventPublisher`, `IdempotencyStore`, `UsersClient`, `StockClient`).
  - `usecases/order/` — one file per use case, injected with ports only.
- **`internal/adapter/`** — framework-specific implementations:
  - `http/` — Gin router, handlers, middlewares.
  - `dynamo/` — DynamoDB single-table repository (`PK/SK` + `GSI1` by status + `GSI2` by customer vehicle).
  - `sns/`, `sqs/` — event publisher and SQS consumers (to be implemented per `migration-v2.md`).
  - `clients/` — outbound HTTP to Users and Stock services (to be implemented).
  - `config/` — env-based config via `godotenv`.
  - `metrics/` — Datadog DogStatsD or noop.
- **`pkg/`** — framework-agnostic utilities (monetary parser, converters).
- **`cmd/api/`** — HTTP server entrypoint.
- **`cmd/worker/`** — SQS consumer worker (to be created per migration plan).

## Migration in progress

The codebase is mid-migration from a monolith (PostgreSQL + GORM) to a standalone microservice (DynamoDB + SNS/SQS). `migration-v2.md` is the authoritative implementation plan. Key current state:

- `internal/adapter/database/` (Postgres/GORM) and `internal/adapter/email/` still exist but are **scheduled for removal**.
- `internal/application/usecases/order/usecase.go` is **entirely commented out** and must be rewritten split per use case file.
- `internal/application/services/order.go` is a GORM-backed service also **scheduled for removal**.
- DynamoDB repository (`internal/adapter/dynamo/`) is implemented but needs optimistic concurrency (`Version` + `ConditionExpression`) and an idempotency store added.

## Auth model

The auth middleware (`internal/adapter/http/middlewares/auth.go`) reads `X-User-Id`, `X-User-Email`, and `X-User-Role` headers — it does **not** validate JWT. Authentication is delegated to an upstream API Gateway + Lambda Authorizer. The `RoleRequired` middleware enforces roles: `administrator`, `attendant`, `mechanic`.

Public routes (`GET /api/orders/:id/approve` and `GET /api/orders/:id/reject`) have no auth middleware — the UUID v4 order ID acts as the only gate.

## Order state machine

```
RECEIVED → IN_ANALYSIS → ANALYSIS_FINISHED → AWAITING_APPROVAL
AWAITING_APPROVAL → REJECTED
AWAITING_APPROVAL → AWAITING_STOCK_CONSULT  (publishes order.approved → SNS)
AWAITING_STOCK_CONSULT → IN_PROGRESS         (consumes stock.available from SQS)
AWAITING_STOCK_CONSULT → AWAITING_STOCK_ORDER (consumes stock.unavailable)
AWAITING_STOCK_ORDER → IN_PROGRESS           (consumes stock.updated)
IN_PROGRESS → FINISHED                       (publishes order.finished → SNS)
FINISHED → AWAITING_PAYMENT                  (consumes payment.generated)
AWAITING_PAYMENT → PAYMENT_APPROVED          (consumes payment.approved)
AWAITING_PAYMENT → PAYMENT_FAILED            (consumes payment.failed)
PAYMENT_APPROVED → DELIVERED
```

Transitions must be validated via a `canTransition map[Status][]Status` table in `internal/domain/status.go`.

## Observability

Datadog tracing (`dd-trace-go`), profiling, and DogStatsD metrics are wired at startup. If `DD_AGENT_HOST` is unset, metrics fall back to `NoopOrderMetrics`. Trace IDs are injected into Gin's zap logger via `ginzap`.

## Deployment

- **Local k8s**: `make up` uses `kind` + kustomize overlays at `k8s/overlays/local/`.
- **AWS**: EKS + RDS (Postgres) via kustomize overlays at `k8s/overlays/aws/`. Terraform in `infra/`.
- Secrets are loaded from `.env` (never committed); the Makefile `include .env` + `export` pattern passes them to kubectl.

## Rules

- Do not excessive commentaries. Just use them only in complex logic cases.
- Do not commit or push automatically
- Use the existing patterns in the system.
