# Migration Notes — monolith → Orders microservice

This document records the key decisions, non-obvious changes, and design
rationale made during the migration from `tech-challenge-s1` (Go + Gin +
PostgreSQL monolith) to this standalone microservice (DynamoDB + SNS/SQS).
Read it alongside the implementation plan in `migration-v2.md`.

---

## Phase 1 — PostgreSQL/GORM removal

**Removed entirely:**
- `internal/adapter/database/` (Postgres + seeder)
- `internal/adapter/email/` (Mailtrap — responsibility moved to Lambda Notification)
- `internal/application/services/order.go` (GORM-backed service)
- `internal/application/ports/base_repository.go` (generic GORM repo)
- All related mocks and the legacy e2e test file

**Auth middleware change:** `X-User-Id` was validated as a numeric ID. Changed to
accept any non-empty string (UUIDs from the API Gateway authorizer). Tests updated
accordingly.

**Config rewrite:** Removed `DataBaseConfig` and `MailTrapConfig` sections.
Added `DynamoDBConfig`, `SNSConfig`, `SQSConfig`, `HTTPClientsConfig`. Fixed a
broken `getEnvInt` that used an invalid type assertion — replaced with
digit-by-digit parsing.

**docker-compose rewrite:** Replaced the `db` (Postgres) service with
`dynamodb-local`, `localstack`, `bootstrap`, `worker`, `users-mock`,
`stock-mock`, and `datadog-agent`.

---

## Phase 2 — Domain rewrite

**State machine:** The monolith had no state machine. A `canTransition
map[Status][]Status` table was added to `status.go`. Every `Order` method calls
the private `transition(to, actor, reason)` helper, which validates the
transition, appends a `HistoryEntry`, increments `Version`, and updates
`LastStatusAt`.

**Status removed — `APPROVED`:** The flow now goes
`AWAITING_APPROVAL → AWAITING_STOCK_CONSULT` (via `Approve()`). Stock service
handles decrement asynchronously and responds with `stock.available` or
`stock.unavailable`. The separate `APPROVED` status from the monolith is gone.

**Snapshots:** `VehicleSnapshot` and `ItemSnapshot` replace the old GORM
associations. Items have a `Kind` field (`product | service | maintenance`) that
determines whether they touch stock (`product` only) and how they are priced.

**`PriceCents int64`:** Total order price in cents, frozen at `CompleteAnalysis`
as the sum of `Quantity × UnitPriceCents` for each item. Never updated after
that — downstream services (Payments) use this value from the SNS event payload.

**History accumulation:** Every transition appends to `order.History`. The
`HistoryEntry` records `From`, `To`, `At`, `Actor`, and an optional `Reason`.
This is stored as a nested list in DynamoDB.

---

## Phase 3 — DynamoDB model and infra

**Optimistic concurrency:** `Save()` uses
`ConditionExpression: attribute_not_exists(PK) OR #ver = :v_prev`
where `:v_prev = order.Version - 1`. The domain increments `Version` on every
transition. `ConditionalCheckFailedException` → `ErrConcurrencyConflict`.

**Idempotency store:** Uses the same `orders` table with
`PK = PROCESSED#<event_id>`, `SK = META`, and a 7-day TTL on `ExpiresAt`.
`MarkProcessed` uses `attribute_not_exists(PK)` — returns `(true, nil)` first
time, `(false, nil)` if already recorded. Called at the top of every
`Handle*` use case before applying any transition.

**LocalStack bootstrap:** `deploy/localstack/bootstrap.sh` creates the DynamoDB
table (with GSI1 by status, GSI2 by customer vehicle, and TTL on `ExpiresAt`),
three SNS topics, and seven SQS queues. Each queue gets a DLQ with
`maxReceiveCount=5` and an SNS subscription with `FilterPolicy` on `event_type`.

**Mockoon stubs:** `deploy/mocks/users.json` and `stock.json` simulate the
Users and Stock APIs for local development. Responses use Mockoon's
`{{urlParam 'id'}}` template helper.

---

## Phase 4 — Ports and use cases

**`CustomerID` added to `Order`:** The domain aggregate stores the customer ID
(from `UsersClient.GetCustomerVehicle`) at creation time. This avoids an extra
`GetCustomerVehicle` call in `RequestApproval` and `CompleteWork`, where only
the customer's name/email (from `GetCustomer`) is needed.

**One struct, methods in separate files:** The `Service` struct is defined in
`service.go`; each use case is a method on it in its own file. This is idiomatic
Go for splitting a large type across files and keeps the `ports.OrderUseCase`
interface satisfied by a single type.

