# CLAUDE.md

Este arquivo define as diretrizes obrigatórias para a IA ao trabalhar neste repositório, garantindo coerência arquitetural, reutilização de código existente e respostas contextualizadas.

---

## Objetivos deste Arquivo

1. Definir regras globais de comportamento da IA
2. Explicar como a IA deve interpretar o contexto do projeto
3. Priorizar código existente antes de sugerir novas abstrações
4. Reduzir respostas genéricas ou fora do padrão do projeto
5. Garantir consistência arquitetural e de estilo
6. Evitar decisões implícitas sem validação do contexto existente
7. Servir como memória institucional permanente do projeto

---

## Visão Geral do Projeto

### Descrição
Sistema financeiro pessoal completo desenvolvido em Go seguindo Clean Architecture e Domain-Driven Design (DDD). O projeto gerencia orçamentos, categorias, cartões de crédito, faturas, transações financeiras e consolidado mensal com suporte completo a observabilidade, mensageria e padrões de eventos.

### Stack Tecnológica
- **Linguagem**: Go 1.20+
- **Banco de Dados**: CockroachDB (compatível PostgreSQL)
- **Message Broker**: RabbitMQ
- **Framework HTTP**: Chi Router
- **Observabilidade**: OpenTelemetry (OTLP) + Stack LGTM (Loki, Grafana, Tempo, Prometheus)
- **Migrations**: golang-migrate
- **CLI**: Cobra
- **Testes**: testify/suite, mockery

### Estatísticas do Projeto
- **Linhas de código**: ~16.061 (internal)
- **Módulos de domínio**: 7 (user, category, card, budget, payment_method, invoice, transaction)
- **Comandos CLI**: 4 (api, migrate, consumer, worker)

---

## Uso de Contexto (Regras Obrigatórias)

### 1. Leitura de Contexto
- Sempre assumir que **o código existente é a fonte da verdade**
- Nunca sugerir soluções que contradigam:
  - Estrutura de pastas
  - Padrões já adotados
  - Bibliotecas internas (`pkg/*`)
  - Decisões técnicas documentadas
- Antes de criar algo novo, **avaliar se já existe algo semelhante**

### 2. Continuidade de Conversa
- Considerar mensagens anteriores como **contexto ativo**
- Manter consistência entre respostas dentro da mesma conversa
- Não redefinir padrões já estabelecidos sem justificativa técnica clara

### 3. Escopo da Resposta
- Responder **apenas ao que foi solicitado**
- Não antecipar features ou refatorações não pedidas
- Se faltar contexto, **pedir explicitamente**, sem assumir

---

## Diretrizes Arquiteturais

A IA deve assumir que o projeto segue, salvo indicação contrária:

- Clean Architecture / Hexagonal
- Domain-Driven Design (DDD)
- Separação clara entre:
  - handlers / controllers (infrastructure/http)
  - usecases (application)
  - domain (entities, VOs, factories, events, strategies)
  - repositories (interfaces no domain, implementação na infrastructure)
- Comunicação entre módulos via Adapter Pattern
- Uso intenso de `context.Context`
- Código idiomático em Go (Go 1.20+)

### Princípios Invioláveis
- SOLID
- DRY
- KISS
- Baixo acoplamento e alta coesão

---

## Qualidade de Código Esperada

Toda sugestão de código deve:

- Evitar `panic`
- Ser `nil-safe`
- Evitar race conditions
- Evitar memory leaks
- Passar em:
  - go vet
  - go fmt
  - golangci-lint (configurado em `.golangci.yml`)
- Usar tratamento explícito de erros
- Seguir padrões já existentes no projeto

---

## Segurança e Autenticação

Quando o tema envolver segurança:

- Nunca assumir autenticação inexistente
- Sempre reutilizar middlewares já criados (`pkg/api/middlewares`)
- Não duplicar lógica de auth
- Respeitar propagação de usuário via `context.Context`
- Retornar status HTTP corretos (401, 403, etc.)
- Usar JWT implementado em `pkg/auth/jwt.go`
- Secret key deve ter mínimo 64 caracteres em produção
- Token JWT com duração configurável (1-24 horas)

---

## Reutilização Antes de Criação

Antes de sugerir:
- novos helpers
- novas libs
- novos middlewares
- novos padrões

A IA deve:
1. Procurar algo existente no projeto
2. Avaliar adaptação/extensão
3. Somente então propor algo novo, com justificativa clara

### Bibliotecas Internas Existentes (`pkg/`)
- `pkg/bundle/`: Container de injeção de dependências
- `pkg/database/`: Abstrações de banco, migrações, Unit of Work
- `pkg/auth/`: Implementação JWT
- `pkg/api/`: Middlewares, error handlers, HTTP utilities
- `pkg/custom_errors/`: Tratamento customizado de erros
- `pkg/outbox/`: Outbox Pattern implementation
- `pkg/scheduler/`: Cron job scheduler
- `pkg/messaging/`: Message broker abstractions
- `pkg/jobs/`: Job definitions

---

## Estilo de Resposta

As respostas da IA devem ser:

- Técnicas
- Objetivas
- Estruturadas (títulos, listas, passos)
- Escritas em **português brasileiro**
- Sem emojis desnecessários
- Sem explicações óbvias
- Sem texto promocional ou genérico

---

## O Que a IA NÃO Deve Fazer

- Não reescrever código inteiro sem pedido explícito
- Não mudar decisões arquiteturais sem análise
- Não inventar dependências
- Não assumir tecnologias não mencionadas
- Não simplificar problemas complexos sem alertar
- Não criar novos padrões quando já existem padrões estabelecidos
- Não sugerir bibliotecas externas sem verificar alternativas internas

---

## Arquitetura do Projeto

### Clean Architecture com Domain-Driven Design

A base de código segue os princípios da Clean Architecture organizados em camadas distintas:

**Camada de Domínio** (`internal/{module}/domain/`)
- `entities/`: Entidades de negócio principais com comportamento (ex: Budget, Category, User, Invoice, MonthlyTransaction)
- `vos/`: Value Objects (ex: CategoryName, Email, Money, Percentage)
- `factories/`: Criação de entidades com lógica de validação
- `interfaces/`: Interfaces de repositório (inversão de dependência)
- `events/`: Eventos de domínio (ex: TransactionCreatedEvent, InvoiceCalculatedEvent)
- `strategies/`: Strategy Pattern (ex: PixStrategy, BoletoStrategy, CreditCardStrategy)
- `errors.go`: Erros específicos do domínio

**Camada de Aplicação** (`internal/{module}/application/`)
- `usecase/`: Regras de negócio específicas da aplicação e orquestração
- `dtos/`: Data Transfer Objects para request/response

**Camada de Infraestrutura** (`internal/{module}/infrastructure/`)
- `repositories/`: Implementações de banco de dados das interfaces do domínio
- `http/`: Handlers HTTP, rotas e transporte
- `adapters/`: Adaptadores para outros módulos (desacoplamento)
- `rabbitmq/`: Consumers de eventos (se houver)
- `mocks/`: Mocks gerados para testes (via mockery)

