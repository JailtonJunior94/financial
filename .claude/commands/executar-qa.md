Voce e um orquestrador para validacao funcional e contratual usando a skill `qa`.

<critical>Nao concluir sem evidencias objetivas de validacao</critical>
<critical>Status final deve ser consistente com o parecer da skill `qa`</critical>

## Contexto Obrigatorio
- `.claude/context/stack.md`
- `.claude/context/tooling.md`
- `.claude/context/paths.md`
- `.claude/rules/`

## Entradas
- `tasks/prd-[nome-funcionalidade]/prd.md`
- `tasks/prd-[nome-funcionalidade]/techspec.md`
- `tasks/prd-[nome-funcionalidade]/tasks.md`
- task alvo

## Saida
- `tasks/prd-[nome-funcionalidade]/qa_report.md`

## Workflow Deterministico
1. Validar disponibilidade dos artefatos de entrada.
2. Executar skill `qa`.
3. Salvar relatorio em `qa_report.md`.
4. Se resultado for `REJECTED`, estado final `blocked`.
5. Se resultado for `APPROVED`, estado final `done`.

## Mandatory Rules
Este comando deve seguir `.claude/rules/`.
Em conflito, regras prevalecem.
