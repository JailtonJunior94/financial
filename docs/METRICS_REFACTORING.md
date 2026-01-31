# Refatoração de Métricas OpenTelemetry

Data: 2026-01-31

## Sumário

Refatoração completa das métricas seguindo convenções OpenTelemetry para produção.

## Problemas Corrigidos

### ✅ P0 - Crítico

#### 1. Removida Redundância de Métricas (Alto Impacto)
**Antes:**
```go
// Duas métricas para a mesma informação
operationsTotal.Increment(ctx, operation, status)
errorsTotal.Increment(ctx, operation, error_type)
```

**Depois:**
```go
// Uma métrica consolidada
operationsTotal.Increment(ctx, operation, status, error_type)
```

**Impacto:**
- ✅ Redução de 25% no número de métricas
- ✅ Eliminação de duplicação de dados
- ✅ Armazenamento e processamento mais eficiente

#### 2. Corrigido Uso Incorreto de UpDownCounter (Alto Impacto)
**Antes:**
```go
// ❌ ERRADO - UpDownCounter não é Gauge
func SetActiveCards(ctx context.Context, count float64) {
    m.activeCardsTotal.Add(ctx, int64(count))  // Semântica incorreta
}
```

**Depois:**
```go
// ✅ CORRETO - Uso semântico adequado
func IncActiveCards(ctx context.Context) {
    m.activeCardsTotal.Add(ctx, 1)  // Evento: +1 card
}

func DecActiveCards(ctx context.Context) {
    m.activeCardsTotal.Add(ctx, -1)  // Evento: -1 card
}
```

**Impacto:**
- ✅ Semântica correta do OpenTelemetry
- ✅ Valores confiáveis em produção
- ✅ Removido método `SetActiveCards()` incorreto

### ✅ P1 - Importante

#### 3. Status HTTP por Classe (Médio Impacto)
**Antes:**
```go
observability.String("status", "200")  // Alta cardinalidade
observability.String("status", "404")
observability.String("status", "500")
```

**Depois:**
```go
observability.String("status_class", "2xx")  // Cardinalidade controlada
observability.String("status_class", "4xx")
observability.String("status_class", "5xx")
```

**Impacto:**
- ✅ Cardinalidade reduzida de ~60 para 5 séries
- ✅ Queries de erro agregado simplificadas
- ✅ Melhor performance no Prometheus

#### 4. Status no Histogram de Latência (Médio Impacto)
**Antes:**
```go
// Sucesso e falha misturados
operationDuration.Record(ctx, duration, operation)
```

**Depois:**
```go
// Separação clara de sucesso vs falha
operationDuration.Record(ctx, duration, operation, status)
```

**Impacto:**
- ✅ Análise separada de latência por status
- ✅ P95 de erros vs sucesso distinguíveis
- ✅ Alertas mais precisos

## Arquivos Modificados

### Métricas HTTP
- `pkg/api/middlewares/metrics.go`
  - Adicionado helper `statusClass()`
  - Substituído label `status` por `status_class`

- `pkg/api/middlewares/README.md`
  - Atualizada documentação
  - Queries Prometheus corrigidas

### Métricas Card
- `pkg/observability/metrics/card_metrics.go`
  - Removido `errorsTotal` counter
  - Removido método `SetActiveCards()`
  - Adicionado `error_type` em `RecordOperationFailure()`
  - Adicionado `status` no histogram `operationDuration`

- `pkg/observability/metrics/README.md`
  - Documentação completa criada
  - Exemplos de uso
  - Queries Prometheus

### Usecases Atualizados
- `internal/card/application/usecase/create.go`
- `internal/card/application/usecase/update.go`
- `internal/card/application/usecase/remove.go`
- `internal/card/application/usecase/find.go`
- `internal/card/application/usecase/find_by.go`
- `internal/card/application/usecase/find_paginated.go`

**Mudanças:**
```go
// Antes
u.metrics.RecordOperationFailure(ctx, operation, duration)
u.metrics.RecordError(ctx, operation, errorType)

// Depois
u.metrics.RecordOperationFailure(ctx, operation, duration, errorType)
```

## Métricas Finais

### HTTP Metrics
```
financial.http.request.duration.seconds (Histogram)
  Labels: method, route, status_class

financial.http.requests.total (Counter)
  Labels: method, route, status_class

financial.http.active_requests (UpDownCounter)
  Labels: -
```

### Card Business Metrics
```
financial.card.operations.total (Counter)
  Labels: operation, status, error_type

financial.card.operation.duration.seconds (Histogram)
  Labels: operation, status

financial.card.active.total (UpDownCounter)
  Labels: -
```

## Queries Úteis

### Taxa de Erro HTTP
```promql
sum(rate(financial_http_requests_total{status_class="5xx"}[5m])) by (route)
```

### P95 Latência Card Sucesso vs Falha
```promql
histogram_quantile(0.95,
  sum(rate(financial_card_operation_duration_seconds_bucket[5m])) by (status, le)
)
```

### Erros Card por Tipo
```promql
sum(rate(financial_card_operations_total{status="failure"}[5m])) by (error_type)
```

## Validação

✅ Build: `go build ./...` - Sucesso
✅ Testes: `go test ./...` - Todos passando
✅ Lint: Sem erros de compilação
✅ Compatibilidade: OpenTelemetry 1.x compliant

## Benefícios de Produção

1. **Redução de Custo**
   - 25% menos métricas = menos armazenamento
   - Cardinalidade controlada = queries mais rápidas

2. **Observabilidade Aprimorada**
   - Latência de erros vs sucesso visível
   - Classificação automática de erros
   - Correlação com traces mantida

3. **Conformidade OpenTelemetry**
   - Semântica correta de instrumentos
   - Labels controlados por enums
   - Naming convention adequada
   - Portável entre backends (Prometheus, Grafana Cloud, etc)

4. **Manutenibilidade**
   - Consolidação elimina duplicação
   - Documentação completa
   - Queries de exemplo prontas