**Padrão de Registro de Módulos**
Cada módulo de domínio possui um arquivo `module.go` que:
- Conecta dependências (repositories, use cases, handlers)
- Retorna rotas HTTP para registro
- Encapsula inicialização do módulo
- Registra adapters quando necessário

### Padrões Arquiteturais Aplicados

1. **Aggregate Root Pattern**
   - `Budget`, `Invoice`, `MonthlyTransaction`: Garantem invariantes e consistência
   - Entidades filhas (`BudgetItem`, `InvoiceItem`, `TransactionItem`) só acessíveis via aggregate
   - Métodos do aggregate recalculam totais automaticamente

2. **Unit of Work Pattern**
   - Implementado em `pkg/database/uow/`
   - Usado para transações multi-entidade (Budget, Invoice, Transaction)
   - Método `Do(ctx, fn)` com rollback automático

3. **Repository Pattern**
   - Interfaces no domínio, implementações na infraestrutura
   - Aceitam `database.DBExecutor` (suporta `*sql.DB` e `*sql.Tx`)
   - Permitem operações em transação sem acoplamento

4. **Factory Pattern**
   - Criação de entidades com validação completa
   - Exemplo: `factories.CreateBudget()`, `factories.CreateCategory()`

5. **Strategy Pattern**
   - `transaction/domain/strategies/`: Validação por tipo de transação
   - `PixStrategy`, `BoletoStrategy`, `CreditCardStrategy`, `TransferStrategy`

6. **Outbox Pattern**
   - Eventos de domínio persistidos em `outbox_events`
   - Dispatcher processa e publica no RabbitMQ
   - Garantia de entrega confiável (at-least-once)
   - Worker independente executando a cada 5 segundos

7. **Adapter Pattern**
   - `CardProviderAdapter`: Expõe Card repository para Invoice module
   - `InvoiceTotalProviderAdapter`: Expõe Invoice totals para Transaction module
   - Desacoplamento entre módulos (evita dependência cíclica)

### Container de Injeção de Dependências

O `pkg/bundle/container.go` fornece gerenciamento centralizado de dependências:
- Conexão com banco de dados (`*sql.DB`)
- Configuração (`configs.Config`)
- Adaptadores JWT e hashing
- Telemetria OpenTelemetry (traces, métricas, logs via devkit-go)
- Middlewares compartilhados (auth, panic recovery)
- Message broker (RabbitMQ)

### Padrão Unit of Work

`pkg/database/uow/` implementa Unit of Work para operações transacionais:
- `Executor()`: Retorna executor de DB atual (transaction ou connection)
- `Do(ctx, fn)`: Envolve operações em transação com rollback automático
- Usado em Budget, Invoice e Transaction para operações multi-entidade

### Estratégia de Banco de Dados

- **Banco Principal**: CockroachDB (compatível com PostgreSQL)
- **Drivers Disponíveis**: MySQL, Postgres, CockroachDB
- **Migrações**: golang-migrate com nomenclatura Unix timestamp
- **Testes**: Testcontainers para testes de integração

### Arquitetura do Servidor HTTP

Construído sobre `github.com/JailtonJunior94/devkit-go/pkg/httpserver`:
- Tratamento customizado de erros via `pkg/custom_errors` e `pkg/httperrors`
- Cadeia de middlewares: RequestID, Auth, Panic Recovery, Metrics, Tracing
- Endpoint de health check em `/health` (verifica database)
- Rotas registradas por módulo com middleware opcional
- Graceful shutdown com timeout de 10 segundos

### Tratamento Customizado de Erros

**Custom Errors** (`pkg/custom_errors/custom_errors.go`):
- `CustomError` envolve erros com contexto adicional (mensagem, detalhes)
- Suporta unwrapping para extrair erro original
- Utilizado em toda a aplicação para contexto rico

**Error Handler** (`pkg/api/httperrors/http_errors.go`):
1. **Unwrapping**: Extrai erro original de CustomError
2. **Mapping**: Mapeia erro para HTTP status + mensagem
3. **Tracing**: Adiciona atributos ao span OpenTelemetry
4. **Logging**: Log único com nível apropriado (ERROR/WARN/INFO)
5. **ProblemDetail**: RFC 7807 compliant response

**Error Mapping**:
- Validation errors → 400 Bad Request
- Not found errors → 404 Not Found
- Conflict errors → 409 Conflict
- Auth errors → 401 Unauthorized
- Unknown → 500 Internal Server Error

**Erros de Domínio Predefinidos**:
- `ErrUserNotFound`, `ErrCategoryNotFound`, `ErrBudgetNotFound`
- `ErrEmailAlreadyExists`, `ErrInvalidParentCategory`
- `ErrUnauthorized`, `ErrInvalidToken`, `ErrTokenExpired`
- `ErrBudgetInvalidTotal`, `ErrCategoryCycle`

### Observabilidade

Integração completa com OpenTelemetry via devkit-go:
- **Tracing**: Rastreamento distribuído (OTLP gRPC para Tempo)
- **Metrics**: Coleta de métricas (Prometheus)
- **Logging**: Logging estruturado (JSON ou texto)
- **Configuração**: Via variáveis de ambiente `OTEL_*`
- **Instrumentação**: Middleware instrumenta automaticamente handlers HTTP

**Stack LGTM** (Docker):
- **Loki**: Logs
- **Grafana**: Dashboards
- **Tempo**: Traces distribuídos
- **Prometheus**: Métricas

**Atributos de Span**:
- `http.method`, `http.route`, `http.status_code`
- `db.system`, `db.statement`, `db.operation`
- `messaging.system`, `messaging.destination`, `messaging.message_id`

---

## Módulos de Domínio (Detalhado)

### Módulo USER

**Responsabilidades**: Autenticação e gestão de usuários

**Entidades**:
- `User`: ID, Name (UserName VO), Email (Email VO), Password (hashed)

**Value Objects**:
- `UserName`: Validação de nome (não vazio, max 100 chars)
- `Email`: Validação de formato de email

**Use Cases**:
1. `CreateUserUseCase`: Criar novo usuário com senha hasheada
2. `TokenUseCase`: Autenticar e gerar JWT

**Rotas HTTP**:
- `POST /api/v1/users` (público) - Criar usuário
- `POST /api/v1/token` (público) - Login e obter token

**Notas Importantes**:
- Senha hasheada com bcrypt via `devkit-go/pkg/encrypt`
- Token JWT com duração configurável (1-24 horas)
- Validação de comprimento mínimo de hash (20 chars)
- Claims JWT: `user_id`, `email`, `roles[]`

---

### Módulo CATEGORY

**Responsabilidades**: Categorias financeiras hierárquicas

**Entidades**:
- `Category`: ID, UserID, ParentID (nullable), Name, Sequence, Children[]

**Value Objects**:
- `CategoryName`: Não vazio, max 255 chars
- `CategorySequence`: 0-1000 (ordem de exibição)

