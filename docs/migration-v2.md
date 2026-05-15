# Plano de implementação — `tech-challenge-orders`

## Contexto

O monolito `tech-challenge-s1` (Go + Gin + PostgreSQL) que gerencia uma oficina mecânica está sendo quebrado em 4 microsserviços (Orders, Payments, Stock, Users). O repositório `tech-challenge-orders` já existe com:

- Arquitetura hexagonal/Clean (Gin + GORM + DynamoDB SDK v2 parcial).
- Domain `Order` com structs portados do monolito (sem state machine ainda).
- Handlers HTTP completos.
- `OrderUseCase` 100% comentado (`/internal/application/usecases/order/usecase.go`).
- Repositório DynamoDB single-table com PK/SK + 2 GSIs já implementado.
- PostgreSQL/GORM coexistindo com DynamoDB (será removido).
- Sem mensageria, sem HTTP clients outbound, sem `cmd/worker`, sem LocalStack.

Decisões já tomadas pelo usuário neste planejamento:

1. **Persistência**: migrar 100% para DynamoDB e remover PostgreSQL/GORM.
2. **Nomenclatura de eventos**: adotar **RFC-Payments como fonte de verdade** — `payment.failed` (não `payment.rejected`), `notification.email.requested` para o link de checkout.
3. **Fluxo de estoque**: SAGA coreografada conforme RFC — `start-work` deixa de existir; após `order.approved`, Stock decrementa e responde com `stock.available`/`stock.unavailable`.
4. **Payloads ambíguos**: schemas propostos neste plano e revisados pelo usuário.

---

## Premissas e decisões arquiteturais

- **Go 1.25 + Gin** mantidos.
- **DynamoDB single-table** (`orders`) com PK/SK + GSI1 (status) + GSI2 (customer vehicle).
- **Snapshot vs. dados vivos**: produtos, serviços, manutenções e veículo congelados no `Order`. Cliente, funcionário e empresa **não** são congelados — consultados sob demanda à API de Users via REST.
- **Autenticação**: API Gateway + Lambda Authorizer (fora do escopo). Orders apenas lê `X-User-ID`, `X-User-Email`, `X-User-Role`.
- **Fan-out SNS → SQS**: cada serviço publica em **um** tópico SNS próprio; consumidores assinam via fila SQS dedicada com `FilterPolicy` por `event_type`.
- **Idempotência**: chave `event_id` (UUID v4) em todo evento; consumidores gravam `PROCESSED#<event_id>` no DynamoDB com TTL de 7 dias antes de aplicar a transição.
- **Concorrência otimista**: campo `Version` no agregado + `ConditionExpression` no `PutItem`.
- **Localstack** para SNS + SQS, **dynamodb-local** para DynamoDB no dev.

---

## Estado atual vs. alvo

| Camada           | Atual                                 | Alvo                                                                  |
| ---------------- | ------------------------------------- | --------------------------------------------------------------------- |
| Domain           | structs do monolito sem state machine | agregado com state machine, snapshots imutáveis, `Version`, `History` |
| Persistência     | PostgreSQL (GORM) + DynamoDB          | DynamoDB only                                                         |
| Use cases        | comentados                            | implementados (state machine + publish + consume)                     |
| HTTP clients     | nenhum                                | `UsersClient`, `StockClient` com timeout/retry/circuit-breaker        |
| Mensageria       | nenhuma                               | publisher SNS + consumers SQS por evento                              |
| Worker           | inexistente                           | `cmd/worker` com goroutines por fila                                  |
| Dev infra        | docker-compose com Postgres           | docker-compose com dynamodb-local + localstack + bootstrap            |
| Email (Mailtrap) | direto do Orders                      | removido (responsabilidade do Lambda notification)                    |

---

## Estrutura final do repositório

