# Resumo: Métricas Prometheus - Módulo Card

## Objetivo

Implementar observabilidade completa do módulo de cartões usando Prometheus, cobrindo eventos de negócio, erros, latência e estado geral.

---

## Arquivos Criados

### 1. `pkg/observability/metrics/card_metrics.go`
Define e registra todas as métricas do módulo de cartões:

**Métricas**:
- `financial_card_operations_total` (Counter) - Total de operações por tipo e status
- `financial_card_errors_total` (Counter) - Erros detalhados por tipo
- `financial_card_operation_duration_seconds` (Histogram) - Latência das operações
- `financial_card_active_total` (Gauge) - Número de cartões ativos

**Constantes**:
- Operações: `create`, `update`, `delete`, `find`, `find_by`
- Tipos de erro: `validation`, `not_found`, `repository`, `parsing`, `unknown`

### 2. `pkg/observability/metrics/error_classifier.go`
Classifica erros automaticamente para as métricas baseado em palavras-chave no erro.

### 3. `pkg/observability/metrics/test_helpers.go`
Helper para criar métricas isoladas em testes unitários.

### 4. `pkg/observability/metrics/README.md`
Documentação completa com queries PromQL, dashboards Grafana e alertas.

---

## Pontos de Instrumentação

### Use Cases Instrumentados

Todos os 5 use cases foram instrumentados:

1. **CreateCardUseCase** (`create.go`)
   - Registra duração total
   - Incrementa `active_cards_total` no sucesso
   - Registra erros de validação/repositório

2. **UpdateCardUseCase** (`update.go`)
   - Registra duração total
   - Registra erros de not_found, validação e repositório

3. **RemoveCardUseCase** (`remove.go`)
   - Registra duração total
   - Decrementa `active_cards_total` no sucesso
   - Registra erros de not_found e repositório

4. **FindCardUseCase** (`find.go`)
   - Registra duração de listagem
   - Registra erros de parsing e repositório

5. **FindCardByUseCase** (`find_by.go`)
   - Registra duração de busca
   - Registra erros de not_found, parsing e repositório

### Padrão de Instrumentação

Cada use case segue este padrão:

```go
func (u *useCase) Execute(ctx context.Context, ...) (..., error) {
    start := time.Now()

    // Proteção contra panics
    defer func() {
        duration := time.Since(start)
        if err := recover(); err != nil {
            u.metrics.RecordOperationFailure(operation, duration)
            panic(err)
        }
    }()

    // Lógica de negócio...
    if err != nil {
        duration := time.Since(start)
        u.metrics.RecordOperationFailure(operation, duration)
        u.metrics.RecordError(operation, metrics.ClassifyError(err))
        return nil, err
    }

    // Sucesso
    duration := time.Since(start)
    u.metrics.RecordOperation(operation, duration)
    // Side-effects: IncActiveCards() ou DecActiveCards()

    return result, nil
}
```

---

## Integração no Módulo

**Arquivo**: `internal/card/module.go`

```go
func NewCardModule(db *sql.DB, o11y observability.Observability, tokenValidator auth.TokenValidator) CardModule {
    // Inicializa métricas
    cardMetrics := metrics.NewCardMetrics(prometheus.DefaultRegisterer)

    // Injeta nas use cases
    createCardUsecase := usecase.NewCreateCardUseCase(o11y, cardRepository, cardMetrics)
    // ... demais use cases
}
```

---

## Métricas em Ação

### Evento → Métrica → Uso

| Evento | Métricas Geradas | Uso no Monitoramento |
|--------|------------------|----------------------|
| Cartão criado com sucesso | `operations_total{operation="create",status="success"}++`<br>`operation_duration_seconds{operation="create"}` observado<br>`active_total++` | Taxa de criação, latência P95, crescimento da base |
| Erro de validação no create | `operations_total{operation="create",status="failure"}++`<br>`errors_total{operation="create",error_type="validation"}++`<br>`operation_duration_seconds` observado | Taxa de erro, identificação de problemas de validação |
| Cartão não encontrado | `operations_total{operation="update",status="failure"}++`<br>`errors_total{operation="update",error_type="not_found"}++` | Taxa de erro 404, possível problema no front-end |
| Erro de repositório | `operations_total{status="failure"}++`<br>`errors_total{error_type="repository"}++` | Problemas de banco, necessidade de escalonamento |
| Cartão deletado | `operations_total{operation="delete",status="success"}++`<br>`active_total--` | Taxa de churn, redução da base |

---

## Queries PromQL Essenciais

### Taxa de Operações
```promql
sum(rate(financial_card_operations_total[5m])) by (operation)
```

