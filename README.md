# Tech Challenge — Microsserviço de Pedidos

Microsserviço de pedidos para uma plataforma de gestão de oficina automotiva. Gerencia o
ciclo de vida completo de uma ordem de serviço: criação, atribuição de mecânico, análise,
aprovação do cliente, reserva de estoque, conclusão do trabalho e pagamento.

Parte de um sistema de quatro microsserviços: **Pedidos** (este), Pagamentos, Catálogo (Estoque), Usuários.

---

## Arquitetura

Arquitetura Hexagonal (Clean) com três camadas:

```
internal/
├── domain/          # Domínio puro: agregado Order, máquina de estados, snapshots
├── application/
│   ├── ports/       # Interfaces (Repository, EventPublisher, UsersClient, ...)
│   └── usecases/order/  # Um arquivo por caso de uso, injetado apenas com ports
└── adapter/
    ├── http/        # Roteador Gin, handlers, middlewares
    ├── dynamo/      # Repositório single-table DynamoDB + idempotency store
    ├── sns/         # Publisher de eventos SNS + publisher de notificações
    ├── sqs/         # Consumidor genérico SQS + factories de handlers
    ├── clients/     # HTTP de saída para os serviços de Usuários e Catálogo
    ├── config/      # Configuração por variáveis de ambiente (godotenv)
    └── metrics/     # Datadog DogStatsD (fallback noop)
```

Dois entrypoints:

| Binário | Caminho | Função |
|---|---|---|
| `api` | `cmd/api/` | Servidor HTTP (Gin) — trata requisições síncronas |
| `worker` | `cmd/worker/` | Consumidor SQS — trata eventos assíncronos do Catálogo e Pagamentos |

### Persistência

Single-table DynamoDB (`orders`):

| Chave | Valor |
|---|---|
| `PK` | `ORDER#<uuid>` |
| `SK` | `META` |
| `GSI1PK/SK` | `STATUS#<status>` / `<dateIn>#<id>` — listagem por status |
| `GSI2PK/SK` | `CUSTOMERVEHICLE#<id>` / `<dateIn>#<id>` — listagem por veículo |
| `PROCESSED#<event_id>` | Registros de idempotência (TTL de 7 dias) |

Concorrência otimista via `ConditionExpression: attribute_not_exists(PK) OR Version = :v_prev`.

### Mensageria

Fan-out SNS → SQS. Pedidos publica no `orders-events-topic` e consome filas
inscritas nos tópicos de Catálogo e Pagamentos.

**API do Catálogo** publica um tipo de evento por tópico (sem filter policy):

| Fila | Tópico de origem | Evento | Efeito |
|---|---|---|---|
| orders-stock-reserved-queue | `catalog-stock-reserved-topic` | `STOCK_RESERVED` | → `IN_PROGRESS` |
| orders-stock-insufficient-queue | `catalog-stock-insufficient-topic` | `STOCK_INSUFFICIENT` | → `AWAITING_STOCK_ORDER` |
| orders-backorder-created-queue | `catalog-backorder-created-topic` | `BACKORDER_CREATED` | → `IN_PROGRESS` |

**API de Pagamentos** publica todos os eventos em um único tópico, filtrados pelo MessageAttribute `event_type`:

| Fila | Tópico de origem | Eventos (filtrados) | Efeito |
|---|---|---|---|
| orders-payment-events-queue | `payments-events-topic` | `payment.checkout_created` | → `AWAITING_PAYMENT` |
| | | `payment.approved` | → `PAYMENT_APPROVED` |
| | | `payment.failed` | → `PAYMENT_FAILED` |

Eventos do Catálogo carregam JSON puro (`order_id`, `type`, etc.) sem envelope — a chave de idempotência é derivada como `order_id#<TYPE>`. Eventos de Pagamentos carregam um envelope `{event_id, event_type, payload}`.

### Máquina de estados do pedido

```
RECEIVED → IN_ANALYSIS → ANALYSIS_FINISHED → AWAITING_APPROVAL
AWAITING_APPROVAL → REJECTED
AWAITING_APPROVAL → AWAITING_STOCK_CONSULT  (publica order.approved)
AWAITING_STOCK_CONSULT → IN_PROGRESS        (consome STOCK_RESERVED do Catálogo)
AWAITING_STOCK_CONSULT → AWAITING_STOCK_ORDER (consome STOCK_INSUFFICIENT do Catálogo)
AWAITING_STOCK_ORDER → IN_PROGRESS          (consome BACKORDER_CREATED do Catálogo)
IN_PROGRESS → FINISHED                      (publica order.finished)
FINISHED → AWAITING_PAYMENT                 (consome payment.checkout_created de Pagamentos)
AWAITING_PAYMENT → PAYMENT_APPROVED         (consome payment.approved)
AWAITING_PAYMENT → PAYMENT_FAILED           (consome payment.failed)
PAYMENT_APPROVED → DELIVERED
```

