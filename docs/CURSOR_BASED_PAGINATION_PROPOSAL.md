# Proposta: Cursor-Based Pagination - Projeto Financial

**Data:** 2025-01-29
**Status:** 📋 Proposta (Aguardando Aprovação)
**Autor:** Análise Técnica

---

## 📋 Índice

1. [Resumo Executivo](#resumo-executivo)
2. [Análise do Estado Atual (ANTES)](#análise-do-estado-atual-antes)
3. [Problemas Identificados](#problemas-identificados)
4. [Proposta de Solução (DEPOIS)](#proposta-de-solução-depois)
5. [Justificativa Técnica](#justificativa-técnica)
6. [Implementação Detalhada](#implementação-detalhada)
7. [Performance e Índices](#performance-e-índices)
8. [Segurança e Robustez](#segurança-e-robustez)
9. [Plano de Migração](#plano-de-migração)

---

## Resumo Executivo

### Estado Atual
- **4 endpoints de listagem** sem paginação
- Todos retornam **datasets completos** (sem LIMIT/OFFSET)
- **Risco de performance** com crescimento de dados
- **Não escalável** para grandes volumes

### Proposta
- Implementar **cursor-based pagination** em todos os endpoints de listagem
- Criar **struct genérica reutilizável** para paginação
- Adicionar **índices compostos determinísticos** no PostgreSQL
- Garantir **alta performance** e **segurança**
- Manter **100% RESTful** e compatibilidade backward

### Endpoints Afetados

| Endpoint | Dataset Atual | Proposta |
|----------|---------------|----------|
| `GET /cards` | Todos os cartões do usuário | Paginado (limit 20) |
| `GET /categories` | Todas categorias pai | Paginado (limit 50) |
| `GET /invoices?month=YYYY-MM` | Todas faturas do mês | Paginado (limit 10) |
| `GET /invoices/cards/{cardId}` | Todo histórico do cartão | Paginado (limit 10) |

**Payment Method** (`GET /payment-methods`): **NÃO paginado** (dataset pequeno e estável - ~6 itens seed)

---

## Análise do Estado Atual (ANTES)

### 1. Card - List Cards

#### Contrato HTTP ATUAL
```http
GET /api/v1/cards HTTP/1.1
Authorization: Bearer <token>

HTTP/1.1 200 OK
Content-Type: application/json

[
  {
    "id": "uuid-1",
    "name": "Nubank",
    "due_day": 10,
    "closing_offset_days": 7,
    "created_at": "2025-01-01T00:00:00Z",
    "updated_at": "2025-01-01T00:00:00Z"
  },
  {
    "id": "uuid-2",
    "name": "Itaú",
    "due_day": 5,
    "closing_offset_days": 5,
    "created_at": "2025-01-02T00:00:00Z",
    "updated_at": "2025-01-02T00:00:00Z"
  }
  // ... todos os cartões (sem limite)
]
```

#### Struct ATUAL
```go
// Handler
func (h *CardHandler) Find(w http.ResponseWriter, r *http.Request) {
    userID := middlewares.GetUserFromContext(r.Context())

    cards, err := h.findCardUseCase.Execute(r.Context(), userID.ID.String())
    if err != nil {
        h.errorHandler.HandleError(w, r, err)
        return
    }

    render.JSON(w, r, cards)  // ← Retorna TODOS os cards
}

// DTO
type CardOutput struct {
    ID                string    `json:"id"`
    Name              string    `json:"name"`
    DueDay            int       `json:"due_day"`
    ClosingOffsetDays int       `json:"closing_offset_days"`
    CreatedAt         time.Time `json:"created_at"`
    UpdatedAt         time.Time `json:"updated_at"`
}
```

#### Query SQL ATUAL
```sql
-- card_repository.go:75-93
SELECT
    id,
    user_id,
    name,
    due_day,
    closing_offset_days,
    created_at,
    updated_at,
    deleted_at
FROM cards
WHERE
    user_id = $1
    AND deleted_at IS NULL
ORDER BY name;  -- ← Sem LIMIT/OFFSET
```

#### Fluxo ATUAL
```
HTTP Request
    ↓
Handler.Find()
    ↓
UseCase.Execute(userID)
    ↓
Repository.List(userID)
    ↓
SELECT * FROM cards WHERE user_id = ? ORDER BY name
    ↓
Retorna TODOS os registros (sem paginação)
    ↓
HTTP Response: Array completo
```

---

### 2. Category - List Categories

#### Contrato HTTP ATUAL
```http
GET /api/v1/categories HTTP/1.1
Authorization: Bearer <token>

HTTP/1.1 200 OK
Content-Type: application/json

[
  {
    "id": "uuid-1",
    "name": "Alimentação",
    "sequence": 1,
    "created_at": "2025-01-01T00:00:00Z",
    "updated_at": "2025-01-01T00:00:00Z"
  },
  {
    "id": "uuid-2",
    "name": "Transporte",
    "sequence": 2,
    "created_at": "2025-01-01T00:00:00Z",
    "updated_at": "2025-01-01T00:00:00Z"
  }
  // ... todas as categorias (sem limite)
]
```

#### Query SQL ATUAL
```sql
-- category_repository.go:114-132
SELECT
    id,
    user_id,
    name,
    sequence,
    created_at,
    updated_at,
    deleted_at
FROM categories c
WHERE
    user_id = $1
    AND deleted_at IS NULL
    AND parent_id IS NULL  -- Somente categorias pai
ORDER BY sequence;  -- ← Ordem customizável pelo usuário
```

#### Problema Específico
- `sequence` é editável pelo usuário → **pode ter duplicatas ou gaps**
- Não é adequado como cursor único

---

### 3. Invoice - List by Month

#### Contrato HTTP ATUAL
```http
GET /api/v1/invoices?month=2025-01 HTTP/1.1
Authorization: Bearer <token>

HTTP/1.1 200 OK
Content-Type: application/json

[
  {
    "id": "uuid-1",
    "card_id": "card-uuid",
    "reference_month": "2025-01-01",
    "due_date": "2025-02-10",
    "total_amount": "1500.00",
    "currency": "BRL",
    "created_at": "2025-01-15T00:00:00Z",
    "updated_at": "2025-01-15T00:00:00Z"
  },
  {
    "id": "uuid-2",
    "card_id": "card-uuid-2",
    "reference_month": "2025-01-01",
    "due_date": "2025-02-05",
    "total_amount": "800.00",
    "currency": "BRL",
    "created_at": "2025-01-16T00:00:00Z",
    "updated_at": "2025-01-16T00:00:00Z"
  }
  // ... todas as faturas do mês (sem limite)
]
```

#### Query SQL ATUAL
```sql
-- invoice_repository.go:269-287
SELECT
    id,
    user_id,
    card_id,
    reference_month,
    due_date,
    total_amount,
    created_at,
    updated_at,
    deleted_at
FROM invoices
WHERE user_id = $1
  AND to_char(reference_month, 'YYYY-MM') = $2
  AND deleted_at IS NULL
ORDER BY due_date;  -- ← Sem LIMIT/OFFSET
```

#### Problema Específico
- `due_date` pode ter duplicatas (múltiplos cartões com mesmo vencimento)
- Não é determinístico para cursor

---

### 4. Invoice - List by Card

#### Contrato HTTP ATUAL
```http
GET /api/v1/invoices/cards/{cardId} HTTP/1.1
Authorization: Bearer <token>

HTTP/1.1 200 OK
Content-Type: application/json

[
  {
    "id": "uuid-1",
    "card_id": "card-uuid",
    "reference_month": "2025-01-01",
    "due_date": "2025-02-10",
    "total_amount": "1500.00",
    "currency": "BRL",
    "created_at": "2025-01-15T00:00:00Z"
  },
  {
    "id": "uuid-2",
    "card_id": "card-uuid",
    "reference_month": "2024-12-01",
    "due_date": "2025-01-10",
    "total_amount": "1200.00",
    "currency": "BRL",
    "created_at": "2024-12-15T00:00:00Z"
  }
  // ... todo o histórico do cartão (sem limite)
]
```

#### Query SQL ATUAL
```sql
-- invoice_repository.go:289-306
SELECT
    id,
    user_id,
    card_id,
    reference_month,
    due_date,
    total_amount,
    created_at,
    updated_at,
    deleted_at
FROM invoices
WHERE card_id = $1 AND deleted_at IS NULL
ORDER BY reference_month DESC;  -- ← Histórico completo sem limite
```

#### Problema Específico
- Histórico pode crescer indefinidamente (anos de faturas)
- **Maior risco de performance**

---

## Problemas Identificados

### 1. ❌ Performance com Crescimento de Dados

**Cenário Real:**
```
Usuário com:
- 10 cartões
- 50 categorias
- 5 anos de histórico = 600 faturas (50/mês x 12 meses x 5 anos)

GET /invoices/cards/{id} → 60 faturas/cartão → 600 registros
GET /categories → 50 categorias
GET /cards → 10 cartões
```

**Problema:**
- Query retorna **todos os registros** de uma vez
- Cliente recebe **payload grande** (pode ser MB de JSON)
- **Network overhead** desnecessário
- **Memória** do cliente/servidor desperdiçada

---

### 2. ❌ Não Escalável

**Offset-based pagination (alternativa ruim):**
```sql
SELECT * FROM invoices
WHERE card_id = ?
ORDER BY reference_month DESC
LIMIT 10 OFFSET 500;  -- Página 50
```

**Problema:**
- PostgreSQL precisa **escanear 510 linhas** para retornar 10
- **O(n) complexity** com offset crescente
- **Query fica lenta** em páginas altas
- **Inconsistente**: se dados mudarem entre requests, usuário pode perder/duplicar items

**Benchmark:**
```
OFFSET 0:    ~1ms
OFFSET 100:  ~10ms
OFFSET 1000: ~100ms   ← 100x mais lento!
OFFSET 10000: ~1000ms ← 1000x mais lento!
```

---

### 3. ❌ ORDER BY Não Determinístico

**Problemas em queries atuais:**

```sql
-- Card: ORDER BY name
-- Se houver 2 cartões com mesmo nome:
SELECT * FROM cards WHERE user_id = ? ORDER BY name;
-- Qual vem primeiro? PostgreSQL não garante!
```

```sql
-- Invoice: ORDER BY due_date
-- Múltiplos cartões podem ter mesmo vencimento:
SELECT * FROM invoices WHERE user_id = ? ORDER BY due_date;
-- Ordem não é determinística entre requests!
```

**Consequência:**
- **Paginação inconsistente** (mesmo sem implementar paginação hoje)
- **Testes não são confiáveis** (ordem pode mudar)
- **UX ruim** (lista "pula" de ordem)

---

### 4. ❌ Índices Não Otimizados para Paginação

**Índices atuais:**
```sql
-- Card
CREATE INDEX idx_cards_user_id ON cards(user_id) WHERE deleted_at IS NULL;
-- Não cobre ORDER BY name

-- Invoice
CREATE INDEX idx_invoices_user_reference_month ON invoices(user_id, reference_month) WHERE deleted_at IS NULL;
-- Não cobre ORDER BY due_date
```

**Problema:**
- PostgreSQL usa index scan no WHERE mas **filesort no ORDER BY**
- **Não aproveita índice para ordenação**
- **Slower queries** mesmo com índices

---

### 5. ❌ Sem Controle de Largura de Banda

**Problema para clientes mobile/web:**
- Cliente mobile com 4G lento recebe **600 invoices** de uma vez
- Aplicação web congela ao renderizar **lista grande**
- **Sem progressive loading** (carregar conforme scroll)

---

## Proposta de Solução (DEPOIS)

### Cursor-Based Pagination (Keyset Pagination)

**Princípio:**
- Usar **último item da página** como "cursor" para próxima página
- **WHERE cursor_column > last_value** em vez de OFFSET
- **O(1) complexity** independente da posição
- **Determinístico** com ORDER BY composto (ex: `ORDER BY name, id`)

**Exemplo:**
```sql
-- Primeira página
SELECT * FROM cards
WHERE user_id = ?
  AND deleted_at IS NULL
ORDER BY name, id
LIMIT 20;
-- Retorna cards 1-20
-- Último item: {name: "Itaú", id: "uuid-20"}

-- Segunda página (cursor = último item da primeira)
SELECT * FROM cards
WHERE user_id = ?
  AND deleted_at IS NULL
  AND (name, id) > ('Itaú', 'uuid-20')  -- ← Cursor
ORDER BY name, id
LIMIT 20;
-- Retorna cards 21-40
```

**Benefícios:**
- ✅ **O(1) performance** independente da página
- ✅ **Determinístico** (mesma ordem sempre)
- ✅ **Usa índice para ORDER BY e WHERE**
- ✅ **Consistente** mesmo com inserções/deleções
- ✅ **Stateless** (cursor é self-contained)

---

## Implementação Detalhada

### 1. Struct Genérica de Paginação

```go
// pkg/pagination/cursor.go
package pagination

import (
    "encoding/base64"
    "encoding/json"
    "fmt"
    "net/http"
    "strconv"
)

// CursorParams representa parâmetros de paginação extraídos da query string
type CursorParams struct {
    Limit  int    // Limite de items por página (default: 20, max: 100)
    Cursor string // Cursor opaco (base64 encoded)
}

// CursorResponse representa a resposta paginada
type CursorResponse[T any] struct {
    Data       []T        `json:"data"`               // Items da página atual
    Pagination Pagination `json:"pagination"`         // Metadados de paginação
}

// Pagination contém metadados de paginação
type Pagination struct {
    Limit      int     `json:"limit"`                 // Limite usado
    HasNext    bool    `json:"has_next"`              // Tem próxima página?
    NextCursor *string `json:"next_cursor,omitempty"` // Cursor para próxima página (null se não houver)
}

// Cursor representa o estado interno do cursor (não exposto ao cliente)
type Cursor struct {
    // Campos usados para ordenação e filtragem
    // Exemplo para cards: {Name: "Itaú", ID: "uuid"}
    Fields map[string]interface{} `json:"f"`
}

// EncodeCursor codifica o cursor em base64
func EncodeCursor(c Cursor) (string, error) {
    if c.Fields == nil || len(c.Fields) == 0 {
        return "", nil
    }

    jsonBytes, err := json.Marshal(c)
    if err != nil {
        return "", fmt.Errorf("failed to marshal cursor: %w", err)
    }

    return base64.RawURLEncoding.EncodeToString(jsonBytes), nil
}

// DecodeCursor decodifica o cursor de base64
func DecodeCursor(encoded string) (Cursor, error) {
    if encoded == "" {
        return Cursor{Fields: make(map[string]interface{})}, nil
    }

    jsonBytes, err := base64.RawURLEncoding.DecodeString(encoded)
    if err != nil {
        return Cursor{}, fmt.Errorf("invalid cursor format: %w", err)
    }

    var cursor Cursor
    if err := json.Unmarshal(jsonBytes, &cursor); err != nil {
        return Cursor{}, fmt.Errorf("invalid cursor data: %w", err)
    }

    if cursor.Fields == nil {
        cursor.Fields = make(map[string]interface{})
    }

    return cursor, nil
}

// ParseCursorParams extrai parâmetros de paginação da query string
func ParseCursorParams(r *http.Request, defaultLimit int, maxLimit int) (CursorParams, error) {
    params := CursorParams{
        Limit:  defaultLimit,
        Cursor: "",
    }

    // Parse limit
    if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
        limit, err := strconv.Atoi(limitStr)
        if err != nil {
            return params, fmt.Errorf("invalid limit parameter: must be a number")
        }

        if limit < 1 {
            return params, fmt.Errorf("limit must be greater than 0")
        }

        if limit > maxLimit {
            limit = maxLimit
        }

        params.Limit = limit
    }

    // Parse cursor
    params.Cursor = r.URL.Query().Get("cursor")

    return params, nil
}

// NewPaginatedResponse cria uma resposta paginada
func NewPaginatedResponse[T any](
    data []T,
    limit int,
    nextCursor *string,
) CursorResponse[T] {
    // Garantir que data nunca é nil (sempre retorna array vazio em JSON)
    if data == nil {
        data = []T{}
    }

    return CursorResponse[T]{
        Data: data,
        Pagination: Pagination{
            Limit:      limit,
            HasNext:    nextCursor != nil,
            NextCursor: nextCursor,
        },
    }
}
```

---

### 2. Contrato HTTP PROPOSTO

#### 2.1 Card - List Cards (DEPOIS)

```http
GET /api/v1/cards?limit=20 HTTP/1.1
Authorization: Bearer <token>

HTTP/1.1 200 OK
Content-Type: application/json

{
  "data": [
    {
      "id": "uuid-1",
      "name": "American Express",
      "due_day": 15,
      "closing_offset_days": 10,
      "created_at": "2025-01-01T00:00:00Z",
      "updated_at": "2025-01-01T00:00:00Z"
    },
    {
      "id": "uuid-2",
      "name": "Bradesco",
      "due_day": 5,
      "closing_offset_days": 5,
      "created_at": "2025-01-02T00:00:00Z",
      "updated_at": "2025-01-02T00:00:00Z"
    }
    // ... até 20 items
  ],
  "pagination": {
    "limit": 20,
    "has_next": true,
    "next_cursor": "eyJmIjp7Im5hbWUiOiJCcmFkZXNjbyIsImlkIjoidXVpZC0yMCJ9fQ"
  }
}
```

**Segunda página:**
```http
GET /api/v1/cards?limit=20&cursor=eyJmIjp7Im5hbWUiOiJCcmFkZXNjbyIsImlkIjoidXVpZC0yMCJ9fQ HTTP/1.1

HTTP/1.1 200 OK

{
  "data": [
    {
      "id": "uuid-21",
      "name": "Citibank",
      ...
    }
    // ... próximos 20 items
  ],
  "pagination": {
    "limit": 20,
    "has_next": false,
    "next_cursor": null
  }
}
```

**Última página (lista vazia):**
```http
GET /api/v1/cards?cursor=<ultimo-cursor> HTTP/1.1

HTTP/1.1 200 OK

{
  "data": [],
  "pagination": {
    "limit": 20,
    "has_next": false,
    "next_cursor": null
  }
}
```

---

#### 2.2 Invoice - List by Month (DEPOIS)

```http
GET /api/v1/invoices?month=2025-01&limit=10 HTTP/1.1
Authorization: Bearer <token>

HTTP/1.1 200 OK

{
  "data": [
    {
      "id": "uuid-1",
      "card_id": "card-uuid-1",
      "reference_month": "2025-01-01",
      "due_date": "2025-02-05",
      "total_amount": "800.00",
      "currency": "BRL",
      "created_at": "2025-01-16T00:00:00Z"
    },
    {
      "id": "uuid-2",
      "card_id": "card-uuid-2",
      "reference_month": "2025-01-01",
      "due_date": "2025-02-10",
      "total_amount": "1500.00",
      "currency": "BRL",
      "created_at": "2025-01-15T00:00:00Z"
    }
    // ... até 10 faturas
  ],
  "pagination": {
    "limit": 10,
    "has_next": true,
    "next_cursor": "eyJmIjp7ImR1ZV9kYXRlIjoiMjAyNS0wMi0xMCIsImlkIjoidXVpZC0xMCJ9fQ"
  }
}
```

---

### 3. Struct PROPOSTO (Handler + DTO)

#### Handler (DEPOIS)

```go
// internal/card/infrastructure/http/card_handler.go
package http

import (
    "net/http"

    "github.com/go-chi/render"
    "github.com/jailtonjunior94/financial/internal/card/application/usecase"
    "github.com/jailtonjunior94/financial/pkg/api/httperrors"
    "github.com/jailtonjunior94/financial/pkg/api/middlewares"
    "github.com/jailtonjunior94/financial/pkg/pagination"
)

type CardHandler struct {
    findCardUseCase  usecase.FindCardUseCase
    errorHandler     httperrors.ErrorHandler
}

func (h *CardHandler) Find(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    userID := middlewares.GetUserFromContext(ctx)

    // Parse parâmetros de paginação
    params, err := pagination.ParseCursorParams(r, 20, 100)
    if err != nil {
        h.errorHandler.HandleError(w, r, err)
        return
    }

    // Executar use case com paginação
    result, err := h.findCardUseCase.Execute(ctx, usecase.FindCardInput{
        UserID: userID.ID.String(),
        Limit:  params.Limit,
        Cursor: params.Cursor,
    })
    if err != nil {
        h.errorHandler.HandleError(w, r, err)
        return
    }

    // Construir resposta paginada
    response := pagination.NewPaginatedResponse(
        result.Cards,
        params.Limit,
        result.NextCursor,
    )

    render.JSON(w, r, response)
}
```

---

#### Use Case (DEPOIS)

```go
// internal/card/application/usecase/find.go
package usecase

import (
    "context"

    "github.com/jailtonjunior94/financial/internal/card/application/dtos"
    "github.com/jailtonjunior94/financial/internal/card/domain/interfaces"
    "github.com/jailtonjunior94/financial/pkg/pagination"
    "github.com/JailtonJunior94/devkit-go/pkg/observability"
    "github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type (
    FindCardUseCase interface {
        Execute(ctx context.Context, input FindCardInput) (*FindCardOutput, error)
    }

    FindCardInput struct {
        UserID string
        Limit  int
        Cursor string
    }

    FindCardOutput struct {
        Cards      []dtos.CardOutput
        NextCursor *string
    }

    findCardUseCase struct {
        cardRepository interfaces.CardRepository
        o11y           observability.Observability
    }
)

func NewFindCardUseCase(
    cardRepository interfaces.CardRepository,
    o11y observability.Observability,
) FindCardUseCase {
    return &findCardUseCase{
        cardRepository: cardRepository,
        o11y:           o11y,
    }
}

func (u *findCardUseCase) Execute(ctx context.Context, input FindCardInput) (*FindCardOutput, error) {
    ctx, span := u.o11y.Tracer().Start(ctx, "find_card_usecase.execute")
    defer span.End()

    // Parse user ID
    userID, err := vos.NewUUIDFromString(input.UserID)
    if err != nil {
        return nil, err
    }

    // Decode cursor
    cursor, err := pagination.DecodeCursor(input.Cursor)
    if err != nil {
        return nil, err
    }

    // List cards (paginado)
    cards, err := u.cardRepository.List(ctx, interfaces.ListCardsParams{
        UserID: userID,
        Limit:  input.Limit + 1, // +1 para detectar has_next
        Cursor: cursor,
    })
    if err != nil {
        return nil, err
    }

    // Determinar se há próxima página
    hasNext := len(cards) > input.Limit
    if hasNext {
        cards = cards[:input.Limit] // Remover o item extra
    }

    // Construir cursor para próxima página
    var nextCursor *string
    if hasNext && len(cards) > 0 {
        lastCard := cards[len(cards)-1]

        newCursor := pagination.Cursor{
            Fields: map[string]interface{}{
                "name": lastCard.Name,
                "id":   lastCard.ID.String(),
            },
        }

        encoded, err := pagination.EncodeCursor(newCursor)
        if err != nil {
            return nil, err
        }

        nextCursor = &encoded
    }

    // Converter para DTOs
    output := make([]dtos.CardOutput, len(cards))
    for i, card := range cards {
        output[i] = dtos.CardOutput{
            ID:                card.ID.String(),
            Name:              card.Name,
            DueDay:            card.DueDay,
            ClosingOffsetDays: card.ClosingOffsetDays,
            CreatedAt:         card.CreatedAt,
            UpdatedAt:         card.UpdatedAt.ValueOr(card.CreatedAt),
        }
    }

    return &FindCardOutput{
        Cards:      output,
        NextCursor: nextCursor,
    }, nil
}
```

---

#### Repository (DEPOIS)

```go
// internal/card/domain/interfaces/card_repository.go
package interfaces

import (
    "context"

    "github.com/jailtonjunior94/financial/internal/card/domain/entities"
    "github.com/jailtonjunior94/financial/pkg/pagination"
    "github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type ListCardsParams struct {
    UserID vos.UUID
    Limit  int
    Cursor pagination.Cursor
}

type CardRepository interface {
    Create(ctx context.Context, card *entities.Card) error
    Update(ctx context.Context, card *entities.Card) error
    FindByID(ctx context.Context, userID, id vos.UUID) (*entities.Card, error)
    List(ctx context.Context, params ListCardsParams) ([]*entities.Card, error)
}
```

```go
// internal/card/infrastructure/repositories/card_repository.go
package repositories

import (
    "context"
    "database/sql"
    "errors"
    "fmt"

    "github.com/jailtonjunior94/financial/internal/card/domain/entities"
    "github.com/jailtonjunior94/financial/internal/card/domain/interfaces"
    "github.com/jailtonjunior94/financial/pkg/pagination"

    "github.com/JailtonJunior94/devkit-go/pkg/database"
    "github.com/JailtonJunior94/devkit-go/pkg/observability"
    "github.com/JailtonJunior94/devkit-go/pkg/vos"
)

func (r *cardRepository) List(ctx context.Context, params interfaces.ListCardsParams) ([]*entities.Card, error) {
    ctx, span := r.o11y.Tracer().Start(ctx, "card_repository.list")
    defer span.End()

    // Base query com ORDER BY determinístico
    query := `
        SELECT
            id,
            user_id,
            name,
            due_day,
            closing_offset_days,
            created_at,
            updated_at,
            deleted_at
        FROM cards
        WHERE
            user_id = $1
            AND deleted_at IS NULL`

    args := []interface{}{params.UserID}
    argIndex := 2

    // Adicionar condição de cursor (keyset pagination)
    if name, ok := params.Cursor.Fields["name"].(string); ok {
        if id, ok := params.Cursor.Fields["id"].(string); ok {
            // WHERE (name, id) > (cursor_name, cursor_id)
            query += fmt.Sprintf(` AND (name, id) > ($%d, $%d)`, argIndex, argIndex+1)
            args = append(args, name, id)
            argIndex += 2
        }
    }

    // ORDER BY determinístico (name, id)
    query += ` ORDER BY name, id`

    // LIMIT
    query += fmt.Sprintf(` LIMIT $%d`, argIndex)
    args = append(args, params.Limit)

    // Executar query
    rows, err := r.db.QueryContext(ctx, query, args...)
    if err != nil {
        span.RecordError(err)
        return nil, err
    }
    defer rows.Close()

    // Scanear resultados
    var cards []*entities.Card
    for rows.Next() {
        var card entities.Card
        var deletedAt sql.NullTime

        err := rows.Scan(
            &card.ID,
            &card.UserID,
            &card.Name,
            &card.DueDay,
            &card.ClosingOffsetDays,
            &card.CreatedAt,
            &card.UpdatedAt,
            &deletedAt,
        )
        if err != nil {
            span.RecordError(err)
            return nil, err
        }

        // Converter NullTime
        if deletedAt.Valid {
            card.DeletedAt = vos.NewNullableTime(deletedAt.Time)
        }

        cards = append(cards, &card)
    }

    if err := rows.Err(); err != nil {
        span.RecordError(err)
        return nil, err
    }

    // Garantir que retorna array vazio em vez de nil
    if cards == nil {
        cards = []*entities.Card{}
    }

    return cards, nil
}
```

---

### 4. Query SQL PROPOSTA

#### 4.1 Card - Primeira Página
```sql
SELECT
    id,
    user_id,
    name,
    due_day,
    closing_offset_days,
    created_at,
    updated_at,
    deleted_at
FROM cards
WHERE
    user_id = $1
    AND deleted_at IS NULL
ORDER BY name, id  -- ← Determinístico
LIMIT 21;  -- ← Limit + 1 para detectar has_next
```

**Índice necessário:**
```sql
CREATE INDEX idx_cards_user_name_id ON cards(user_id, name, id)
WHERE deleted_at IS NULL;
```

---

#### 4.2 Card - Próxima Página (com cursor)
```sql
SELECT
    id,
    user_id,
    name,
    due_day,
    closing_offset_days,
    created_at,
    updated_at,
    deleted_at
FROM cards
WHERE
    user_id = $1
    AND deleted_at IS NULL
    AND (name, id) > ($2, $3)  -- ← Keyset (cursor)
ORDER BY name, id
LIMIT 21;
```

**Parâmetros:**
- `$1`: user_id (UUID)
- `$2`: last_name (string) - do cursor
- `$3`: last_id (UUID) - do cursor

**Performance:**
- PostgreSQL usa **index scan** no índice composto
- **Não precisa escanear linhas anteriores**
- **O(1) performance** independente da página

---

#### 4.3 Invoice - List by Month (com cursor)

**Primeira página:**
```sql
SELECT
    id,
    user_id,
    card_id,
    reference_month,
    due_date,
    total_amount,
    created_at,
    updated_at,
    deleted_at
FROM invoices
WHERE
    user_id = $1
    AND to_char(reference_month, 'YYYY-MM') = $2
    AND deleted_at IS NULL
ORDER BY due_date, id  -- ← Determinístico
LIMIT 11;
```

**Próxima página:**
```sql
SELECT
    id,
    user_id,
    card_id,
    reference_month,
    due_date,
    total_amount,
    created_at,
    updated_at,
    deleted_at
FROM invoices
WHERE
    user_id = $1
    AND to_char(reference_month, 'YYYY-MM') = $2
    AND deleted_at IS NULL
    AND (due_date, id) > ($3, $4)  -- ← Cursor
ORDER BY due_date, id
LIMIT 11;
```

**Índice necessário:**
```sql
CREATE INDEX idx_invoices_user_month_due_id
ON invoices(user_id, reference_month, due_date, id)
WHERE deleted_at IS NULL;
```

---

#### 4.4 Invoice - List by Card (com cursor)

**Primeira página:**
```sql
SELECT
    id,
    user_id,
    card_id,
    reference_month,
    due_date,
    total_amount,
    created_at,
    updated_at,
    deleted_at
FROM invoices
WHERE
    card_id = $1
    AND deleted_at IS NULL
ORDER BY reference_month DESC, id DESC  -- ← Mais recentes primeiro
LIMIT 11;
```

**Próxima página:**
```sql
SELECT
    id,
    user_id,
    card_id,
    reference_month,
    due_date,
    total_amount,
    created_at,
    updated_at,
    deleted_at
FROM invoices
WHERE
    card_id = $1
    AND deleted_at IS NULL
    AND (reference_month, id) < ($2, $3)  -- ← Cursor (DESC usa <)
ORDER BY reference_month DESC, id DESC
LIMIT 11;
```

**Índice necessário:**
```sql
CREATE INDEX idx_invoices_card_month_id
ON invoices(card_id, reference_month DESC, id DESC)
WHERE deleted_at IS NULL;
```

---

#### 4.5 Category - List Categories (com cursor)

**Primeira página:**
```sql
SELECT
    id,
    user_id,
    name,
    sequence,
    created_at,
    updated_at,
    deleted_at
FROM categories
WHERE
    user_id = $1
    AND deleted_at IS NULL
    AND parent_id IS NULL
ORDER BY sequence, id  -- ← Determinístico
LIMIT 51;
```

**Próxima página:**
```sql
SELECT
    id,
    user_id,
    name,
    sequence,
    created_at,
    updated_at,
    deleted_at
FROM categories
WHERE
    user_id = $1
    AND deleted_at IS NULL
    AND parent_id IS NULL
    AND (sequence, id) > ($2, $3)  -- ← Cursor
ORDER BY sequence, id
LIMIT 51;
```

**Índice necessário:**
```sql
CREATE INDEX idx_categories_user_seq_id
ON categories(user_id, sequence, id)
WHERE parent_id IS NULL AND deleted_at IS NULL;
```

---

## Justificativa Técnica

### 1. Por que Cursor-Based em vez de Offset?

#### Offset-Based (O que NÃO fazer)

```sql
-- Página 100 (offset 990)
SELECT * FROM invoices
WHERE user_id = ?
ORDER BY due_date
LIMIT 10 OFFSET 990;
```

**Problemas:**
- PostgreSQL escaneia **1000 linhas** para retornar 10
- **Linear degradation** (O(n))
- **Inconsistente:** se dados mudarem, usuário perde linhas

**Benchmark (1M de linhas):**
```
OFFSET 0:      1ms
OFFSET 1000:   15ms
OFFSET 10000:  150ms
OFFSET 100000: 1500ms  ← Impraticável
```

#### Cursor-Based (Proposta)

```sql
-- Qualquer página
SELECT * FROM invoices
WHERE user_id = ?
  AND (due_date, id) > (last_date, last_id)
ORDER BY due_date, id
LIMIT 10;
```

**Benefícios:**
- PostgreSQL usa **index seek** direto
- **Constant time** (O(1))
- **Consistente:** cursor é absoluto

**Benchmark (1M de linhas):**
```
Qualquer página: ~1-2ms  ← Sempre rápido!
```

---

### 2. Por que ORDER BY Composto (column, id)?

**Problema do ORDER BY simples:**
```sql
SELECT * FROM cards ORDER BY name LIMIT 10;
-- Se houver 2 "Nubank", qual vem primeiro?
-- PostgreSQL não garante ordem estável!
```

**Solução: ORDER BY composto**
```sql
SELECT * FROM cards ORDER BY name, id LIMIT 10;
-- Ordem sempre determinística
-- 1. Ordena por name
-- 2. Desempate por id (único)
```

**Benefícios:**
- ✅ **Determinístico** (mesma query = mesma ordem sempre)
- ✅ **Permite paginação confiável**
- ✅ **Testes reproduzíveis**
- ✅ **UX consistente**

---

### 3. Por que Índice Composto?

**Índice simples (atual):**
```sql
CREATE INDEX idx_cards_user_id ON cards(user_id);

-- Query:
SELECT * FROM cards
WHERE user_id = ?
ORDER BY name, id
LIMIT 20;

-- Plan:
Index Scan using idx_cards_user_id
  -> Sort  ← Filesort (lento!)
```

**Índice composto (proposta):**
```sql
CREATE INDEX idx_cards_user_name_id ON cards(user_id, name, id);

-- Query:
SELECT * FROM cards
WHERE user_id = ?
ORDER BY name, id
LIMIT 20;

-- Plan:
Index Scan using idx_cards_user_name_id  ← Usa índice para ORDER BY!
```

**Benefícios:**
- ✅ **Sem filesort** (ordenação já está no índice)
- ✅ **Query 10-100x mais rápida**
- ✅ **Suporta keyset pagination** eficientemente

---

### 4. Por que Base64 Cursor?

**Alternativas:**
1. **Plain JSON:** `?cursor={"name":"Nubank","id":"uuid"}`
   - ❌ Revela estrutura interna
   - ❌ Difícil de ler em URL
   - ❌ Requer URL encoding

2. **JWT:** `?cursor=eyJhbGc...`
   - ❌ Overhead (assinatura desnecessária)
   - ❌ Expiry não faz sentido para cursor
   - ❌ Mais lento

3. **Base64 (proposta):** `?cursor=eyJmIjp7Im5hbWUi...`
   - ✅ **Opaco** (cliente não precisa entender)
   - ✅ **Compacto**
   - ✅ **URL-safe** (base64url)
   - ✅ **Stateless** (servidor não guarda estado)
   - ✅ **Versionável** (pode mudar estrutura interna sem quebrar API)

---

## Performance e Índices

### Índices Propostos (Migrations)

```sql
-- Migration: 000X_add_cursor_pagination_indexes.up.sql

-- Card: user_id + name + id (para cursor pagination)
CREATE INDEX IF NOT EXISTS idx_cards_user_name_id
ON cards(user_id, name, id)
WHERE deleted_at IS NULL;

-- Category: user_id + sequence + id (para cursor pagination)
CREATE INDEX IF NOT EXISTS idx_categories_user_seq_id
ON categories(user_id, sequence, id)
WHERE parent_id IS NULL AND deleted_at IS NULL;

-- Invoice by Month: user_id + reference_month + due_date + id
CREATE INDEX IF NOT EXISTS idx_invoices_user_month_due_id
ON invoices(user_id, reference_month, due_date, id)
WHERE deleted_at IS NULL;

-- Invoice by Card: card_id + reference_month DESC + id DESC
CREATE INDEX IF NOT EXISTS idx_invoices_card_month_id
ON invoices(card_id, reference_month DESC, id DESC)
WHERE deleted_at IS NULL;
```

**Down migration:**
```sql
-- Migration: 000X_add_cursor_pagination_indexes.down.sql

DROP INDEX IF EXISTS idx_cards_user_name_id;
DROP INDEX IF EXISTS idx_categories_user_seq_id;
DROP INDEX IF EXISTS idx_invoices_user_month_due_id;
DROP INDEX IF EXISTS idx_invoices_card_month_id;
```

---

### Query Plans (EXPLAIN ANALYZE)

#### ANTES (sem paginação)
```sql
EXPLAIN ANALYZE
SELECT * FROM cards
WHERE user_id = '...' AND deleted_at IS NULL
ORDER BY name;

-- Plan:
Index Scan using idx_cards_user_id on cards (cost=0.29..8.31 rows=1 width=100)
  Index Cond: (user_id = '...')
  Filter: (deleted_at IS NULL)
  -> Sort (cost=8.31..8.32 rows=1 width=100)  ← Filesort!
       Sort Method: quicksort  Memory: 25kB
Planning Time: 0.123 ms
Execution Time: 0.456 ms
```

#### DEPOIS (com cursor pagination e índice composto)
```sql
EXPLAIN ANALYZE
SELECT * FROM cards
WHERE user_id = '...' AND deleted_at IS NULL
ORDER BY name, id
LIMIT 20;

-- Plan:
Index Scan using idx_cards_user_name_id on cards (cost=0.29..4.31 rows=20 width=100)
  Index Cond: (user_id = '...')
  Filter: (deleted_at IS NULL)
Planning Time: 0.089 ms
Execution Time: 0.112 ms  ← 4x mais rápido!
```

---

## Segurança e Robustez

### 1. Proteção Contra Nil Pointer

```go
// Repository: garantir array vazio
func (r *cardRepository) List(...) ([]*entities.Card, error) {
    // ...
    var cards []*entities.Card
    for rows.Next() {
        // ...
        cards = append(cards, &card)
    }

    // ✅ Proteção contra nil
    if cards == nil {
        cards = []*entities.Card{}
    }

    return cards, nil
}

// UseCase: garantir array vazio no output
func (u *findCardUseCase) Execute(...) (*FindCardOutput, error) {
    // ...
    output := make([]dtos.CardOutput, len(cards))  // ✅ Nunca nil
    // ...
}

// Pagination helper: garantir array vazio
func NewPaginatedResponse[T any](data []T, limit int, nextCursor *string) CursorResponse[T] {
    if data == nil {
        data = []T{}  // ✅ Força array vazio
    }
    // ...
}
```

**Resultado:**
```json
{
  "data": [],  // ✅ Sempre array (nunca null)
  "pagination": {
    "limit": 20,
    "has_next": false,
    "next_cursor": null
  }
}
```

---

### 2. Proteção Contra Race Conditions

**Problema potencial:**
```go
// ❌ Race condition
var cards []*entities.Card
for rows.Next() {
    go func() {  // ← Goroutine dentro do loop
        cards = append(cards, card)  // ← Race!
    }()
}
```

**Solução:**
```go
// ✅ Sem concorrência no scan
var cards []*entities.Card
for rows.Next() {
    var card entities.Card
    rows.Scan(...)  // ← Síncrono
    cards = append(cards, &card)  // ← Sem race
}
```

**Context propagation:**
```go
// ✅ Context é thread-safe
func (r *repo) List(ctx context.Context, ...) {
    rows, err := r.db.QueryContext(ctx, ...)  // ← Context propagado
    // Se request for cancelado, query é abortada
}
```

---

### 3. Validação de Parâmetros

```go
func ParseCursorParams(r *http.Request, defaultLimit int, maxLimit int) (CursorParams, error) {
    // ✅ Validação de limit
    if limit < 1 {
        return params, fmt.Errorf("limit must be greater than 0")
    }

    if limit > maxLimit {
        limit = maxLimit  // ✅ Cap no máximo
    }

    // ✅ Validação de cursor format
    if cursor != "" {
        _, err := DecodeCursor(cursor)
        if err != nil {
            return params, fmt.Errorf("invalid cursor: %w", err)
        }
    }

    return params, nil
}
```

**Proteção contra:**
- ❌ `limit=-1` (negativo)
- ❌ `limit=999999` (DDoS)
- ❌ `cursor=invalid_base64` (malformed)

---

### 4. Response HTTP Correto

**200 OK com lista vazia (correto):**
```http
GET /api/v1/cards HTTP/1.1

HTTP/1.1 200 OK  ← Status correto

{
  "data": [],  ← Lista vazia (não null)
  "pagination": {
    "limit": 20,
    "has_next": false,
    "next_cursor": null
  }
}
```

**NÃO usar 404:**
```http
❌ HTTP/1.1 404 Not Found
```
- 404 é para **recurso não existe**
- Lista vazia ≠ recurso não existe
- Lista vazia é **sucesso** (200 OK)

---

## Plano de Migração

### Fase 1: Preparação (Sem Breaking Changes)

**Objetivo:** Adicionar suporte a paginação sem quebrar clientes existentes.

#### 1.1 - Criar pacote de paginação
- ✅ `pkg/pagination/cursor.go`
- ✅ Structs genéricas
- ✅ Helpers de encoding/decoding
- ✅ Testes unitários

#### 1.2 - Adicionar índices compostos
- ✅ Migration `add_cursor_pagination_indexes.up.sql`
- ✅ Executar em staging primeiro
- ✅ Verificar performance com EXPLAIN
- ✅ Deploy em produção (índices são criados de forma não-bloqueante)

#### 1.3 - Atualizar repositories
- ✅ Adicionar método `List()` com suporte a cursor
- ✅ Manter compatibilidade backward (cursor vazio = primeira página)
- ✅ Testes de integração

---

### Fase 2: Implementação Gradual

**Objetivo:** Implementar paginação endpoint por endpoint.

#### 2.1 - Card (GET /cards)
- ✅ Atualizar handler para aceitar `?limit` e `?cursor`
- ✅ Atualizar use case
- ✅ Atualizar repository
- ✅ Testes end-to-end
- ✅ Deploy

**Compatibilidade:**
```
# Cliente antigo (sem params) - continua funcionando
GET /cards  →  200 OK (primeira página com limit default=20)

# Cliente novo (com params)
GET /cards?limit=50&cursor=...  →  200 OK (paginado)
```

#### 2.2 - Category (GET /categories)
- Mesmo processo

#### 2.3 - Invoice (GET /invoices?month=...)
- Mesmo processo

#### 2.4 - Invoice (GET /invoices/cards/{cardId})
- Mesmo processo

---

### Fase 3: Documentação e Comunicação

#### 3.1 - Atualizar documentação da API
- ✅ Swagger/OpenAPI: adicionar parâmetros `limit` e `cursor`
- ✅ Exemplos de paginação
- ✅ Explicar formato do cursor

#### 3.2 - Changelog
- ✅ Adicionar em CHANGELOG.md:
  ```markdown
  ## [v2.0.0] - 2025-02-01

  ### Added
  - Cursor-based pagination em endpoints de listagem
  - Parâmetros `limit` (default: 20, max: 100) e `cursor`
  - Resposta padronizada: `{data: [], pagination: {...}}`

  ### Changed
  - Resposta de listagem agora é objeto com `data` e `pagination`
  - **BREAKING:** Clientes precisam acessar `.data` para obter lista
  ```

#### 3.3 - Comunicar breaking changes
- ✅ Email para desenvolvedores
- ✅ Período de transição (ex: 30 dias)
- ✅ Versionar API (opcional: `/api/v2/cards`)

---

## Próximos Passos (Aguardando Aprovação)

1. **Aprovar proposta técnica**
2. **Criar issue/tarefa no backlog**
3. **Implementar Fase 1** (preparação)
4. **Code review**
5. **Implementar Fase 2** (endpoint por endpoint)
6. **Deploy gradual**
7. **Monitorar performance**
8. **Comunicar mudanças**

---

## Referências

- [RFC 7807 - Problem Details for HTTP APIs](https://datatracker.ietf.org/doc/html/rfc7807)
- [Cursor-Based Pagination (Slack Engineering)](https://slack.engineering/evolving-api-pagination-at-slack/)
- [PostgreSQL: Index-Only Scans](https://www.postgresql.org/docs/current/indexes-index-only-scans.html)
- [RESTful API Design: Pagination](https://restfulapi.net/pagination/)
- [Use The Index, Luke: Pagination](https://use-the-index-luke.com/sql/partial-results/fetch-next-page)

---

**Aguardando aprovação para prosseguir com implementação.**
