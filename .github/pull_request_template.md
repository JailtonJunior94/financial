# Pull Request

## 📋 Descrição

<!-- Descreva brevemente as mudanças deste PR -->

## 🔗 Issue Relacionada

<!-- Se aplicável, link para a issue: Closes #123 -->

## 🎯 Tipo de Mudança

<!-- Marque o tipo de mudança -->

- [ ] 🐛 Bug fix (correção de bug)
- [ ] ✨ New feature (nova funcionalidade)
- [ ] 💥 Breaking change (mudança que quebra compatibilidade)
- [ ] 📚 Documentation (apenas documentação)
- [ ] ♻️ Refactoring (refatoração sem mudança de comportamento)
- [ ] ⚡ Performance (melhoria de performance)
- [ ] ✅ Test (adição ou correção de testes)
- [ ] 🔧 Chore (configuração, dependências, etc)

## ✅ Checklist

### Geral

- [ ] Código segue os padrões do projeto (Go idiomático, Clean Architecture)
- [ ] Fiz self-review do meu código
- [ ] Código está comentado em partes complexas (quando necessário)
- [ ] Fiz as mudanças correspondentes na documentação
- [ ] Minhas mudanças não geram novos warnings
- [ ] Testes unitários passam localmente (`go test ./...`)
- [ ] Lint passa sem erros (`golangci-lint run`)

### Error Handling (se aplicável)

**Se este PR adiciona/modifica tratamento de erros:**

- [ ] ✅ Novos erros de domínio foram adicionados a `errors.go` (não usei `fmt.Errorf()` genérico)
- [ ] ✅ Erros foram mapeados em `pkg/api/httperrors/error_mapping.go`
- [ ] ✅ Usei erros de domínio em vez de `fmt.Errorf()` para cenários de negócio
- [ ] ✅ Erros "not found" retornam **404**, não 500
- [ ] ✅ Erros de validação retornam **400**, não 500
- [ ] ✅ Conflitos de estado retornam **409**, não 500
- [ ] ✅ Usei `err == sql.ErrNoRows` ou `errors.Is()` em vez de comparação de string
- [ ] ✅ Repositórios retornam `nil, nil` para not found (não erro de domínio)
- [ ] ✅ Use cases convertem `nil` para erro de domínio apropriado
- [ ] ✅ Teste de integração verifica status HTTP correto
- [ ] ✅ Li o guia: [`docs/ERROR_HANDLING_GUIDE.md`](../docs/ERROR_HANDLING_GUIDE.md)

**Exemplo do que NÃO fazer:**
```go
❌ if user == nil { return fmt.Errorf("user not found") }  // Retorna 500
✅ if user == nil { return domain.ErrUserNotFound }        // Retorna 404
```

### RESTful API (se aplicável)

**Se este PR adiciona/modifica endpoints:**

- [ ] Endpoints seguem convenções RESTful (GET, POST, PUT, DELETE)
- [ ] Status codes estão corretos:
  - `200 OK` - Sucesso (GET, PUT)
  - `201 Created` - Recurso criado (POST)
  - `204 No Content` - Sucesso sem body (DELETE)
  - `400 Bad Request` - Validação de entrada
  - `401 Unauthorized` - Autenticação necessária
  - `404 Not Found` - Recurso não existe
  - `409 Conflict` - Conflito de estado/constraint
  - `500 Internal Server Error` - Erro técnico inesperado
- [ ] Respostas de erro seguem RFC 7807 (ProblemDetail)
- [ ] Endpoints são idempotentes quando apropriado (GET, PUT, DELETE)

### Database (se aplicável)

**Se este PR modifica schema ou queries:**

- [ ] Migration criada (up + down)
- [ ] Índices apropriados adicionados
- [ ] Queries otimizadas (sem N+1, selects desnecessários)
- [ ] Usei `sql.ErrNoRows` corretamente (não string comparison)
- [ ] Transações são usadas quando necessário (múltiplas operações)

### Testing

- [ ] Adicionei testes que provam que minha correção funciona / nova feature funciona
- [ ] Testes unitários cobrem casos de sucesso
- [ ] Testes cobrem casos de erro (validação, not found, conflito)
- [ ] Testes de integração verificam status HTTP corretos
- [ ] Cobertura de testes mantida ou melhorada

### Performance & Security

- [ ] Não introduzi N+1 queries
- [ ] Validei entrada do usuário (prevent XSS, SQL injection, etc)
- [ ] Secrets não estão hardcoded
- [ ] Logs não expõem dados sensíveis (senhas, tokens, PII)
- [ ] Rate limiting considerado (se endpoint público)

## 🧪 Como Testar

<!-- Descreva como revisar/testar as mudanças -->

**Passos:**
1.
2.
3.

**Comandos:**
```bash
# Exemplo
go test ./internal/budget/...
curl -X GET http://localhost:8080/api/v1/budgets/{id}
```

## 📸 Screenshots (se aplicável)

<!-- Se mudanças visuais, adicione screenshots -->

## 📝 Notas Adicionais

<!-- Informações adicionais para os reviewers -->

## ⚠️ Breaking Changes (se aplicável)

<!-- Descreva breaking changes e migration path -->

---

**Checklist do Reviewer:**

- [ ] Código está limpo e legível
- [ ] Lógica de negócio está correta
- [ ] Tratamento de erros está adequado
- [ ] Testes são suficientes
- [ ] Performance é aceitável
- [ ] Segurança foi considerada
- [ ] Documentação está atualizada