```
tech-challenge-orders/
├── cmd/
│   ├── api/main.go              # HTTP server (Gin)
│   └── worker/main.go           # SQS consumers
├── internal/
│   ├── domain/
│   │   ├── order.go             # agregado raiz + Version + History
│   │   ├── status.go            # enum + canTransition map
│   │   ├── snapshot.go          # VehicleSnapshot, ItemSnapshot (kind: product|service|maintenance)
│   │   ├── history.go           # HistoryEntry
│   │   └── events.go            # tipos dos eventos de domínio
│   ├── application/
│   │   ├── ports/
│   │   │   ├── order_repository.go
│   │   │   ├── event_publisher.go
│   │   │   ├── idempotency_store.go
│   │   │   ├── users_client.go
│   │   │   └── stock_client.go
│   │   └── usecases/order/
│   │       ├── create.go
│   │       ├── assign.go
│   │       ├── complete_analysis.go
│   │       ├── request_approval.go
│   │       ├── approve.go            # transita p/ AWAITING_STOCK_CONSULT + publica order.approved
│   │       ├── reject.go
│   │       ├── complete_work.go      # transita p/ FINISHED + publica order.finished
│   │       ├── archive.go
│   │       ├── handle_stock_available.go
│   │       ├── handle_stock_unavailable.go
│   │       ├── handle_stock_updated.go
│   │       ├── handle_payment_generated.go
│   │       ├── handle_payment_approved.go
│   │       └── handle_payment_failed.go
│   └── adapter/
│       ├── config/config.go
│       ├── http/
│       │   ├── handlers/order.go
│       │   ├── middlewares/auth.go    # lê X-User-* (sem validação JWT)
│       │   ├── middlewares/role.go
│       │   ├── response/
│       │   └── router.go
│       ├── dynamo/
│       │   ├── client.go              # já existe; aceitar endpoint custom p/ localstack
│       │   ├── model/order_item.go    # já existe (estender com Version, History)
│       │   ├── repository/order_repository.go  # já existe (adicionar ConditionExpression)
│       │   └── idempotency_store.go   # novo
│       ├── sns/publisher.go           # SNS publisher
│       ├── sqs/
│       │   ├── consumer.go            # consumer genérico c/ long-polling + idempotência
│       │   └── consumers/             # cada um aciona um handler de use case
│       └── clients/
│           ├── users_client.go        # GET /customer-vehicles/:id, /customers/:id, /users/:id, /maintenances/check-batch
│           └── stock_client.go        # POST /products/check-batch
├── pkg/                               # manter o que é agnóstico (monetary, converters)
├── deploy/localstack/
│   ├── bootstrap.sh                   # cria tabela + tópicos + filas
│   └── policies/                      # JSON de FilterPolicy por subscription
├── docker-compose.yml                 # api + worker + dynamodb-local + localstack + bootstrap
├── Dockerfile                         # multi-stage (já existe)
├── Makefile
└── README.md
```

---

## Schemas de eventos (payloads completos)

Todo evento publicado segue o envelope:

```json
{
  "event_id": "uuid-v4",
  "event_type": "order.approved",
  "event_version": "1",
  "occurred_at": "2026-05-14T13:42:11Z",
  "correlation_id": "uuid-v4",
  "data": {
    /* específico por evento */
  }
}
```

`event_type` é replicado como **MessageAttribute** no SNS para suportar `FilterPolicy` por subscription.

### Tópico `orders.events` (publicado por Orders)

**`order.approval-requested`** — consumido por Lambda Notification

```json
{
  "order_id": "uuid",
  "customer": { "id": "uuid", "name": "...", "email": "..." },
  "vehicle": { "plate": "...", "name": "...", "year": 2020, "brand": "..." },
  "diagnostic_note": "...",
  "items": [
    {
      "id": "uuid",
      "kind": "product|service|maintenance",
      "name": "...",
      "quantity": 2,
      "unit_price_cents": 12000
    }
  ],
  "total_cents": 24000,
  "approval_url": "https://api.x/orders/<id>/approve?token=...",
  "reject_url": "https://api.x/orders/<id>/reject?token=...",
  "expires_at": "2026-05-21T13:42:11Z"
}
```

**`order.approved`** — consumido por Stock

```json
{
  "order_id": "uuid",
  "items": [{ "product_id": "uuid", "quantity": 2 }],
  "approved_at": "2026-05-14T13:42:11Z"
}
```

> Apenas itens do tipo `product` entram aqui — services/maintenances não tocam estoque.

**`order.finished`** — consumido por Payments