### Taxa de Erro Geral
```promql
sum(rate(financial_card_operations_total{status="failure"}[5m]))
/
sum(rate(financial_card_operations_total[5m])) * 100
```

### Latência P95 por Operação
```promql
histogram_quantile(0.95,
  sum(rate(financial_card_operation_duration_seconds_bucket[5m])) by (le, operation)
)
```

### Top Erros
```promql
topk(5, sum(rate(financial_card_errors_total[5m])) by (error_type))
```

### Cartões Ativos
```promql
financial_card_active_total
```

---

## Alertas Recomendados

### 1. Alta Taxa de Erro
```yaml
expr: |
  sum(rate(financial_card_errors_total[5m]))
  /
  sum(rate(financial_card_operations_total[5m])) > 0.05
for: 5m
```

### 2. Latência Alta
```yaml
expr: |
  histogram_quantile(0.95,
    sum(rate(financial_card_operation_duration_seconds_bucket[5m])) by (le)
  ) > 1
for: 5m
```

### 3. Spike de Validação
```yaml
expr: |
  rate(financial_card_errors_total{error_type="validation"}[5m]) > 10
for: 2m
```

---

## Dashboard Grafana

### Painéis Sugeridos

1. **Visão Geral**
   - Taxa de operações por tipo (Graph)
   - Cartões ativos (Stat + Sparkline)
   - Taxa de erro geral (Gauge com threshold)

2. **Performance**
   - Latência P50/P95/P99 (Graph com múltiplas séries)
   - Distribuição de latência (Heatmap)
   - Operações lentas (> 1s)

3. **Erros**
   - Top 5 tipos de erro (Bar Gauge)
   - Erros por operação (Graph empilhado)
   - Taxa de erro percentual (Graph com threshold 5%)

4. **Tendências**
   - Crescimento de cartões (Graph)
   - Taxa de criação vs deleção (Graph)
   - Operações por dia (Graph)

---

## Testes

### Exemplo de Teste com Métricas

```go
func (s *CreateCardUseCaseSuite) TestExecute() {
    // ...
    cardMetrics := metrics.NewTestCardMetrics()
    uc := usecase.NewCreateCardUseCase(s.obs, repo, cardMetrics)

    output, err := uc.Execute(s.ctx, userID, input)

    s.NoError(err)
    s.NotNil(output)
}
```

**Helper**: `NewTestCardMetrics()` cria registry isolado para evitar conflitos.

---

## Boas Práticas Aplicadas

✅ **Labels com baixa cardinalidade**: Apenas `operation`, `status`, `error_type`
✅ **Sem IDs ou dados sensíveis**: Nunca user_id ou card_id em labels
✅ **Tipos corretos**: Counter para contadores, Histogram para latência, Gauge para estado
✅ **Namespace/Subsystem**: `financial_card_*` para organização
✅ **Buckets apropriados**: [1ms, 5ms, 10ms, 25ms, 50ms, 100ms, 250ms, 500ms, 1s, 2.5s, 5s]
✅ **Classificação automática de erros**: Baseado em palavras-chave
✅ **Proteção contra panics**: Registra métricas mesmo em panics
✅ **Testes isolados**: Registry por teste para evitar conflitos

---

## Dependências Adicionadas

```go
go get github.com/prometheus/client_golang/prometheus
go get github.com/prometheus/client_golang/prometheus/promauto
```

---

## Próximos Passos

### Implementação
- [x] Definir métricas
- [x] Instrumentar use cases
- [x] Integrar no módulo
- [x] Criar testes
- [x] Documentar

### Deploy
- [ ] Configurar scraping do Prometheus
- [ ] Importar dashboard no Grafana
- [ ] Configurar alertas no Alertmanager
- [ ] Definir SLIs/SLOs

### Expansão
- [ ] Replicar padrão para outros módulos (invoice, transaction, budget, etc.)
- [ ] Adicionar métricas de cache (se houver)
- [ ] Adicionar métricas de eventos de domínio
- [ ] Dashboard consolidado multi-módulo

---

## Impacto

### Observabilidade Ganha
- **Antes**: Apenas traces (OpenTelemetry)
- **Depois**: Traces + Métricas + Alertas

### Visibilidade
- **Taxa de operações**: Identificar picos e padrões de uso
- **Performance**: Detectar degradação de latência
- **Erros**: Identificar tipos específicos de falha
- **Estado**: Acompanhar crescimento da base de cartões

### Tempo de Resposta a Incidentes
- **Antes**: Reativo (usuários reportam)
- **Depois**: Proativo (alertas automáticos)

---

## Contato

Para dúvidas sobre as métricas, consulte:
- `pkg/observability/metrics/README.md` - Documentação completa
- `pkg/observability/metrics/card_metrics.go` - Código fonte das métricas
