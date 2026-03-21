# Domain-Driven Design

- Rule ID: R-DDD-001
- Severidade: hard
- Escopo: Todo codigo-fonte Go em `internal/*/domain/`, `internal/*/application/`, `internal/*/domain/factories/`.

## Objetivo
Garantir modelagem de dominio rica, protecao de invariantes e eliminar dominio anemico.

## Requisitos

### Entidades (hard)
- DEVEM proteger invariantes no construtor (`New*`) e em todo metodo de mutacao.
- DEVEM expor comportamento (metodos com logica), nao apenas getters/setters.
- Mutacao de estado DEVE ocorrer apenas via metodos da propria entidade.
- Campos DEVEM ser nao exportados; acesso externo via metodos.

```go
// CORRETO
func (t *Transaction) Cancel() {
    t.status = NewTransactionStatus("cancelled")
    t.updatedAt = time.Now()
}

// PROIBIDO — setter anemico
func (t *Transaction) SetStatus(s string) { t.status = s }
```

### Value Objects (hard)
- DEVEM ser imutaveis — sem metodos que alterem estado interno.
- DEVEM se autovalidar no construtor; retornar erro se input invalido.
- DEVEM ser comparados por valor, nao por referencia.
- Nao DEVEM conter identidade (ID).

```go
// CORRETO
func NewEmail(value string) (*Email, error) {
    if value == "" { return nil, errors.New("email is required") }
    if _, err := mail.ParseAddress(value); err != nil {
        return nil, errors.New("invalid email format")
    }
    return &Email{value: value}, nil
}

// PROIBIDO — VO sem validacao
func NewEmail(v string) *Email { return &Email{value: v} }
```

### Aggregates (hard)
- DEVEM ter uma unica entidade raiz (Aggregate Root).
- Entidades filhas DEVEM ser mutadas exclusivamente via metodos do Aggregate Root.
- Invariantes que envolvem filhos DEVEM ser validadas no metodo do Root.
- Repositorios DEVEM persistir e carregar o aggregate inteiro.

```go
// CORRETO — mutacao via aggregate root
func (b *Budget) AddItem(item *BudgetItem) error {
    if b.hasCategoryID(item.CategoryID()) {
        return ErrDuplicateCategory
    }
    if b.TotalPercentageAllocated()+item.Percentage() > 100 {
        return ErrPercentageExceeded
    }
    b.items = append(b.items, item)
    return nil
}

// PROIBIDO — mutacao direta no filho
budget.Items[0].SetPercentage(50)
```

### Factories (hard)
- DEVEM converter input bruto (strings, floats) em tipos seguros do dominio (VOs, entidades).
- DEVEM executar toda validacao e parsing antes de construir a entidade.
- DEVEM retornar aggregate/entidade totalmente construida e valida ou erro.
- Input de factory DEVE ser struct de params, nao parametros posicionais quando > 3 campos.

```go
// CORRETO
func (f *TransactionFactory) Create(params CreateParams) (*Transaction, error) {
    amount, err := NewMoney(params.Amount, params.Currency)
    if err != nil { return nil, err }
    status, err := NewTransactionStatus(params.Status)
    if err != nil { return nil, err }
    return NewTransaction(TransactionParams{Amount: amount, Status: status}), nil
}

// PROIBIDO — parsing no use case
amount, _ := strconv.ParseFloat(input.Amount, 64)
tx := &Transaction{Amount: amount}
```

### Domain Services (hard)
- DEVEM ser stateless — sem estado mutavel entre chamadas.
- DEVEM encapsular logica que nao pertence a uma unica entidade.
- NAO DEVEM acessar repositorios ou infraestrutura diretamente.
- DEVEM receber dependencias como parametros, nao como campos injetados.

