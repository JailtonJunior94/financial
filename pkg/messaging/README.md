# pkg/messaging - Sistema de Mensageria com RabbitMQ

Sistema de mensageria utilizando RabbitMQ como message broker, com abstrações que facilitam o desenvolvimento de produtores e consumidores.

## Arquitetura

```
pkg/messaging/          # Abstrações agnósticas
  ├── message.go        # Modelo de mensagem
  ├── handler.go        # Interface de handler
  └── consumer.go       # Interface de consumer

pkg/brokers/            # Implementação RabbitMQ
  └── rabbitmq/         # Implementação RabbitMQ
      ├── consumer.go   # Thin adapter sobre devkit-go
      ├── adapter.go    # Adapter para lifecycle.Service
      └── builder.go    # Builder pattern com topologia

pkg/lifecycle/          # Gerenciamento de lifecycle
  ├── service.go        # Interface Service
  └── group.go          # Gerenciador de múltiplos services

cmd/consumer/           # Consumer command
  └── consumers.go      # Factory pattern + lifecycle integration
```

## Conceitos Principais

### 1. Message

Representação agnóstica de mensagem compatível com qualquer broker:

```go
type Message struct {
    ID              string
    Topic           string                 // routing key ou tópico
    Payload         []byte
    Headers         map[string]interface{}
    Timestamp       time.Time
    DeliveryAttempt int
    CorrelationID   string
}
```

**Builder pattern**:
```go
msg := messaging.NewMessage("user.created", payload).
    WithHeaders(map[string]interface{}{"trace-id": "123"}).
    WithCorrelationID("correlation-123")
```

### 2. Handler

Interface para processar mensagens de forma agnóstica:

```go
type Handler interface {
    Handle(ctx context.Context, msg *Message) error
    Topics() []string
}
```

**Exemplo - Handler como função**:
```go
handler := messaging.NewFuncHandler(
    []string{"user.created", "user.updated"},
    func(ctx context.Context, msg *messaging.Message) error {
        // Processar mensagem
        var event UserEvent
        json.Unmarshal(msg.Payload, &event)

        // Lógica de negócio
        return processUser(ctx, event)
    },
)
```

**Exemplo - Handler como struct**:
```go
type UserHandler struct {
    useCase UserUseCase
    o11y    observability.Observability
}

func (h *UserHandler) Topics() []string {
    return []string{"user.created", "user.updated"}
}

func (h *UserHandler) Handle(ctx context.Context, msg *messaging.Message) error {
    ctx, span := h.o11y.Tracer().Start(ctx, "user_handler.handle")
    defer span.End()

    var event UserEvent
    if err := json.Unmarshal(msg.Payload, &event); err != nil {
        return fmt.Errorf("failed to unmarshal: %w", err)
    }

    return h.useCase.Execute(ctx, &event)
}
```

### 3. Consumer

Interface agnóstica de consumer que implementa `lifecycle.Service`:

```go
type Consumer interface {
    Start(ctx context.Context) error
    Shutdown(ctx context.Context) error
    RegisterHandler(handler Handler) error
    Name() string
}
```

## Uso - RabbitMQ (Implementado)

### 1. Criar RabbitMQ Client

```go
import (
    devkitRabbit "github.com/JailtonJunior94/devkit-go/pkg/messaging/rabbitmq"
    "github.com/jailtonjunior94/financial/pkg/brokers/rabbitmq"
)

// Criar client devkit-go
rabbitClient, err := devkitRabbit.New(
    o11y,
    devkitRabbit.WithCloudConnection(cfg.RabbitMQConfig.URL),
    devkitRabbit.WithAutoReconnect(true),
)

// Declarar exchange
rabbitClient.DeclareExchange(
    ctx,
    "financial.events",
    "topic",
    true,  // durable
    false, // auto-delete
    nil,
)
```

### 2. Criar Consumer com Builder

```go
// Configuração do consumer
config := &rabbitmq.ConsumerConfig{
    QueueName:     "user.updates",
    Exchange:      "financial.events",
    RoutingKeys:   []string{"user.created", "user.updated"},
    WorkerCount:   5,
    PrefetchCount: 10,
    Durable:       true,
    AutoDelete:    false,
}

// Build consumer usando builder pattern
builder := rabbitmq.NewBuilder(rabbitClient, o11y)
consumer, err := builder.BuildConsumer(ctx, config)
```