**Use Cases**:
1. `FindCategoryUseCase`: Listar todas (com hierarquia)
2. `FindCategoryByUseCase`: Buscar por ID
3. `CreateCategoryUseCase`: Criar categoria
4. `UpdateCategoryUseCase`: Atualizar categoria
5. `RemoveCategoryUseCase`: Soft delete

**Rotas HTTP** (autenticadas):
- `GET /api/v1/categories` - Listar categorias do usuário
- `GET /api/v1/categories/{id}` - Buscar por ID
- `POST /api/v1/categories` - Criar categoria
- `PUT /api/v1/categories/{id}` - Atualizar
- `DELETE /api/v1/categories/{id}` - Remover (soft delete)

**Regras de Negócio**:
- Suporta hierarquia (parent-child)
- ParentID pode ser null (categoria raiz)
- Sequence determina ordem de exibição
- Repository carrega children recursivamente
- Soft deletes com `deleted_at`

**Índices de Banco**:
- `idx_categories_list`: (user_id, sequence) WHERE parent_id IS NULL
- `idx_categories_user_active`: (user_id, deleted_at)
- `idx_categories_parent`: (parent_id)
- `idx_categories_user_parent`: (user_id, parent_id)

---

### Módulo CARD

**Responsabilidades**: Gestão de cartões de crédito

**Entidades**:
- `Card`: ID, UserID, Name, DueDay (1-31), ClosingOffsetDays (padrão: 7 dias)

**Value Objects**:
- `CardName`: Nome do cartão (não vazio, max 255)
- `DueDay`: Dia de vencimento (1-31)
- `ClosingOffsetDays`: Dias antes do vencimento para fechamento (1-31)

**Métodos Críticos da Entidade**:
1. `CalculateClosingDay(year, month)`: Calcula dia de fechamento da fatura
   - **Regra**: `closingDay = dueDay - offset`
   - Se <= 0, volta para mês anterior

2. `DetermineInvoiceMonth(purchaseDate)`: Determina a qual fatura pertence uma compra
   - **REGRA CRÍTICA**: `purchaseDate < closingDate` → fatura do mês
   - `purchaseDate >= closingDate` → fatura do próximo mês

**Use Cases**:
1. `FindCardUseCase`: Listar cartões do usuário
2. `FindCardByUseCase`: Buscar por ID
3. `CreateCardUseCase`: Criar cartão
4. `UpdateCardUseCase`: Atualizar cartão
5. `RemoveCardUseCase`: Soft delete

**Rotas HTTP** (autenticadas):
- `GET /api/v1/cards`
- `GET /api/v1/cards/{id}`
- `POST /api/v1/cards`
- `PUT /api/v1/cards/{id}`
- `DELETE /api/v1/cards/{id}`

**Adapter Exposto**:
- `CardProviderAdapter`: Implementa `invoice.CardProvider`
- Permite Invoice module buscar Card sem acoplamento direto

---

### Módulo BUDGET

**Responsabilidades**: Orçamentos mensais com distribuição percentual por categoria

**Aggregate Root**:
- `Budget`: ID, UserID, ReferenceMonth, TotalAmount, SpentAmount, PercentageUsed, Items[]

**Entidades Filhas**:
- `BudgetItem`: ID, BudgetID, CategoryID, PercentageGoal, AmountGoal, SpentAmount

**Value Objects**:
- `ReferenceMonth`: Formato DATE (primeiro dia do mês)
- `Money`: Valor monetário com precisão (devkit-go)
- `Percentage`: Percentual com scale 3 (25.5% = 25500)

**Regras de Negócio Críticas**:
1. **100% Rule**: Soma de `PercentageGoal` dos items DEVE ser exatamente 100%
2. **Categoria Única**: Não pode haver items duplicados com mesma categoria
3. **Recalculo Automático**: `SpentAmount` e `PercentageUsed` recalculados automaticamente
4. **Permite Estourar**: `SpentAmount` pode exceder `AmountGoal` (RemainingAmount fica negativo)

**Use Cases**:
1. `CreateBudgetUseCase`: Criar budget com items (validação 100%)
2. `FindBudgetUseCase`: Buscar budgets do usuário
3. `UpdateBudgetUseCase`: Atualizar budget e items
4. `DeleteBudgetUseCase`: Soft delete
5. `CreateItemUseCase`: Adicionar item individual (valida não exceder 100%)

**Rotas HTTP** (autenticadas):
- `GET /api/v1/budgets`
- `GET /api/v1/budgets/{id}`
- `POST /api/v1/budgets`
- `PUT /api/v1/budgets/{id}`
- `DELETE /api/v1/budgets/{id}`
- `POST /api/v1/budgets/{id}/items`

**Uso de Unit of Work**:
- Cria Budget e BudgetItems em transação única
- Rollback automático se alguma validação falhar

**Consumer** (pendente implementação):
- `BudgetConsumer`: Consome `transaction.transaction_created`
- Atualiza `spent_amount` do BudgetItem correspondente

---

### Módulo PAYMENT_METHOD

**Responsabilidades**: Métodos de pagamento disponíveis no sistema

**Entidades**:
- `PaymentMethod`: ID, Name, Code (único), Description

**Value Objects**:
- `PaymentMethodName`: Nome do método (max 255)
- `PaymentMethodCode`: Código único (PIX, CREDIT_CARD, etc.)
- `Description`: Descrição (max 500)

**Métodos Pré-cadastrados** (seed):
- PIX
- Cartão de Crédito (CREDIT_CARD)
- Cartão de Débito (DEBIT_CARD)
- Boleto Bancário (BOLETO)
- Dinheiro (CASH)
- Transferência Bancária (BANK_TRANSFER)

**Use Cases**:
1. `FindPaymentMethodUseCase`: Listar todos
2. `FindPaymentMethodByUseCase`: Buscar por ID
3. `FindPaymentMethodByCodeUseCase`: Buscar por código
4. `CreatePaymentMethodUseCase`: Criar novo método
5. `UpdatePaymentMethodUseCase`: Atualizar
6. `RemovePaymentMethodUseCase`: Soft delete

**Rotas HTTP** (SEM autenticação - dados públicos):
- `GET /api/v1/payment-methods`
- `GET /api/v1/payment-methods/{id}`
- `GET /api/v1/payment-methods/code/{code}`
- `POST /api/v1/payment-methods`
- `PUT /api/v1/payment-methods/{id}`
- `DELETE /api/v1/payment-methods/{id}`

---

### Módulo INVOICE

**Responsabilidades**: Faturas de cartão de crédito e compras parceladas

**Aggregate Root**:
- `Invoice`: ID, UserID, CardID, ReferenceMonth, DueDate, TotalAmount, Items[]

**Entidades Filhas**:
- `InvoiceItem`: ID, InvoiceID, CategoryID, PurchaseDate, Description, TotalAmount, InstallmentNumber, InstallmentTotal, InstallmentAmount

**Value Objects**:
- `ReferenceMonth`: Formato DATE (primeiro dia do mês de vencimento)

