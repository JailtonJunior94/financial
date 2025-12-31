# CLAUDE.md

Este arquivo define as diretrizes obrigatórias para a IA ao trabalhar neste repositório, garantindo coerência arquitetural, reutilização de código existente e respostas contextualizadas.

---

## Objetivos deste Arquivo

1. Definir regras globais de comportamento da IA
2. Explicar como a IA deve interpretar o contexto do projeto
3. Priorizar código existente antes de sugerir novas abstrações
4. Reduzir respostas genéricas ou fora do padrão do projeto
5. Garantir consistência arquitetural e de estilo
6. Evitar decisões implícitas sem validação do contexto existente

---

## Uso de Contexto (Regras Obrigatórias)

### 1. Leitura de Contexto
- Sempre assumir que **o código existente é a fonte da verdade**
- Nunca sugerir soluções que contradigam:
  - Estrutura de pastas
  - Padrões já adotados
  - Bibliotecas internas (`pkg/*`)
- Antes de criar algo novo, **avaliar se já existe algo semelhante**

### 2. Continuidade de Conversa
- Considerar mensagens anteriores como **contexto ativo**
- Manter consistência entre respostas dentro da mesma conversa
- Não redefinir padrões já estabelecidos sem justificativa técnica clara

### 3. Escopo da Resposta
- Responder **apenas ao que foi solicitado**
- Não antecipar features ou refatorações não pedidas
- Se faltar contexto, **pedir explicitamente**, sem assumir

---

## Diretrizes Arquiteturais

A IA deve assumir que o projeto segue, salvo indicação contrária:

- Clean Architecture / Hexagonal
- Domain-Driven Design (DDD)
- Separação clara entre:
  - handlers / controllers (infrastructure/http)
  - usecases (application)
  - domain (entities, VOs, factories)
  - repositories (interfaces no domain, implementação na infrastructure)
- Comunicação entre módulos bem definida
- Uso intenso de `context.Context`
- Código idiomático em Go (Go 1.20+)

### Princípios Invioláveis
- SOLID
- DRY
- KISS
- Baixo acoplamento e alta coesão

---

## Qualidade de Código Esperada

Toda sugestão de código deve:

- Evitar `panic`
- Ser `nil-safe`
- Evitar race conditions
- Evitar memory leaks
- Passar em:
  - go vet
  - go fmt
  - golangci-lint (configurado em `.golangci.yml`)
- Usar tratamento explícito de erros
- Seguir padrões já existentes no projeto

---

## Segurança e Autenticação

Quando o tema envolver segurança:

- Nunca assumir autenticação inexistente
- Sempre reutilizar middlewares já criados (`pkg/api/middlewares`)
- Não duplicar lógica de auth
- Respeitar propagação de usuário via `context.Context`
- Retornar status HTTP corretos (401, 403, etc.)
- Usar JWT implementado em `pkg/auth/jwt.go`

---

## Reutilização Antes de Criação

Antes de sugerir:
- novos helpers
- novas libs
- novos middlewares
- novos padrões

A IA deve:
1. Procurar algo existente no projeto
2. Avaliar adaptação/extensão
3. Somente então propor algo novo, com justificativa clara

### Bibliotecas Internas Existentes (`pkg/`)
- `pkg/bundle/`: Container de injeção de dependências
- `pkg/database/`: Abstrações de banco, migrações, Unit of Work
- `pkg/auth/`: Implementação JWT
- `pkg/api/`: Middlewares, error handlers, HTTP utilities
- `pkg/custom_errors/`: Tratamento customizado de erros
- `pkg/linq/`: Utilidades para slices

---

## Estilo de Resposta

As respostas da IA devem ser:

- Técnicas
- Objetivas
- Estruturadas (títulos, listas, passos)
- Escritas em **português brasileiro**
- Sem emojis desnecessários
- Sem explicações óbvias
- Sem texto promocional ou genérico

---

## O Que a IA NÃO Deve Fazer

- Não reescrever código inteiro sem pedido explícito
- Não mudar decisões arquiteturais sem análise
- Não inventar dependências
- Não assumir tecnologias não mencionadas
- Não simplificar problemas complexos sem alertar
- Não criar novos padrões quando já existem padrões estabelecidos
- Não sugerir bibliotecas externas sem verificar alternativas internas

---

## Arquitetura do Projeto

### Clean Architecture com Domain-Driven Design

