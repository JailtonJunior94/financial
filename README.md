# Financial API

Sistema de gestão financeira pessoal desenvolvido em Go, seguindo princípios de Clean Architecture e Domain-Driven Design (DDD).

## Visão Geral

Aplicação modular para controle financeiro que gerencia cartões de crédito, faturas, transações, orçamentos e categorias. Implementa padrões como Event-Driven Architecture com Outbox Pattern, garantindo confiabilidade na publicação de eventos.

### Características Principais

- **Arquitetura Limpa**: Separação clara entre Domain, Application e Infrastructure
- **Modular Monolith**: Módulos independentes com baixo acoplamento
- **Event-Driven**: Publicação confiável de eventos com Outbox Pattern
- **Observabilidade**: OpenTelemetry com traces, métricas e logs estruturados
- **Cursor-Based Pagination**: Paginação estável e performática
- **API RESTful**: Endpoints bem definidos com RFC 7807 para tratamento de erros

## Arquitetura

```mermaid
graph TB
    subgraph "Client Layer"
        HTTP[HTTP Client]
    end

    subgraph "API Server"
        Router[Chi Router]
        Auth[Auth Middleware]
        Handlers[HTTP Handlers]
    end

    subgraph "Application Layer"
        UC[Use Cases]
        DTO[DTOs]
    end

    subgraph "Domain Layer"
        Entities[Entities]
        VOs[Value Objects]
        Interfaces[Domain Interfaces]
        Events[Domain Events]
    end

    subgraph "Infrastructure Layer"
        Repos[Repositories]
        Adapters[Adapters]
        Outbox[Outbox Publisher]
    end

    subgraph "External Services"
        DB[(CockroachDB)]
        MQ[RabbitMQ]
        OTEL[OpenTelemetry]
    end

    subgraph "Background Services"
        Worker[Worker - Cron Jobs]
        Consumer[Event Consumer]
    end

    HTTP --> Router
    Router --> Auth
    Auth --> Handlers
    Handlers --> UC
    UC --> Entities
    UC --> Interfaces
    Entities --> Events
    Interfaces --> Repos
    Repos --> DB
    Events --> Outbox
    Outbox --> DB
    Worker --> Outbox
    Outbox --> MQ
    Consumer --> MQ
    Handlers --> OTEL
    UC --> OTEL
```

### Fluxo de Dados

```mermaid
sequenceDiagram
    participant Client
    participant API
    participant UseCase
    participant Repository
    participant Outbox
    participant DB
    participant Worker
    participant RabbitMQ

    Client->>API: HTTP Request
    API->>API: Auth Middleware
    API->>UseCase: Execute Business Logic
    UseCase->>Repository: Persist Data
    Repository->>DB: Begin Transaction
    Repository->>DB: Insert/Update Entity
    UseCase->>Outbox: Store Event
    Outbox->>DB: Insert Event (same txn)
    DB->>Repository: Commit Transaction
    Repository->>UseCase: Success
    UseCase->>API: Return DTO
    API->>Client: HTTP Response (200/201)

    Note over Worker: Every 5s
    Worker->>Outbox: Poll Pending Events
    Outbox->>DB: Query Pending Events
    DB->>Outbox: Return Events
    Outbox->>RabbitMQ: Publish Events
    RabbitMQ->>Outbox: Acknowledge
    Outbox->>DB: Mark as Published
```

## Tecnologias

### Core
- **Go 1.25.6** - Linguagem de programação
- **Chi v5** - HTTP router minimalista e performático
- **Cobra** - CLI framework para comandos
- **Viper** - Gerenciamento de configurações

### Database & Messaging
- **CockroachDB** - Banco de dados distribuído (PostgreSQL-compatible)
- **golang-migrate** - Migrations de banco de dados
- **RabbitMQ** - Message broker para eventos

### Security
- **JWT** - Autenticação stateless (golang-jwt/jwt)
- **bcrypt** - Hash de senhas

### Observability
- **OpenTelemetry** - Tracing distribuído
- **Prometheus** - Métricas
- **Grafana LGTM** - Logs, métricas e traces (Grafana + Loki + Tempo)
- **slog** - Logs estruturados

### Development
- **testify** - Testing framework
- **mockery** - Mock generation
- **golangci-lint** - Linter

## Estrutura de Pastas