---

## Executando localmente

**Pré-requisitos:** Docker, Docker Compose.

```bash
cp .env.example .env   # preencha DD_API_KEY se quiser Datadog
docker compose up -d
docker compose logs bootstrap  # confirme que tabela e filas foram criadas
```

O stack inicializa:
- `dynamodb-local` na porta `8000`
- `localstack` (SNS + SQS) na porta `4566`
- `bootstrap` — container one-shot que cria a tabela, tópicos e filas
- `api` na porta `8080` (hot-reload via `air`)
- `worker` — consumidor SQS
- `users-mock` na porta `8081` (Mockoon)
- `stock-mock` na porta `8082` (Mockoon)
- `datadog-agent` — opcional, ativo apenas quando `DD_API_KEY` está definido

Health check: `curl http://localhost:8080/health`

---

## Comandos de desenvolvimento

```bash
go build ./...                            # compila todos os pacotes
go test ./...                             # executa todos os testes
go test ./internal/domain/... -run TestX  # executa um teste específico
air                                       # hot-reload da API localmente (sem Docker)
```

---

## HTTP API

Todas as rotas sob `/v1/`. Rotas protegidas exigem os headers `X-User-Id`, `X-User-Email`
e `X-User-Role` (injetados pelo autorizador do API Gateway upstream — sem validação de JWT neste serviço).

| Método | Caminho | Auth | Descrição |
|---|---|---|---|
| POST | `/v1/orders` | obrigatória | Criar pedido |
| POST | `/v1/orders/:id/assign` | obrigatória | Atribuir mecânico |
| POST | `/v1/orders/:id/complete-analysis` | obrigatória | Concluir análise + congelar itens |
| POST | `/v1/orders/:id/request-approval` | obrigatória | Enviar solicitação de aprovação ao cliente |
| GET | `/v1/orders/:id/approve` | nenhuma | Cliente aprova (UUID funciona como token) |
| GET | `/v1/orders/:id/reject` | nenhuma | Cliente rejeita |
| POST | `/v1/orders/:id/complete-work` | obrigatória | Marcar trabalho como concluído |
| POST | `/v1/orders/:id/archive` | obrigatória | Arquivar após pagamento |
| GET | `/v1/orders` | obrigatória | Listar todos os pedidos |
| GET | `/v1/orders/:id` | obrigatória | Buscar por ID |
| GET | `/v1/orders/in-progress` | obrigatória | Listar pedidos em andamento |

Documentação interativa: `GET /v1/docs/` (Swagger UI) · `GET /v1/redoc` (ReDoc)

---

## Modelo de autenticação

A autenticação é delegada a um API Gateway + Lambda Authorizer upstream.
Este serviço lê três headers injetados pelo autorizador:

- `X-User-Id` — ID do usuário
- `X-User-Email` — e-mail do usuário
- `X-User-Role` — um dos valores: `administrator`, `attendant`, `mechanic`

O middleware `RoleRequired` aplica controle de acesso por papel. Rotas públicas
(`/approve`, `/reject`) não possuem middleware — o UUID do pedido é o único controle de acesso.

Esses headers também são repassados automaticamente em todas as chamadas de saída para os
serviços de Usuários e Catálogo.

---

## Clientes HTTP de saída

`UsersHTTPClient` e `StockHTTPClient` compartilham um `baseClient` com:

- **Retry** — até 3 tentativas com backoff exponencial (somente erros 5xx e de rede)
- **Circuit breaker** — abre após 5 falhas consecutivas; half-open após 30 s
- **Propagação de headers** — `X-Request-ID`, `traceparent`, `X-User-Id`, `X-User-Email`, `X-User-Role`
- **Log de requisições** — cada tentativa registra método + URL; falhas registram o motivo

---

## Variáveis de ambiente