A base de código segue os princípios da Clean Architecture organizados em camadas distintas:

**Camada de Domínio** (`internal/{module}/domain/`)
- `entities/`: Entidades de negócio principais com comportamento (ex: Budget, Category, User)
- `vos/`: Value Objects (ex: CategoryName, Email, Money, Percentage)
- `factories/`: Criação de entidades com lógica de validação
- `interfaces/`: Interfaces de repositório (inversão de dependência)

**Camada de Aplicação** (`internal/{module}/application/`)
- `usecase/`: Regras de negócio específicas da aplicação e orquestração
- `dtos/`: Data Transfer Objects para request/response

**Camada de Infraestrutura** (`internal/{module}/infrastructure/`)
- `repositories/`: Implementações de banco de dados das interfaces do domínio
- `http/`: Handlers HTTP, rotas e transporte
- `mocks/`: Mocks gerados para testes (via mockery)

**Padrão de Registro de Módulos**
Cada módulo de domínio (user, category, budget, card, invoice, payment_method) possui um arquivo `module.go` que:
- Conecta dependências (repositories, use cases, handlers)
- Retorna rotas HTTP para registro
- Encapsula inicialização do módulo

### Container de Injeção de Dependências

O `pkg/bundle/container.go` fornece gerenciamento centralizado de dependências:
- Conexão com banco de dados (`*sql.DB`)
- Configuração (`configs.Config`)
- Adaptadores JWT e hashing
- Telemetria OpenTelemetry (traces, métricas, logs via devkit-go)
- Middlewares compartilhados (auth, panic recovery)

### Padrão Unit of Work

`pkg/database/uow/` implementa Unit of Work para operações transacionais:
- `Executor()`: Retorna executor de DB atual (transaction ou connection)
- `Do(ctx, fn)`: Envolve operações em transação com rollback automático
- Usado principalmente no módulo Budget para operações multi-entidade

### Estratégia de Banco de Dados

- **Banco Principal**: CockroachDB (compatível com PostgreSQL)
- **Suporte**: Drivers MySQL e Postgres disponíveis
- **Migrações**: Usa golang-migrate com nomenclatura Unix timestamp
- **Testes**: Testcontainers para testes de integração (módulos CockroachDB e Postgres)

### Arquitetura do Servidor HTTP

Construído sobre `github.com/JailtonJunior94/devkit-go/pkg/httpserver`:
- Tratamento customizado de erros via `pkg/custom_errors` e `pkg/httperrors`
- Cadeia de middlewares: RequestID, Auth, Panic Recovery, Metrics, Tracing
- Endpoint de health check em `/health` (verifica database)
- Rotas registradas por módulo com middleware opcional

### Tratamento Customizado de Erros

`pkg/custom_errors/custom_errors.go`:
- `CustomError` envolve erros com contexto adicional (mensagem, detalhes)
- `pkg/httperrors/http_errors.go` mapeia erros de domínio para códigos HTTP
- Servidor extrai erro original de CustomError para mapeamento HTTP correto
- Suporta detalhes de erro em respostas JSON

### Autenticação

Autenticação baseada em JWT (`pkg/auth/jwt.go`):
- Geração de token com duração configurável
- Middleware valida tokens e extrai claims do usuário
- Configuração via `AUTH_SECRET_KEY` e `AUTH_TOKEN_DURATION`
- Propagação de dados do usuário via `context.Context`

### Observabilidade

Integração completa com OpenTelemetry via devkit-go:
- Rastreamento distribuído (OTLP gRPC)
- Coleta de métricas
- Logging estruturado
- Configuração via variáveis de ambiente `OTEL_*`
- Middleware instrumenta automaticamente handlers HTTP

### Estratégia de Testes

- Entidades de domínio possuem testes unitários (`*_test.go`)
- Testes de repositório usam testcontainers para cenários reais de banco
- Mocks gerados via mockery para testes de use case
- Relatórios de cobertura gerados com `make cover`

---

## Padrões de Teste

### Padrão AAA (Arrange-Act-Assert)

Todos os testes devem seguir o padrão AAA para clareza e manutenibilidade:

- **Arrange**: Configurar dados de teste, mocks e dependências
- **Act**: Executar a função/método sendo testado
- **Assert**: Verificar os resultados esperados

### Testes Unitários com Mocks

Para testes unitários, sempre usar a configuração `.mockery.yml` e gerar mocks com:

