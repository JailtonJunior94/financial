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

### Comportamento de Not Found
- Retornar `(nil, nil)` quando a ausência não é um erro para aquele contrato.
- Detectar e tratar `sql.ErrNoRows` explicitamente em queries de linha única.

### Erro e Observabilidade no Repositório
- Retornar erros brutos de infraestrutura; não converter para erros de domínio aqui.
- Seguir `error-handling.md` e `o11y.md` para comportamento de erro e telemetria.

## Proibido
- Lógica de negócio em handlers/repositórios.
- Dependências circulares.
- SQL bruto em use cases.
- Estado global mutável para fluxo de negócio.
- Pular cleanup de recursos (`rows.Close`, `stmt.Close`).
- Usar tipo concreto de DB diretamente quando abstração compartilhada existe.
