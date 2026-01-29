# Guia de Tratamento de Erros - Financial API

## 📋 Índice

1. [Princípios Fundamentais](#princípios-fundamentais)
2. [Arquitetura de Erros](#arquitetura-de-erros)
3. [Tipos de Erros e Status HTTP](#tipos-de-erros-e-status-http)
4. [Como Adicionar Novos Erros](#como-adicionar-novos-erros)
5. [Boas Práticas](#boas-práticas)
6. [Anti-Patterns (O que NÃO fazer)](#anti-patterns-o-que-não-fazer)
7. [Exemplos Práticos](#exemplos-práticos)
8. [Troubleshooting](#troubleshooting)

---

## Princípios Fundamentais

### 1. **Sempre Use Erros de Domínio Predefinidos**

✅ **CORRETO:**
```go
if budget == nil {
    return nil, domain.ErrBudgetNotFound
}
```

❌ **ERRADO:**
```go
if budget == nil {
    return nil, fmt.Errorf("budget not found")
}
```

**Por quê?**
- Erros de domínio são mapeados para status HTTP corretos (404, 409, etc)
- `fmt.Errorf()` genérico sempre retorna 500 (Internal Server Error)
- DRY: mensagens centralizadas
- Refactoring seguro: mudanças em um único lugar

### 2. **Mantenha 1:1 entre Definição e Mapeamento**

Sempre que adicionar um erro em `errors.go`, adicione o mapeamento HTTP correspondente em `error_mapping.go`.

### 3. **Erros de Domínio vs Erros Técnicos**

| Tipo | Quando Usar | Status HTTP | Exemplo |
|------|-------------|-------------|---------|
| **Domínio** | Regra de negócio violada, recurso não existe | 400, 404, 409 | `ErrBudgetNotFound` |
| **Técnico** | Falha de infraestrutura, DB down, panic | 500 | `sql.ErrConnDone` |
| **Validação** | Entrada inválida do usuário | 400 | `ErrInvalidEmail` |
| **Autenticação** | Token inválido/expirado | 401 | `ErrTokenExpired` |

---

## Arquitetura de Erros

### Componentes

```
┌─────────────────────────────────────────────────────────────┐
│                     HTTP Request                            │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                    Handler (HTTP Layer)                     │
│  - Valida entrada                                           │
│  - Chama Use Case                                           │
│  - Se erro: errorHandler.HandleError(w, r, err)             │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                 Use Case (Application Layer)                │
│  - Executa lógica de negócio                               │
│  - Retorna erros de domínio                                │
│  - Exemplo: return domain.ErrBudgetNotFound                │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│              ErrorHandler (pkg/api/httperrors)              │
│  1. Desembrulha CustomError                                │
│  2. Mapeia erro → status HTTP (ErrorMapper)                │
│  3. Registra no OpenTelemetry                              │
│  4. Faz log apropriado (ERROR/WARN/INFO)                   │
│  5. Retorna ProblemDetail (RFC 7807)                       │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                   HTTP Response (JSON)                      │
│  {                                                          │
│    "type": "https://httpstatuses.com/404",                 │
│    "title": "Not Found",                                   │
│    "status": 404,                                          │
│    "detail": "Budget not found",                           │
│    "instance": "/api/v1/budgets/123",                      │
│    "timestamp": "2025-01-29T12:00:00Z",                    │
│    "request_id": "req-xyz",                                │
│    "trace_id": "trace-abc"                                 │
│  }                                                          │
└─────────────────────────────────────────────────────────────┘
```

### Fluxo de Decisão

```
Erro ocorreu?
    │
    ├─ Sim → É erro de domínio/negócio?
    │        │
    │        ├─ Sim → Use erro predefinido (ErrXxxNotFound, ErrXxxInvalid, etc)
    │        │        └─ Verifique se está mapeado em error_mapping.go
    │        │
    │        └─ Não → É erro de validação?
    │                 │
    │                 ├─ Sim → Use erro de validação (ErrInvalidEmail, etc)
    │                 │
    │                 └─ Não → É erro SQL?
    │                          │
    │                          ├─ sql.ErrNoRows → Retorne nil, nil (repositório)
    │                          │                   Use case converte para ErrXxxNotFound
    │                          │
    │                          └─ Outro SQL error → Propague (será 500)
    │
    └─ Não → Sucesso
```

---

## Tipos de Erros e Status HTTP

### 400 Bad Request - Validação de Entrada

**Quando usar:** Entrada do usuário é inválida ou malformada.

**Exemplos:**
```go
// pkg/custom_errors/errors.go
ErrInvalidEmail         = errors.New("invalid email format")
ErrPasswordIsRequired   = errors.New("password is required")
ErrNameCannotBeEmpty    = errors.New("name cannot be empty")
ErrCategoryCycle        = errors.New("category cannot be its own parent or create a cycle")

// invoice/domain/errors.go
ErrInvalidPurchaseDate  = errors.New("purchase date cannot be in the future")
ErrNegativeAmount       = errors.New("amount cannot be negative")
ErrEmptyDescription     = errors.New("description cannot be empty")
```

**Uso:**
```go
func (e *Entity) Validate() error {
    if e.Email == "" {
        return customErrors.ErrInvalidEmail
    }
    if e.Amount < 0 {
        return domain.ErrNegativeAmount
    }
    return nil
}
```

---

### 401 Unauthorized - Autenticação

**Quando usar:** Token ausente, inválido ou expirado.

**Exemplos:**
```go
ErrUnauthorized       = errors.New("unauthorized: user not found in context")
ErrMissingAuthHeader  = errors.New("missing authorization header")
ErrInvalidToken       = errors.New("invalid or malformed token")
ErrTokenExpired       = errors.New("token has expired")
```

**Uso:**
```go
func (m *authMiddleware) Authorization(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            m.errorHandler.HandleError(w, r, customErrors.ErrMissingAuthHeader)
            return
        }
        // ...
    })
}
```

---

### 404 Not Found - Recurso Inexistente

**Quando usar:** Recurso solicitado não existe no sistema.

**Exemplos:**
```go
// pkg/custom_errors/errors.go
ErrBudgetNotFound        = errors.New("budget not found")
ErrCardNotFound          = errors.New("card not found")
ErrCategoryNotFound      = errors.New("category not found")
ErrPaymentMethodNotFound = errors.New("payment method not found")
ErrUserNotFound          = errors.New("user not found")

// invoice/domain/errors.go
ErrInvoiceNotFound     = errors.New("invoice not found")
ErrInvoiceItemNotFound = errors.New("invoice item not found")
```

**Uso (Use Case):**
```go
func (u *findBudgetUseCase) Execute(ctx context.Context, budgetID string) (*dtos.BudgetOutput, error) {
    budget, err := u.budgetRepository.FindByID(ctx, id)
    if err != nil {
        return nil, err
    }

    if budget == nil {
        return nil, domain.ErrBudgetNotFound  // ← Retorna 404
    }

    return toDTO(budget), nil
}
```

**Uso (Repository):**
```go
func (r *budgetRepository) FindByID(ctx context.Context, id vos.UUID) (*entities.Budget, error) {
    var budget entities.Budget
    err := r.db.QueryRowContext(ctx, query, id).Scan(&budget)

    if err == sql.ErrNoRows {  // ← Use comparação de erro, não string
        return nil, nil         // ← Repositório retorna nil, não erro
    }

    if err != nil {
        return nil, err         // ← Outros erros SQL são propagados
    }

    return &budget, nil
}
```

---

### 409 Conflict - Conflito de Estado

**Quando usar:** Operação viola constraint de integridade ou estado do sistema.

**Exemplos:**
```go
ErrEmailAlreadyExists           = errors.New("email already exists")
ErrInvalidParentCategory        = errors.New("parent category not found or belongs to different user")

// invoice/domain/errors.go
ErrInvoiceAlreadyExistsForMonth = errors.New("invoice already exists for this card and month")
```

**Uso:**
```go
func (u *createUserUseCase) Execute(ctx context.Context, input *CreateUserInput) error {
    existing, _ := u.userRepository.FindByEmail(ctx, input.Email)
    if existing != nil {
        return customErrors.ErrEmailAlreadyExists  // ← Retorna 409
    }
    // ...
}
```

---

### 500 Internal Server Error - Falha Técnica

**Quando usar:** Erro inesperado de infraestrutura (DB down, panic, etc).

**NÃO use para:**
- ❌ Recurso não encontrado (use 404)
- ❌ Validação de entrada (use 400)
- ❌ Conflito de dados (use 409)

**Exemplos legítimos:**
```go
// Erro SQL não esperado (não ErrNoRows)
if err != nil {
    return nil, err  // ← Será mapeado para 500
}

// Erro de conexão com serviço externo
if err := externalService.Call(); err != nil {
    return err  // ← 500
}
```

---

## Como Adicionar Novos Erros

### Checklist

- [ ] **1. Definir o erro de domínio**
  - Arquivo: `pkg/custom_errors/errors.go` (global) ou `<module>/domain/errors.go` (módulo)
  - Nome: `ErrXxxNotFound`, `ErrXxxInvalid`, etc
  - Mensagem clara e acionável

- [ ] **2. Mapear para status HTTP**
  - Arquivo: `pkg/api/httperrors/error_mapping.go`
  - Adicionar import se necessário (erros de módulo)
  - Adicionar ao mapa `buildDomainErrorMappings()`

- [ ] **3. Usar nos use cases**
  - Substituir `fmt.Errorf()` por erro de domínio
  - Importar o pacote de domínio

- [ ] **4. Testar**
  - Criar teste de integração verificando status HTTP correto
  - Validar formato ProblemDetail (RFC 7807)

### Exemplo Completo: Adicionar ErrTransactionNotFound

#### Passo 1: Definir Erro

**Opção A: Erro Global** (`pkg/custom_errors/errors.go`)
```go
var (
    // ... outros erros
    ErrTransactionNotFound = errors.New("transaction not found")
)
```

**Opção B: Erro de Módulo** (`internal/transaction/domain/errors.go`)
```go
package domain

import "errors"

var (
    ErrTransactionNotFound = errors.New("transaction not found")
    ErrInvalidAmount       = errors.New("amount cannot be negative")
    // ...
)
```

#### Passo 2: Mapear HTTP

**Se erro global:**
```go
// pkg/api/httperrors/error_mapping.go
func buildDomainErrorMappings() map[error]ErrorMapping {
    return map[error]ErrorMapping{
        // ...

        // Not found errors → 404 Not Found
        customerrors.ErrTransactionNotFound: {
            Status:  http.StatusNotFound,
            Message: "Transaction not found",
        },

        // ...
    }
}
```

**Se erro de módulo:**
```go
// pkg/api/httperrors/error_mapping.go

// 1. Adicionar import
import (
    // ...
    transactiondomain "github.com/jailtonjunior94/financial/internal/transaction/domain"
)

// 2. Adicionar ao mapa
func buildDomainErrorMappings() map[error]ErrorMapping {
    return map[error]ErrorMapping {
        // ...

        transactiondomain.ErrTransactionNotFound: {
            Status:  http.StatusNotFound,
            Message: "Transaction not found",
        },
        transactiondomain.ErrInvalidAmount: {
            Status:  http.StatusBadRequest,
            Message: "Amount cannot be negative",
        },

        // ...
    }
}
```

#### Passo 3: Usar no Use Case

```go
// internal/transaction/application/usecase/find_transaction.go
package usecase

import (
    "context"
    "fmt"

    "github.com/jailtonjunior94/financial/internal/transaction/application/dtos"
    "github.com/jailtonjunior94/financial/internal/transaction/domain"
    "github.com/jailtonjunior94/financial/internal/transaction/domain/interfaces"
    // ...
)

func (u *findTransactionUseCase) Execute(ctx context.Context, txID string) (*dtos.TransactionOutput, error) {
    transaction, err := u.transactionRepository.FindByID(ctx, id)
    if err != nil {
        return nil, err
    }

    if transaction == nil {
        return nil, domain.ErrTransactionNotFound  // ✅ Usa erro de domínio
    }

    return toDTO(transaction), nil
}
```

#### Passo 4: Testar

```go
// internal/transaction/infrastructure/http/transaction_handler_test.go
func TestTransactionHandler_FindBy_NotFound(t *testing.T) {
    // Setup
    handler := setupTestHandler(t)
    nonExistentID := "99999999-9999-9999-9999-999999999999"

    // Execute
    req := httptest.NewRequest("GET", "/api/v1/transactions/"+nonExistentID, nil)
    req = req.WithContext(withAuthUser(req.Context(), testUser))
    rr := httptest.NewRecorder()

    handler.ServeHTTP(rr, req)

    // Assert
    assert.Equal(t, http.StatusNotFound, rr.Code)

    var problem httperrors.ProblemDetail
    json.Unmarshal(rr.Body.Bytes(), &problem)

    assert.Equal(t, "Transaction not found", problem.Detail)
    assert.Equal(t, 404, problem.Status)
    assert.Equal(t, "https://httpstatuses.com/404", problem.Type)
    assert.Equal(t, "Not Found", problem.Title)
}
```

---

## Boas Práticas

### 1. Use `errors.Is()` para Wrapped Errors

✅ **CORRETO:**
```go
if errors.Is(err, sql.ErrNoRows) {
    return nil, nil
}
```

⚠️ **ACEITÁVEL (mas não recomendado):**
```go
if err == sql.ErrNoRows {  // Funciona, mas não suporta wrapped errors
    return nil, nil
}
```

❌ **ERRADO:**
```go
if err.Error() == "sql: no rows in result set" {
    return nil, nil
}
```

**Por quê?**
- `errors.Is()` funciona com wrapped errors (`fmt.Errorf("...: %w", err)`)
- String comparison é frágil (quebra se mensagem mudar)
- Go idiomático e preparado para o futuro
- Comparação direta (`==`) só funciona se erro não foi wrapped

### 2. Repositórios: Retornar `nil, nil` para NotFound

✅ **CORRETO:**
```go
func (r *repo) FindByID(ctx context.Context, id UUID) (*Entity, error) {
    // ...
    if err == sql.ErrNoRows {
        return nil, nil  // ← Não é erro do repositório
    }
    return entity, err
}
```

❌ **ERRADO:**
```go
func (r *repo) FindByID(ctx context.Context, id UUID) (*Entity, error) {
    // ...
    if err == sql.ErrNoRows {
        return nil, domain.ErrEntityNotFound  // ← Repositório não decide isso
    }
    return entity, err
}
```

**Por quê?**
- Separação de responsabilidades (SRP)
- Repositório não sabe se "não encontrado" é erro de domínio
- Use case decide a semântica

### 3. Use Cases: Converter `nil` para Erro de Domínio

✅ **CORRETO:**
```go
func (u *useCase) Execute(ctx context.Context, id string) (*Output, error) {
    entity, err := u.repo.FindByID(ctx, id)
    if err != nil {
        return nil, err
    }

    if entity == nil {
        return nil, domain.ErrEntityNotFound  // ← Use case decide
    }

    return toDTO(entity), nil
}
```

### 4. Mensagens de Erro Claras e Acionáveis

✅ **CORRETO:**
```go
ErrInvalidEmail = errors.New("invalid email format")
ErrInvoiceAlreadyExistsForMonth = errors.New("invoice already exists for this card and month")
```

❌ **ERRADO:**
```go
ErrBadInput = errors.New("bad input")
ErrFailed = errors.New("failed")
```

### 5. Agrupe Erros por Categoria

```go
// pkg/custom_errors/errors.go
var (
    // Authentication errors
    ErrUnauthorized      = errors.New("...")
    ErrInvalidToken      = errors.New("...")
    ErrTokenExpired      = errors.New("...")

    // Domain errors - Not Found
    ErrBudgetNotFound    = errors.New("...")
    ErrCardNotFound      = errors.New("...")

    // Domain errors - Validation
    ErrInvalidEmail      = errors.New("...")
    ErrPasswordIsRequired = errors.New("...")
)
```

---

## Anti-Patterns (O que NÃO fazer)

### ❌ 1. Usar `fmt.Errorf()` para Erros de Domínio

```go
// ERRADO
if budget == nil {
    return nil, fmt.Errorf("budget not found")  // ← Retorna 500
}

// CORRETO
if budget == nil {
    return nil, domain.ErrBudgetNotFound  // ← Retorna 404
}
```

### ❌ 2. Comparar Erros por String

```go
// ERRADO
if err.Error() == "sql: no rows in result set" {
    // ...
}

// CORRETO (recomendado - suporta wrapped errors)
if errors.Is(err, sql.ErrNoRows) {
    // ...
}

// ACEITÁVEL (mas não suporta wrapped errors)
if err == sql.ErrNoRows {
    // ...
}
```

### ❌ 3. Definir Erro mas Não Mapear

```go
// pkg/custom_errors/errors.go
ErrCardNotFound = errors.New("card not found")  // ✅ Definido

// pkg/api/httperrors/error_mapping.go
// ❌ AUSENTE no mapa → retorna 500 em vez de 404
```

### ❌ 4. Retornar Erro de Domínio do Repositório

```go
// ERRADO - Repositório não deve decidir semântica de domínio
func (r *repo) FindByID(ctx context.Context, id UUID) (*Entity, error) {
    if err == sql.ErrNoRows {
        return nil, domain.ErrEntityNotFound  // ❌
    }
}

// CORRETO - Use case decide
func (r *repo) FindByID(ctx context.Context, id UUID) (*Entity, error) {
    if err == sql.ErrNoRows {
        return nil, nil  // ✅
    }
}
```

### ❌ 5. Ignorar Context de Erros

```go
// ERRADO - Perde contexto
if err != nil {
    return err
}

// CORRETO - Adiciona contexto
if err != nil {
    return fmt.Errorf("failed to create invoice: %w", err)  // ← %w preserva erro original
}
```

---

## Exemplos Práticos

### Exemplo 1: Endpoint GET (Not Found)

```go
// Handler
func (h *BudgetHandler) FindBy(w http.ResponseWriter, r *http.Request) {
    budgetID := chi.URLParam(r, "id")

    budget, err := h.findBudgetUseCase.Execute(r.Context(), budgetID)
    if err != nil {
        h.errorHandler.HandleError(w, r, err)  // ← Delega para error handler
        return
    }

    render.JSON(w, r, budget)
}

// Use Case
func (u *findBudgetUseCase) Execute(ctx context.Context, budgetID string) (*dtos.BudgetOutput, error) {
    id, err := vos.NewUUIDFromString(budgetID)
    if err != nil {
        return nil, fmt.Errorf("invalid budget_id: %w", err)  // ← 400 (validação heurística)
    }

    budget, err := u.budgetRepository.FindByID(ctx, id)
    if err != nil {
        return nil, err  // ← Propaga erro SQL (500)
    }

    if budget == nil {
        return nil, domain.ErrBudgetNotFound  // ← 404
    }

    return toDTO(budget), nil
}

// Repository
func (r *budgetRepository) FindByID(ctx context.Context, id vos.UUID) (*entities.Budget, error) {
    var budget entities.Budget
    err := r.db.QueryRowContext(ctx, query, id).Scan(&budget)

    if errors.Is(err, sql.ErrNoRows) {
        return nil, nil  // ← Não é erro
    }

    if err != nil {
        return nil, err  // ← Erro SQL real
    }

    return &budget, nil
}
```

**Resultado:**
- `GET /api/v1/budgets/invalid-uuid` → **400** (UUID inválido)
- `GET /api/v1/budgets/99999999-9999-9999-9999-999999999999` → **404** (não existe)
- `GET /api/v1/budgets/valid-uuid` quando DB down → **500** (falha técnica)

### Exemplo 2: Endpoint POST (Validação + Conflito)

```go
// Use Case
func (u *createUserUseCase) Execute(ctx context.Context, input *CreateUserInput) error {
    // Validação
    if input.Email == "" {
        return customErrors.ErrInvalidEmail  // ← 400
    }

    // Verifica conflito
    existing, _ := u.userRepository.FindByEmail(ctx, input.Email)
    if existing != nil {
        return customErrors.ErrEmailAlreadyExists  // ← 409
    }

    user := factories.NewUser(input.Name, input.Email, input.Password)

    if err := u.userRepository.Create(ctx, user); err != nil {
        return err  // ← 500 (falha ao salvar)
    }

    return nil
}
```

**Resultado:**
- `POST /api/v1/users` com email vazio → **400**
- `POST /api/v1/users` com email duplicado → **409**
- `POST /api/v1/users` quando DB down → **500**

---

## Troubleshooting

### Problema: Erro retornando 500 em vez de 404

**Causa:** Erro não está mapeado em `error_mapping.go`.

**Solução:**
1. Verifique se o erro está definido em `errors.go`
2. Adicione o mapeamento em `error_mapping.go`
3. Se for erro de módulo, adicione o import

### Problema: `import not used` após adicionar erro de domínio

**Causa:** Você adicionou o import mas ainda não usou o erro no código.

**Solução:**
1. Substitua `fmt.Errorf()` pelo erro de domínio
2. Ou remova o import se não for usar ainda

### Problema: Teste falhando com 500 em vez de 404

**Causa:** Use case usando `fmt.Errorf()` ou erro não mapeado.

**Solução:**
1. Verifique o use case - deve usar erro de domínio
2. Verifique o error_mapping.go - erro deve estar no mapa
3. Execute o teste com verbose para ver o erro exato

### Problema: `sql.ErrNoRows` não compilando

**Causa:** Import `database/sql` faltando.

**Solução:**
```go
import (
    "context"
    "database/sql"  // ← Adicione esta linha
    // ...
)
```

---

## Referências

- [RFC 7807 - Problem Details for HTTP APIs](https://datatracker.ietf.org/doc/html/rfc7807)
- [HTTP Status Codes](https://httpstatuses.com/)
- [Go Error Handling Best Practices](https://go.dev/blog/error-handling-and-go)
- [Effective Go - Errors](https://go.dev/doc/effective_go#errors)

---

**Última atualização:** 2025-01-29
**Versão:** 1.0.0