**Regras de Negócio**:
1. **Fatura Única**: Uma fatura por cartão por mês (unique constraint)
2. **Cálculo Automático**: `TotalAmount` = soma de `InstallmentAmount` dos items
3. **Parcelamento**: Item armazena `InstallmentNumber/InstallmentTotal` (ex: 3/12)
4. **Compra Única Gera Múltiplos Items**: Uma compra de 12x gera 12 InvoiceItems em faturas diferentes

**Use Cases**:
1. `CreatePurchaseUseCase`: Criar compra (distribui parcelas entre faturas)
2. `UpdatePurchaseUseCase`: Atualizar compra existente
3. `DeletePurchaseUseCase`: Remover compra (soft delete)
4. `GetInvoiceUseCase`: Buscar fatura por ID
5. `ListInvoicesByMonthUseCase`: Listar faturas de um mês
6. `ListInvoicesByCardUseCase`: Listar faturas de um cartão

**Rotas HTTP** (autenticadas):
- `POST /api/v1/purchases` - Criar compra
- `PUT /api/v1/purchases/{id}` - Atualizar compra
- `DELETE /api/v1/purchases/{id}` - Remover compra
- `GET /api/v1/invoices?month=YYYY-MM` - Listar por mês
- `GET /api/v1/invoices/{id}` - Buscar fatura
- `GET /api/v1/invoices/cards/{cardId}` - Listar por cartão

**Fluxo de Criação de Compra**:
1. Recebe: `CardID`, `CategoryID`, `PurchaseDate`, `TotalAmount`, `Installments`
2. Busca Card (via CardProvider adapter)
3. Determina fatura de cada parcela usando `Card.DetermineInvoiceMonth()`
4. Cria/busca Invoice para cada mês
5. Adiciona InvoiceItem em cada Invoice
6. Recalcula `TotalAmount` de cada Invoice

**Adapter Exposto**:
- `InvoiceTotalProviderAdapter`: Implementa `transaction.InvoiceTotalProvider`
- Calcula total de faturas de um mês para Transaction module

**Eventos de Domínio**:
- `InvoiceCalculatedEvent`: Emitido quando fatura é recalculada
- Consumido por Transaction module para atualizar item CREDIT_CARD

---

### Módulo TRANSACTION

**Responsabilidades**: Consolidado financeiro mensal do usuário

**Aggregate Root**:
- `MonthlyTransaction`: ID, UserID, ReferenceMonth (YYYY-MM), TotalIncome, TotalExpense, TotalAmount, Items[]

**Entidades Filhas**:
- `TransactionItem`: ID, MonthlyID, CategoryID, Title, Description, Amount, Direction (INCOME/EXPENSE), Type (PIX/BOLETO/TRANSFER/CREDIT_CARD), IsPaid

**Value Objects**:
- `ReferenceMonth`: Formato "YYYY-MM" (string)
- `TransactionDirection`: INCOME ou EXPENSE
- `TransactionType`: PIX, BOLETO, TRANSFER, CREDIT_CARD
- `TransactionTitle`: Título da transação

**Strategies (Strategy Pattern)**:
1. `PixStrategy`: Valida transações PIX
2. `BoletoStrategy`: Valida boletos
3. `TransferStrategy`: Valida transferências
4. `CreditCardStrategy`: Valida faturas de cartão
   - **REGRA CRÍTICA**: Apenas um item CREDIT_CARD por mês
   - Sempre despesa
   - Gerenciado automaticamente (não pode ser editado manualmente)

**Regras de Negócio**:
1. **Item CREDIT_CARD Único**: Apenas um por mês (validado no aggregate)
2. **Recalculo Automático**: `TotalIncome`, `TotalExpense`, `TotalAmount` recalculados sempre
3. **TotalAmount** = TotalIncome - TotalExpense
4. **Soft Delete**: Items deletados não entram nos cálculos

**Use Cases**:
1. `RegisterTransactionUseCase`: Criar transaction item
   - Se tipo = CREDIT_CARD, busca total de faturas via `InvoiceTotalProvider`
   - Usa `UpdateOrCreateCreditCardItem()` do aggregate
2. `UpdateTransactionItemUseCase`: Atualizar item existente
3. `DeleteTransactionItemUseCase`: Soft delete de item

**Rotas HTTP** (autenticadas):
- `POST /api/v1/transactions` - Registrar transação
- `PUT /api/v1/transactions/items/{id}` - Atualizar item
- `DELETE /api/v1/transactions/items/{id}` - Remover item

**Eventos de Domínio**:
- `TransactionCreatedEvent`: Emitido quando item é criado
- Payload: transaction_id, user_id, category_id, amount, direction, type, reference_month
- **Consumidor**: Budget Consumer (atualiza `spent_amount` do BudgetItem)

---

## Infraestrutura e PKG

### Autenticação e Autorização

**JWT** (`pkg/auth/jwt.go`):
- Geração de token com duração configurável (1-24 horas)
- Claims: `user_id`, `email`, `roles[]`
- Secret key mínimo 64 caracteres (validado em produção)
- Implementa interface `TokenValidator`

**Middleware de Autenticação** (`pkg/api/middlewares/auth.go`):
- Extrai header `Authorization: Bearer <token>`
- Valida formato e token
- Injeta `AuthenticatedUser` no contexto
- Helper: `GetUserFromContext(ctx)`

**Fluxo de Autenticação**:
1. Login: `POST /api/v1/token` com email/senha
2. Retorna JWT com duração configurada
3. Cliente inclui header em requests protegidos
4. Middleware valida e injeta usuário no contexto
5. Handlers acessam `GetUserFromContext(r.Context())`

---

### Outbox Pattern

**Modelo** (`pkg/outbox/model.go`):
```go
type OutboxEvent struct {
    ID            uuid.UUID
    AggregateID   uuid.UUID
    AggregateType string
    EventType     string
    Payload       JSONBPayload (map[string]any)
    Status        OutboxStatus (pending/published/failed)
    RetryCount    int (max 3)
    PublishedAt   *time.Time
    FailedAt      *time.Time
    CreatedAt     time.Time
}
```

**Dispatcher** (`pkg/outbox/dispatcher.go`):
1. Busca eventos `status = pending` em batch (100)
2. Para cada evento:
   - Serializa payload para JSON
   - Publica no RabbitMQ com routing key `{aggregate_type}.{event_type}`
   - Se sucesso: marca `status = published`
   - Se falha: incrementa `retry_count` (max 3)
   - Se esgotou retries: marca `status = failed`
3. Toda operação em transação (Unit of Work)

**Cleanup** (`pkg/outbox/cleanup.go`):
- Remove eventos `status = published` e `published_at < 30 dias`
- Remove eventos `status = failed` e `failed_at < 30 dias`
- Executado diariamente pelo Worker

**Fluxo Completo**:
1. Use Case cria evento de domínio
2. Repository persiste evento na tabela `outbox_events` (mesma transação)
3. Worker executa Dispatcher a cada 5 segundos
4. Dispatcher publica no RabbitMQ
5. Consumer processa mensagem
6. Cleanup limpa eventos antigos

---

### Messaging (RabbitMQ)

**Configuração**:
- Exchange: `financial.events` (tipo: topic)
- Routing Key: `{aggregate_type}.{event_type}`
- Exemplo: `transaction.transaction_created`

