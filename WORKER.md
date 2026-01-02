# Worker - Módulo de Cron Jobs

## Visão Geral

O módulo Worker implementa execução de tarefas agendadas (cron jobs) com:
- Agendamento robusto via `robfig/cron/v3`
- Graceful shutdown completo
- Recovery automático de panics
- Controle de concorrência
- Timeout configurável por job
- Logging estruturado e observabilidade
- Reutilização de infraestrutura existente (DB, RabbitMQ)

## Arquitetura

### Estrutura de Pastas

```
.
├── cmd/
│   └── worker/
│       └── worker.go                    # Entry point do worker
├── internal/
│   └── worker/
│       └── jobs/
│           ├── database_cleanup_job.go  # Exemplo: limpeza de DB
│           └── report_generator_job.go  # Exemplo: relatório + RabbitMQ
└── pkg/
    ├── jobs/
    │   └── job.go                       # Interface Job e Config
    └── scheduler/
        └── scheduler.go                 # Scheduler central com lifecycle
```

### Componentes Principais

#### 1. Interface Job (`pkg/jobs/job.go`)

Define o contrato que todo cron job deve implementar:

```go
type Job interface {
    Run(ctx context.Context) error  // Lógica do job
    Name() string                    // Identificador único
    Schedule() string                // Expressão cron
}
```

#### 2. Scheduler (`pkg/scheduler/scheduler.go`)

Gerencia o ciclo de vida dos jobs:
- Registro de jobs
- Execução agendada
- Recovery de panics
- Controle de concorrência
- Graceful shutdown

Características:
- Encapsula jobs com timeout, logging e recovery
- Aguarda jobs em execução durante shutdown
- Thread-safe
- Suporta limite de execuções concorrentes

#### 3. Worker Entry Point (`cmd/worker/worker.go`)

Responsabilidades:
- Inicializar dependências (DB, RabbitMQ, O11y)
- Criar e configurar scheduler
- Registrar jobs
- Implementar graceful shutdown

## Configuração

### Variáveis de Ambiente

Adicione ao `.env`:

```bash
# Service Names
SERVICE_NAME_WORKER=financial-worker

# Worker Configuration
WORKER_DEFAULT_TIMEOUT_SECONDS=300   # Timeout padrão dos jobs (5min)
WORKER_MAX_CONCURRENT_JOBS=10        # Máximo de jobs simultâneos
```

### Config Struct

```go
type WorkerConfig struct {
    ServiceName           string `mapstructure:"SERVICE_NAME_WORKER"`
    DefaultTimeoutSeconds int    `mapstructure:"WORKER_DEFAULT_TIMEOUT_SECONDS"`
    MaxConcurrentJobs     int    `mapstructure:"WORKER_MAX_CONCURRENT_JOBS"`
}
```

## Criando um Novo Job

### 1. Implementar a Interface Job

```go
package jobs

import (
    "context"
    "database/sql"
    "github.com/jailtonjunior94/financial/pkg/jobs"
    "github.com/JailtonJunior94/devkit-go/pkg/observability"
)

type MyCustomJob struct {
    db   *sql.DB
    o11y observability.Observability
}

func NewMyCustomJob(db *sql.DB, o11y observability.Observability) jobs.Job {
    return &MyCustomJob{
        db:   db,
        o11y: o11y,
    }
}

// Name retorna identificador único
func (j *MyCustomJob) Name() string {
    return "my_custom_job"
}

// Schedule retorna expressão cron
// Formatos suportados:
//   - "* * * * *" (minuto hora dia mês dia-da-semana)
//   - "@hourly", "@daily", "@weekly", "@monthly"
//   - "@every 30m", "@every 1h30m"
func (j *MyCustomJob) Schedule() string {
    return "0 0 * * *" // Diariamente à meia-noite
}

// Run executa a lógica do job
func (j *MyCustomJob) Run(ctx context.Context) error {
    j.o11y.Logger().Info(ctx, "executing my custom job")
    
    // Sua lógica aqui
    // O contexto será cancelado em caso de:
    // - Timeout (configurado em WORKER_DEFAULT_TIMEOUT_SECONDS)
    // - Shutdown do worker
    
    return nil
}
```

### 2. Registrar o Job

Edite `cmd/worker/worker.go`:

