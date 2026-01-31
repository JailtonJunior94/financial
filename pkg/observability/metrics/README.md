# Card Business Metrics

Métricas de negócio para operações do módulo de cartões seguindo convenções OpenTelemetry.

## Métricas Expostas

### `financial.card.operations.total`
- **Tipo:** Counter
- **Descrição:** Total de operações de cartão executadas
- **Labels:**
  - `operation`: Tipo de operação (`create`, `update`, `delete`, `find`, `find_by`)
  - `status`: Status da operação (`success`, `failure`)
  - `error_type`: Tipo de erro quando `status=failure` (`validation`, `not_found`, `repository`, `parsing`, `unknown`)

**Exemplo de uso:**
```go
// Sucesso
metrics.RecordOperation(ctx, metrics.OperationCreate, duration)

// Falha
metrics.RecordOperationFailure(ctx, metrics.OperationCreate, duration, metrics.ClassifyError(err))
```

### `financial.card.operation.duration.seconds`
- **Tipo:** Histogram
- **Descrição:** Latência de operações de cartão em segundos
- **Labels:**
  - `operation`: Tipo de operação
  - `status`: Status da operação (`success`, `failure`)
- **Buckets:** `.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5`

**Nota:** Inclui `status` para permitir análise separada de latência de sucesso vs falha.

### `financial.card.active.total`
- **Tipo:** UpDownCounter
- **Descrição:** Contador de cartões ativos (incrementado em create, decrementado em delete)
- **Labels:** Nenhum

**Uso correto:**
```go
// Em CreateCard
metrics.IncActiveCards(ctx)  // +1

// Em DeleteCard
metrics.DecActiveCards(ctx)  // -1
```

**❌ Não fazer:**
```go
metrics.SetActiveCards(ctx, 100)  // Removido - semântica incorreta
```

## Constantes

### Tipos de Operação
```go
OperationCreate = "create"
OperationUpdate = "update"
OperationDelete = "delete"
OperationFind   = "find"
OperationFindBy = "find_by"
```

### Tipos de Erro
```go
ErrorTypeValidation = "validation"
ErrorTypeNotFound   = "not_found"
ErrorTypeRepository = "repository"
ErrorTypeUnknown    = "unknown"
ErrorTypeParsing    = "parsing"
```

## Queries Prometheus

### Taxa de erros por tipo
```promql
sum(rate(financial_card_operations_total{status="failure"}[5m])) by (error_type)
```

### P95 latência de operações bem-sucedidas
```promql
histogram_quantile(0.95,
  sum(rate(financial_card_operation_duration_seconds_bucket{status="success"}[5m])) by (operation, le)
)
```

### Comparar latência sucesso vs falha
```promql
histogram_quantile(0.95,
  sum(rate(financial_card_operation_duration_seconds_bucket[5m])) by (status, le)
)
```

### Cartões ativos atual
```promql
financial_card_active_total
```

### Taxa de criação de cartões
```promql
rate(financial_card_operations_total{operation="create", status="success"}[5m])
```

## Mudanças vs Versão Anterior

**Removido:**
- ❌ `financial.card.errors.total` - Redundante, consolidado em `operations.total` com label `error_type`
- ❌ `SetActiveCards()` - Uso incorreto de UpDownCounter como Gauge

**Adicionado:**
- ✅ Label `error_type` em `operations.total` (apenas quando `status=failure`)
- ✅ Label `status` em `operation.duration` (permite análise separada de latência)
- ✅ Parâmetro `errorType` em `RecordOperationFailure()`

**Benefícios:**
- Redução de 25% no número de métricas
- Cardinalidade controlada
- Correlação direta entre erros e operações
- Análise de latência por status (sucesso vs falha)
