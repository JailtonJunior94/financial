# Regras para Agentes de IA

Este diretorio centraliza regras para uso com agentes de IA em tarefas reais de analise, alteracao e validacao de codigo.

## Objetivo

Use estas instrucoes para manter consistencia, seguranca e qualidade ao trabalhar com codigo, configuracao, validacao e evolucao de sistemas.

## Arquitetura: monolito modular

O projeto aparenta ser um monolito modular, com separacao relevante por modulos, dominios ou componentes internos. A governanca deve proteger essas fronteiras e evitar dependencias circulares.

Stack detectada: Go.
Frameworks detectados: Chi, gRPC, Cobra.

## Estrutura de Pastas

```
.
cmd/                              # Pontos de entrada (Cobra CLI: api, consumer, worker, migrate)
  main.go
  server/server.go                # Composicao e bootstrap do servidor HTTP
  consumer/consumers.go           # Bootstrap dos consumers de mensageria
  worker/worker.go                # Bootstrap do worker (cron/outbox)
  grpc/                           # Ponto de entrada gRPC (reservado)
  http/                           # Ponto de entrada HTTP alternativo (reservado)
  soap/                           # Ponto de entrada SOAP (reservado)
configs/
  config.go                       # Configuracao centralizada via Viper
internal/                         # Modulos de dominio (bounded contexts)
  budget/                         # Modulo: orcamento domestico
    domain/                       # Entidades, VOs, interfaces, factories, erros
    application/                  # DTOs, use cases
    infrastructure/               # HTTP handlers, repositories, messaging, adapters
    module.go                     # Composicao do modulo (wiring de dependencias)
    errormappings.go              # Mapeamento de erros de dominio para HTTP
  card/                           # Modulo: cartoes de credito
    domain/ application/ infrastructure/ module.go errormappings.go
  category/                       # Modulo: categorias e subcategorias
    domain/ application/ infrastructure/ module.go errormappings.go
  invoice/                        # Modulo: faturas
    domain/ application/ infrastructure/ module.go errormappings.go
  payment_method/                 # Modulo: metodos de pagamento
    domain/ application/ infrastructure/ module.go errormappings.go
  transaction/                    # Modulo: transacoes financeiras
    domain/ application/ infrastructure/ module.go errormappings.go
  user/                           # Modulo: usuarios e autenticacao
    domain/ application/ infrastructure/ module.go errormappings.go
pkg/                              # Pacotes compartilhados (cross-cutting concerns)
  api/                            # HTTP helpers, middlewares, error handlers
  auth/                           # JWT adapter e validacao de tokens
  constants/                      # Constantes globais
  custom_errors/                  # Erros customizados compartilhados
  database/                       # Database manager e conexao
  domain/                         # Interfaces e VOs compartilhados entre modulos
  helpers/                        # Utilitarios genericos
  jobs/                           # Abstracoes de jobs
  lifecycle/                      # Gerenciamento de ciclo de vida da aplicacao
  messaging/                      # Abstracoes de mensageria (consumer, handler)
  money/                          # Operacoes monetarias
  observability/                  # Observabilidade e metricas customizadas
  outbox/                         # Transactional outbox pattern
  pagination/                     # Paginacao por cursor
  scheduler/                      # Agendamento de tarefas (cron)
  validation/                     # Validacao de entrada
database/
  migrations/                     # SQL migrations (golang-migrate, formato unix)
deployment/
  Dockerfile                      # Build multi-stage (Go 1.26 + Alpine)
  docker-compose.yml              # Stack completa (API, Consumer, Worker, CockroachDB, RabbitMQ, OTEL)
  financial/                      # Helm chart para Kubernetes
  telemetry/                      # Configuracao de telemetria (Grafana, OTEL Collector)
docs/                             # Swagger/OpenAPI, Postman collections, dashboards Grafana
tasks/                            # PRDs e tarefas de implementacao por feature
.github/
  agents/                         # Wrappers de agentes para GitHub Copilot
  skills/                         # Skills procedurais (fonte de verdade)
  copilot-instructions.md         # Instrucoes para GitHub Copilot CLI
  pull_request_template.md        # Template de PR
```

## Padrao Arquitetural

Clean Architecture por modulo: cada bounded context em `internal/` segue a separacao `domain/` → `application/` → `infrastructure/`, com composicao explicitada em `module.go`. O `pkg/` agrupa cross-cutting concerns compartilhados entre modulos sem conter logica de dominio.

### Fluxo de Dependencias