```
financial/
├── cmd/                          # Entry points da aplicação
│   ├── main.go                  # CLI principal (Cobra)
│   ├── server/                  # Comando: API server
│   ├── consumer/                # Comando: Event consumer
│   └── worker/                  # Comando: Background jobs
├── configs/                      # Configurações
│   └── config.go                # Viper config loader
├── internal/                     # Módulos de domínio
│   ├── user/                    # Autenticação e usuários
│   │   ├── application/         # Use cases e DTOs
│   │   ├── domain/              # Entities, VOs, Interfaces
│   │   ├── infrastructure/      # HTTP, Repositories
│   │   └── module.go            # DI e setup do módulo
│   ├── card/                    # Gestão de cartões
│   ├── category/                # Categorias hierárquicas
│   ├── payment_method/          # Métodos de pagamento
│   ├── invoice/                 # Faturas e compras
│   ├── transaction/             # Transações mensais
│   └── budget/                  # Planejamento orçamentário
├── pkg/                          # Shared libraries
│   ├── api/                     # HTTP utilities
│   │   ├── http/                # Response helpers
│   │   ├── httperrors/          # RFC 7807 errors
│   │   └── middlewares/         # Auth middleware
│   ├── auth/                    # JWT handling
│   ├── custom_errors/           # Domain errors
│   ├── database/                # Database utilities
│   ├── pagination/              # Cursor pagination
│   ├── observability/           # Metrics e tracing
│   ├── outbox/                  # Outbox pattern
│   ├── scheduler/               # Cron jobs
│   └── messaging/               # Message utilities
├── database/
│   └── migrations/              # SQL migrations
├── deployment/
│   ├── docker-compose.yml       # Local environment
│   └── Dockerfile               # Container image
├── Makefile                     # Build scripts
├── go.mod                       # Dependencies
└── .golangci.yml               # Linter config
```

### Estrutura de Módulo (Padrão)

Cada módulo segue a mesma estrutura:

```
internal/{module}/
├── application/
│   ├── usecase/                 # Business logic
│   │   ├── create.go
│   │   ├── update.go
│   │   └── list.go
│   └── dtos/                    # Data Transfer Objects
│       └── {module}.go
├── domain/
│   ├── entities/                # Aggregate Roots
│   │   └── {module}.go
│   ├── vos/                     # Value Objects
│   ├── interfaces/              # Contracts
│   │   └── {module}_repository.go
│   ├── strategies/              # Strategy pattern (opcional)
│   └── events/                  # Domain events (opcional)
├── infrastructure/
│   ├── http/
│   │   ├── {module}_handler.go  # HTTP handlers
│   │   └── {module}_routes.go   # Route registration
│   ├── repositories/
│   │   └── {module}_repository.go
│   └── adapters/                # Integration adapters (opcional)
└── module.go                    # Dependency injection
```

## Como Rodar

### Pré-requisitos

- Go 1.25.6+
- Docker & Docker Compose
- Make

### 1. Clone o Repositório

```bash
git clone <repository-url>
cd financial
```

### 2. Configure as Variáveis de Ambiente

```bash
cp cmd/.env.example cmd/.env
```

Edite `cmd/.env` e configure as variáveis **obrigatórias**:

```env
# SECURITY: Gere valores seguros para produção
AUTH_SECRET_KEY=your-secret-key-minimum-64-chars
DB_PASSWORD=your-strong-database-password
RABBITMQ_URL=amqp://user:password@localhost:5672/
```

### 3. Inicie a Infraestrutura

```bash
# Apenas dependências (DB, RabbitMQ, Observability)
make start_minimal

# Ou todos os serviços (incluindo aplicação)
make start_docker
```

Serviços disponíveis:
- **CockroachDB**: http://localhost:8080 (UI)
- **RabbitMQ**: http://localhost:15672 (user: guest, pass: guest)
- **Grafana**: http://localhost:3000 (metrics, traces, logs)

### 4. Execute as Migrations

```bash
# Executar migrations
./bin/financial migrate

# Ou via Docker
docker compose -f deployment/docker-compose.yml run migrate
```

### 5. Inicie a Aplicação

```bash
# Build
make build

# Executar API Server
./bin/financial api

# Executar Worker (background jobs)
./bin/financial worker

# Executar Consumer (message processing)
./bin/financial consumer
```

### Desenvolvimento Local (Sem Docker)

Se preferir rodar apenas a aplicação localmente:

```bash
# 1. Inicie apenas a infraestrutura
make start_minimal

# 2. Execute as migrations
go run cmd/main.go migrate

# 3. Inicie os serviços em terminais separados
go run cmd/main.go api      # Terminal 1
go run cmd/main.go worker   # Terminal 2
go run cmd/main.go consumer # Terminal 3
```

## API Endpoints

### Autenticação

```http
POST /api/v1/token        # Gerar JWT token
POST /api/v1/users        # Registrar usuário
```

### Cards (Auth Required)

```http
GET    /api/v1/cards           # Listar cartões (paginado)
GET    /api/v1/cards/{id}      # Buscar cartão
POST   /api/v1/cards           # Criar cartão
PUT    /api/v1/cards/{id}      # Atualizar cartão
DELETE /api/v1/cards/{id}      # Deletar cartão
```