```json
{
  "order_id": "uuid",
  "customer_id": "uuid",
  "customer_email": "cliente@exemplo.com",
  "amount_cents": 24000,
  "currency": "BRL",
  "items": [
    {
      "id": "uuid",
      "kind": "product|service|maintenance",
      "title": "...",
      "quantity": 2,
      "unit_price_cents": 12000
    }
  ],
  "finished_at": "2026-05-14T15:10:00Z"
}
```

### Tópico `stock.events` (consumido por Orders)

**`stock.available`**

```json
{ "order_id": "uuid", "checked_at": "2026-05-14T13:43:00Z" }
```

**`stock.unavailable`**

```json
{
  "order_id": "uuid",
  "missing_items": [
    { "product_id": "uuid", "missing_quantity": 1, "available_quantity": 0 }
  ],
  "detected_at": "2026-05-14T13:43:00Z"
}
```

**`stock.updated`** — Stock obtém `order_id` da tabela Backorder (RFC-Stock §Backorder)

```json
{
  "order_id": "uuid",
  "restocked_items": [{ "product_id": "uuid", "quantity": 1 }],
  "updated_at": "2026-05-14T16:00:00Z"
}
```

### Tópico `payments.events` (consumido por Orders)

**`payment.generated`**

```json
{
  "order_id": "uuid",
  "payment_id": "uuid",
  "preference_id": "mp-pref-id",
  "checkout_url": "https://mpago.la/...",
  "amount_cents": 24000,
  "currency": "BRL",
  "generated_at": "2026-05-14T15:10:30Z"
}
```

**`payment.approved`**

```json
{
  "order_id": "uuid",
  "payment_id": "uuid",
  "preference_id": "mp-pref-id",
  "amount_cents": 24000,
  "currency": "BRL",
  "approved_at": "2026-05-14T15:20:00Z"
}
```

**`payment.failed`** (nomenclatura RFC-Payments)

```json
{
  "order_id": "uuid",
  "payment_id": "uuid",
  "preference_id": "mp-pref-id",
  "amount_cents": 24000,
  "currency": "BRL",
  "reason": "rejected|cancelled|expired",
  "failed_at": "2026-05-14T15:25:00Z"
}
```

---

## Topologia de filas (Fan-out SNS → SQS)

Tópicos SNS:

- `orders-events-topic`
- `stock-events-topic`
- `payments-events-topic`

Filas SQS (Orders consome) — uma por evento, cada uma com sua DLQ (`maxReceiveCount=5`):

| Fila                             | Assina tópico         | FilterPolicy `event_type` |
| -------------------------------- | --------------------- | ------------------------- |
| `orders-stock-available-queue`   | stock-events-topic    | `["stock.available"]`     |
| `orders-stock-unavailable-queue` | stock-events-topic    | `["stock.unavailable"]`   |
| `orders-stock-updated-queue`     | stock-events-topic    | `["stock.updated"]`       |
| `orders-payment-generated-queue` | payments-events-topic | `["payment.generated"]`   |
| `orders-payment-approved-queue`  | payments-events-topic | `["payment.approved"]`    |
| `orders-payment-failed-queue`    | payments-events-topic | `["payment.failed"]`      |

Fila do Lambda Notification (fora do escopo, mas declarada na infra):

- `notification-approval-requested-queue` → orders-events-topic, filtro `order.approval-requested`

---

## Máquina de estados final

```
RECEIVED          → IN_ANALYSIS           (POST /orders/:id/assign)
IN_ANALYSIS       → ANALYSIS_FINISHED     (POST /orders/:id/complete-analysis)
ANALYSIS_FINISHED → AWAITING_APPROVAL     (POST /orders/:id/request-approval)  + publish order.approval-requested
AWAITING_APPROVAL → REJECTED              (GET/POST /orders/:id/reject)
AWAITING_APPROVAL → AWAITING_STOCK_CONSULT(GET/POST /orders/:id/approve)        + publish order.approved
AWAITING_STOCK_CONSULT → IN_PROGRESS      (consume stock.available)
AWAITING_STOCK_CONSULT → AWAITING_STOCK_ORDER (consume stock.unavailable)
AWAITING_STOCK_ORDER → IN_PROGRESS        (consume stock.updated, match por order_id)
IN_PROGRESS       → FINISHED              (POST /orders/:id/complete-work)     + publish order.finished
FINISHED          → AWAITING_PAYMENT      (consume payment.generated)
AWAITING_PAYMENT  → PAYMENT_APPROVED      (consume payment.approved)
AWAITING_PAYMENT  → PAYMENT_FAILED        (consume payment.failed)
PAYMENT_APPROVED  → DELIVERED             (POST /orders/:id/archive)
```