```bash
make mocks
```

Isso garante geração consistente de mocks em toda a base de código.

### Estrutura de Teste com testify/suite

Usar `testify/suite` para organizar testes relacionados com setup compartilhado:

```go
package usecase_test

import (
    "context"
    "errors"
    "testing"

    "github.com/jailtonjunior94/financial/internal/domain/interfaces/mocks"
    "github.com/jailtonjunior94/financial/internal/application/usecase"
    "github.com/jailtonjunior94/financial/internal/application/dtos"

    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/suite"
)

type CreateCategoryUseCaseSuite struct {
    suite.Suite

    ctx                context.Context
    categoryRepository *mocks.CategoryRepository
}

func TestCreateCategoryUseCaseSuite(t *testing.T) {
    suite.Run(t, new(CreateCategoryUseCaseSuite))
}

func (s *CreateCategoryUseCaseSuite) SetupTest() {
    s.ctx = context.Background()
    s.categoryRepository = mocks.NewCategoryRepository(s.T())
}

func (s *CreateCategoryUseCaseSuite) TestExecute() {
    type (
        args struct {
            input *dtos.CreateCategoryInput
        }

        dependencies struct {
            categoryRepository *mocks.CategoryRepository
        }
    )

    scenarios := []struct {
        name         string
        args         args
        dependencies dependencies
        expect       func(output *dtos.CategoryOutput, err error)
    }{
        {
            name: "deve criar uma nova categoria com sucesso",
            args: args{
                input: &dtos.CreateCategoryInput{
                    Name: "Transport",
                },
            },
            dependencies: dependencies{
                categoryRepository: func() *mocks.CategoryRepository {
                    // Arrange: configurar mock
                    s.categoryRepository.
                        EXPECT().
                        Create(s.ctx, mock.Anything).
                        Return(nil).
                        Once()
                    return s.categoryRepository
                }(),
            },
            expect: func(output *dtos.CategoryOutput, err error) {
                // Assert: verificar resultados
                s.NoError(err)
                s.NotNil(output)
                s.Equal("Transport", output.Name)
            },
        },
        {
            name: "deve retornar erro ao falhar ao criar categoria",
            args: args{
                input: &dtos.CreateCategoryInput{
                    Name: "Transport",
                },
            },
            dependencies: dependencies{
                categoryRepository: func() *mocks.CategoryRepository {
                    // Arrange: configurar mock para falhar
                    s.categoryRepository.
                        EXPECT().
                        Create(s.ctx, mock.Anything).
                        Return(errors.New("database error")).
                        Once()
                    return s.categoryRepository
                }(),
            },
            expect: func(output *dtos.CategoryOutput, err error) {
                // Assert: verificar erro
                s.Error(err)
                s.Nil(output)
                s.Contains(err.Error(), "database error")
            },
        },
    }

    for _, scenario := range scenarios {
        s.T().Run(scenario.name, func(t *testing.T) {
            // Act: executar o use case
            uc := usecase.NewCreateCategoryUseCase(scenario.dependencies.categoryRepository)
            output, err := uc.Execute(s.ctx, scenario.args.input)

            // Assert: chamar função de verificação
            scenario.expect(output, err)
        })
    }
}
```

### Princípios-Chave de Testes

1. **Isolamento**: Cada teste deve ser independente e não depender de outros testes
2. **Testes Orientados a Tabelas**: Usar abordagem baseada em cenários para múltiplos casos de teste
3. **Configuração de Mocks**: Configurar mocks dentro de funções de dependência para cada cenário
4. **Nomenclatura Clara**: Usar nomes descritivos de teste em português explicando o comportamento esperado
5. **Cobertura Abrangente**: Testar cenários de sucesso e erro
6. **Verificação de Mocks**: Usar `.Once()` para garantir que mocks sejam chamados exatamente uma vez
7. **Setup e Teardown**: Usar `SetupTest()` para inicialização de testes

### Executando Testes Específicos

```bash
# Executar todos os testes em um pacote
go test -v -count=1 ./internal/category/application/usecase/

# Executar uma suite de testes específica
go test -v -count=1 -run TestCreateCategoryUseCaseSuite ./internal/category/application/usecase/

# Executar um caso de teste específico
go test -v -count=1 -run TestCreateCategoryUseCaseSuite/deve_criar_uma_nova_categoria ./internal/category/application/usecase/
```

---