### Categories (Auth Required)

```http
GET    /api/v1/categories      # Listar categorias (paginado)
GET    /api/v1/categories/{id} # Buscar categoria
POST   /api/v1/categories      # Criar categoria
PUT    /api/v1/categories/{id} # Atualizar categoria
DELETE /api/v1/categories/{id} # Deletar categoria
```

### Payment Methods

```http
GET    /api/v1/payment-methods      # Listar métodos
GET    /api/v1/payment-methods/{id} # Buscar método
POST   /api/v1/payment-methods      # Criar método
PUT    /api/v1/payment-methods/{id} # Atualizar método
DELETE /api/v1/payment-methods/{id} # Deletar método
```

### Invoices (Auth Required)

```http
POST   /api/v1/invoice-items          # Criar compra
PUT    /api/v1/invoice-items/{id}     # Atualizar compra
DELETE /api/v1/invoice-items/{id}     # Deletar compra
GET    /api/v1/invoices               # Listar faturas (paginado)
GET    /api/v1/invoices/{id}          # Buscar fatura
GET    /api/v1/invoices/card/{cardId} # Faturas por cartão (paginado)
```

### Transactions (Auth Required)

```http
GET    /api/v1/transactions                                 # Listar transações mensais (paginado)
GET    /api/v1/transactions/{id}                            # Buscar transação
POST   /api/v1/transactions                                 # Registrar transação
PUT    /api/v1/transactions/{transactionId}/items/{itemId}  # Atualizar item
DELETE /api/v1/transactions/{transactionId}/items/{itemId}  # Deletar item
```

### Budgets (Auth Required)

```http
GET    /api/v1/budgets      # Listar orçamentos (paginado)
POST   /api/v1/budgets      # Criar orçamento
GET    /api/v1/budgets/{id} # Buscar orçamento
PUT    /api/v1/budgets/{id} # Atualizar orçamento
DELETE /api/v1/budgets/{id} # Deletar orçamento
```

### Formato de Resposta

#### Sucesso (2xx)

```json
{
  "data": { ... },
  "pagination": {
    "limit": 20,
    "has_next": true,
    "next_cursor": "eyJm..."
  }
}
```

#### Erro (4xx/5xx) - RFC 7807

```json
{
  "type": "https://httpstatuses.com/400",
  "title": "Validation Error",
  "status": 400,
  "detail": "Invalid request parameters",
  "instance": "/api/v1/cards",
  "timestamp": "2026-01-30T10:30:00Z",
  "request_id": "req-123",
  "trace_id": "trace-456",
  "errors": {
    "name": ["field is required"]
  }
}
```

### Autenticação

Todas as rotas protegidas requerem header de autenticação:

```http
Authorization: Bearer {jwt_token}
```

**Obter token:**

```bash
curl -X POST http://localhost:8000/api/v1/token \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "yourpassword"
  }'
```

### Paginação (Cursor-Based)

Todas as listagens suportam:

```http
GET /api/v1/cards?limit=20&cursor=eyJm...
```

Parâmetros:
- `limit`: Número de resultados (default: 20, max: 100)
- `cursor`: Token para próxima página (retornado em `pagination.next_cursor`)

## Comandos Make

```bash
# Build
make build                # Compilar binário

# Database
make migrate              # Criar nova migration

# Qualidade de Código
make lint                 # Executar linter
make test                 # Testes unitários
make test-integration     # Testes de integração
make test-all             # Todos os testes
make cover                # Coverage HTML
make check                # Lint + todos os testes

# Mocks
make mocks                # Gerar mocks com mockery

# Docker
make start_minimal        # Infraestrutura apenas
make start_docker         # Todos os serviços
make stop_docker          # Parar serviços
```

## Configuração

### Variáveis de Ambiente

| Variável | Descrição | Padrão |
|----------|-----------|--------|
| **HTTP_PORT** | Porta da API | 8000 |
| **DB_HOST** | Host do CockroachDB | localhost |
| **DB_PORT** | Porta do banco | 26257 |
| **DB_NAME** | Nome do banco | financial |
| **AUTH_SECRET_KEY** | Chave JWT (min 64 chars) | *required* |
| **AUTH_TOKEN_DURATION** | Duração do token (minutos) | 60 |
| **RABBITMQ_URL** | URL do RabbitMQ | amqp://guest:guest@localhost:5672/ |
| **OTEL_EXPORTER_OTLP_ENDPOINT** | Endpoint OpenTelemetry | localhost:4317 |
| **LOG_LEVEL** | Nível de log (debug, info, warn, error) | info |
| **OUTBOX_POLL_INTERVAL_SECONDS** | Intervalo de polling do outbox | 5 |
| **CONSUMER_WORKER_COUNT** | Workers do consumer | 5 |