Tabela `canTransition` em `internal/domain/status.go` valida cada transição antes de aplicar.

---

## Fases de implementação

### Fase 1 — Limpeza e bootstrap (PostgreSQL out)

Arquivos/diretórios a **remover**:

- `internal/adapter/database/` (Postgres + seeder)
- `internal/adapter/database/model/order/`
- `internal/adapter/database/repository/`
- `internal/adapter/email/` (Mailtrap — vai para o Lambda Notification)
- `internal/application/services/order.go` (lógica vai para use cases)
- `internal/application/ports/base_repository.go` (genérico GORM)
- `internal/application/ports/order/repository.go` (substituir pela port única abaixo)
- Dependências em `go.mod`: `gorm.io/gorm`, `gorm.io/driver/postgres`

Ajustes:

- `cmd/api/main.go`: remover `database.Init`, `AutoMigrate`, `Seed`, `email.NewSendtrap`. Adicionar `dynamo.NewClient`, `sns.NewPublisher`.
- `config/config.go`: remover seção `Database` e `MailTrap`. Adicionar `DynamoDB.Endpoint`, `DynamoDB.TableName`, `SNS.OrdersTopicARN`, `SQS.*QueueURL`, `Users.BaseURL`, `Stock.BaseURL`.
- `docker-compose.yml`: trocar `db` (postgres) por `dynamodb-local` + `localstack`. Adicionar serviço `bootstrap` que executa `deploy/localstack/bootstrap.sh` (cria tabela + tópicos + filas).

### Fase 2 — Domain (state machine + snapshots + history)

Editar `internal/domain/order.go`:

- Adicionar `Version int`, `CreatedAt`, `UpdatedAt`, `PaymentDate *time.Time`, `History []HistoryEntry`.
- Converter `Vehicle` → `VehicleSnapshot` e `Item` → `ItemSnapshot` (com `Kind` = `product|service|maintenance`).
- Remover backref circular `OrderItems.Order`.
- Métodos com validação de transição: `Assign`, `CompleteAnalysis`, `RequestApproval`, `Approve`, `Reject`, `CompleteWork`, `Archive`, `MarkStockAvailable`, `MarkStockUnavailable`, `MarkStockReady`, `MarkPaymentGenerated`, `MarkPaymentApproved`, `MarkPaymentFailed`. Cada método:
  - chama `canTransition(from, to)` → retorna `ErrInvalidTransition` se inválido;
  - aplica transição, atualiza `LastStatusAt` + datas-chave;
  - anexa `HistoryEntry{from, to, at, actor, reason}`;
  - incrementa `Version`.

Criar:

- `internal/domain/status.go` — adicionar `REJECTED`, `AWAITING_STOCK_CONSULT`, `AWAITING_STOCK_ORDER`, `AWAITING_PAYMENT`, `PAYMENT_APPROVED`, `PAYMENT_FAILED` ao enum + tabela `canTransition map[Status][]Status`.
- `internal/domain/snapshot.go` — `VehicleSnapshot`, `ItemSnapshot`.
- `internal/domain/history.go` — `HistoryEntry`.
- `internal/domain/events.go` — `OrderApprovalRequested`, `OrderApproved`, `OrderFinished` (com `Marshal() []byte`).

Testes unitários por transição (table-driven), incluindo casos inválidos.

### Fase 3 — Modelagem DynamoDB (estender o existente)

`internal/adapter/dynamo/model/order_item.go`:

- Adicionar `Version`, `CreatedAt`, `UpdatedAt`, `PaymentDate`, `History`.
- Estender `FromDomain` / `ToDomain`.

`internal/adapter/dynamo/repository/order_repository.go`:

