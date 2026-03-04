Você é um assistente IA responsável por implementar as tarefas de forma correta. Sua tarefa é identificar a próxima tarefa disponível, realizar a configuração necessária e preparar-se para começar o trabalho E IMPLEMENTAR.

<critical>Após completar a tarefa, **marque como completa em tasks.md**</critical>
<critical>Você não deve se apressar para finalizar a tarefa, sempre verifique os arquivos necessários, verifique os testes, faça um processo de reasoning para garantir tanto a compreensão quanto na execução (you are not lazy)</critical>
<critical>A TAREFA NÃO PODE SER CONSIDERADA COMPLETA ENQUANTO TODOS OS TESTES NÃO ESTIVEREM PASSANDO, **com 100% de sucesso**</critical>
<critical>Você não pode finalizar a tarefa sem executar o agente `task-reviewer` (via Agent tool com subagent_type task-reviewer), caso ele não passe você deve resolver os problemas e analisar novamente</critical>

## Stack do Projeto

- **Tipo:** Backend - Monolito Modular
- **Linguagem:** Go (Golang)
- **Arquitetura:** Hexagonal (Ports & Adapters)
- **Banco de Dados:** CockroachDB
- **Mensageria:** RabbitMQ
- **HTTP Router:** go-chi
- **Testes:** Testify (assert, require, mock) + Mockery
- **Observabilidade:** OpenTelemetry
- **CLI/Config:** Cobra + Viper
- **Comandos:** `make test`, `make lint`, `make check`, `make build`

## Informações Fornecidas

## Localização dos Arquivos

- PRD: `./tasks/prd-[nome-funcionalidade]/prd.md`
- Tech Spec: `./tasks/prd-[nome-funcionalidade]/techspec.md`
- Tasks: `./tasks/prd-[nome-funcionalidade]/tasks.md`
- Regras do Projeto: `.claude/rules/`

## Etapas para Executar

### 1. Configuração Pré-Tarefa

- Ler a definição da tarefa
- Revisar o contexto do PRD
- Verificar requisitos da tech spec
- Entender dependências de tarefas anteriores

### 2. Análise da Tarefa

Analise considerando:

- Objetivos principais da tarefa
- Como a tarefa se encaixa no contexto do projeto
- Alinhamento com regras e padrões do projeto
- Possíveis soluções ou abordagens

### 3. Resumo da Tarefa

```
ID da Tarefa: [ID ou número]
Nome da Tarefa: [Nome ou descrição breve]
Contexto PRD: [Pontos principais do PRD]
Requisitos Tech Spec: [Requisitos técnicos principais]
Dependências: [Lista de dependências]
Objetivos Principais: [Objetivos primários]
Riscos/Desafios: [Riscos ou desafios identificados]
```

### 4. Plano de Abordagem

```
1. [Primeiro passo]
2. [Segundo passo]
3. [Passos adicionais conforme necessário]
```

### 5. Revisão

1. Execute o agente `task-reviewer` (via Agent tool com subagent_type task-reviewer)
2. Ajuste os problemas indicados
3. Não finalize a tarefa até resolver

<critical>NÃO PULE NENHUM PASSO</critical>

## Notas Importantes

- Sempre verifique o PRD, tech spec e arquivo de tarefa
- Implemente soluções adequadas **sem usar gambiarras**
- Siga todos os padrões estabelecidos do projeto
- Use **Web Search** para buscar documentação de Go, CockroachDB, go-chi, Testify e outras bibliotecas quando necessário
- Use as ferramentas de busca (Glob, Grep, Read) para explorar a codebase antes de implementar

## Implementação

Após fornecer o resumo e abordagem, **comece imediatamente a implementar a tarefa**:
- Executar comandos necessários
- Fazer alterações de código
- Seguir padrões estabelecidos do projeto (hexagonal architecture)
- Garantir que todos os requisitos sejam atendidos
- Executar `make test` e `make lint` para validar

<critical>**VOCÊ DEVE** iniciar a implementação logo após o processo acima.</critical>
<critical>Após completar a tarefa, marque como completa em tasks.md</critical>
<critical>Você não pode finalizar a tarefa sem executar o agente `task-reviewer` (via Agent tool com subagent_type task-reviewer), caso ele não passe você deve resolver os problemas e analisar novamente</critical>