### Camada de Application / Use Cases (hard)
- DEVE orquestrar: carregar agregados, invocar metodos de dominio, persistir.
- NAO DEVE conter regras de negocio — delegar para entidades, VOs, factories ou domain services.
- NAO DEVE realizar parsing de tipos primitivos — delegar para factories.
- DEVE ser o unico ponto de chamada a repositorios.

```go
// CORRETO — use case orquestra, dominio decide
func (u *CreateBudgetItemUseCase) Execute(ctx context.Context, input Input) error {
    budget, err := u.repo.FindByID(ctx, input.BudgetID)
    if err != nil { return err }
    item, err := NewBudgetItem(parsedParams)
    if err != nil { return err }
    if err := budget.AddItem(item); err != nil { return err }
    return u.repo.Save(ctx, budget)
}

// PROIBIDO — regra de negocio no use case
if totalPercentage + input.Percentage > 100 { return ErrExceeded }
```

### Direcao de Dependencia (hard)
- Domain NAO DEVE importar application ou infrastructure.
- Application depende de interfaces definidas em domain.
- Infrastructure implementa interfaces de domain.
- Fluxo: infrastructure -> application -> domain.

### Isolamento por Feature / Vertical Slice (hard)
- Cada bounded context em `internal/{module}/` DEVE ser autocontido.
- Comunicacao entre bounded contexts DEVE usar interfaces (providers/ports).
- NAO DEVE haver import direto entre `internal/a/domain` e `internal/b/domain`.
- Tipos compartilhados residem em `pkg/`.

### Fail-Fast (hard)
- Construtores e factories DEVEM retornar erro no primeiro input invalido.
- NAO DEVE existir entidade em estado invalido apos construcao.
- VOs DEVEM rejeitar input invalido imediatamente, sem defaults silenciosos.

### Nomenclatura e Intencao (guideline)
- Metodos de entidade DEVEM expressar intencao de negocio: `Cancel()`, `Approve()`, nao `SetStatus("cancelled")`.
- Nomes de erros de dominio DEVEM comecar com `Err` e descrever a violacao: `ErrDuplicateCategory`, `ErrPercentageExceeded`.
- Factories DEVEM residir em `domain/factories/`.
- VOs DEVEM residir em `domain/vos/`.
- Entidades DEVEM residir em `domain/entities/`.

## Red Flags (Anti-Patterns)

| Anti-Pattern | Sinal | Correcao |
|---|---|---|
| Dominio Anemico | Entidade so tem getters/setters | Adicionar metodos de comportamento |
| Logica no Use Case | `if/else` de regra de negocio em application | Mover para entidade ou domain service |
| Aggregate Vazado | Filho mutado fora do root | Expor metodo no root para a operacao |
| VO sem Validacao | Construtor aceita qualquer valor | Adicionar validacao com erro |
| Factory Bypassada | Use case cria entidade com struct literal | Usar factory ou construtor `New*` |
| Import Cruzado | `internal/a` importa `internal/b/domain` | Usar interface provider em `pkg/` ou no consumidor |
| Parsing no Use Case | `strconv`, `uuid.Parse` em application | Delegar para factory |

## Validacao Final (hard)

Antes de considerar qualquer implementacao de dominio completa:

1. Toda entidade tem construtor `New*` que valida invariantes?
2. Todo VO se autovalida e e imutavel?
3. Todo aggregate protege invariantes de filhos via root?
4. Toda factory converte input bruto em tipos seguros?
5. O use case esta livre de regras de negocio?
6. Nenhum campo exportado permite bypass de invariantes?

Se QUALQUER resposta for NAO, a implementacao esta INCOMPLETA.

## Proibido
- Struct literals para criar entidades fora de testes.
- Campos exportados em entidades e VOs (exceto para serialization tags com campo nao exportado).
- Regras de negocio em handlers, repositorios ou use cases.
- Entidades sem comportamento (dominio anemico).
- Mutacao de filhos de aggregate fora do aggregate root.
- Defaults silenciosos que mascarem input invalido.
- Import direto entre bounded contexts.