- `Save` passa a usar `ConditionExpression: attribute_not_exists(PK) OR Version = :v_prev`, com `:v_prev = order.Version - 1`. Em falha, retornar `ErrConcurrencyConflict`.

Criar `internal/adapter/dynamo/idempotency_store.go`:

- Tabela única (mesma `orders`) com PK `PROCESSED#<event_id>`, SK `META`, TTL 7 dias.
- Métodos `MarkProcessed(ctx, eventID) (bool, error)` — retorna `false` se já existia (uso de `attribute_not_exists` no `PutItem`).

Bootstrap LocalStack (`deploy/localstack/bootstrap.sh`):

- `awslocal dynamodb create-table` com `orders`, GSI1 (`GSI1PK`/`GSI1SK`), GSI2 (`GSI2PK`/`GSI2SK`).
- Habilitar TTL no atributo `ExpiresAt`.

### Fase 4 — Ports e use cases (descomentar e reescrever)

Substituir `internal/application/ports/order.go` e o legado `order/repository.go` por:

```go
type OrderRepository interface {
    Save(ctx, *Order) error
    FindByID(ctx, id) (*Order, error)
    ListByStatus(ctx, status, page) (PageResult, error)
    ListByCustomerVehicle(ctx, customerVehicleID, page) (PageResult, error)
}

type EventPublisher interface {
    Publish(ctx, topicARN string, evt DomainEvent) error
}

type IdempotencyStore interface {
    MarkProcessed(ctx, eventID string) (bool, error)
}

type UsersClient interface {
    GetCustomerVehicle(ctx, id) (CustomerVehicle, error)  // veículo + customer associado
    GetCustomer(ctx, id) (Customer, error)
    GetUser(ctx, id) (User, error)
    GetMaintenancesBatch(ctx, ids []string) ([]ItemSnapshot, error)
}

type StockClient interface {
    CheckProductsBatch(ctx, items []ItemQty) (CheckBatchResult, error)
}
```

Reescrever `internal/application/usecases/order/usecase.go` removendo o `/* … */` e dividindo em arquivos por caso. Mapeamento principal:

| Use case           | Lógica                                                                                                                                                          |
| ------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `CreateOrder`      | Busca veículo via `UsersClient.GetCustomerVehicle`. Cria `Order{Status: RECEIVED}` com `VehicleSnapshot`. `Save`.                                               |
| `AssignOrder`      | Lê `X-User-ID` do contexto. `order.Assign(employeeID)`. `Save`.                                                                                                 |
| `CompleteAnalysis` | Chama `StockClient.CheckProductsBatch` e `UsersClient.GetMaintenancesBatch` em paralelo. Congela `Items`. `order.CompleteAnalysis(items, note, price)`. `Save`. |
| `RequestApproval`  | `order.RequestApproval()`. `Save`. `Publisher.Publish(ordersTopic, OrderApprovalRequested{...})`.                                                               |
| `ApproveOrder`     | `order.Approve()`. `Save`. `Publisher.Publish(ordersTopic, OrderApproved{...})`.                                                                                |
| `RejectOrder`      | `order.Reject()`. `Save`.                                                                                                                                       |
| `CompleteWork`     | `order.CompleteWork()`. `Save`. `Publisher.Publish(ordersTopic, OrderFinished{...})`.                                                                           |
| `ArchiveOrder`     | Verifica `Status == PAYMENT_APPROVED`. `order.Archive()`. `Save`.                                                                                               |
| `HandleStock*`     | `idempotency.MarkProcessed(eventID)`. `FindByID`. Aplica transição correspondente. `Save`.                                                                      |
| `HandlePayment*`   | Idem.                                                                                                                                                           |

Cada use case recebe ports via construtor — sem instanciar SDK direto.

### Fase 5 — HTTP clients outbound

`internal/adapter/clients/users_client.go` e `stock_client.go`:

