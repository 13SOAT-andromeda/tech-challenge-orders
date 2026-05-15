# Tech Challenge — Orders Microservice

Orders microservice for an automotive workshop management platform. Handles the
full lifecycle of a service order: creation, mechanic assignment, analysis,
customer approval, stock reservation, work completion, and payment.

Part of a four-microservice system: **Orders** (this), Payments, Stock, Users.

---

## Architecture

Hexagonal (Clean) architecture with three layers:

```
internal/
├── domain/          # Pure domain: Order aggregate, state machine, snapshots
├── application/
│   ├── ports/       # Interfaces (Repository, EventPublisher, UsersClient, ...)
│   └── usecases/order/  # One file per use case, injected with ports only
└── adapter/
    ├── http/        # Gin router, handlers, middlewares
    ├── dynamo/      # DynamoDB single-table repository + idempotency store
    ├── sns/         # SNS event publisher
    ├── sqs/         # Generic SQS consumer + 6 specific handlers
    ├── clients/     # Outbound HTTP to Users and Stock services
    ├── config/      # Env-based config
    └── metrics/     # Datadog DogStatsD (noop fallback)
```

Two entrypoints:

| Binary | Path | Role |
|---|---|---|
| `api` | `cmd/api/` | HTTP server (Gin) — handles synchronous requests |
| `worker` | `cmd/worker/` | SQS consumer — handles async events from Stock and Payments |

### Persistence

DynamoDB single-table (`orders`):

| Key | Value |
|---|---|
| `PK` | `ORDER#<uuid>` |
| `SK` | `META` |
| `GSI1PK/SK` | `STATUS#<status>` / `<dateIn>#<id>` — list by status |
| `GSI2PK/SK` | `CUSTOMERVEHICLE#<id>` / `<dateIn>#<id>` — list by vehicle |
| `PROCESSED#<event_id>` | Idempotency records (7-day TTL) |

Optimistic concurrency via `ConditionExpression: attribute_not_exists(PK) OR Version = :v_prev`.

### Messaging

Fan-out SNS → SQS. Orders publishes to `orders-events-topic` and consumes from
queues subscribed to `stock-events-topic` and `payments-events-topic`.

| Queue | Event | Effect |
|---|---|---|
| orders-stock-available-queue | `stock.available` | → `IN_PROGRESS` |
| orders-stock-unavailable-queue | `stock.unavailable` | → `AWAITING_STOCK_ORDER` |
| orders-stock-updated-queue | `stock.updated` | → `IN_PROGRESS` |
| orders-payment-generated-queue | `payment.generated` | → `AWAITING_PAYMENT` |
| orders-payment-approved-queue | `payment.approved` | → `PAYMENT_APPROVED` |
| orders-payment-failed-queue | `payment.failed` | → `PAYMENT_FAILED` |

### Order state machine

```
RECEIVED → IN_ANALYSIS → ANALYSIS_FINISHED → AWAITING_APPROVAL
AWAITING_APPROVAL → REJECTED
AWAITING_APPROVAL → AWAITING_STOCK_CONSULT  (publishes order.approved)
AWAITING_STOCK_CONSULT → IN_PROGRESS        (consumes stock.available)
AWAITING_STOCK_CONSULT → AWAITING_STOCK_ORDER (consumes stock.unavailable)
AWAITING_STOCK_ORDER → IN_PROGRESS          (consumes stock.updated)
IN_PROGRESS → FINISHED                      (publishes order.finished)
FINISHED → AWAITING_PAYMENT                 (consumes payment.generated)
AWAITING_PAYMENT → PAYMENT_APPROVED         (consumes payment.approved)
AWAITING_PAYMENT → PAYMENT_FAILED           (consumes payment.failed)
PAYMENT_APPROVED → DELIVERED
```

---

## Running locally

**Prerequisites:** Docker, Docker Compose.

```bash
cp .env.example .env   # fill in DD_API_KEY if you want Datadog
docker compose up -d
docker compose logs bootstrap  # confirm table + queues were created
```

The stack starts:
- `dynamodb-local` on port `8000`
- `localstack` (SNS + SQS) on port `4566`
- `bootstrap` — one-shot container that creates the table, topics and queues
- `api` on port `8080` (hot-reload via `air`)
- `worker` — SQS consumer (via `go run`)
- `users-mock` on port `8081` (Mockoon)
- `stock-mock` on port `8082` (Mockoon)

Health check: `curl http://localhost:8080/api/health`

---

## Development commands

```bash
go build ./...                            # build all packages
go test ./...                             # run all tests
go test ./internal/domain/... -run TestX  # run a single test
air                                       # hot-reload API locally (no Docker)
```

---

## HTTP API

All routes under `/api/`. Protected routes require `X-User-Id`, `X-User-Email`,
and `X-User-Role` headers (set by the API Gateway authorizer upstream — no JWT
validation in this service).

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/orders` | required | Create order |
| POST | `/orders/:id/assign` | required | Assign mechanic |
| POST | `/orders/:id/complete-analysis` | required | Complete analysis + freeze items |
| POST | `/orders/:id/request-approval` | required | Send approval request to customer |
| GET | `/orders/:id/approve` | none | Customer approves (UUID acts as token) |
| GET | `/orders/:id/reject` | none | Customer rejects |
| POST | `/orders/:id/complete-work` | required | Mark work done |
| POST | `/orders/:id/archive` | required | Archive after payment |
| GET | `/orders` | required | List all (not yet implemented) |
| GET | `/orders/:id` | required | Get by ID (not yet implemented) |
| GET | `/orders/in-progress` | required | List in-progress (not yet implemented) |

---

## Auth model

Authentication is delegated to an upstream API Gateway + Lambda Authorizer.
This service reads three headers injected by the authorizer:

- `X-User-Id` — user UUID
- `X-User-Email` — user email
- `X-User-Role` — one of `administrator`, `attendant`, `mechanic`

The `RoleRequired` middleware enforces role-based access. Public routes
(`/approve`, `/reject`) have no middleware — the order UUID is the only gate.

---

## Project layout

```
cmd/
  api/       HTTP server entrypoint
  worker/    SQS consumer entrypoint
internal/
  domain/    Order, Status, VehicleSnapshot, ItemSnapshot, HistoryEntry, events
  application/
    ports/   Repository, EventPublisher, IdempotencyStore, UsersClient, StockClient
    usecases/order/  create, assign, complete_analysis, request_approval,
                     approve, reject, complete_work, archive,
                     handle_stock_*, handle_payment_*
  adapter/
    config/        Env-based config (godotenv)
    dynamo/        DynamoDB client, repository, idempotency store
    http/          Router, handlers, middlewares, response helpers
    sns/           SNS publisher
    sqs/           Generic consumer + 6 handler factories
    clients/       UsersHTTPClient, StockHTTPClient (retry + circuit breaker)
    metrics/       Datadog DogStatsD or noop
deploy/
  localstack/  bootstrap.sh — creates DynamoDB table, SNS topics, SQS queues
  mocks/       Mockoon stubs for Users and Stock APIs
docs/
  migration-v2.md    Implementation plan (authoritative)
  migration-notes.md Key decisions and rationale per phase
```
