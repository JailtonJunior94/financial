# Testes

- Rule ID: R-TEST-001
- Severidade: hard para correção/isolamento, guideline para estilo.
- Escopo: Todos os arquivos `*_test.go`.

## Objetivo
Garantir qualidade, estrutura e confiabilidade consistentes nos testes.

## Requisitos

### Classificação (hard)
- `make test` é o gate padrão para testes unitários.
- `make integration-test` é o gate padrão para testes de integração.
- Testes unitários devem cobrir `usecase`, entidades e serviços de domínio.
- Testes de integração devem ser usados para repositórios e adapters que dependem de banco, broker ou infraestrutura real.
- Testes de integração devem usar build tag `integration` e não podem rodar no fluxo padrão de `make test`.

### Framework (guideline)
- Usar testify (`suite`, `require`, `mock`) onde melhorar a consistência.
- Testes unitários no nível de domínio podem usar `testing` padrão + `require`.

### Isolamento e Determinismo (hard)
- Testes não devem compartilhar estado mutável.
- Setup por teste deve resetar dependências.
- Testes não devem depender da ordem de execução.

### Estrutura (guideline)
- Preferir AAA (Arrange, Act, Assert).
- Usar cenários table-driven para lógica com múltiplas variações.
- Nomes de teste devem descrever o comportamento esperado em inglês.

### Mocks e Fakes (guideline)
- Usar mocks gerados por `mockery` com configuração centralizada em `.mockery.yml`.
- Ao mockar dependências de `usecase`, entidades ou serviços de domínio, gerar/reutilizar mocks via `make mocks`.
- Não criar mocks manuais quando a dependência puder ser gerada pelo `mockery.yml`.
- Expectativas de mock devem definir contagem de chamadas quando relevante.

### Cobertura e Execução (hard)
- Use cases e comportamento de domínio devem ser cobertos por testes.
- Repositórios devem ter testes de integração com `testcontainers`.
- Testes de integração devem ficar em arquivos `*_test.go` com `//go:build integration` quando dependerem de infraestrutura real.
- Gate de execução padrão:
  `make test` para unitários e `make integration-test` para integração.

## Proibido
- Testes sem asserções.
- Chamadas reais de rede/banco em testes unitários puros.
- Usar mocks em substituição a testes de integração de repositório.
- Estado mutável compartilhado entre casos de teste.
- Suposições de timing não determinísticas sem controle explícito.