Veja todas as variáveis em `cmd/.env.example`.

## Boas Práticas

### Clean Architecture

1. **Domain Layer**: Sem dependências externas, apenas lógica de negócio
2. **Application Layer**: Orquestra domínio e infraestrutura via interfaces
3. **Infrastructure Layer**: Implementa interfaces do domínio

### Domain-Driven Design

- **Entities**: Objetos com identidade única (ex: `Card`, `Invoice`)
- **Value Objects**: Objetos imutáveis sem identidade (ex: `DueDay`, `CardName`)
- **Aggregates**: Entidades com filhos (ex: `Invoice` com `InvoiceItems`)
- **Repository Pattern**: Abstração de persistência
- **Domain Events**: Eventos de domínio para comunicação assíncrona

### Padrões Implementados

#### Outbox Pattern
Garante publicação confiável de eventos:
1. Evento salvo no banco junto com entidade (mesma transação)
2. Worker processa eventos pendentes periodicamente
3. Publica no RabbitMQ e marca como publicado
4. Retry automático até 3 vezes

#### Unit of Work
Garante consistência transacional:
```go
uow.Do(ctx, func(uow *unitofwork.UoW) error {
    // Múltiplas operações na mesma transação
    repo.Create(ctx, entity)
    outbox.Save(ctx, event)
    return nil
})
```

#### Cursor-Based Pagination
Paginação estável mesmo com inserções:
- Baseada em campos únicos (ID + timestamps)
- Não usa OFFSET (performance constante)
- Cursor base64-encoded com valores dos campos

### Tratamento de Erros

Use os erros personalizados em `pkg/custom_errors`:

```go
// Domínio
if err != nil {
    return custom_errors.NewInternalError("operation failed", err)
}

// Not Found
if entity == nil {
    return custom_errors.NewNotFoundError("entity", entityID)
}

// Validation
if invalid {
    return custom_errors.NewValidationError(map[string][]string{
        "field": {"validation error message"},
    })
}
```

### Observabilidade

Todas as operações são instrumentadas:

```go
// Traces são propagados via context
ctx = otel.ExtractTraceContext(r.Context())

// Logs estruturados com contexto
logger.InfoContext(ctx, "operation executed", "entity_id", id)

// Métricas customizadas
observability.RecordCustomMetric("cards_created", 1)
```

### Testes

- **Unit Tests**: Testam use cases isoladamente com mocks
- **Integration Tests**: Testam handlers + repositories com banco real
- Use tag `//go:build integration` para testes de integração
- Gere mocks com `make mocks`

```bash
make test              # Unit tests apenas
make test-integration  # Integration tests
make test-all          # Todos os testes
```

## Desenvolvimento

### Criar Nova Migration

```bash
make migrate NAME=create_new_table
```

### Adicionar Novo Módulo

1. Crie estrutura em `internal/novo_modulo/`
2. Siga o padrão dos módulos existentes (application/domain/infrastructure)
3. Implemente `module.go` com setup de DI
4. Registre no `cmd/server/server.go`

### Gerar Mocks

```bash
make mocks
```

Configuração em `.mockery.yaml`.

### Executar Linter

```bash
make lint
```

Configuração em `.golangci.yml`.

## Monitoramento

### Grafana Dashboards

Acesse http://localhost:3000 após `make start_docker`:

- **API Metrics**: Latência, throughput, taxa de erro
- **Database Metrics**: Conexões, queries, latência
- **Business Metrics**: Cards criados, faturas geradas

### Logs

Logs estruturados em JSON:

```json
{
  "time": "2026-01-30T10:30:00Z",
  "level": "INFO",
  "msg": "request completed",
  "request_id": "req-123",
  "trace_id": "trace-456",
  "method": "POST",
  "path": "/api/v1/cards",
  "status": 201,
  "duration_ms": 45
}
```

### Traces

Traces distribuídos com OpenTelemetry visualizados no Grafana Tempo.

## Contribuindo

1. Fork o projeto
2. Crie uma branch para sua feature (`git checkout -b feature/nova-feature`)
3. Commit suas mudanças (`git commit -m 'feat: adiciona nova feature'`)
4. Push para a branch (`git push origin feature/nova-feature`)
5. Abra um Pull Request

### Convenção de Commits

Seguimos [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` Nova funcionalidade
- `fix:` Correção de bug
- `docs:` Documentação
- `refactor:` Refatoração de código
- `test:` Adicionar/modificar testes
- `chore:` Tarefas de manutenção

## Licença

Este projeto é privado e proprietário.

---

**Desenvolvido com Go 1.25.6**
