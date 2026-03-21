# Arquitetura e Persistência

- Rule ID: R-ARCH-001
- Severidade: hard
- Escopo: Todo código-fonte Go em `internal/*/`.

## Objetivo
Garantir design Go idiomático, Clean Architecture, fronteiras DDD e implementação consistente de repositórios.

## Requisitos

### Estrutura de Bounded Context
Cada bounded context deve residir em `internal/{module}/` com domain, application, infrastructure e wiring do módulo.

### Responsabilidades por Camada
- Domain: entidades, VOs, factories, interfaces, erros de domínio.
- Application: orquestra use cases, sem regras de negócio core.
- Infrastructure: transporte/persistência/adapters, sem regras de negócio.
- Wiring do módulo: composition root e setup de dependências.

### Direção de Dependência
- Deve apontar para dentro: infrastructure -> application -> domain.
- Domain não deve importar application/infrastructure.
- Application depende de interfaces do domain.
- Infrastructure implementa interfaces do domain.

### Contexto e Interfaces
- Métodos públicos devem aceitar `context.Context` como primeiro parâmetro.
- Interfaces devem estar no pacote consumidor e permanecer focadas.

### Modelagem de Domínio
- Value objects se autovalidam e são imutáveis por design.
- Entidades protegem invariantes na criação e mutação.
- Factories fazem a ponte entre input bruto e tipos seguros do domínio.

### Estrutura de Repositório
- Implementação do repositório deve ser struct privada implementando interface do domínio.
- Construtor deve aceitar `database.DBTX` e provedor de observabilidade.
- Construtor deve retornar o tipo da interface do domínio.

### Acesso a Banco de Dados
- Usar abstração `database.DBTX`.
- Segurança SQL: ver R-SEC-001.
- Usar chamadas de DB com context.
- Fechar rows/statements e verificar `rows.Err()`.
- **(hard)** Ver seção "Defer Close com Observabilidade" abaixo para regra de cleanup de recursos.

### Comportamento de Not Found
- Retornar `(nil, nil)` quando a ausência não é um erro para aquele contrato.
- Detectar e tratar `sql.ErrNoRows` explicitamente em queries de linha única.

### Defer Close com Observabilidade (hard)
Todo recurso que implemente `io.Closer` ou retorne erro no `Close` deve ter o erro capturado e registrado via o11y. Aplica-se a, mas não se limita a:
- `sql.Rows`, `sql.Stmt` (banco de dados)
- `http.Response.Body` (chamadas HTTP externas)
- `amqp.Channel`, `amqp.Connection` (RabbitMQ)
- Qualquer `io.ReadCloser`, `io.WriteCloser` ou recurso similar

Padrão obrigatório:
```go
defer func() {
    if closeErr := resource.Close(); closeErr != nil {
        span.RecordError(closeErr)
        r.o11y.Logger().Error(ctx, "{MethodName}: failed to close {resource}",
            observability.Error(closeErr),
        )
    }
}()
```

- `{MethodName}`: nome do método onde o defer está (ex.: `ListPaginated`, `FindByID`).
- `{resource}`: tipo do recurso sendo fechado (ex.: `rows`, `response body`, `channel`).
- Quando não houver span disponível, registrar apenas via logger.
- Nunca usar `_ = resource.Close()` ou `defer resource.Close()` sem captura de erro.

### Erro e Observabilidade no Repositório
- Retornar erros brutos de infraestrutura; não converter para erros de domínio aqui.
- Seguir `error-handling.md` e `o11y.md` para comportamento de erro e telemetria.

### Design Patterns (guideline)
Referência canônica: https://refactoring.guru/design-patterns

Usar design patterns quando simplificarem o design ou resolverem problema recorrente. Não forçar pattern onde código direto resolve. Abaixo estão os patterns aplicáveis a este projeto, organizados por categoria.

#### Creational Patterns

**Factory Method / Abstract Factory**
- Já em uso: `internal/*/domain/factories/`.
- Usar para converter input bruto em entidades/VOs validados.
- Preferir factory function (`CreateCard()`) para criação simples e factory struct (`TransactionFactory`) quando houver dependências injetadas.

**Builder**
- Usar quando construção de objeto exigir múltiplos passos opcionais ou configurações complexas.
- Preferir params struct para até 5 campos; Builder para objetos com muitas combinações opcionais.
```go
// CORRETO — Builder para configuração complexa
invoice := NewInvoiceBuilder().
    WithCard(card).
    WithItems(items).
    WithDueDate(dueDate).
    Build()

// DESNECESSÁRIO — Builder para objeto simples (usar params struct)
category := NewCategoryBuilder().WithName(name).Build() // overengineering
```