**Publisher** (usado pelo Dispatcher):
- Modo persistente (delivery_mode=2)
- Confirmações habilitadas (publisher confirms)
- Headers: aggregate_id, aggregate_type, event_type
- MessageID: event.ID

**Consumer** (`cmd/consumer/`):
- Worker pool: 10 workers
- Prefetch count: 10 mensagens
- Queue: `financial_queue`
- Auto-reconnect habilitado

**Handlers** (futuro):
- `BudgetConsumer`: Consome `transaction.transaction_created`
- Atualiza `spent_amount` do BudgetItem correspondente

---

### Scheduler (Worker Cron Jobs)

**Implementação** (`pkg/scheduler/`):
- Baseado em `robfig/cron/v3`
- Suporte a expressões cron com segundos
- Graceful shutdown
- Recovery automático de panics
- Controle de concorrência (max jobs simultâneos)
- Timeout por job

**Jobs Registrados**:
1. **OutboxDispatcher**: `@every 5s`
   - Processa eventos pendentes
   - Publica no RabbitMQ

2. **OutboxCleanup**: `@daily`
   - Remove eventos publicados/falhos antigos
   - Mantém banco limpo

**Configuração**:
```go
type Config struct {
    DefaultTimeout      time.Duration  // Timeout padrão (60s)
    MaxConcurrentJobs   int           // Max execuções simultâneas (10)
    EnableRecovery      bool          // Recovery de panics
}
```

---

## Banco de Dados

### Schema Completo

**Tabelas Principais**:
1. `users`: Usuários (root entity)
2. `categories`: Categorias hierárquicas
3. `payment_methods`: Métodos de pagamento (global)
4. `cards`: Cartões de crédito
5. `budgets`: Orçamentos mensais
6. `budget_items`: Items do orçamento
7. `invoices`: Faturas de cartão
8. `invoice_items`: Compras na fatura
9. `monthly_transactions`: Consolidado mensal
10. `transaction_items`: Transações individuais
11. `outbox_events`: Eventos para publicação
12. `processed_events`: Rastreamento de idempotência

### Tipos de Dados

**Monetários**: `NUMERIC(19,2)` (máxima precisão)
- TotalAmount: R$ 9.999.999.999.999.999,99
- Suporta valores até 17 dígitos + 2 decimais

**Percentuais**: `NUMERIC(6,3)`
- Range: 0.000% até 100.000%
- Exemplo: 25.5% = 25.500

**Timestamps**: `TIMESTAMPTZ` (com fuso horário)
- `created_at`: NOT NULL
- `updated_at`: Nullable
- `deleted_at`: Nullable (soft deletes)

**Datas de Referência**:
- Budgets e Invoices: `DATE` (primeiro dia do mês)
- Transactions: `VARCHAR(7)` formato "YYYY-MM"

### Constraints Importantes

**Checks**:
- `budgets.amount_goal > 0`
- `budgets.amount_used >= 0`
- `budgets.percentage_used BETWEEN 0 AND 100`
- `budget_items.percentage_goal BETWEEN 0 AND 100`
- `cards.due_day BETWEEN 1 AND 31`
- `cards.closing_offset_days BETWEEN 1 AND 31`
- `invoice_items.installment_number <= installment_total`
- `outbox_events.retry_count <= 3`

**Unique Constraints**:
- `users.email`
- `payment_methods.code`
- `budgets(user_id, date)`
- `invoices(user_id, card_id, reference_month)`
- `budget_items(budget_id, category_id)`
- `monthly_transactions(user_id, reference_month)`
- `processed_events(event_id, consumer_name)`

### Índices Otimizados

**Índices Compostos**:
- `categories(user_id, sequence)` WHERE parent_id IS NULL AND deleted_at IS NULL
- `invoices(user_id, reference_month)` WHERE deleted_at IS NULL
- `transaction_items(monthly_id)` WHERE deleted_at IS NULL

**Índices para Queries Comuns**:
- Budget por usuário e mês
- Categorias ativas por usuário
- Faturas por cartão e mês
- Transações por tipo
- Eventos pendentes no outbox

---

## Configuração

### Variáveis de Ambiente

**Servidor HTTP**:
- `HTTP_PORT`: Porta do servidor (padrão: 8000)
- `SERVICE_NAME_API`: Nome do serviço API

**Banco de Dados**:
- `DB_DRIVER`: postgres (compatível CockroachDB)
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
- `DB_MAX_IDLE_CONNS`: 6
- `DB_MAX_OPEN_CONNS`: 25
- `DB_CONN_MAX_LIFE_TIME_MINUTES`: 5
- `DB_CONN_MAX_IDLE_TIME_MINUTES`: 2
- `MIGRATE_PATH`: file://database/migrations

**Autenticação**:
- `AUTH_SECRET_KEY`: Mínimo 64 caracteres (validado em produção)
- `AUTH_TOKEN_DURATION`: Duração do token em horas (1-24)

**RabbitMQ**:
- `RABBITMQ_URL`: amqp://guest:guest@localhost:5672/
- `RABBITMQ_EXCHANGE`: financial.events
- `RABBITMQ_QUEUE`: financial_queue

**Observabilidade (OpenTelemetry)**:
- `OTEL_SERVICE_VERSION`: 1.0.0
- `OTEL_EXPORTER_OTLP_ENDPOINT`: localhost:4317
- `OTEL_EXPORTER_OTLP_PROTOCOL`: grpc
- `OTEL_EXPORTER_OTLP_INSECURE`: true
- `OTEL_TRACE_SAMPLE_RATE`: 1.0 (100%)
- `LOG_LEVEL`: info (debug/info/warn/error)
- `LOG_FORMAT`: json (json/text)

**Worker**:
- `SERVICE_NAME_WORKER`: financial-worker
- `WORKER_DEFAULT_TIMEOUT_SECONDS`: 60
- `WORKER_MAX_CONCURRENT_JOBS`: 10

**Consumer**:
- `SERVICE_NAME_CONSUMER`: financial-consumer
- `CONSUMER_BROKER_TYPE`: rabbitmq (kafka/sqs futuro)
- `CONSUMER_EXCHANGE`: financial.events
- `CONSUMER_WORKER_COUNT`: 5
- `CONSUMER_PREFETCH_COUNT`: 10

### Validação de Configuração

**Validações em Produção** (`configs/config.go`):
1. **DB_PASSWORD**: Mínimo 16 caracteres, não pode ser valor default
2. **AUTH_SECRET_KEY**: Mínimo 64 caracteres, não pode ser valor default
3. **RABBITMQ_URL**: Não pode usar credenciais default (guest:guest)
4. **AUTH_TOKEN_DURATION**: Entre 1 e 24 horas

**SafeDSN()**: Método que retorna DSN sem senha para logs (evita exposição)

---

## Padrões de Teste

### Padrão AAA (Arrange-Act-Assert)

Todos os testes devem seguir o padrão AAA para clareza e manutenibilidade:

- **Arrange**: Configurar dados de teste, mocks e dependências
- **Act**: Executar a função/método sendo testado
- **Assert**: Verificar os resultados esperados

