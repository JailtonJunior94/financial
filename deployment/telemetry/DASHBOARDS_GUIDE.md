# Grafana Dashboards - Guia Completo

## 📊 Dashboards Disponíveis

### 1. Financial - API HTTP Metrics
**Arquivo**: `financial-api-http.json`
**UID**: `financial-api-http`

**Métricas Monitadas**:
- Taxa de requisições (requests/second)
- Latência HTTP (P50, P95, P99)
- Taxa de erro (5xx)
- Distribuição de status codes
- Goroutines ativas
- Uso de memória
- CPU e GC

**Painéis**:
- 🌐 API Overview (6 stats principais)
- 📊 HTTP Traffic (por método e status)
- ⚡ Performance (latência e heatmap)
- 💾 System Resources (memória, goroutines, GC)

---

### 2. Financial - Card Metrics
**Arquivo**: `financial-cards.json`
**UID**: `financial-card-metrics`

**Métricas Monitadas**:
- Cartões ativos no sistema
- Taxa de operações (create, update, delete, find)
- Taxa de sucesso/falha
- Erros por tipo (validation, not_found, repository)
- Latência das operações
- Crescimento de cartões

**Painéis**:
- 📊 Visão Geral (KPIs principais)
- ⚡ Performance (latência P50/P95/P99 por operação)
- ❌ Erros e Falhas (por tipo e operação)
- 📈 Tendências (crescimento e taxa de criação/deleção)

---

### 3. Financial - Database Metrics
**Arquivo**: `financial-database.json`
**UID**: `financial-database-metrics`

**Métricas Monitadas**:
- Pool de conexões (open, idle, usage %)
- Tempo de espera por conexão (P95)
- Taxa de queries (por operação: SELECT, INSERT, UPDATE)
- Latência de queries (P50, P95, P99)
- Erros de queries (por tipo)
- Taxa de erro geral

**Painéis**:
- 🗄️ Connection Pool (4 stats)
- ⚡ Query Performance (rate e latência)
- ❌ Errors (por tipo e gauge de taxa)
- 📈 Trends (pool trend e heatmap)

---

## 🚀 Como Importar os Dashboards

### Método 1: Importação Manual via UI

1. **Acesse Grafana**:
   ```
   http://localhost:3100
   Login: admin / admin
   ```

2. **Vá para Dashboards**:
   - Menu lateral → Dashboards → Import

3. **Cole o JSON**:
   - Clique em "Import dashboard"
   - Copie o conteúdo de um dos arquivos JSON
   - Cole no campo "Import via panel json"
   - Clique em "Load"

4. **Configure o Datasource**:
   - Selecione "Prometheus" no dropdown
   - Clique em "Import"

5. **Repita** para os 3 dashboards

---

### Método 2: Provisioning Automático (Recomendado)

Os dashboards já estão configurados para serem carregados automaticamente via provisioning do Grafana.

**Verificar se está funcionando**:
```bash
# Ver logs do Grafana
docker logs deployment-otel-lgtm-1 2>&1 | grep -i dashboard

# Você deve ver algo como:
# "Dashboard provisioning completed"
```

**Acessar dashboards**:
- Menu lateral → Dashboards → Browse
- Filtrar por tag: `financial`

---

## 🔍 Verificando se os Dashboards Estão Funcionando

### 1. Verificar Datasource Prometheus

```bash
# Via API do Grafana
curl -s http://admin:admin@localhost:3100/api/datasources | jq '.[] | select(.type=="prometheus")'
```

**Esperado**: Datasource "Prometheus" com URL http://prometheus:9090

---

### 2. Testar Queries Manualmente

Acesse: http://localhost:9090/graph

**Query 1 - HTTP Metrics**:
```promql
rate(http_server_requests_total[1m])
```

**Query 2 - Database Metrics**:
```promql
sql_client_connections_open
```

**Query 3 - Card Metrics**:
```promql
financial_card_active_total
```

Se todas retornarem dados, os dashboards funcionarão corretamente.

---

### 3. Gerar Dados de Teste

Para popular os dashboards com dados reais:

```bash
# 1. Criar usuário
curl -X POST http://localhost:8000/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Teste Dashboard",
    "email": "dashboard@test.com",
    "password": "senha123456"
  }'

# 2. Fazer login
TOKEN=$(curl -s -X POST http://localhost:8000/api/v1/token \
  -H "Content-Type: application/json" \
  -d '{
    "email": "dashboard@test.com",
    "password": "senha123456"
  }' | jq -r '.token')

# 3. Criar alguns cartões
for i in {1..5}; do
  curl -X POST http://localhost:8000/api/v1/cards \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
      \"name\": \"Cartão $i\",
      \"due_day\": $((5 + i))
    }"
done

# 4. Aguardar 60 segundos (exportação OTLP)
sleep 60

# 5. Acessar dashboards no Grafana
```

