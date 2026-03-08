Voce e um orquestrador para revisao tecnica de alteracoes usando a skill `code-reviewer`.

<critical>Nao aprovar quando houver violacao de regra hard</critical>
<critical>Somente concluir com relatorio de review salvo</critical>

## Contexto Obrigatorio
- `.claude/context/stack.md`
- `.claude/context/tooling.md`
- `.claude/context/paths.md`
- `.claude/rules/`

## Entradas
- Diff das alteracoes
- Arquivos impactados

## Saida
- `tasks/prd-[nome-funcionalidade]/review_report.md`

## Workflow Deterministico
1. Ler regras aplicaveis em `.claude/rules/`.
2. Executar skill `code-reviewer`.
3. Salvar o relatorio em `review_report.md` com veredito e findings.
4. Se veredito for `REJECTED`, estado final `blocked`.
5. Se veredito for `APPROVED` ou `APPROVED_WITH_REMARKS`, estado final `done`.

## Mandatory Rules
Este comando deve seguir `.claude/rules/`.
Em conflito, regras prevalecem.