### Testes Unitários com Mocks

Para testes unitários, sempre usar a configuração `.mockery.yml` e gerar mocks com:

```bash
make mocks
```

Isso garante geração consistente de mocks em toda a base de código.

### Estrutura de Teste com testify/suite

Usar `testify/suite` para organizar testes relacionados com setup compartilhado:

```go
package usecase_test

import (
    "context"
    "errors"
    "testing"

    "github.com/jailtonjunior94/financial/internal/domain/interfaces/mocks"
    "github.com/jailtonjunior94/financial/internal/application/usecase"
    "github.com/jailtonjunior94/financial/internal/application/dtos"

    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/suite"
)

type CreateCategoryUseCaseSuite struct {
    suite.Suite

    ctx                context.Context
    categoryRepository *mocks.CategoryRepository
}

func TestCreateCategoryUseCaseSuite(t *testing.T) {
    suite.Run(t, new(CreateCategoryUseCaseSuite))
}

func (s *CreateCategoryUseCaseSuite) SetupTest() {
    s.ctx = context.Background()
    s.categoryRepository = mocks.NewCategoryRepository(s.T())
}

func (s *CreateCategoryUseCaseSuite) TestExecute() {
    type (
        args struct {
            input *dtos.CreateCategoryInput
        }

        dependencies struct {
            categoryRepository *mocks.CategoryRepository
        }
    )

    scenarios := []struct {
        name         string
        args         args
        dependencies dependencies
        expect       func(output *dtos.CategoryOutput, err error)
    }{
        {
            name: "deve criar uma nova categoria com sucesso",
            args: args{
                input: &dtos.CreateCategoryInput{
                    Name: "Transport",
                },
            },
            dependencies: dependencies{
                categoryRepository: func() *mocks.CategoryRepository {
                    // Arrange: configurar mock
                    s.categoryRepository.
                        EXPECT().
                        Create(s.ctx, mock.Anything).
                        Return(nil).
                        Once()
                    return s.categoryRepository
                }(),
            },
            expect: func(output *dtos.CategoryOutput, err error) {
                // Assert: verificar resultados
                s.NoError(err)
                s.NotNil(output)
                s.Equal("Transport", output.Name)
            },
        },
        {
            name: "deve retornar erro ao falhar ao criar categoria",
            args: args{
                input: &dtos.CreateCategoryInput{
                    Name: "Transport",
                },
            },
            dependencies: dependencies{
                categoryRepository: func() *mocks.CategoryRepository {
                    // Arrange: configurar mock para falhar
                    s.categoryRepository.
                        EXPECT().
                        Create(s.ctx, mock.Anything).
                        Return(errors.New("database error")).
                        Once()
                    return s.categoryRepository
                }(),
            },
            expect: func(output *dtos.CategoryOutput, err error) {
                // Assert: verificar erro
                s.Error(err)
                s.Nil(output)
                s.Contains(err.Error(), "database error")
            },
        },
    }

    for _, scenario := range scenarios {
        s.T().Run(scenario.name, func(t *testing.T) {
            // Act: executar o use case
            uc := usecase.NewCreateCategoryUseCase(scenario.dependencies.categoryRepository)
            output, err := uc.Execute(s.ctx, scenario.args.input)

            // Assert: chamar função de verificação
            scenario.expect(output, err)
        })
    }
}
```

### Princípios-Chave de Testes

1. **Isolamento**: Cada teste deve ser independente e não depender de outros testes
2. **Testes Orientados a Tabelas**: Usar abordagem baseada em cenários para múltiplos casos de teste
3. **Configuração de Mocks**: Configurar mocks dentro de funções de dependência para cada cenário
4. **Nomenclatura Clara**: Usar nomes descritivos de teste em português explicando o comportamento esperado
5. **Cobertura Abrangente**: Testar cenários de sucesso e erro
6. **Verificação de Mocks**: Usar `.Once()` para garantir que mocks sejam chamados exatamente uma vez
7. **Setup e Teardown**: Usar `SetupTest()` para inicialização de testes

### Executando Testes Específicos

```bash
# Executar todos os testes em um pacote
go test -v -count=1 ./internal/category/application/usecase/

# Executar uma suite de testes específica
go test -v -count=1 -run TestCreateCategoryUseCaseSuite ./internal/category/application/usecase/

# Executar um caso de teste específico
go test -v -count=1 -run TestCreateCategoryUseCaseSuite/deve_criar_uma_nova_categoria ./internal/category/application/usecase/
```

---

## Estrutura do Projeto

```
.
├── cmd/
│   ├── main.go           # CLI Cobra (comandos api, migrate, consumer, worker)
│   ├── server/           # Setup do servidor HTTP e conexão de módulos
│   ├── consumer/         # Consumer RabbitMQ para eventos
│   ├── worker/           # Worker com cron jobs (Outbox Dispatcher, Cleanup)
│   └── .env.example      # Template de variáveis de ambiente
├── internal/
│   ├── user/             # Domínio User (auth, criação)
│   ├── category/         # Domínio Category (hierárquico com children)
│   ├── budget/           # Domínio Budget (com items, validação de porcentagem)
│   ├── card/             # Domínio Card (cartões de crédito)
│   ├── invoice/          # Domínio Invoice (faturas e compras)
│   ├── payment_method/   # Domínio PaymentMethod (métodos de pagamento)
│   └── transaction/      # Domínio Transaction (transações mensais consolidadas)
├── pkg/
│   ├── bundle/           # Container de injeção de dependências
│   ├── database/         # Abstração de DB, migrações, UoW, helpers de teste
│   ├── auth/             # Implementação JWT
│   ├── api/              # Utilitários HTTP, middlewares, error handlers
│   ├── custom_errors/    # Wrapping e mapeamento de erros
│   ├── outbox/           # Outbox Pattern implementation
│   ├── scheduler/        # Cron job scheduler
│   ├── messaging/        # Message broker abstractions
│   └── jobs/             # Job definitions
├── configs/              # Carregamento de configuração via Viper
├── database/migrations/  # Migrações SQL (formato Unix timestamp)
└── deployment/           # Docker Compose e Kubernetes
```

---

## Decisões Técnicas Importantes

### CockroachDB vs PostgreSQL

**Escolha**: CockroachDB
- Compatibilidade total com Postgres wire protocol
- Escalabilidade horizontal (futuro)
- Resiliência e distribuição geográfica
- Suporte a transações ACID

### Valores Monetários

**Decisão**: `NUMERIC(19,2)` em banco, `vos.Money` em código
- Precisão máxima (sem erros de ponto flutuante)
- Suporta valores até 17 dígitos + 2 decimais
- `Money` VO usa centavos internamente (int64)
- Métodos: `Add()`, `Subtract()`, `Float()`, `Cents()`

### Soft Deletes vs Hard Deletes

**Decisão**: Soft Deletes com `deleted_at`
- Auditoria completa
- Possibilidade de recuperação
- Índices com `WHERE deleted_at IS NULL`
- Métodos `.Delete()` nas entidades

### Referência Mensal