---

## 📈 Queries Úteis por Dashboard

### HTTP Dashboard

**Top 5 endpoints mais lentos**:
```promql
topk(5, histogram_quantile(0.95, sum by(path, le) (rate(http_server_request_duration_seconds_bucket[5m]))))
```

**Taxa de erro por endpoint**:
```promql
sum by(path) (rate(http_server_requests_total{status=~"5.."}[5m])) / sum by(path) (rate(http_server_requests_total[5m])) * 100
```

**Throughput geral**:
```promql
sum(rate(http_server_requests_total[1m]))
```

---

### Card Dashboard

**Taxa de sucesso**:
```promql
(sum(rate(financial_card_operations_total{status="success"}[5m])) / sum(rate(financial_card_operations_total[5m]))) * 100
```

**Operações lentas (P99 > 500ms)**:
```promql
histogram_quantile(0.99, sum by(operation, le) (rate(financial_card_operation_duration_seconds_bucket[5m]))) > 0.5
```

**Top 3 tipos de erro**:
```promql
topk(3, sum by(error_type) (rate(financial_card_errors_total[5m])))
```

---

### Database Dashboard

**Pool saturado (> 95%)**:
```promql
(sql_client_connections_open / 25) * 100 > 95
```

**Queries lentas (P95 > 100ms)**:
```promql
histogram_quantile(0.95, sum by(db_operation, le) (rate(sql_client_query_duration_seconds_bucket[5m]))) > 0.1
```

**Taxa de erro de queries**:
```promql
(sum(rate(sql_client_query_errors_total[5m])) / sum(rate(sql_client_query_duration_seconds_count[5m]))) * 100
```

---

## 🚨 Alertas Recomendados

### Criar Alertas no Grafana

1. Abra um dashboard
2. Edite um painel
3. Vá para a aba "Alert"
4. Configure regras de alerta

**Exemplo - Alta Taxa de Erro**:
```yaml
Condition: WHEN last() OF query(A, 1m, now) IS ABOVE 5

Query A:
(sum(rate(http_server_requests_total{status=~"5.."}[5m])) / sum(rate(http_server_requests_total[5m]))) * 100

Notifications:
  - Send to: default channel
  - Message: "Taxa de erro HTTP acima de 5%"
```

---

## 🐛 Troubleshooting

### Dashboard não mostra dados

**1. Verificar se o Prometheus está coletando métricas**:
```bash
curl -s http://localhost:9090/api/v1/targets | jq '.data.activeTargets[] | {job: .labels.job, health: .health}'
```

**2. Verificar exportação OTLP**:
```bash
# Ver logs do OTEL Collector
docker logs deployment-otel-lgtm-1 2>&1 | tail -50
```

**3. Verificar endpoint /metrics da aplicação**:
```bash
curl http://localhost:8000/metrics | head -20
```

---

### Queries retornam "No data"

**Causa comum**: Métricas ainda não foram exportadas (intervalo de 60s).

**Solução**:
1. Aguardar 60 segundos após iniciar a aplicação
2. Gerar alguma atividade (criar cartão, fazer requisições)
3. Atualizar o dashboard (botão refresh no topo)

---

### Datasource não conecta

**Verificar conectividade**:
```bash
# De dentro do container do Grafana
docker exec -it deployment-otel-lgtm-1 sh
wget -O- http://prometheus:9090/api/v1/status/config
```

**Se falhar**:
- Verificar se Prometheus está rodando: `docker ps | grep prometheus`
- Verificar redes Docker: `docker network inspect deployment_default`

---

## 📊 Personalizações

### Adicionar Variáveis (Templates)

Exemplo - Filtro por operação no Card Dashboard:

1. Dashboard Settings → Variables → Add variable
2. Name: `operation`
3. Type: Query
4. Query: `label_values(financial_card_operations_total, operation)`
5. Use na query: `financial_card_operations_total{operation="$operation"}`

---

### Ajustar Time Range

Por padrão, dashboards mostram última hora (`now-1h` até `now`).

**Mudar para 24h**:
- Dashboard Settings → Time options
- Default: `now-24h to now`
- Refresh: `1m`

---

## 📚 Recursos Adicionais

- **Documentação de Métricas**: `/METRICS_REFERENCE.md`
- **Referência Rápida**: `/METRICS_QUICK_REFERENCE.md`
- **PromQL Guide**: https://prometheus.io/docs/prometheus/latest/querying/basics/
- **Grafana Docs**: https://grafana.com/docs/grafana/latest/

---

**Última atualização**: 2026-01-27
**Versão**: 1.0.0