**Singleton**
- Evitar singleton clássico com `sync.Once` para serviços de negócio.
- Instâncias únicas devem ser gerenciadas pelo composition root (`module.go`, `server.go`) via injeção de dependência.

#### Structural Patterns

**Adapter**
- Já em uso: `internal/*/infrastructure/adapters/`.
- Usar para conectar bounded contexts sem acoplamento direto.
- Adapter DEVE implementar interface do consumidor e delegar para repositório/serviço do módulo de origem.
```go
// CORRETO — adapter entre bounded contexts
type cardProviderAdapter struct {
    repo card.CardRepository
}
func (a *cardProviderAdapter) FindCardByID(ctx context.Context, id uuid.UUID) (*Card, error) {
    return a.repo.FindByID(ctx, id)
}
```

**Decorator**
- Já em uso: middlewares HTTP em `pkg/api/middlewares/`.
- Usar para adicionar comportamento cross-cutting (auth, métricas, logging) sem alterar o handler original.
- Decorators DEVEM preservar a assinatura da interface original.

**Composite**
- Considerar quando objetos formarem hierarquias tree-like que devem ser tratadas uniformemente.
- Aplicável a: árvore de categorias, regras compostas de validação.

**Facade**
- Usar para simplificar subsistemas complexos de infraestrutura (ex.: wrapper sobre SDK externo).
- Facade NÃO DEVE conter lógica de negócio — apenas simplificar chamadas.

#### Behavioral Patterns

**Strategy**
- Já em uso via interfaces de domínio com implementações intercambiáveis.
- Usar quando comportamento precisar variar em runtime sem alterar o consumidor.
- Cada strategy DEVE implementar interface focada (ISP).
```go
// CORRETO — strategy via interface
type NotificationSender interface {
    Send(ctx context.Context, msg Notification) error
}
// Implementações: EmailSender, SMSSender, PushSender
// Use case recebe interface, não implementação concreta
```

**Observer / Event-Driven**
- Já em uso: transactional outbox (`pkg/outbox/`) + consumers (`infrastructure/messaging/`).
- Usar para comunicação assíncrona entre bounded contexts.
- Eventos DEVEM ser definidos em `domain/events/` do módulo publicador.
- Consumidores DEVEM ser idempotentes.

**Command**
- Já em uso: use cases com método `Execute()` e messaging handlers com `Handle()`.
- Cada operação encapsulada como struct com dependências injetadas e método único de execução.
- Usar para operações que precisam ser enfileiradas, logadas ou revertidas.

**Chain of Responsibility**
- Já em uso: middleware chains via chi router groups.
- Usar para pipelines de processamento onde cada etapa pode tratar ou delegar.
- Cada handler na cadeia DEVE ter responsabilidade única.

**Template Method**
- Usar quando repositórios ou serviços seguirem fluxo comum com passos variáveis.
- Em Go, preferir composição com funções/closures ao invés de herança via embedding.
```go
// CORRETO — template via closure em Go
func withSpanAndLog(ctx context.Context, op string, fn func(ctx context.Context) error) error {
    ctx, span := tracer.Start(ctx, op)
    defer span.End()
    if err := fn(ctx); err != nil {
        span.RecordError(err)
        return err
    }
    return nil
}
```

**State**
- Considerar quando entidade tiver transições de estado complexas com regras por estado.
- Implementar como VO de status com métodos que validam transições permitidas.
```go
// CORRETO — transição validada no VO
func (s InvoiceStatus) CanTransitionTo(target InvoiceStatus) bool {
    allowed := transitions[s]
    return slices.Contains(allowed, target)
}
```

#### Quando Usar e Quando Evitar

| Situação | Pattern Recomendado |
|---|---|
| Converter input bruto em entidade | Factory |
| Conectar bounded contexts | Adapter |
| Comportamento cross-cutting HTTP | Decorator / Middleware |
| Comunicação assíncrona entre módulos | Observer (Outbox + Consumer) |
| Múltiplas implementações de contrato | Strategy |
| Construção com muitas opções | Builder |
| Transições de estado complexas | State (via VO) |
| Pipeline de processamento sequencial | Chain of Responsibility |

**Regra de ouro**: se o pattern não simplifica o código atual ou não resolve um problema concreto de extensibilidade, NÃO usar. Três linhas diretas são melhores que uma abstração prematura.

## Proibido
- Lógica de negócio em handlers/repositórios.
- Dependências circulares.
- SQL bruto em use cases.
- Estado global mutável para fluxo de negócio.
- Pular cleanup de recursos (`rows.Close`, `stmt.Close`).
- Engolir erros de `Close()` com `_ =` ou `defer resource.Close()` sem captura — erros de close devem ser registrados via o11y.
- Usar tipo concreto de DB diretamente quando abstração compartilhada existe.