### 3. Registrar Handler

```go
// Handler de domínio
userHandler := messaging.NewFuncHandler(
    []string{"user.created", "user.updated"},
    func(ctx context.Context, msg *messaging.Message) error {
        // Processar mensagem
        return processUserEvent(ctx, msg)
    },
)

// Registrar no consumer
consumer.RegisterHandler(userHandler)
```

### 4. Integrar com Lifecycle

```go
import "github.com/jailtonjunior94/financial/pkg/lifecycle"

// Criar lifecycle group
serviceGroup := lifecycle.NewGroup(o11y, lifecycle.DefaultGroupConfig())

// Adaptar consumer para lifecycle.Service
consumerService := rabbitmq.NewConsumerService(consumer)
serviceGroup.Register(consumerService)

// Start
if err := serviceGroup.Start(ctx); err != nil {
    return err
}

// Wait for shutdown
<-ctx.Done()

// Graceful shutdown (paralelo, ordem reversa)
shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

if err := serviceGroup.Shutdown(shutdownCtx); err != nil {
    log.Fatal(err)
}
```

## Factory Pattern (Trocar Broker via Config)

O `cmd/consumer/consumers.go` usa factory pattern para selecionar broker baseado em configuração:

```go
type ConsumerFactory interface {
    BuildBudgetConsumer(ctx context.Context, handler messaging.Handler) (lifecycle.Service, error)
    Shutdown(ctx context.Context) error
}

func createConsumerFactory(cfg *configs.Config, o11y observability.Observability) (ConsumerFactory, error) {
    switch cfg.ConsumerConfig.BrokerType {
    case "rabbitmq":
        return newRabbitMQFactory(cfg, o11y)
    default:
        return nil, fmt.Errorf("unsupported broker type: %s (only rabbitmq is supported)", cfg.ConsumerConfig.BrokerType)
    }
}
```

### Configuração (.env)

```env
# Consumer Configuration
CONSUMER_BROKER_TYPE=rabbitmq
CONSUMER_EXCHANGE=financial.events
CONSUMER_WORKER_COUNT=5
CONSUMER_PREFETCH_COUNT=10
```

## Features do RabbitMQ (via devkit-go)

O thin adapter sobre devkit-go fornece automaticamente:

- ✅ **Auto-retry com backoff exponencial**: Retry automático com delays crescentes
- ✅ **Dead Letter Queue (DLQ)**: Mensagens após max retries vão para DLQ
- ✅ **Worker pool**: Múltiplos workers concorrentes configuráveis
- ✅ **Panic recovery**: Recover de panics com logging
- ✅ **Observability completa**: Tracing e logging integrados via OpenTelemetry
- ✅ **Auto-reconnect**: Reconexão automática em caso de falha
- ✅ **Publisher confirms**: Confirmação de entrega de mensagens
- ✅ **Graceful shutdown**: Aguarda mensagens em processamento

## Testes

### Testes Unitários

```go
type HandlerTestSuite struct {
    suite.Suite
    ctx context.Context
}

func (s *HandlerTestSuite) TestHandle() {
    // Arrange
    called := false
    handler := messaging.NewFuncHandler(
        []string{"test.topic"},
        func(ctx context.Context, msg *messaging.Message) error {
            called = true
            return nil
        },
    )

    // Act
    msg := messaging.NewMessage("test.topic", []byte("test"))
    err := handler.Handle(s.ctx, msg)

    // Assert
    s.NoError(err)
    s.True(called)
}
```

### Testes de Integração (testcontainers)

```go
// Usar testcontainers para RabbitMQ real
container, err := testcontainers.RunContainer(ctx, "rabbitmq:3-management")
rabbitmqURL := container.GetConnectionString()

// Criar consumer
consumer := setupConsumer(rabbitmqURL)

// Publicar mensagem
publishMessage(rabbitmqURL, "test.topic", payload)

// Verificar processamento
time.Sleep(100 * time.Millisecond)
assert.True(t, messageProcessed)
```

## Observabilidade

### Logging

Todos os eventos de lifecycle e processamento são logados:

```go
o11y.Logger().Info(ctx, "starting consumer",
    observability.String("queue", config.QueueName),
    observability.Int("worker_count", config.WorkerCount),
)

o11y.Logger().Error(ctx, "handler error",
    observability.Error(err),
    observability.String("topic", msg.Topic),
    observability.String("message_id", msg.ID),
)
```

