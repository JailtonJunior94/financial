# Métricas Prometheus - Módulo Card

Sistema de métricas para observabilidade do módulo de cartões usando Prometheus.

## Métricas Disponíveis

### 1. Contador de Operações (`financial_card_operations_total`)

**Tipo**: Counter

**Descrição**: Total de operações executadas no módulo de cartões

**Labels**:
- `operation`: Tipo da operação (create, update, delete, find, find_by)
- `status`: Status da operação (success, failure)

**Uso**: Rastrear o volume de operações e taxa de sucesso/falha

**Queries PromQL úteis**:
```promql
# Taxa de sucesso por operação (últimos 5 minutos)
rate(financial_card_operations_total{status="success"}[5m])

# Taxa de erro geral
rate(financial_card_operations_total{status="failure"}[5m])

# Percentual de erro por operação
sum(rate(financial_card_operations_total{status="failure"}[5m])) by (operation)
/
sum(rate(financial_card_operations_total[5m])) by (operation) * 100
```

---

### 2. Contador de Erros (`financial_card_errors_total`)

**Tipo**: Counter

**Descrição**: Total de erros detalhados por operação e tipo

**Labels**:
- `operation`: Tipo da operação (create, update, delete, find, find_by)
- `error_type`: Tipo do erro (validation, not_found, repository, parsing, unknown)

**Uso**: Identificar tipos específicos de erros e onde ocorrem

**Queries PromQL úteis**:
```promql
# Top 5 tipos de erro
topk(5, sum(rate(financial_card_errors_total[5m])) by (error_type))

# Erros de validação por operação
rate(financial_card_errors_total{error_type="validation"}[5m])

# Distribuição de erros (últimas 24h)
sum(increase(financial_card_errors_total[24h])) by (error_type, operation)
```

---

### 3. Histograma de Latência (`financial_card_operation_duration_seconds`)

**Tipo**: Histogram

**Descrição**: Duração das operações em segundos

**Labels**:
- `operation`: Tipo da operação (create, update, delete, find, find_by)

**Buckets**: [.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5] segundos

**Uso**: Monitorar performance e identificar operações lentas

**Queries PromQL úteis**:
```promql
# P50, P95, P99 de latência por operação
histogram_quantile(0.50, sum(rate(financial_card_operation_duration_seconds_bucket[5m])) by (le, operation))
histogram_quantile(0.95, sum(rate(financial_card_operation_duration_seconds_bucket[5m])) by (le, operation))
histogram_quantile(0.99, sum(rate(financial_card_operation_duration_seconds_bucket[5m])) by (le, operation))

# Latência média por operação
rate(financial_card_operation_duration_seconds_sum[5m]) / rate(financial_card_operation_duration_seconds_count[5m])

# Operações lentas (> 1s)
sum(rate(financial_card_operation_duration_seconds_bucket{le="1"}[5m])) by (operation)
```

---

### 4. Gauge de Cartões Ativos (`financial_card_active_total`)

**Tipo**: Gauge

**Descrição**: Número atual de cartões ativos no sistema

**Labels**: Nenhum (métrica agregada)

**Uso**: Monitorar crescimento da base de cartões

**Queries PromQL úteis**:
```promql
# Número atual de cartões ativos
financial_card_active_total

# Crescimento de cartões (últimas 24h)
financial_card_active_total - financial_card_active_total offset 24h

# Taxa de crescimento diário
deriv(financial_card_active_total[24h])
```

---

## Pontos de Instrumentação

### Create Card
- **Início**: Ao iniciar execução do use case
- **Sucesso**: Após salvar no repositório com sucesso
- **Falha**: Em qualquer erro (validação, parsing, repository)
- **Side-effect**: Incrementa `active_cards_total`

### Update Card
- **Início**: Ao iniciar execução do use case
- **Sucesso**: Após atualizar no repositório com sucesso
- **Falha**: Em qualquer erro (not_found, validation, repository)

### Delete Card
- **Início**: Ao iniciar execução do use case
- **Sucesso**: Após soft delete no repositório com sucesso
- **Falha**: Em qualquer erro (not_found, repository)
- **Side-effect**: Decrementa `active_cards_total`

### Find Cards (List)
- **Início**: Ao iniciar execução do use case
- **Sucesso**: Após listar cards do repositório com sucesso
- **Falha**: Em erros de parsing ou repository

### Find Card By ID
- **Início**: Ao iniciar execução do use case
- **Sucesso**: Após buscar card no repositório com sucesso
- **Falha**: Em erros de parsing, not_found ou repository

---

## Integração com Grafana

### Dashboard Recomendado

**Painéis sugeridos**:

1. **Taxa de Operações** (Graph)
   - Query: `sum(rate(financial_card_operations_total[5m])) by (operation)`
   - Agrupado por operação

2. **Taxa de Erro** (Graph)
   - Query: Taxa de erro percentual por operação
   - Com threshold de alerta em 5%

3. **Latência P95/P99** (Graph)
   - Query: Percentis 95 e 99 por operação
   - Com threshold de SLA

