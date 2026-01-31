# HTTP Metrics Middleware

Middleware para instrumentação automática de métricas HTTP por rota.

## Métricas Expostas

### `financial.http.request.duration.seconds`
- **Tipo:** Histogram
- **Descrição:** Latência das requisições HTTP em segundos
- **Labels:**
  - `method`: HTTP method (GET, POST, PUT, DELETE)
  - `route`: Chi route pattern (ex: `/api/v1/cards/{id}`)
  - `status_class`: HTTP status class (2xx, 4xx, 5xx)
- **Buckets:** `.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5`

### `financial.http.requests.total`
- **Tipo:** Counter
- **Descrição:** Total de requisições HTTP processadas
- **Labels:**
  - `method`: HTTP method
  - `route`: Chi route pattern
  - `status_class`: HTTP status class (2xx, 4xx, 5xx)

### `financial.http.active_requests`
- **Tipo:** UpDownCounter
- **Descrição:** Número de requisições HTTP ativas no momento

## Uso

O middleware é registrado automaticamente no servidor HTTP em `cmd/server/server.go`:

```go
metricsMiddleware := middlewares.NewMetricsMiddleware(o11y)

srv, err := httpserver.New(
    o11y,
    httpserver.WithMiddleware(metricsMiddleware.Handler),
)
```

## Queries Prometheus

### Taxa de erros por rota
```promql
sum(rate(financial_http_requests_total{status_class="5xx"}[5m])) by (route, method)
```

### P95 latência por rota
```promql
histogram_quantile(0.95,
  sum(rate(financial_http_request_duration_seconds_bucket[5m])) by (route, method, status_class, le)
)
```

### Requests/segundo
```promql
sum(rate(financial_http_requests_total[5m])) by (route, method)
```

### Requisições ativas
```promql
financial_http_active_requests
```
