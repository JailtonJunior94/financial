# Métricas - Referência Rápida

## 📊 Métricas HTTP (Automáticas)

| Métrica | Tipo | Descrição | Labels |
|---------|------|-----------|--------|
| `http_server_requests_total` | Counter | Total de requisições | method, path, status |
| `http_server_request_duration_seconds` | Histogram | Latência das requisições | method, path, status |
| `http_server_active_requests` | Gauge | Requisições ativas | method, path |
| `http_server_request_size_bytes` | Histogram | Tamanho das requisições | method, path |
| `http_server_response_size_bytes` | Histogram | Tamanho das respostas | method, path, status |

**Query Útil**:
```promql
# Taxa de requisições/s
rate(http_server_requests_total[5m])

# P95 de latência
histogram_quantile(0.95, rate(http_server_request_duration_seconds_bucket[5m]))

# Taxa de erro (%)
(sum(rate(http_server_requests_total{status=~"5.."}[5m])) / sum(rate(http_server_requests_total[5m]))) * 100
```

---

## 🗄️ Métricas de Banco de Dados (Automáticas)

| Métrica | Tipo | Descrição | Labels |
|---------|------|-----------|--------|
| `sql_client_connections_open` | Gauge | Conexões abertas no pool | db_system, db_name |
| `sql_client_connections_idle` | Gauge | Conexões ociosas | db_system, db_name |
| `sql_client_connections_wait_duration_seconds` | Histogram | Tempo de espera por conexão | db_system, db_name |
| `sql_client_query_duration_seconds` | Histogram | Duração das queries | db_system, db_name, db_operation |
| `sql_client_query_errors_total` | Counter | Erros em queries | db_system, db_name, db_operation, error_type |

**Query Útil**:
```promql
# Uso do pool (% de 25 max)
(sql_client_connections_open / 25) * 100

# P99 de latência de queries
histogram_quantile(0.99, rate(sql_client_query_duration_seconds_bucket[5m]))

# Taxa de erros de query
rate(sql_client_query_errors_total[5m])
```

---

## 💳 Métricas do Módulo Card (Customizadas)

| Métrica | Tipo | Descrição | Labels |
|---------|------|-----------|--------|
| `financial_card_operations_total` | Counter | Total de operações | operation, status |
| `financial_card_errors_total` | Counter | Erros detalhados | operation, error_type |
| `financial_card_operation_duration_seconds` | Histogram | Duração das operações | operation |
| `financial_card_active_total` | UpDownCounter | Cartões ativos | - |

**Operations**: create, update, delete, find, find_by
**Error Types**: validation, not_found, repository, parsing, unknown

**Query Útil**:
```promql
# Success rate (%)
(sum(rate(financial_card_operations_total{status="success"}[5m])) / sum(rate(financial_card_operations_total[5m]))) * 100

# P95 de latência
histogram_quantile(0.95, sum by(operation, le) (rate(financial_card_operation_duration_seconds_bucket[5m])))

# Total de cartões
financial_card_active_total
```

---

## 🚀 Comandos Rápidos

### Listar Todas as Métricas
```bash
curl -s "http://localhost:9090/api/v1/label/__name__/values" | jq
```

### Query Específica
```bash
curl -s "http://localhost:9090/api/v1/query?query=http_server_requests_total" | jq
```

### Verificar Endpoint de Métricas (Prometheus Format)
```bash
curl http://localhost:8000/metrics | grep financial_card
```

---

## 📈 Dashboards Grafana

**Acesso**: http://localhost:3100 (admin/admin)

**Datasources**:
- Prometheus: http://prometheus:9090
- Tempo (Traces): http://tempo:3200
- Loki (Logs): http://loki:3100

---

## ⚠️ Alertas Críticos

| Alerta | Condição | Severidade |
|--------|----------|------------|
| Alta taxa de erro | Error rate > 5% | Crítico |
| Latência alta | P99 > 1s | Alto |
| Pool saturado | Conexões >= 24 | Alto |
| Erros de validação | Rate > 0.5/s | Médio |

---

## 🔄 Pipeline de Métricas

```
┌─────────────┐
│ Application │
│  (Go App)   │
└──────┬──────┘
       │ OpenTelemetry SDK
       │ (Context propagation)
       ▼
┌─────────────┐
│ OTLP Exporter│ ──gRPC:4317──┐
└─────────────┘               │
                              ▼
                       ┌─────────────┐
                       │OTEL Collector│
                       └──────┬──────┘
                              │
              ┌───────────────┼───────────────┐
              ▼               ▼               ▼
       ┌──────────┐    ┌──────────┐   ┌──────────┐
       │Prometheus│    │  Tempo   │   │   Loki   │
       │ (Metrics)│    │ (Traces) │   │  (Logs)  │
       └──────────┘    └──────────┘   └──────────┘
              │               │               │
              └───────────────┴───────────────┘
                              │
                              ▼
                       ┌──────────┐
                       │ Grafana  │
                       │:3100     │
                       └──────────┘
```

---

## 📝 Configuração

**Arquivo**: `cmd/.env`

```env
# OpenTelemetry
OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317
OTEL_EXPORTER_OTLP_PROTOCOL=grpc
OTEL_EXPORTER_OTLP_INSECURE=true
OTEL_TRACE_SAMPLE_RATE=1.0

# Servidor
HTTP_PORT=8000
SERVICE_NAME_API=financial-api

# Banco
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=6
```

---

**Ver documentação completa**: [METRICS_REFERENCE.md](./METRICS_REFERENCE.md)
