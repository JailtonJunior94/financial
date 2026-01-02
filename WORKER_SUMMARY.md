# Worker Module - Resumo da Implementação

## ✅ Estrutura Criada

```
financial/
├── cmd/
│   ├── main.go (✅ atualizado com comando 'worker')
│   ├── .env.example (✅ atualizado com configs do worker)
│   └── worker/
│       └── worker.go (✅ entry point completo)
│
├── internal/
│   └── worker/
│       └── jobs/
│           ├── database_cleanup_job.go (✅ exemplo DB)
│           └── report_generator_job.go (✅ exemplo RabbitMQ)
│
├── pkg/
│   ├── jobs/
│   │   └── job.go (✅ interface e config)
│   └── scheduler/
│       └── scheduler.go (✅ scheduler com lifecycle)
│
├── configs/
│   └── config.go (✅ atualizado com WorkerConfig)
│
├── WORKER.md (✅ documentação completa)
└── WORKER_SUMMARY.md (✅ este arquivo)
```

## 🎯 Comandos Disponíveis

```bash
# Build
make build
# ou
go build -o bin/financial cmd/main.go

# Executar Worker
./bin/financial worker

# Executar API (existente)
./bin/financial api

# Executar Consumers (existente)
./bin/financial consumers

# Executar Migrações (existente)
./bin/financial migrate
```

## 🔧 Configuração (.env)

```bash
# Adicione ao seu .env:

# Service Names
SERVICE_NAME_API=financial-api
SERVICE_NAME_CONSUMER=financial-consumer
SERVICE_NAME_WORKER=financial-worker

# Worker Configuration
WORKER_DEFAULT_TIMEOUT_SECONDS=300
WORKER_MAX_CONCURRENT_JOBS=10
```

## 📝 Jobs Criados (Exemplos)

### 1. DatabaseCleanupJob
- **Schedule**: Diariamente às 2h (`0 2 * * *`)
- **Função**: Remove registros soft-deleted há mais de 90 dias
- **Demonstra**: Uso de banco de dados, queries, timeout

### 2. ReportGeneratorJob
- **Schedule**: Segundas-feiras às 8h (`0 8 * * 1`)
- **Função**: Gera relatório semanal e publica no RabbitMQ
- **Demonstra**: Integração DB + RabbitMQ, serialização JSON

## 🚀 Como Adicionar um Novo Job

### Passo 1: Criar o Job

Crie `internal/worker/jobs/meu_job.go`:

```go
package jobs

import (
    "context"
    "github.com/jailtonjunior94/financial/pkg/jobs"
    "github.com/JailtonJunior94/devkit-go/pkg/observability"
)

type MeuJob struct {
    o11y observability.Observability
}

func NewMeuJob(o11y observability.Observability) jobs.Job {
    return &MeuJob{o11y: o11y}
}

func (j *MeuJob) Name() string {
    return "meu_job"
}

func (j *MeuJob) Schedule() string {
    return "@hourly" // A cada hora
}

func (j *MeuJob) Run(ctx context.Context) error {
    j.o11y.Logger().Info(ctx, "executando meu job")
    // Sua lógica aqui
    return nil
}
```

### Passo 2: Registrar

Edite `cmd/worker/worker.go` (linha ~118):

```go
jobsToRegister := []pkgjobs.Job{
    jobs.NewDatabaseCleanupJob(dbManager.DB(), o11y),
    jobs.NewReportGeneratorJob(dbManager.DB(), rabbitClient, cfg.RabbitMQConfig.Exchange, o11y),
    jobs.NewMeuJob(o11y), // ← Adicione aqui
}
```

### Passo 3: Build e Execute

```bash
make build
./bin/financial worker
```

## 🛡️ Características Implementadas

### ✅ Graceful Shutdown Completo
- Captura SIGINT/SIGTERM
- Para scheduler
- Aguarda jobs em execução (timeout: 30s)
- Encerra conexões ordenadamente (RabbitMQ → DB → O11y)

### ✅ Recovery Automático
- Dupla proteção contra panics
- Logs detalhados com stack trace
- Job não derruba a aplicação

### ✅ Controle de Concorrência
- Limite configurável por `WORKER_MAX_CONCURRENT_JOBS`
- Skip automático se limite atingido
- Logs de warning