- `net/http` com `http.Client{Timeout: 2s}`.
- Retry exponencial (3 tentativas) só em 5xx/timeout (usar `cenkalti/backoff/v4`).
- Circuit breaker (`sony/gobreaker`) por cliente.
- Headers propagados: `X-Request-ID`, `traceparent`.
- Base URLs vêm de `cfg.Users.BaseURL` e `cfg.Stock.BaseURL`.
- Endpoints (baseados no swagger atual + RFCs):
  - Users: `GET /customer-vehicles/:id`, `GET /customers/:id`, `GET /users/:id`, `GET /maintenances/check-batch?ids=…`
  - Stock: `POST /products/check-batch` body `{"items":[{"id":"…","qty":1}]}`

Mocks gerados via `mockery` para testes unitários dos use cases.

### Fase 6 — SNS publisher + SQS consumers

`internal/adapter/sns/publisher.go`:

- `Publish(ctx, topicARN, event)` com `MessageAttributes` = `{ event_type: <string>, event_version: <string> }`.
- Usa `aws-sdk-go-v2/service/sns`.

`internal/adapter/sqs/consumer.go` (genérico):

- Long-polling 20s.
- Para cada mensagem: decodifica envelope SNS → evento → chama handler injetado.
- Em sucesso: `DeleteMessage`. Em erro: deixa expirar para retry/DLQ.

`internal/adapter/sqs/consumers/*.go` — 6 consumers, cada um:

- Recebe `QueueURL`, `IdempotencyStore`, use case correspondente.
- Em `Run(ctx)`: loop até `ctx.Done()`.

`cmd/worker/main.go`:

- Inicializa config + AWS clients + repo + idempotency store + 6 consumers.
- Cada consumer em sua goroutine (`errgroup`).
- Graceful shutdown em SIGTERM/SIGINT.

### Fase 7 — Handlers HTTP (ajustes)

`internal/adapter/http/handlers/order.go`:

- Substituir dependência `ports.OrderService` por use cases concretos (ou um `OrderUseCases` agregador).
- Remover métodos `start-work` (não existe mais — Stock que dispara).
- Rota pública `GET /orders/:id/approve` e `GET /orders/:id/reject`: sem autenticação. A obscuridade do UUID v4 do `order_id` é o único gate — aceitável no contexto pós-grad. URLs no evento `order.approval-requested` são geradas concatenando `cfg.PublicBaseURL` + `/api/orders/<id>/approve|reject`.
- Middleware `AuthRequired` continua lendo `X-User-*`, mas pula nas rotas públicas.
- Remover handler/rota `DELETE /orders/:id` (não há semântica de exclusão no fluxo SAGA).
- Manter `GET /orders/in-progress` como conveniência (atalho para listagem por status).

`internal/adapter/http/router.go`:

- Manter as rotas atuais menos `POST /orders/:id/start-work` e `DELETE /orders/:id` (removidas).
- Manter `GET /orders/in-progress`.

### Fase 8 — Dev infra (docker-compose + localstack)

`docker-compose.yml`:

```yaml
services:
  dynamodb-local:
    image: amazon/dynamodb-local
    ports: ["8000:8000"]

  localstack:
    image: localstack/localstack:3
    environment:
      SERVICES: sns,sqs
      DEBUG: 1
    ports: ["4566:4566"]

  bootstrap:
    image: amazon/aws-cli
    depends_on: [dynamodb-local, localstack]
    volumes: ["./deploy/localstack:/bootstrap"]
    entrypoint: ["/bootstrap/bootstrap.sh"]

  api:
    build: { context: ., target: development }
    depends_on: [bootstrap]
    environment:
      DYNAMODB_ENDPOINT: http://dynamodb-local:8000
      AWS_ENDPOINT_URL: http://localstack:4566
      AWS_REGION: us-east-1
      AWS_ACCESS_KEY_ID: test
      AWS_SECRET_ACCESS_KEY: test
      ORDERS_TOPIC_ARN: arn:aws:sns:us-east-1:000000000000:orders-events-topic
      USERS_BASE_URL: http://users-mock:8081
      STOCK_BASE_URL: http://stock-mock:8082
      PUBLIC_BASE_URL: http://localhost:8080
    ports: ["8080:8080"]

  worker:
    build: { context: ., target: development }
    command: ["./worker"]
    depends_on: [bootstrap]
    environment: # mesmo env do api + URLs das 6 filas

  users-mock:
    image: mockoon/cli
    command: ["--data", "/data/users.json", "--port", "8081"]
    volumes: ["./deploy/mocks:/data"]

  stock-mock:
    image: mockoon/cli
    command: ["--data", "/data/stock.json", "--port", "8082"]
    volumes: ["./deploy/mocks:/data"]
```