**`Handle*` methods are not part of `OrderUseCase`:** The HTTP-facing interface
only exposes the eight synchronous use cases. The six async event handlers
(`HandleStock*`, `HandlePayment*`) are methods on `Service` but called directly
by SQS consumers — no interface wrapping needed.

**Parallel calls in `CompleteAnalysis`:** `CheckProductsBatch` (Stock) and
`GetMaintenancesBatch` (Users) are called concurrently via `sync.WaitGroup`.
Both errors are checked after `wg.Wait()` — if either fails, the use case returns
before transitioning the order.

---

## Phase 5 — HTTP clients

**Shared `baseClient`:** Both `UsersHTTPClient` and `StockHTTPClient` embed
`baseClient`, which wraps every call with:
1. Retry — exponential backoff, max 3 attempts, only on 5xx or network/timeout
   errors. 4xx errors use `backoff.Permanent` to stop retrying immediately.
2. Circuit breaker — opens after 5 consecutive failures, resets after 30 s.

**Header propagation:** `X-Request-ID` and `traceparent` are read from context
keys injected by `clients.WithRequestHeaders(ctx, requestID, traceparent)`.
The HTTP middleware should call this helper before passing ctx to use cases.

**Maintenance quantity defaults to 1:** The Users API returns maintenance
details (id, name, price) but no quantity. `GetMaintenancesBatch` maps each
result to `ItemSnapshot{Quantity: 1}`. If future requirements need configurable
quantities for maintenances, the input type would need to change.

---

## Phase 6 — SNS publisher and SQS consumers

**Event type resolution via type switch:** The SNS publisher maps domain structs
to `event_type` strings using a type switch in `eventTypeOf`. Adding a new event
type requires updating this switch. The alternative (an `EventType()` method on
domain structs) was rejected to keep the domain free of infrastructure concerns.

**SNS → SQS envelope unwrapping:** When SNS delivers to SQS, the SQS message
body is a JSON notification envelope with the actual payload in a nested
`Message` string field. The generic consumer unwraps this two-level structure
before calling the handler.

**Delete-on-success, leave-on-error:** The consumer only calls `DeleteMessage`
if the handler returns `nil`. Errors leave the message visible so SQS retries it
up to `maxReceiveCount` times before sending to the DLQ. This is intentional —
a transient DynamoDB failure will be retried; a persistent one will surface in
the DLQ for manual inspection.

**Worker entrypoint:** `cmd/worker/main.go` uses `signal.NotifyContext` for
graceful shutdown and `errgroup` to run all six consumers concurrently. If any
consumer returns a non-nil error (not `ctx.Err()`), the `errgroup` cancels the
others.

**`AWS_ENDPOINT_URL`:** Both the API and worker read this env var to override the
SNS/SQS endpoint. When set (as in docker-compose), requests go to LocalStack
instead of AWS. The DynamoDB endpoint uses a separate `DYNAMODB_ENDPOINT` var
because DynamoDB Local runs on a different port than LocalStack.

---

## Phase 7 — HTTP handler cleanup

**Routes removed:** `POST /orders/:id/start-work` and `DELETE /orders/:id`.
The start-work concept no longer exists — stock transition to `IN_PROGRESS` is
driven by the `stock.available` SQS event after the customer approves the order.
Order deletion has no semantic in the SAGA flow.

**Public approval routes:** `GET /orders/:id/approve` and
`GET /orders/:id/reject` have no auth middleware. The UUID v4 of the order ID
acts as the only gate — acceptable for this academic context.

---

## Phase 8 — Dev infra finalization

**`condition: service_completed_successfully`:** The `api` and `worker`
docker-compose services use this condition for `depends_on: bootstrap`. This
ensures they only start after the bootstrap script has fully created the
DynamoDB table, SNS topics, and SQS queues. Without this, the services would
race to use infrastructure that doesn't exist yet.

**Worker command in dev:** Uses `go run ./cmd/worker` instead of a pre-built
binary. The `development` Dockerfile stage runs `air` (which only watches the
API), so there is no `./worker` binary available in that stage.

**Dockerfile:** The old `COPY .../adapter/email/templates` line was left over
from the monolith and pointed to a directory deleted in Phase 1. Removed.
Both `api` and `worker` binaries are now built in `production_builder` and
copied to the `production` stage. The default `CMD` is `/app/api`; the worker
deployment overrides it with `/app/worker`.