## Estrutura do Projeto

```
.
├── cmd/
│   ├── main.go           # CLI Cobra (comandos api, migrate, consumers)
│   ├── server/           # Setup do servidor HTTP e conexão de módulos
│   └── .env.example      # Template de variáveis de ambiente
├── internal/
│   ├── user/             # Domínio User (auth, criação)
│   ├── category/         # Domínio Category (hierárquico com children)
│   ├── budget/           # Domínio Budget (com items, validação de porcentagem)
│   ├── card/             # Domínio Card (cartões de crédito)
│   ├── invoice/          # Domínio Invoice (faturas e compras)
│   └── payment_method/   # Domínio PaymentMethod (métodos de pagamento)
├── pkg/
│   ├── bundle/           # Container de injeção de dependências
│   ├── database/         # Abstração de DB, migrações, UoW, helpers de teste
│   ├── auth/             # Implementação JWT
│   ├── api/              # Utilitários HTTP, middlewares, error handlers
│   ├── custom_errors/    # Wrapping e mapeamento de erros
│   └── linq/             # Funções utilitárias para slices
├── configs/              # Carregamento de configuração via Viper
├── database/migrations/  # Migrações SQL (formato Unix timestamp)
└── deployment/           # Docker Compose e configurações de container
```

---

## Configuração

Configuração carregada do `.env` via Viper (ver `cmd/.env.example`):
- Porta HTTP
- Conexão com banco de dados (suporta Postgres/CockroachDB/MySQL)
- Endpoints OpenTelemetry
- Secret JWT e duração do token

---

## Padrões Importantes do Domínio

### Value Objects
Amplamente usados para validação de domínio (Money, Percentage, Email, CategoryName).
Criados via funções factory que enforçam regras de negócio.

### Padrão Repository
Toda persistência abstraída por interfaces na camada de domínio.
Repositories aceitam interface `database.DBExecutor` (suporta `*sql.DB` e `*sql.Tx`).

### Regras do Domínio Budget
- Budget deve ter itens com porcentagens totalizando exatamente 100%
- `AddItem()` e `AddItems()` retornam `bool` indicando se a regra de porcentagem é satisfeita
- AmountUsed e PercentageUsed calculados automaticamente quando itens são adicionados

### Hierarquia de Categorias
Categorias suportam relacionamentos pai-filho via `ParentID *UUID`.
Repository carrega children recursivamente e anexa via `AddChildrens()`.

### Soft Deletes
Entidades usam timestamps `DeletedAt` (via `sharedVos.NullableTime`).
Chamar método `Delete()` da entidade para definir timestamp.

---

## Comandos de Build e Desenvolvimento

### Setup do Ambiente
```bash
make dotenv  # Gerar arquivo .env do .env.example em cmd/
```

### Build
```bash
make build   # Compila para ./bin/financial (CGO_ENABLED=0, otimizado com -ldflags="-w -s")
```

### Testes
```bash
make test-unit           # Executar testes unitários com detecção de race e cobertura
make test-integration    # Executar testes de integração
make test-all           # Executar todos os testes
make cover              # Gerar e visualizar relatório HTML de cobertura
make cover-html         # Gerar relatórios HTML de cobertura
```

### Linting
```bash
make lint    # Executar golangci-lint com configuração .golangci.yml
```

### Mocks
```bash
make mocks   # Gerar mocks usando mockery (configurado em .mockery.yml)
```

### Migrações de Banco de Dados
```bash
make migrate NAME=migration_name  # Criar novos arquivos de migração em database/migrations
```

### Executando a Aplicação
```bash
# Executar servidor API
./bin/financial api

# Executar migrações de banco (CockroachDB)
./bin/financial migrate

# Consumers (ainda não implementado)
./bin/financial consumers
```

### Docker
```bash
make start_minimal  # Iniciar CockroachDB, migration, RabbitMQ e OTEL
make start_docker   # Iniciar todos os serviços
make stop_docker    # Parar todos os serviços
```

---

## Resultado Esperado

Ao seguir este CLAUDE.md, a IA deve:

- Produzir respostas altamente contextualizadas
- Reduzir retrabalho e correções manuais
- Manter coerência entre módulos
- Ajudar na evolução do projeto sem quebrar padrões existentes
- Respeitar a arquitetura e decisões técnicas já tomadas
- Priorizar reutilização sobre criação
- Manter consistência de estilo e qualidade de código
