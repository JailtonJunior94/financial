# 🔍 Diagnóstico: Problema com Métricas (NoData)

## ❌ PROBLEMA IDENTIFICADO

Os dashboards estão mostrando **"No Data"** porque as métricas esperadas **NÃO estão sendo exportadas**.

---

## 📊 Métricas Disponíveis vs Esperadas

### Database Metrics

#### ✅ Métricas que EXISTEM (Prometheus atual)
```promql
db_sql_connection_open                              # Conexões abertas
db_sql_connection_max_open                          # Máximo de conexões
db_sql_connection_wait_duration_milliseconds_total  # Tempo de espera total
db_sql_connection_wait_total                        # Total de waits
db_sql_latency_milliseconds_bucket                  # Latência (histogram)
db_sql_latency_milliseconds_count
db_sql_latency_milliseconds_sum
db_sql_connection_closed_max_idle_time_total
db_sql_connection_closed_max_idle_total
db_sql_connection_closed_max_lifetime_total
```

#### ❌ Métricas ESPERADAS pelos dashboards (mas NÃO existem)
```promql
db_client_connections_usage                         # ← NÃO EXISTE
db_client_connections_max                           # ← NÃO EXISTE
db_client_connections_wait_time                     # ← NÃO EXISTE
db_client_operation_duration                        # ← NÃO EXISTE
```

**Causa:** A biblioteca `otelsql` usada pelo `postgres_otelsql` exporta métricas com nomenclatura `db_sql_*` ao invés de `db_client_*` (semantic conventions antigas).

---

### HTTP Server Metrics

#### ❌ Métricas HTTP NÃO EXISTEM!
```promql
http_server_request_count                           # ← NÃO EXISTE
http_server_duration                                # ← NÃO EXISTE
http_server_active_requests                         # ← NÃO EXISTE
```

**Causa:** O `httpserver` do devkit-go **NÃO está exportando métricas** ou:
1. Não está configurado para exportar via OTLP
2. Está exportando mas com nomenclatura diferente
3. Precisa de configuração adicional

---

### Métricas Alternativas Disponíveis

#### ✅ Traces Service Graph (geradas automaticamente)
```promql
traces_service_graph_request_total                  # Requests via traces
traces_service_graph_request_server_seconds_bucket  # Latência server
traces_service_graph_request_client_seconds_bucket  # Latência client
```

#### ✅ Métricas Custom da Aplicação
```promql
financial_card_active_total                         # Cards ativos
financial_card_operations_operation_total           # Operações de card
financial_card_operation_duration_seconds_bucket    # Latência operações
financial_card_errors_error_total                   # Erros
```

---

## 🔧 SOLUÇÃO

### Opção 1: Atualizar Dashboards para Usar Métricas Reais ✅ (RECOMENDADO)

Criar novos dashboards usando as métricas que **REALMENTE existem**:

#### Dashboard Database - Métricas Reais
```promql
# Conexões Abertas
db_sql_connection_open

# Pool Máximo
db_sql_connection_max_open

# Utilização do Pool (%)
(db_sql_connection_open / db_sql_connection_max_open) * 100

# Latência P95
histogram_quantile(0.95, rate(db_sql_latency_milliseconds_bucket[5m]))

# Tempo de Espera Total
rate(db_sql_connection_wait_duration_milliseconds_total[5m])

# Total de Waits por Segundo
rate(db_sql_connection_wait_total[5m])
```

#### Dashboard HTTP - Usar Traces Service Graph
```promql
# Requests por Segundo (via traces)
sum(rate(traces_service_graph_request_total[1m]))

# Latência P95 (via traces - server side)
histogram_quantile(0.95,
  rate(traces_service_graph_request_server_seconds_bucket[5m])
)

# Latência P95 (via traces - client side)
histogram_quantile(0.95,
  rate(traces_service_graph_request_client_seconds_bucket[5m])
)
```

---

### Opção 2: Investigar Por Que HTTP Metrics Não Estão Sendo Exportadas

**Possíveis causas:**

1. **Servidor não está exportando métricas OTLP**
   - Verificar se `WithMetrics()` está configurado (✅ JÁ ESTÁ)
   - Verificar se métricas estão sendo criadas mas não enviadas

2. **Endpoint OTLP incorreto**
   - App deve exportar para `otel-lgtm:4317` (gRPC)
   - Verificar variável de ambiente `OTEL_EXPORTER_OTLP_ENDPOINT`

3. **Nomenclatura diferente**
   - Procurar por métricas com prefixo diferente
   - Verificar logs do OTLP Collector

---

## 🔍 Verificação da Configuração

### 1. Variáveis de Ambiente da Aplicação

Verificar se a aplicação está configurada para exportar métricas:

```bash
# Deve estar configurado no .env ou variáveis de ambiente
OTEL_EXPORTER_OTLP_ENDPOINT=otel-lgtm:4317
OTEL_EXPORTER_OTLP_PROTOCOL=grpc
OTEL_EXPORTER_OTLP_INSECURE=true
OTEL_SERVICE_VERSION=1.0.0
OTEL_TRACE_SAMPLE_RATE=1.0
```

### 2. Verificar Logs do Collector

```bash
docker logs deployment-otel-lgtm-1 2>&1 | grep -i "metric"
```

### 3. Verificar Se App Está Conectando no Collector

```bash
docker logs deployment-otel-lgtm-1 2>&1 | grep -i "connection"
```

---

## ✅ Solução Imediata: Dashboards Corrigidos

Vou criar dashboards usando as métricas que **REALMENTE existem**:

### 1. `financial-database-real.json`
- Usar `db_sql_*` ao invés de `db_client_*`
- Conexões: `db_sql_connection_open` / `db_sql_connection_max_open`
- Latência: `db_sql_latency_milliseconds_bucket`
- Wait: `db_sql_connection_wait_duration_milliseconds_total`

### 2. `financial-http-traces.json`
- Usar `traces_service_graph_*` (métricas geradas de traces)
- Requests: `traces_service_graph_request_total`
- Latência: `traces_service_graph_request_server_seconds_bucket`

### 3. `financial-cards.json` (custom metrics)
- Usar `financial_card_*`
- Já existe e funciona

---

## 📋 Checklist de Debugging

### Métricas Database
- [x] Identificadas métricas reais: `db_sql_*`
- [ ] Dashboard atualizado para usar `db_sql_*`
- [ ] Testado no Grafana

### Métricas HTTP
- [x] Identificado que não existem métricas diretas
- [x] Encontradas alternativas: `traces_service_graph_*`
- [ ] Dashboard criado usando traces
- [ ] Investigar por que `http_server_*` não está sendo exportado

### Configuração
- [ ] Verificar `OTEL_EXPORTER_OTLP_ENDPOINT` na aplicação
- [ ] Verificar logs do OTLP Collector
- [ ] Confirmar se `httpserver.WithMetrics()` está ativo

---

## 🎯 Próximos Passos

1. **IMEDIATO:** Criar dashboards usando métricas reais (`db_sql_*` e `traces_service_graph_*`)
2. **CURTO PRAZO:** Investigar por que métricas HTTP diretas não estão sendo exportadas
3. **LONGO PRAZO:** Atualizar `postgres_otelsql` do devkit-go para usar semantic conventions corretas

---

**Diagnóstico em:** 2026-01-29
**Status:** Problema identificado - Nomenclatura de métricas diferente
**Prioridade:** 🔴 ALTA - Dashboards não funcionam
**Solução:** Criar dashboards com métricas reais