```go
jobsToRegister := []pkgjobs.Job{
    jobs.NewDatabaseCleanupJob(dbManager.DB(), o11y),
    jobs.NewReportGeneratorJob(dbManager.DB(), rabbitClient, cfg.RabbitMQConfig.Exchange, o11y),
    
    // Adicione seu job aqui
    jobs.NewMyCustomJob(dbManager.DB(), o11y),
}
```

## Exemplos de Jobs

### 1. DatabaseCleanupJob

Limpa registros soft-deleted há mais de 90 dias:

```go
// Schedule: Diariamente às 2h
func (j *DatabaseCleanupJob) Schedule() string {
    return "0 2 * * *"
}

// Deleta permanentemente categorias antigas
func (j *DatabaseCleanupJob) Run(ctx context.Context) error {
    cutoffDate := time.Now().Add(-90 * 24 * time.Hour)
    
    query := `
        DELETE FROM categories 
        WHERE deleted_at IS NOT NULL 
          AND deleted_at < $1
    `
    
    result, err := j.db.ExecContext(ctx, query, cutoffDate)
    // ...
}
```

### 2. ReportGeneratorJob

Gera relatórios e publica no RabbitMQ:

```go
// Schedule: Segundas-feiras às 8h
func (j *ReportGeneratorJob) Schedule() string {
    return "0 8 * * 1"
}

// Coleta dados, serializa e publica
func (j *ReportGeneratorJob) Run(ctx context.Context) error {
    // 1. Coletar dados do DB
    report := j.collectReportData(ctx)
    
    // 2. Serializar para JSON
    payload, _ := json.Marshal(report)
    
    // 3. Publicar no RabbitMQ
    err := j.publisher.Publish(ctx, j.exchange, "reports.weekly.generated", payload,
        rabbitmq.WithContentType("application/json"),
        rabbitmq.WithDeliveryMode(2), // Persistent
    )
    
    return err
}
```

## Executando o Worker

### Build

```bash
make build
# ou
go build -o bin/financial cmd/main.go
```

### Executar

```bash
./bin/financial worker
```

### Logs Esperados

```
INFO  initializing worker service=financial-worker environment=development
INFO  database connection established
INFO  rabbitmq initialized exchange=financial.events
INFO  job registered: database_cleanup schedule="0 2 * * *"
INFO  job registered: report_generator schedule="0 8 * * 1"
INFO  starting scheduler jobs_count=2 default_timeout_seconds=300
INFO  worker started successfully jobs_registered=2
```

## Graceful Shutdown

### Comportamento

Quando recebe `SIGINT` ou `SIGTERM`, o worker:

1. **Para de aceitar novos jobs** - O scheduler não agenda mais execuções
2. **Aguarda jobs em execução** - Jobs ativos têm até 30s para finalizar
3. **Encerra conexões ordenadamente**:
   - Scheduler (aguarda jobs)
   - RabbitMQ
   - Database
   - Observability (flush de métricas/traces)

### Timeout de Shutdown

Configurado em `cmd/worker/worker.go`:

```go
shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
```

### Logs de Shutdown

```
INFO  shutdown signal received, initiating graceful shutdown...
INFO  shutting down scheduler...
INFO  all running jobs completed
INFO  worker shutdown completed
```

## Controle de Concorrência

### Limite Global

Configurado em `WORKER_MAX_CONCURRENT_JOBS`:
- `0` = sem limite (não recomendado)
- `> 0` = máximo de jobs executando simultaneamente

### Comportamento

Se limite for atingido:
- Novas execuções são **puladas** (skip)
- Log de warning é gerado
- Job será agendado novamente no próximo ciclo

```
WARN  job skipped (max concurrent executions reached) job=my_job max_concurrent=10
```

## Recovery de Panics

### Proteção Dupla

1. **Recovery do cron** - Captura panics durante agendamento
2. **Recovery customizado** - Captura panics durante execução e loga stack trace

### Log de Panic

```
ERROR job panic recovered job=my_job panic="runtime error: index out of range"
      stack="goroutine 42 [running]:\ngithub.com/..."
```

## Timeout de Jobs

### Configuração

- Padrão: `WORKER_DEFAULT_TIMEOUT_SECONDS` (300s = 5min)
- Pode ser customizado por job (futura extensão)

### Comportamento

Quando timeout expira:
- Contexto é cancelado
- Job deve respeitar cancelamento (`ctx.Done()`)
- Scheduler aguarda job finalizar ou shutdown timeout

## Expressões Cron Suportadas

### Formato Padrão

