# Prompt: Geração de PRD - Gestão de Transações e Orçamento

Você é um Product Manager Sênior especialista em sistemas financeiros e arquitetura baseada em eventos.
Seu objetivo é gerar um **Documento de Requisitos do Produto (PRD)** detalhado seguindo rigorosamente o template em `.claude/templates/prd-template.md`.

## Contexto do Objetivo
O sistema precisa de um novo módulo/feature para gerenciar transações financeiras com as seguintes características:
- **Tipos de Transação:** Entrada (Receita) e Saída (Despesa).
- **Classificação:** Categoria (Obrigatória) e Subcategoria (Opcional).
- **Dados:** Data da transação (ISO 8601) e Breve descrição/memo.
- **Ciclo de Vida:** Criação, Edição e Remoção (utilizando Soft Delete).
- **Integração:** Para cada operação no ciclo de vida (Criação, Edição, Remoção), o sistema DEVE disparar um evento para o RabbitMQ para que o módulo de orçamento (Budget) seja atualizado de forma assíncrona para a categoria correspondente.

## Instruções para o PRD
Ao preencher as seções do template, considere:

1.  **Visão Geral:** Explique a importância do registro preciso de transações e como a atualização em tempo quase real do orçamento agrega valor ao usuário.
2.  **Histórias de Usuário:** Crie histórias cobrindo fluxos felizes (criar transação válida) e fluxos de correção (editar valor/categoria e ver o orçamento se ajustar).
3.  **Funcionalidades Core:**
    - Detalhe a validação de obrigatoriedade da categoria.
    - Descreva o comportamento do Soft Delete (não apagar fisicamente, mas marcar como removido).
    - Descreva a regra de negócio para os eventos: qual dado mínimo deve ir no evento (ID, valor anterior, valor novo, categoria, timestamp).
4.  **Restrições Técnicas:** Foque na necessidade de persistência transacional (garantir que a transação salve e o evento seja enviado - Outbox Pattern pode ser sugerido como requisito de confiabilidade).
5.  **Fora de Escopo:** Defina limites claros, como conciliação bancária automática ou orçamentos multi-moeda (se não forem solicitados agora).

## Output Esperado
Gere o conteúdo completo em Markdown, formatado de acordo com o `.claude/templates/prd-template.md`.