`deploy/localstack/bootstrap.sh` — cria:

- Tabela DynamoDB `orders` com 2 GSIs + TTL.
- Tópicos: `orders-events-topic`, `stock-events-topic`, `payments-events-topic`.
- 7 filas (6 Orders + 1 Notification) cada uma com DLQ + RedrivePolicy.
- Subscriptions SNS → SQS com `FilterPolicy` por `event_type`.

Mocks `deploy/mocks/users.json` e `stock.json` para Mockoon — opcional, simulam o que será real depois.

### Fase 9 — Testes

| Camada                      | Tipo                                         | Como                                         |
| --------------------------- | -------------------------------------------- | -------------------------------------------- |
| Domain                      | unit table-driven                            | `testing` + `testify`                        |
| Use cases síncronos         | unit com mocks de ports                      | `mockery`                                    |
| Use cases async (`Handle*`) | unit com mocks; cobrir replay (idempotência) | `mockery`                                    |
| Repository Dynamo           | integration                                  | dynamodb-local via `docker-compose.test.yml` |
| Consumers SQS               | integration                                  | localstack                                   |
| Handlers                    | integration                                  | `httptest`                                   |
| E2E                         | cenário aprovar→consumir stock→pagar         | dynamodb-local + localstack + mockoon        |

CI: unit em todo PR; integration + e2e em `main`.

---

## Configuração de variáveis de ambiente (`.env.example`)

```
HTTP_PORT=8080
ENV=development

AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=test
AWS_SECRET_ACCESS_KEY=test

# DynamoDB
DYNAMODB_ENDPOINT=http://localhost:8000
DYNAMODB_TABLE=orders

# SNS / SQS (via localstack)
AWS_ENDPOINT_URL=http://localhost:4566
ORDERS_TOPIC_ARN=arn:aws:sns:us-east-1:000000000000:orders-events-topic
SQS_STOCK_AVAILABLE_URL=http://localhost:4566/000000000000/orders-stock-available-queue
SQS_STOCK_UNAVAILABLE_URL=http://localhost:4566/000000000000/orders-stock-unavailable-queue
SQS_STOCK_UPDATED_URL=http://localhost:4566/000000000000/orders-stock-updated-queue
SQS_PAYMENT_GENERATED_URL=http://localhost:4566/000000000000/orders-payment-generated-queue
SQS_PAYMENT_APPROVED_URL=http://localhost:4566/000000000000/orders-payment-approved-queue
SQS_PAYMENT_FAILED_URL=http://localhost:4566/000000000000/orders-payment-failed-queue

# Clients
USERS_BASE_URL=http://localhost:8081
STOCK_BASE_URL=http://localhost:8082
HTTP_CLIENT_TIMEOUT_MS=2000

# Base pública (compõe approval_url no evento order.approval-requested)
PUBLIC_BASE_URL=http://localhost:8080
```

---

## Arquivos críticos a serem modificados

- `internal/domain/order.go` — adicionar campos + state machine.
- `internal/domain/status.go` — novos estados + `canTransition`.
- `internal/domain/snapshot.go` (novo) e `internal/domain/history.go` (novo).
- `internal/adapter/dynamo/model/order_item.go` — incluir `Version`, `History`, `PaymentDate`.
- `internal/adapter/dynamo/repository/order_repository.go` — adicionar `ConditionExpression`.
- `internal/adapter/dynamo/idempotency_store.go` (novo).
- `internal/adapter/sns/publisher.go` (novo).
- `internal/adapter/sqs/consumer.go` (novo) e 6 consumers em `internal/adapter/sqs/consumers/`.
- `internal/adapter/clients/users_client.go` (novo) e `stock_client.go` (novo).
- `internal/application/ports/*` — substituir interfaces antigas.
- `internal/application/usecases/order/*.go` — descomentar e dividir em arquivos por use case.
- `internal/adapter/http/handlers/order.go` — remover `StartWork`, ajustar dependências.
- `internal/adapter/config/config.go` — remover Postgres/Mailtrap, adicionar SNS/SQS/Clients.
- `cmd/api/main.go` — remover GORM, adicionar SNS publisher e clients.
- `cmd/worker/main.go` (novo).
- `docker-compose.yml` — trocar Postgres por dynamodb-local + localstack.
- `deploy/localstack/bootstrap.sh` (novo).
- `go.mod` — remover `gorm.io/*`, adicionar `aws-sdk-go-v2/service/sns` e `service/sqs`, `cenkalti/backoff/v4`, `sony/gobreaker`.