```
┌───────────── segundo (0 - 59) - OPCIONAL (habilitado com cron.WithSeconds())
│ ┌───────────── minuto (0 - 59)
│ │ ┌───────────── hora (0 - 23)
│ │ │ ┌───────────── dia do mês (1 - 31)
│ │ │ │ ┌───────────── mês (1 - 12)
│ │ │ │ │ ┌───────────── dia da semana (0 - 6) (Domingo=0)
│ │ │ │ │ │
* * * * * *
```

### Exemplos

```go
"0 2 * * *"        // Diariamente às 2h
"0 */6 * * *"      // A cada 6 horas
"0 0 * * 1"        // Segundas às 00h
"0 9-17 * * 1-5"   // Dias úteis, das 9h às 17h
"*/30 * * * *"     // A cada 30 minutos

// Especiais
"@hourly"          // A cada hora
"@daily"           // Diariamente à meia-noite
"@weekly"          // Domingos à meia-noite
"@monthly"         // Dia 1 à meia-noite
"@every 1h30m"     // A cada 1h30min
```

## Boas Práticas

### 1. Jobs Idempotentes

Jobs devem ser idempotentes (executar múltiplas vezes = mesmo resultado):

```go
// ❌ Ruim
func (j *Job) Run(ctx context.Context) error {
    j.db.Exec("INSERT INTO log VALUES ('executed')")
}

// ✅ Bom
func (j *Job) Run(ctx context.Context) error {
    j.db.Exec("INSERT INTO log VALUES (?) ON CONFLICT DO NOTHING", time.Now())
}
```

### 2. Respeitar Context

Sempre verificar cancelamento em operações longas:

```go
func (j *Job) Run(ctx context.Context) error {
    for _, item := range items {
        select {
        case <-ctx.Done():
            return ctx.Err() // Shutdown ou timeout
        default:
            j.processItem(item)
        }
    }
}
```

### 3. Logging Estruturado

Use campos estruturados para facilitar busca:

```go
j.o11y.Logger().Info(ctx, "processing batch",
    observability.Int("batch_size", len(items)),
    observability.String("entity_type", "invoice"),
)
```

### 4. Tratamento de Erros

Jobs devem retornar erros (não fazer `log.Fatal`):

```go
func (j *Job) Run(ctx context.Context) error {
    if err := j.doWork(); err != nil {
        return fmt.Errorf("failed to do work: %w", err)
    }
    return nil
}
```

### 5. Transações

Use transações para operações atômicas:

```go
func (j *Job) Run(ctx context.Context) error {
    tx, err := j.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    // Operações...
    
    return tx.Commit()
}
```

## Observabilidade

### Métricas Disponíveis

O scheduler automaticamente loga:
- Início de execução
- Duração (em ms)
- Sucesso/falha
- Erros e stack traces

### Rastreamento

Jobs são rastreados via OpenTelemetry:
- Cada execução gera span
- Correlação com outras operações
- Visualização em ferramentas de APM

### Logs Estruturados

Formato JSON com campos:
- `job`: Nome do job
- `duration_ms`: Tempo de execução
- `error`: Mensagem de erro (se houver)

## Monitoramento

### Health Checks

Não implementado diretamente no worker (é stateless).
Para monitorar:
- Verificar logs de execução
- Alertas em erros consecutivos
- Métricas de duração anormal

### Alertas Sugeridos

- Job falhando por N execuções consecutivas
- Duração > percentil 99
- Job não executou no horário esperado
- Panic recovery

## Troubleshooting

### Job não está executando

1. Verificar expressão cron: `https://crontab.guru/`
2. Verificar logs de registro do job
3. Verificar fuso horário do servidor

### Shutdown lento

- Aumentar timeout de shutdown (30s padrão)
- Jobs devem respeitar `ctx.Done()`
- Verificar queries/operações longas

### Panic contínuo

- Verificar stack trace nos logs
- Adicionar tratamento de nil-safety
- Validar dados de entrada

### Concorrência excessiva

- Reduzir `WORKER_MAX_CONCURRENT_JOBS`
- Ajustar schedules para evitar overlap
- Aumentar timeout se jobs são lentos

## Roadmap

Melhorias futuras:
- [ ] Timeout customizado por job
- [ ] Retry automático com backoff
- [ ] Métricas prometheus nativas
- [ ] Health check endpoint
- [ ] Distributed locking (evitar execução duplicada em múltiplas instâncias)
- [ ] Job history/audit trail
- [ ] Dynamic job registration (via API)
- [ ] Cron UI dashboard