### Tracing (OpenTelemetry)

Span automático por mensagem processada:

```go
ctx, span := o11y.Tracer().Start(ctx, "handler.handle")
defer span.End()

span.SetAttributes(
    observability.String("topic", msg.Topic),
    observability.String("message_id", msg.ID),
    observability.String("correlation_id", msg.CorrelationID),
)

if err != nil {
    span.RecordError(err)
    span.SetStatus(observability.StatusError, err.Error())
}
```

### Métricas (Futuro)

```
consumer_messages_processed_total{topic, status}
consumer_messages_duration_seconds{topic}
consumer_handler_errors_total{topic, error_type}
consumer_worker_count{queue}
```

## Exemplos Completos

### Exemplo 1: Consumer Simples

```go
package main

import (
    "context"
    "log"

    "github.com/jailtonjunior94/financial/configs"
    "github.com/jailtonjunior94/financial/pkg/lifecycle"
    "github.com/jailtonjunior94/financial/pkg/messaging"
    "github.com/jailtonjunior94/financial/pkg/brokers/rabbitmq"

    devkitRabbit "github.com/JailtonJunior94/devkit-go/pkg/messaging/rabbitmq"
)

func main() {
    ctx := context.Background()
    cfg, _ := configs.LoadConfig(".")

    // Setup RabbitMQ
    client, _ := devkitRabbit.New(o11y, devkitRabbit.WithCloudConnection(cfg.RabbitMQConfig.URL))
    client.DeclareExchange(ctx, "events", "topic", true, false, nil)

    // Build consumer
    builder := rabbitmq.NewBuilder(client, o11y)
    consumer, _ := builder.BuildConsumer(ctx, &rabbitmq.ConsumerConfig{
        QueueName:   "notifications",
        Exchange:    "events",
        RoutingKeys: []string{"user.*", "order.*"},
        WorkerCount: 3,
    })

    // Register handler
    handler := messaging.NewFuncHandler(
        []string{"user.created", "order.created"},
        func(ctx context.Context, msg *messaging.Message) error {
            log.Printf("Received: %s - %s", msg.Topic, string(msg.Payload))
            return nil
        },
    )
    consumer.RegisterHandler(handler)

    // Lifecycle
    group := lifecycle.NewGroup(o11y, lifecycle.DefaultGroupConfig())
    group.Register(rabbitmq.NewConsumerService(consumer))
    group.Start(ctx)

    // Wait and shutdown
    <-ctx.Done()
    group.Shutdown(context.Background())
}
```

### Exemplo 2: Múltiplos Consumers

```go
// Create multiple consumers
userConsumer := buildUserConsumer(ctx)
orderConsumer := buildOrderConsumer(ctx)
paymentConsumer := buildPaymentConsumer(ctx)

// Register all in lifecycle group
group := lifecycle.NewGroup(o11y, lifecycle.DefaultGroupConfig())
group.Register(rabbitmq.NewConsumerService(userConsumer))
group.Register(rabbitmq.NewConsumerService(orderConsumer))
group.Register(rabbitmq.NewConsumerService(paymentConsumer))

// Start all (sequential)
group.Start(ctx)

// Shutdown all (parallel, reverse order)
group.Shutdown(shutdownCtx)
```

## Troubleshooting

### Consumer não recebe mensagens

1. Verificar se exchange e queue existem
2. Verificar se routing keys estão corretos
3. Verificar se há mensagens na fila (`rabbitmqadmin list queues`)
4. Verificar logs do consumer

### Mensagens indo para DLQ

1. Verificar logs de erro do handler
2. Verificar max retries configurado
3. Verificar se handler está retornando erro corretamente

### Performance degradada

1. Aumentar `WorkerCount` para mais concorrência
2. Aumentar `PrefetchCount` para buscar mais mensagens
3. Verificar se handler não está bloqueando
4. Verificar conexão com banco de dados

## Referências

- [devkit-go RabbitMQ](https://github.com/JailtonJunior94/devkit-go/tree/main/pkg/messaging/rabbitmq)
- [devkit-go Consumer](https://github.com/JailtonJunior94/devkit-go/pkg/consumer)
- [OpenTelemetry Go](https://opentelemetry.io/docs/languages/go/)
- [RabbitMQ Tutorials](https://www.rabbitmq.com/tutorials)