---

## Verificação ponta-a-ponta

1. **Subir ambiente local**:

   ```bash
   docker compose up -d
   docker compose logs bootstrap   # garantir tabela + filas criadas
   ```

2. **Fluxo feliz via curl**:

   ```bash
   # criar ordem
   curl -X POST localhost:8080/api/orders -H "X-User-ID: emp-1" \
     -d '{"customer_vehicle_id":"cv-1","vehicle_kilometers":50000,"company_id":"co-1","note":"Barulho no motor"}'
   # → 201, status RECEIVED
   curl -X POST localhost:8080/api/orders/<id>/assign -H "X-User-ID: emp-1"
   curl -X POST localhost:8080/api/orders/<id>/complete-analysis -H "X-User-ID: emp-1" \
     -d '{"diagnostic_note":"trocar correia","products":[{"id":"p1","quantity":1}],"maintenances":["m1"]}'
   curl -X POST localhost:8080/api/orders/<id>/request-approval -H "X-User-ID: emp-1"
   # → publica order.approval-requested no SNS; verifique via:
   awslocal --endpoint-url=http://localhost:4566 sqs receive-message \
     --queue-url http://localhost:4566/000000000000/notification-approval-requested-queue
   curl "localhost:8080/api/orders/<id>/approve?token=…"
   # → publica order.approved
   # injeta stock.available manualmente:
   awslocal sns publish --topic-arn …stock-events-topic \
     --message '{"event_id":"e1","event_type":"stock.available","occurred_at":"…","data":{"order_id":"<id>"}}' \
     --message-attributes 'event_type={DataType=String,StringValue=stock.available}'
   # worker consome → status IN_PROGRESS
   curl -X POST localhost:8080/api/orders/<id>/complete-work -H "X-User-ID: emp-1"
   # injeta payment.generated, payment.approved seguidos
   curl -X POST localhost:8080/api/orders/<id>/archive -H "X-User-ID: emp-1"
   # → status DELIVERED
   ```

3. **Idempotência**: republicar o mesmo `stock.available` (mesmo `event_id`) → ordem permanece em `IN_PROGRESS`, nenhum erro, log indica skip.

4. **Concorrência otimista**: simular `Save` com `Version` desatualizada → retorna `ErrConcurrencyConflict`, mensagem volta para fila e é retentada.

5. **Testes**:

   ```bash
   go test ./...                              # unit
   docker compose -f docker-compose.test.yml up --abort-on-container-exit  # integration + e2e
   ```

6. **DLQ**: simular consumer com erro persistente (ex.: derrubar dynamodb-local) → mensagem deve aparecer em `*-dlq` após 5 tentativas. Verificar com `awslocal sqs receive-message`.

---

## Itens fora do escopo (delegar/alinhar com outros times)

- Implementação real das APIs de Users e Stock (este plano usa Mockoon no dev).
- Implementação do microsserviço Payments (apenas consumimos seus eventos).
- Lambda de notificação (`tech-challenge-notification`) — repositório separado.
- Provisionamento AWS real (Terraform) — pós-grad sem ambiente produtivo, fora do escopo.
- Migração de dados PG → DynamoDB — sem necessidade neste contexto acadêmico.

---

## Decisões finais consolidadas (após Q&A)

- **Aprovação pública**: sem token. UUID v4 do `order_id` é o gate.
- **`DELETE /orders/:id`**: removido.
- **`GET /orders/in-progress`**: mantido.
- **Estado interno de falha de pagamento**: `PAYMENT_FAILED` (alinha 1:1 com `payment.failed`).