4. **Cartões Ativos** (Stat + Graph)
   - Query: `financial_card_active_total`
   - Mostra valor atual + histórico

5. **Top Erros** (Bar Gauge)
   - Query: `topk(5, sum(rate(financial_card_errors_total[5m])) by (error_type))`

6. **Distribuição de Latência** (Heatmap)
   - Query: Buckets do histograma
   - Visualização de distribuição temporal

---

## Alertas Prometheus

### Exemplos de Regras de Alerta

```yaml
groups:
  - name: card_alerts
    interval: 30s
    rules:
      # Alta taxa de erro
      - alert: CardHighErrorRate
        expr: |
          sum(rate(financial_card_errors_total[5m]))
          /
          sum(rate(financial_card_operations_total[5m])) > 0.05
        for: 5m
        labels:
          severity: warning
          component: card
        annotations:
          summary: "Alta taxa de erro no módulo de cartões"
          description: "Taxa de erro acima de 5% nos últimos 5 minutos"

      # Latência alta
      - alert: CardHighLatency
        expr: |
          histogram_quantile(0.95,
            sum(rate(financial_card_operation_duration_seconds_bucket[5m])) by (le, operation)
          ) > 1
        for: 5m
        labels:
          severity: warning
          component: card
        annotations:
          summary: "Alta latência nas operações de cartão"
          description: "P95 de latência acima de 1 segundo"

      # Spike de erros de validação
      - alert: CardValidationErrorSpike
        expr: |
          rate(financial_card_errors_total{error_type="validation"}[5m]) > 10
        for: 2m
        labels:
          severity: info
          component: card
        annotations:
          summary: "Spike de erros de validação"
          description: "Aumento anormal de erros de validação"

      # Crescimento anormal de cartões
      - alert: CardAbnormalGrowth
        expr: |
          deriv(financial_card_active_total[1h]) > 1000
        for: 5m
        labels:
          severity: info
          component: card
        annotations:
          summary: "Crescimento anormal de cartões"
          description: "Taxa de crescimento acima de 1000 cartões/hora"
```

---

## Boas Práticas

### ✅ DO
- Usar labels com cardinalidade limitada (operation, error_type)
- Registrar métricas em todos os caminhos de código (sucesso + erro)
- Medir duração total da operação (incluindo chamadas ao repository)
- Incrementar/decrementar gauges atomicamente

### ❌ DON'T
- Nunca use IDs de usuário ou cartão como labels (alta cardinalidade)
- Não crie métricas por usuário individual
- Evite labels dinâmicos ou com valores ilimitados
- Não meça apenas casos de sucesso

---

## Classificação de Erros

O sistema classifica erros automaticamente:

| Tipo | Descrição | Exemplos |
|------|-----------|----------|
| `validation` | Erros de validação de entrada | "invalid due day", "name required" |
| `not_found` | Recurso não encontrado | "card not found" |
| `parsing` | Erros de parsing/conversão | UUID inválido, JSON malformado |
| `repository` | Erros de banco de dados | Conexão, SQL errors, constraints |
| `unknown` | Erros não classificados | Panics, erros inesperados |

---

## Exemplo de Uso Completo

```go
// Em um use case
func (u *createCardUseCase) Execute(ctx context.Context, userID string, input *dtos.CardInput) (*dtos.CardOutput, error) {
    start := time.Now()

    // Proteção contra panics
    defer func() {
        duration := time.Since(start)
        if err := recover(); err != nil {
            u.metrics.RecordOperationFailure(metrics.OperationCreate, duration)
            panic(err)
        }
    }()

    // Lógica de negócio
    card, err := factories.CreateCard(userID, input.Name, input.DueDay)
    if err != nil {
        duration := time.Since(start)
        u.metrics.RecordOperationFailure(metrics.OperationCreate, duration)
        u.metrics.RecordError(metrics.OperationCreate, metrics.ClassifyError(err))
        return nil, err
    }

    if err := u.repository.Save(ctx, card); err != nil {
        duration := time.Since(start)
        u.metrics.RecordOperationFailure(metrics.OperationCreate, duration)
        u.metrics.RecordError(metrics.OperationCreate, metrics.ClassifyError(err))
        return nil, err
    }

    // Sucesso
    duration := time.Since(start)
    u.metrics.RecordOperation(metrics.OperationCreate, duration)
    u.metrics.IncActiveCards()

    return output, nil
}
```

---

## Roadmap

### Próximas Melhorias
- [ ] Adicionar métricas de cache (se implementado)
- [ ] Métricas de eventos de domínio (se houver)
- [ ] SLI/SLO tracking automático
- [ ] Integração com Alertmanager
- [ ] Dashboard Grafana completo exportável

---

## Referências

- [Prometheus Best Practices](https://prometheus.io/docs/practices/naming/)
- [Metric and Label Naming](https://prometheus.io/docs/practices/naming/)
- [Instrumentation](https://prometheus.io/docs/practices/instrumentation/)