### ✅ Timeout Configurável
- Padrão: 300s (5 minutos)
- Configurável via `WORKER_DEFAULT_TIMEOUT_SECONDS`
- Context cancelado automaticamente

### ✅ Observabilidade Completa
- Logs estruturados (JSON)
- OpenTelemetry traces
- Métricas de duração
- Correlação de contexto

### ✅ Reutilização de Infraestrutura
- Usa mesmas conexões de DB
- Usa mesmo client RabbitMQ
- Usa mesma stack de O11y
- Configurações centralizadas

## 📚 Documentação

- **WORKER.md**: Documentação completa e detalhada
  - Arquitetura
  - Exemplos
  - Boas práticas
  - Troubleshooting
  - Expressões cron

## 🔍 Testando o Worker

### 1. Verificar Compilação

```bash
make build
# Deve compilar sem erros
```

### 2. Testar Startup

```bash
./bin/financial worker
```

**Saída esperada:**
```
INFO  initializing worker service=financial-worker
INFO  database connection established
INFO  rabbitmq initialized exchange=financial.events
INFO  job registered: database_cleanup schedule="0 2 * * *"
INFO  job registered: report_generator schedule="0 8 * * 1"
INFO  starting scheduler jobs_count=2 default_timeout_seconds=300
INFO  worker started successfully jobs_registered=2
```

### 3. Testar Graceful Shutdown

Pressione `Ctrl+C`:

```
INFO  shutdown signal received, initiating graceful shutdown...
INFO  shutting down scheduler...
INFO  cron scheduler stopped
INFO  all running jobs completed
INFO  worker shutdown completed
```

## 🎨 Decisões Arquiteturais

### 1. Separação de Responsabilidades
- **pkg/jobs**: Interface e contratos (reutilizável)
- **pkg/scheduler**: Lógica de agendamento (reutilizável)
- **internal/worker/jobs**: Jobs específicos da aplicação
- **cmd/worker**: Entry point e wiring

### 2. Reutilização vs. Duplicação
- **Reutiliza**: DB, RabbitMQ, O11y, Configs
- **Não duplica**: Lógica de conexão, middlewares
- **Extrai**: Componentes genéricos para `pkg/`

### 3. Graceful Shutdown Robusto
- Ordem clara de shutdown
- Timeouts em cada etapa
- Logs detalhados
- Não bloqueia indefinidamente

### 4. Testabilidade
- Interface clara (Job)
- Dependências injetáveis
- Context-aware
- Fácil criar mocks

### 5. Observabilidade First-Class
- Logs estruturados em toda execução
- Traces automáticos
- Métricas de duração
- Stack traces em panics

## 🚨 Limitações Conhecidas

1. **Sem Distributed Locking**
   - Se múltiplas instâncias do worker rodarem, jobs executarão em paralelo
   - Solução futura: Redis lock, DB advisory locks

2. **Sem Job History**
   - Não há persistência de execuções passadas
   - Solução futura: Tabela de audit trail

3. **Timeout Global**
   - Todos jobs usam mesmo timeout
   - Solução futura: Timeout por job

4. **Sem Retry Automático**
   - Jobs falhados não são retentados automaticamente
   - Cron agendará novamente no próximo ciclo
   - Solução futura: Retry com backoff

## 📊 Próximos Passos Recomendados

### Curto Prazo
1. Criar jobs reais para seu domínio
2. Ajustar schedules conforme necessidade
3. Monitorar logs em produção
4. Configurar alertas de falhas

### Médio Prazo
1. Implementar distributed locking
2. Adicionar job history/audit
3. Criar health check endpoint
4. Métricas prometheus customizadas

### Longo Prazo
1. Dashboard de jobs
2. Registro dinâmico de jobs
3. Job queue (complementar ao cron)
4. Retry policies configuráveis

## 📞 Suporte

- Documentação completa: `WORKER.md`
- Exemplos de código: `internal/worker/jobs/`
- Interface: `pkg/jobs/job.go`
- Scheduler: `pkg/scheduler/scheduler.go`

---

**Status**: ✅ Implementação completa e testada
**Build**: ✅ Compila sem erros
**Documentação**: ✅ Completa e detalhada
**Pronto para produção**: ✅ Sim (com observações sobre distributed locking)
