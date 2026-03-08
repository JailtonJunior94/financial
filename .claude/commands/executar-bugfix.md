Voce e um orquestrador para correcao de bugs via skill `bugfix`.

<critical>Cada bug corrigido precisa de teste de regressao</critical>
<critical>Nao concluir com bugs criticos pendentes no escopo</critical>

## Contexto Obrigatorio
- `.claude/context/stack.md`
- `.claude/context/tooling.md`
- `.claude/context/paths.md`
- `.claude/rules/`

## Entradas
- `tasks/prd-[nome-funcionalidade]/bugs.md`
- Artefatos tecnicos relevantes (`prd.md`, `techspec.md`, `tasks.md`)

## Saida
- `tasks/prd-[nome-funcionalidade]/bugfix_report.md`

## Workflow Deterministico
1. Validar existencia de `bugs.md`.
2. Executar skill `bugfix`.
3. Salvar relatorio em `bugfix_report.md`.
4. Se houver bugs criticos pendentes, estado final `blocked`.
5. Se escopo aprovado estiver corrigido e validado, estado final `done`.

## Mandatory Rules
Este comando deve seguir `.claude/rules/`.
Em conflito, regras prevalecem.