| Variável | Descrição |
|---|---|
| `HTTP_PORT` | Porta em que a API escuta (padrão `8080`) |
| `ENV` | `development` ou `production` |
| `AWS_REGION` | Região AWS |
| `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY` | Credenciais AWS |
| `AWS_ENDPOINT_URL` | Override para o localstack (omitir na AWS) |
| `DYNAMODB_TABLE` | Nome da tabela DynamoDB |
| `ORDERS_TOPIC_ARN` | ARN do tópico SNS para eventos de pedidos |
| `NOTIFICATION_TOPIC_ARN` | ARN do tópico SNS para eventos de notificação |
| `SQS_CATALOG_EVENTS_URL` | URL da fila SQS para eventos do catálogo |
| `SQS_PAYMENT_EVENTS_URL` | URL da fila SQS para eventos de pagamento |
| `USERS_BASE_URL` | URL base do serviço de Usuários |
| `STOCK_BASE_URL` | URL base do serviço de Catálogo |
| `HTTP_CLIENT_TIMEOUT_MS` | Timeout para chamadas HTTP de saída (ms) |
| `PUBLIC_BASE_URL` | URL pública base usada nos links de e-mail de aprovação. Na AWS, defina como `https://<api-gateway-id>.execute-api.<region>.amazonaws.com` — o prefixo de stage `/dev/v1` é adicionado automaticamente pelo pipeline de deploy. |
| `DD_AGENT_HOST` | Host do agente Datadog para métricas DogStatsD (noop se não definido) |
| `DD_API_KEY` / `DD_SITE` / `DD_SERVICE` | Configuração do Datadog |

---

## Deploy

### Kubernetes local (kind)

```bash
make up   # build da imagem → carrega no kind → aplica k8s/overlays/local
```

Requer um cluster `kind` em execução com o nome `tech-challenge-local` e o stack local
do Docker Compose já ativo (localstack + dynamodb-local).

### AWS (EKS)

O deploy é automatizado via GitHub Actions a cada push na branch `main`.

**Etapas do pipeline:**
1. Build da imagem Docker → push para o ECR (`tech-challenge-orders-api-repo`)
2. Atualização do kubeconfig para `eks-tech-challenge`
3. Geração de `k8s/overlays/aws/.env.secrets` e `.env.host` a partir dos Secrets/Variables do GitHub
4. Aplicação do overlay kustomize com a tag da imagem ECR
5. Rollout restart dos dois deployments

**GitHub Secrets necessários:**

| Secret | Descrição |
|---|---|
| `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY` / `AWS_SESSION_TOKEN` | Credenciais AWS |
| `AWS_REGION` | Região AWS |
| `AWS_ACCOUNT_ID` | ID da conta AWS |
| `PUBLIC_BASE_URL` | URL base do API Gateway sem caminho, ex.: `https://abc.execute-api.us-east-1.amazonaws.com` |
| `DD_API_KEY` | Chave de API do Datadog |

**GitHub Variables necessárias:**

| Variável | Descrição |
|---|---|
| `SNS_ORDERS_TOPIC_NAME` | Nome do tópico SNS para eventos de pedidos |
| `SQS_CATALOG_QUEUE_NAME` | Nome da fila SQS para eventos do catálogo |
| `SQS_ORDERS_QUEUE_NAME` | Nome da fila SQS para eventos de pagamento |
| `DD_SERVICE` / `DD_VERSION` / `DD_SITE` | Configuração do Datadog |

**Service discovery no cluster** (DNS do Kubernetes):

| Serviço | URL |
|---|---|
| Usuários | `http://tech-challenge-users-svc.default.svc.cluster.local` |
| Catálogo | `http://catalog-api-svc.default.svc.cluster.local` |

### Terraform

A infraestrutura (EKS, DynamoDB, SNS/SQS) é gerenciada com Terraform em `infra/`.

```bash
make create-tfstate-bucket   # único: cria bucket S3 para estado do Terraform
make apply-terraform         # provisiona / atualiza a infraestrutura
```

---

## Estrutura do projeto

```
cmd/
  api/        Entrypoint do servidor HTTP
  worker/     Entrypoint do consumidor SQS
internal/
  domain/     Order, Status, VehicleSnapshot, ItemSnapshot, HistoryEntry, eventos
  application/
    ports/    Repository, EventPublisher, IdempotencyStore, UsersClient, StockClient
    usecases/order/  create, assign, complete_analysis, request_approval,
                     approve, reject, complete_work, archive,
                     handle_stock_*, handle_payment_*
  adapter/
    config/        Configuração por env (godotenv)
    dynamo/        Cliente DynamoDB, repositório, idempotency store
    http/          Roteador, handlers, middlewares, helpers de resposta
    sns/           Publisher SNS, publisher de notificações
    sqs/           Consumidor genérico + factories de handlers
    clients/       UsersHTTPClient, StockHTTPClient (retry + circuit breaker)
    metrics/       Datadog DogStatsD ou noop
infra/           Terraform — EKS, DynamoDB, SNS, SQS
k8s/
  base/          Manifestos base do Kustomize
  overlays/
    local/       Overlay kind + localstack
    aws/         Overlay EKS (secrets gerados, URLs DNS do cluster)
deploy/
  localstack/  bootstrap.sh — cria tabela DynamoDB, tópicos SNS, filas SQS
  mocks/       Stubs Mockoon para as APIs de Usuários e Catálogo
swagger/       Spec OpenAPI + página estática ReDoc
```