- `domain/` e o nucleo: define entidades, value objects, interfaces (ports) e erros. Nao importa nada de `application/`, `infrastructure/` ou `pkg/`.
- `application/` orquestra use cases consumindo interfaces de `domain/`. Nao conhece detalhes de transporte ou persistencia.
- `infrastructure/` implementa os contratos de `domain/` (repositories, adapters, messaging) e expoe HTTP handlers. Depende de `domain/` e `application/`.
- `module.go` faz a composicao (wiring) de todas as dependencias do modulo e expoe routers e consumers prontos para registro.
- `pkg/` fornece contratos e utilitarios transversais. Modulos podem depender de `pkg/`, mas `pkg/` nao importa `internal/`.
- Comunicacao entre modulos ocorre via interfaces (adapters) — ex: `invoice` expoe `InvoiceProviderAdapter` consumido por `transaction`.

## Modo de trabalho

1. Entender o contexto antes de editar qualquer arquivo.
2. Preferir a menor mudanca segura que resolva a causa raiz.
3. Preservar arquitetura, convencoes e fronteiras ja existentes no contexto analisado.
4. Nao introduzir abstracoes, camadas ou dependencias sem demanda concreta.
5. Atualizar ou adicionar testes quando houver mudanca de comportamento.
6. Rodar validacoes proporcionais a mudanca.
7. Registrar bloqueios e suposicoes explicitamente quando o contexto estiver incompleto.

## Diretrizes de Estrutura

1. Priorize entendimento do codigo e do contexto atual antes de propor refatoracoes.
2. Respeite padroes existentes de nomenclatura, organizacao e tratamento de erro.
3. Defina estrutura simples, evolutiva e com defaults explicitos.
4. Evite reescritas amplas quando uma alteracao localizada resolver o problema.
5. Estabeleca contratos, testes e comandos de validacao cedo quando eles ainda nao existirem.
6. Considere risco de regressao como restricao principal.
7. Evite overengineering disfarcado de arquitetura futura.

## Regras por Arquitetura

1. Respeitar fronteiras entre modulos e bounded contexts.
2. Evitar dependencia circular entre packages internos.
3. Nao extrair shared helpers sem demanda comprovada de mais de um modulo.

## Regras por Linguagem

Para tarefas que alteram codigo, carregar a skill:

- `.agents/skills/agent-governance/SKILL.md`

Para tarefas que alteram codigo Go, carregar tambem:

- `.agents/skills/go-implementation/SKILL.md`

Para tarefas de revisao ou refatoracao incremental de design em Go guiadas por heuristicas de object calisthenics, carregar tambem:

- `.agents/skills/object-calisthenics-go/SKILL.md`

Para tarefas de correcao de bugs com remediacao e teste de regressao, carregar tambem:

- `.agents/skills/bugfix/SKILL.md`

### Composicao Multi-Linguagem

Em projetos com mais de uma linguagem (ex: monorepo Go + Node), carregar apenas a skill da linguagem afetada pela mudanca. Se a tarefa cruzar linguagens, carregar ambas e aplicar a validacao de cada stack nos arquivos correspondentes. Nao misturar convencoes de uma linguagem em arquivos de outra.

## Referencias

Cada skill lista suas proprias referencias em `references/` com gatilhos de carregamento no respectivo `SKILL.md`. Nao duplicar a listagem aqui — consultar o SKILL.md da skill ativa para saber quais referencias carregar e em que condicao.

## Validacao

Antes de concluir uma alteracao:

Seguir Etapa 4 de `.agents/skills/agent-governance/SKILL.md` como base canonica.

Comandos detectados no projeto (Go):
1. Rodar fmt: `go fmt ./...`.
2. Rodar vet: `go vet ./...`.
3. Rodar test unitario: `go test -short -count=1 -race ./...`.
4. Rodar test integracao: `go test -tags=integration -count=1 -race ./...`.
5. Rodar lint: `GOGC=20 golangci-lint run --config .golangci.yml ./...`.
6. Gerar mocks: `mockery` (quando interfaces forem alteradas).
7. Gerar docs: `swag init -g cmd/main.go --output docs --parseInternal` (quando anotacoes Swagger forem alteradas).

## Restricoes

1. Nao inventar contexto ausente.
2. Nao assumir versao de linguagem, framework ou runtime sem verificar.
3. Nao alterar comportamento publico sem deixar isso explicito.
4. Nao usar exemplos como copia cega; adaptar ao contexto real.
