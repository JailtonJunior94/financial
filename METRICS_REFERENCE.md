# Referência de Métricas - Financial API

Este documento lista todas as métricas coletadas pela aplicação Financial API usando OpenTelemetry.

## Índice
- [Métricas HTTP (Servidor)](#métricas-http-servidor)
- [Métricas de Banco de Dados](#métricas-de-banco-de-dados)
- [Métricas Customizadas por Módulo](#métricas-customizadas-por-módulo)
- [Como Consultar](#como-consultar)

---

## Métricas HTTP (Servidor)

Fornecidas automaticamente pelo `devkit-go/pkg/http_server` com `WithMetrics()`.

### http_server_requests_total
**Tipo**: Counter
**Descrição**: Total de requisições HTTP recebidas
**Labels**:
- `method` - Método HTTP (GET, POST, PUT, DELETE)
- `path` - Rota HTTP (ex: /api/v1/cards)
- `status` - Status code HTTP (200, 404, 500, etc.)

**Exemplo PromQL**:
```promql
# Taxa de requisições por segundo
rate(http_server_requests_total[5m])

# Requisições com erro (5xx)
http_server_requests_total{status=~"5.."}

# Top 5 endpoints mais acessados
topk(5, sum by(path) (rate(http_server_requests_total[5m])))
```

---

### http_server_request_duration_seconds
**Tipo**: Histogram
**Descrição**: Latência das requisições HTTP em segundos
**Labels**:
- `method` - Método HTTP
- `path` - Rota HTTP
- `status` - Status code HTTP

**Buckets**: [.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10] segundos

**Exemplo PromQL**:
```promql
# P95 de latência geral
histogram_quantile(0.95, rate(http_server_request_duration_seconds_bucket[5m]))

# P99 por endpoint
histogram_quantile(0.99, sum by(path, le) (rate(http_server_request_duration_seconds_bucket[5m])))

# Latência média por método
avg by(method) (rate(http_server_request_duration_seconds_sum[5m]) / rate(http_server_request_duration_seconds_count[5m]))
```

---

### http_server_active_requests
**Tipo**: Gauge
**Descrição**: Número de requisições HTTP ativas no momento
**Labels**:
- `method` - Método HTTP
- `path` - Rota HTTP

**Exemplo PromQL**:
```promql
# Requisições ativas no momento
http_server_active_requests

# Pico de requisições simultâneas (últimas 24h)
max_over_time(http_server_active_requests[24h])
```

---

### http_server_request_size_bytes
**Tipo**: Histogram
**Descrição**: Tamanho das requisições HTTP em bytes
**Labels**:
- `method` - Método HTTP
- `path` - Rota HTTP

**Exemplo PromQL**:
```promql
# Tamanho médio de requisições
avg(rate(http_server_request_size_bytes_sum[5m]) / rate(http_server_request_size_bytes_count[5m]))

# Maior requisição recebida
max(http_server_request_size_bytes)
```

---

### http_server_response_size_bytes
**Tipo**: Histogram
**Descrição**: Tamanho das respostas HTTP em bytes
**Labels**:
- `method` - Método HTTP
- `path` - Rota HTTP
- `status` - Status code HTTP

**Exemplo PromQL**:
```promql
# Tamanho médio de respostas
avg(rate(http_server_response_size_bytes_sum[5m]) / rate(http_server_response_size_bytes_count[5m]))

# Bandwidth de saída (bytes/s)
rate(http_server_response_size_bytes_sum[5m])
```

---

## Métricas de Banco de Dados

### sql_client_connections_open
**Tipo**: Gauge
**Descrição**: Número de conexões abertas no pool de conexões
**Labels**:
- `db_system` - Sistema de banco (postgres, cockroachdb)
- `db_name` - Nome do banco de dados

**Exemplo PromQL**:
```promql
# Conexões abertas no momento
sql_client_connections_open

# Uso do pool de conexões (% de max_open_conns=25)
(sql_client_connections_open / 25) * 100
```

---

### sql_client_connections_idle
**Tipo**: Gauge
**Descrição**: Número de conexões ociosas no pool
**Labels**:
- `db_system` - Sistema de banco
- `db_name` - Nome do banco

**Exemplo PromQL**:
```promql
# Conexões ociosas
sql_client_connections_idle

# Taxa de utilização (conexões ativas / total)
(sql_client_connections_open - sql_client_connections_idle) / sql_client_connections_open
```

---

### sql_client_connections_wait_duration_seconds
**Tipo**: Histogram
**Descrição**: Tempo de espera para obter conexão do pool
**Labels**:
- `db_system` - Sistema de banco
- `db_name` - Nome do banco

**Exemplo PromQL**:
```promql
# P95 de tempo de espera por conexão
histogram_quantile(0.95, rate(sql_client_connections_wait_duration_seconds_bucket[5m]))

# Alertar se P99 > 100ms (pool saturado)
histogram_quantile(0.99, rate(sql_client_connections_wait_duration_seconds_bucket[5m])) > 0.1
```

---

### sql_client_query_duration_seconds
**Tipo**: Histogram
**Descrição**: Duração das queries SQL em segundos
**Labels**:
- `db_system` - Sistema de banco
- `db_name` - Nome do banco
- `db_operation` - Operação (SELECT, INSERT, UPDATE, DELETE)

**Buckets**: [.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5] segundos

**Exemplo PromQL**:
```promql
# P99 de latência de queries
histogram_quantile(0.99, rate(sql_client_query_duration_seconds_bucket[5m]))

# Queries lentas (P95 > 1s)
histogram_quantile(0.95, sum by(db_operation) (rate(sql_client_query_duration_seconds_bucket[5m]))) > 1

# Taxa de queries por segundo
rate(sql_client_query_duration_seconds_count[5m])
```

---

### sql_client_query_errors_total
**Tipo**: Counter
**Descrição**: Total de erros em queries SQL
**Labels**:
- `db_system` - Sistema de banco
- `db_name` - Nome do banco
- `db_operation` - Operação SQL
- `error_type` - Tipo de erro (timeout, constraint_violation, deadlock, etc.)

**Exemplo PromQL**:
```promql
# Taxa de erros por segundo
rate(sql_client_query_errors_total[5m])

# Erros por tipo
sum by(error_type) (rate(sql_client_query_errors_total[5m]))

# Error rate (%)
(rate(sql_client_query_errors_total[5m]) / rate(sql_client_query_duration_seconds_count[5m])) * 100
```

---

## Métricas Customizadas por Módulo

### Módulo: Card (Cartões de Crédito)

#### financial.card.operations.total
**Tipo**: Counter
**Descrição**: Total de operações executadas no módulo de cartões
**Labels**:
- `operation` - Tipo de operação (create, update, delete, find, find_by)
- `status` - Status da operação (success, failure)

**Exemplo PromQL**:
```promql
# Taxa de operações bem-sucedidas
rate(financial_card_operations_total{status="success"}[5m])

# Taxa de falha por operação
rate(financial_card_operations_total{status="failure"}[5m])

# Success rate (%)
(sum(rate(financial_card_operations_total{status="success"}[5m])) / sum(rate(financial_card_operations_total[5m]))) * 100
```

---

#### financial.card.errors.total
**Tipo**: Counter
**Descrição**: Total de erros detalhados por operação
**Labels**:
- `operation` - Tipo de operação (create, update, delete, find, find_by)
- `error_type` - Tipo de erro (validation, not_found, repository, parsing, unknown)

**Exemplo PromQL**:
```promql
# Taxa de erros por tipo
sum by(error_type) (rate(financial_card_errors_total[5m]))

# Erros de validação
rate(financial_card_errors_total{error_type="validation"}[5m])

# Top erros por operação
topk(5, sum by(operation, error_type) (rate(financial_card_errors_total[5m])))
```

**Classificação de Erros**:
- `validation` - Erros de validação de entrada (nome vazio, due_day inválido)
- `not_found` - Cartão não encontrado (404)
- `repository` - Erros de banco de dados
- `parsing` - Erros de parsing de UUID ou dados
- `unknown` - Erros não classificados

---

#### financial.card.operation.duration.seconds
**Tipo**: Histogram
**Descrição**: Duração das operações de cartão em segundos (use case completo)
**Labels**:
- `operation` - Tipo de operação (create, update, delete, find, find_by)

**Buckets**: [.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5] segundos

**Exemplo PromQL**:
```promql
# P95 de latência por operação
histogram_quantile(0.95, sum by(operation, le) (rate(financial_card_operation_duration_seconds_bucket[5m])))

# Operações lentas (P99 > 500ms)
histogram_quantile(0.99, sum by(operation, le) (rate(financial_card_operation_duration_seconds_bucket[5m]))) > 0.5

# Latência média
avg(rate(financial_card_operation_duration_seconds_sum[5m]) / rate(financial_card_operation_duration_seconds_count[5m]))
```

---

#### financial.card.active.total
**Tipo**: UpDownCounter (behaves like Gauge)
**Descrição**: Número total de cartões ativos no sistema
**Labels**: (nenhum)

**Exemplo PromQL**:
```promql
# Total de cartões ativos no momento
financial_card_active_total

# Crescimento de cartões (últimas 24h)
financial_card_active_total - financial_card_active_total offset 24h

# Taxa de crescimento (%)
((financial_card_active_total - financial_card_active_total offset 24h) / financial_card_active_total offset 24h) * 100
```

---

## Como Consultar

### Via Prometheus API

```bash
# Listar todas as métricas disponíveis
curl -s "http://localhost:9090/api/v1/label/__name__/values" | jq

# Query específica
curl -s "http://localhost:9090/api/v1/query?query=http_server_requests_total" | jq

# Query com range (últimos 5 minutos)
curl -s "http://localhost:9090/api/v1/query_range?query=rate(http_server_requests_total[5m])&start=$(date -u -v-5M +%s)&end=$(date -u +%s)&step=15s" | jq
```

---

### Via Grafana

**URL**: http://localhost:3100
**Datasource**: Prometheus (http://prometheus:9090)

**Dashboards Recomendados**:
1. **HTTP Overview**: Requisições, latência, erros
2. **Database Performance**: Pool de conexões, query latency
3. **Card Module**: Métricas customizadas do módulo de cartões
4. **System Health**: CPU, memória, goroutines

---

### Alertas Recomendados

#### Alta Taxa de Erros HTTP
```promql
(sum(rate(http_server_requests_total{status=~"5.."}[5m])) / sum(rate(http_server_requests_total[5m]))) * 100 > 5
```

#### Latência Alta (P99 > 1s)
```promql
histogram_quantile(0.99, rate(http_server_request_duration_seconds_bucket[5m])) > 1
```

#### Pool de Conexões Saturado
```promql
sql_client_connections_open >= 24  # 96% de max_open_conns=25
```

#### Alta Taxa de Erros de Validação (Card)
```promql
rate(financial_card_errors_total{error_type="validation"}[5m]) > 0.5
```

#### Operações Card Lentas
```promql
histogram_quantile(0.99, sum by(operation) (rate(financial_card_operation_duration_seconds_bucket[5m]))) > 1
```

---

## Exportação e Formato

### OpenTelemetry OTLP
- **Protocolo**: gRPC
- **Endpoint**: localhost:4317
- **Intervalo**: ~60 segundos
- **Formato**: OTLP (OpenTelemetry Protocol)

### Prometheus Exposition
- **Endpoint**: http://localhost:8000/metrics
- **Formato**: Prometheus text format
- **Scrape Interval**: Configurável (padrão: 15s)

### Compatibilidade
- Nomes de métricas OpenTelemetry são automaticamente normalizados para formato Prometheus
- Dots (.) são convertidos para underscores (_)
- Prefixos mantidos para namespace adequado

---

## Próximos Passos

### Métricas Futuras (Planejadas)
- [ ] Métricas customizadas para Budget module
- [ ] Métricas customizadas para Invoice module
- [ ] Métricas customizadas para Transaction module
- [ ] Métricas de negócio (taxa de conversão, valor médio de faturas)
- [ ] Métricas de RabbitMQ (mensagens publicadas/consumidas)
- [ ] Métricas de Worker (jobs executados, falhas)

### Melhorias
- [ ] Exemplar support (correlação traces ↔ métricas)
- [ ] Resource detection automática (hostname, container_id)
- [ ] Métricas de runtime Go (goroutines, GC, heap)
- [ ] Service Level Objectives (SLOs) com alertas

---

**Última atualização**: 2026-01-27
**Versão do devkit-go**: v1.7.8
**OpenTelemetry SDK**: v1.39.0