**Decisão**: Formatos diferentes por domínio
- **Budget**: `DATE` (primeiro dia do mês)
- **Invoice**: `DATE` (primeiro dia do mês de vencimento)
- **Transaction**: `VARCHAR(7)` formato "YYYY-MM"

**Motivo**:
- Budget e Invoice: Precisa de operações de data (comparação, ordenação)
- Transaction: Simplicidade de agrupamento e query

### Outbox Pattern

**Decisão**: Implementar Transactional Outbox
- Garante consistência entre banco e mensageria
- At-least-once delivery
- Idempotência via `processed_events`
- Worker independente (não bloqueia API)

### Padrão de Fatura Brasileira

**Decisão**: Implementar lógica determinística
- `closingDay = dueDay - offset`
- Compra ANTES do fechamento → fatura do mês
- Compra NO DIA ou APÓS → fatura do próximo mês
- Métodos testados: `CalculateClosingDay()`, `DetermineInvoiceMonth()`

### Item CREDIT_CARD Único

**Decisão**: Apenas um item CREDIT_CARD por mês no MonthlyTransaction
- Representa o total consolidado de faturas
- Atualizado automaticamente via eventos
- Não pode ser criado/editado manualmente
- Validado no aggregate root

### Validação de Budget 100%

**Decisão**: Soma de percentuais DEVE ser exatamente 100%
- Validado em `AddItems()` (criação)
- `AddItem()` permite parcial (< 100%)
- Factory valida em criação
- Permite usuário distribuir gradualmente

---

## Fluxos de Negócio Principais

### Criar Orçamento

1. User faz POST /api/v1/budgets com:
   - `total_amount`, `currency`, `reference_month`, `items[]`
2. Handler valida JWT e extrai UserID
3. Factory cria Budget e BudgetItems
4. Factory valida soma de percentuais = 100%
5. Use Case inicia Unit of Work
6. Repository persiste Budget e Items (transação única)
7. Commit ou rollback automático

### Criar Compra Parcelada

1. User faz POST /api/v1/purchases com:
   - `card_id`, `category_id`, `purchase_date`, `total_amount`, `installments`
2. Handler extrai UserID do contexto
3. Use Case busca Card via CardProvider
4. Para cada parcela (1 a N):
   - Calcula data de compra da parcela (purchase_date + (i-1) meses)
   - Usa `Card.DetermineInvoiceMonth(purchaseDate)` → (year, month)
   - Busca ou cria Invoice (user_id, card_id, reference_month)
   - Cria InvoiceItem com installment_number/total
   - Adiciona item à Invoice
5. Recalcula `TotalAmount` de cada Invoice
6. Persiste tudo em transação (UoW)
7. Emite `InvoiceCalculatedEvent` para cada fatura afetada

### Registrar Transação

1. User faz POST /api/v1/transactions com:
   - `category_id`, `title`, `amount`, `direction`, `type`, `is_paid`
2. Handler extrai UserID e ReferenceMonth
3. Use Case busca ou cria MonthlyTransaction
4. Se tipo = CREDIT_CARD:
   - Busca total de faturas via InvoiceTotalProvider
   - Chama `aggregate.UpdateOrCreateCreditCardItem()`
5. Senão:
   - Obtém strategy apropriada (`GetStrategy(type)`)
   - Valida com strategy
   - Cria TransactionItem
   - Adiciona ao aggregate via `AddItem()`
6. Aggregate recalcula totais automaticamente
7. Persiste em transação (UoW)
8. Emite `TransactionCreatedEvent` (Outbox)

### Processar Evento de Transação (Budget)

1. Worker executa OutboxDispatcher (@every 5s)
2. Busca eventos pendentes
3. Publica `transaction.transaction_created` no RabbitMQ
4. Consumer recebe mensagem
5. BudgetHandler:
   - Verifica se já processou (idempotência via `processed_events`)
   - Extrai `category_id`, `amount`, `reference_month`
   - Busca Budget do mês correspondente
   - Busca BudgetItem com category_id
   - Atualiza `spent_amount` do item
   - Budget recalcula totais automaticamente
6. Persiste mudanças
7. Marca evento como processado

---

## Rotas HTTP Completas

### Rotas Públicas (sem autenticação)

```
POST   /api/v1/users               # Criar usuário
POST   /api/v1/token               # Login (obter JWT)
GET    /health                     # Health check
```

### Rotas de Categorias (autenticadas)

```
GET    /api/v1/categories          # Listar categorias do usuário
GET    /api/v1/categories/{id}     # Buscar por ID
POST   /api/v1/categories          # Criar categoria
PUT    /api/v1/categories/{id}     # Atualizar
DELETE /api/v1/categories/{id}     # Remover (soft delete)
```

### Rotas de Cartões (autenticadas)

```
GET    /api/v1/cards               # Listar cartões
GET    /api/v1/cards/{id}          # Buscar por ID
POST   /api/v1/cards               # Criar cartão
PUT    /api/v1/cards/{id}          # Atualizar
DELETE /api/v1/cards/{id}          # Remover
```

### Rotas de Orçamentos (autenticadas)

```
GET    /api/v1/budgets             # Listar orçamentos
GET    /api/v1/budgets/{id}        # Buscar por ID
POST   /api/v1/budgets             # Criar orçamento
PUT    /api/v1/budgets/{id}        # Atualizar
DELETE /api/v1/budgets/{id}        # Remover
POST   /api/v1/budgets/{id}/items  # Adicionar item
```

### Rotas de Faturas (autenticadas)

```
POST   /api/v1/purchases                     # Criar compra parcelada
PUT    /api/v1/purchases/{id}                # Atualizar compra
DELETE /api/v1/purchases/{id}                # Remover compra
GET    /api/v1/invoices?month=YYYY-MM        # Listar faturas por mês
GET    /api/v1/invoices/{id}                 # Buscar fatura específica
GET    /api/v1/invoices/cards/{cardId}       # Listar faturas por cartão
```

### Rotas de Transações (autenticadas)

```
POST   /api/v1/transactions                  # Registrar transação
PUT    /api/v1/transactions/items/{id}       # Atualizar item
DELETE /api/v1/transactions/items/{id}       # Remover item
```

### Rotas de Métodos de Pagamento (públicas)

```
GET    /api/v1/payment-methods               # Listar todos
GET    /api/v1/payment-methods/{id}          # Buscar por ID
GET    /api/v1/payment-methods/code/{code}   # Buscar por código
POST   /api/v1/payment-methods               # Criar
PUT    /api/v1/payment-methods/{id}          # Atualizar
DELETE /api/v1/payment-methods/{id}          # Remover
```

---

## Value Objects (Resumo)

### Shared VOs (devkit-go)

- `UUID`: Identificador único
- `Money`: Valor monetário com precisão (centavos)
- `Currency`: BRL, USD, EUR
- `Percentage`: Percentual com scale 3 (0-100.000%)
- `NullableTime`: Timestamp nullable (soft deletes)

### Domain-Specific VOs

**User**:
- `UserName`: Nome do usuário (max 100 chars)
- `Email`: Email com validação de formato

**Category**:
- `CategoryName`: Nome da categoria (max 255 chars)
- `CategorySequence`: Sequência de ordenação (0-1000)

**Card**:
- `CardName`: Nome do cartão (max 255 chars)
- `DueDay`: Dia de vencimento (1-31)
- `ClosingOffsetDays`: Offset de fechamento (1-31)

**Budget**:
- `ReferenceMonth`: Mês de referência (DATE)

**PaymentMethod**:
- `PaymentMethodName`: Nome do método (max 255)
- `PaymentMethodCode`: Código único (max 50)
- `Description`: Descrição (max 500)

**Invoice**:
- `ReferenceMonth`: Mês de referência (DATE)

**Transaction**:
- `ReferenceMonth`: Mês de referência (string "YYYY-MM")
- `TransactionTitle`: Título da transação (max 255)
- `TransactionDirection`: INCOME ou EXPENSE
- `TransactionType`: PIX, BOLETO, TRANSFER, CREDIT_CARD

---

## Eventos de Domínio

### TransactionCreatedEvent

**Emitido quando**: Um novo TransactionItem é criado

**Aggregate**: `transaction`

**Event Type**: `transaction_created`

**Routing Key**: `transaction.transaction_created`

**Payload**:
```json
{
  "transaction_id": "uuid",
  "user_id": "uuid",
  "category_id": "uuid",
  "amount": 10000,          // centavos
  "currency": "BRL",
  "direction": "EXPENSE",
  "type": "PIX",
  "reference_month": "2026-01"
}
```

**Consumidores**:
- `BudgetConsumer`: Atualiza `spent_amount` do BudgetItem correspondente

### InvoiceCalculatedEvent (futuro)

**Emitido quando**: Invoice TotalAmount é recalculado

**Aggregate**: `invoice`

**Event Type**: `invoice_calculated`

**Routing Key**: `invoice.invoice_calculated`

**Payload**:
```json
{
  "invoice_id": "uuid",
  "user_id": "uuid",
  "card_id": "uuid",
  "reference_month": "2026-02-01",
  "total_amount": 250000,   // centavos
  "currency": "BRL"
}
```

**Consumidores**:
- `TransactionConsumer`: Atualiza item CREDIT_CARD no MonthlyTransaction

---

## Convenções de Código

### Naming Conventions

**Packages**:
- Singular: `user`, `category`, `budget`
- Lowercase, sem underscores

**Structs**:
- PascalCase: `User`, `BudgetItem`, `MonthlyTransaction`

**Interfaces**:
- PascalCase com sufixo: `UserRepository`, `TokenValidator`, `ErrorHandler`

**Métodos**:
- PascalCase (exported): `Create()`, `Update()`, `Delete()`
- camelCase (private): `recalculateTotals()`, `findItemByID()`

**Constantes**:
- Exported: `MaxRetryCount`, `StatusPending`
- Private: `hundredPercent`, `zeroPercentage`

### Error Handling

**Domain Errors**:
```go
var (
    ErrUserNotFound = errors.New("user not found")
    ErrEmailAlreadyExists = errors.New("email already exists")
)
```

**Custom Errors**:
```go
return customerrors.New("failed to create user", err)
return customerrors.NewWithDetails("validation failed", err, details)
```

**Error Wrapping**:
```go
return fmt.Errorf("repository: %w", err)
```

### Context Propagation

**Sempre passar `context.Context`**:
```go
func (r *repository) Create(ctx context.Context, entity *Entity) error
func (uc *usecase) Execute(ctx context.Context, input *Input) (*Output, error)
```

**User no contexto**:
```go
user, err := middlewares.GetUserFromContext(r.Context())
```

### Comentários

**Structs e Interfaces**:
```go
// User representa um usuário do sistema.
// Aggregate Root do módulo de autenticação.
type User struct { ... }
```

**Métodos Críticos**:
```go
// DetermineInvoiceMonth determina a qual fatura uma compra pertence.
// REGRA CRÍTICA: purchaseDate < closingDate → fatura do mês
func (c *Card) DetermineInvoiceMonth(purchaseDate time.Time) (int, time.Month)
```

---

## Comandos de Build e Desenvolvimento

### Setup do Ambiente
```bash
make dotenv  # Gerar arquivo .env do .env.example em cmd/
```

### Build
```bash
make build   # Compila para ./bin/financial (CGO_ENABLED=0, otimizado com -ldflags="-w -s")
```

### Testes
```bash
make test-unit           # Executar testes unitários com detecção de race e cobertura
make test-integration    # Executar testes de integração
make test-all           # Executar todos os testes
make cover              # Gerar e visualizar relatório HTML de cobertura
make cover-html         # Gerar relatórios HTML de cobertura
```

### Linting
```bash
make lint    # Executar golangci-lint com configuração .golangci.yml
```

### Mocks
```bash
make mocks   # Gerar mocks usando mockery (configurado em .mockery.yml)
```

### Migrações de Banco de Dados
```bash
make migrate NAME=migration_name  # Criar novos arquivos de migração em database/migrations
```

### Executando a Aplicação
```bash
# Executar servidor API
./bin/financial api

# Executar migrações de banco (CockroachDB)
./bin/financial migrate

# Consumer RabbitMQ
./bin/financial consumer

# Worker (cron jobs)
./bin/financial worker
```

### Docker
```bash
make start_minimal  # Iniciar CockroachDB, RabbitMQ e OTEL
make start_docker   # Iniciar todos os serviços + migration
make stop_docker    # Parar todos os serviços
```

---

## Implementações Pendentes

### Consumers de Eventos

1. **BudgetConsumer**:
   - Consome `transaction.transaction_created`
   - Atualiza `spent_amount` do BudgetItem
   - Já tem estrutura base, precisa implementação

2. **TransactionConsumer**:
   - Consome `invoice.invoice_calculated`
   - Atualiza item CREDIT_CARD no MonthlyTransaction
   - Ainda não criado

### Features Futuras

1. **Pagination**:
   - Listar categorias, budgets, invoices com paginação
   - Cursor-based ou offset-based

2. **Filtros Avançados**:
   - Filtrar transações por período, categoria, tipo
   - Filtrar faturas por status de pagamento

3. **Dashboards**:
   - Endpoint para resumo financeiro mensal
   - Gráficos de evolução de gastos

4. **Cache Redis**:
   - Cache de categorias (hierarquia)
   - Cache de payment_methods (dados estáticos)

---

## Resultado Esperado

Ao seguir este CLAUDE.md, a IA deve:

- Produzir respostas altamente contextualizadas baseadas no estado real do projeto
- Reduzir retrabalho e correções manuais
- Manter coerência entre todos os 7 módulos de domínio
- Ajudar na evolução do projeto sem quebrar padrões existentes
- Respeitar todas as decisões técnicas documentadas
- Priorizar reutilização sobre criação
- Manter consistência de estilo e qualidade de código
- Seguir os fluxos de negócio estabelecidos
- Utilizar corretamente os padrões arquiteturais implementados
- Considerar as regras de negócio críticas de cada módulo
