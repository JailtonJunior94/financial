# Middlewares de Autenticação

Este pacote contém middlewares HTTP seguros e reutilizáveis para autenticação de requisições.

## Middleware de Autenticação

O middleware `Authorization` valida tokens Bearer e injeta o usuário autenticado no contexto da requisição.

### Características

- ✅ Extração segura de Bearer Token do header `Authorization`
- ✅ Validação robusta do formato do token
- ✅ Interface `TokenValidator` para desacoplamento de implementações
- ✅ Propagação via `context.Context` com tipos seguros
- ✅ Logs estruturados (sem expor dados sensíveis)
- ✅ Mensagens de erro claras e seguras
- ✅ Cobertura completa de testes

### Uso Básico

```go
package main

import (
    "github.com/jailtonjunior94/financial/pkg/api/middlewares"
    "github.com/jailtonjunior94/financial/pkg/auth"
    "github.com/JailtonJunior94/devkit-go/pkg/observability"
)

func main() {
    // 1. Criar um TokenValidator (ex: JwtAdapter)
    jwtAdapter := auth.NewJwtAdapter(config, observability)

    // 2. Criar o middleware de autenticação
    authMiddleware := middlewares.NewAuthorization(jwtAdapter, observability)

    // 3. Aplicar o middleware nas rotas protegidas
    router.With(authMiddleware.Authorization).Post("/api/v1/protected", handler)
}
```

### Recuperando o Usuário Autenticado

Em handlers, use `GetUserFromContext` para obter o usuário autenticado:

```go
func MyHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // Recuperar usuário autenticado do contexto
    user, err := middlewares.GetUserFromContext(ctx)
    if err != nil {
        // Usuário não autenticado (não deveria acontecer se o middleware foi aplicado)
        responses.Error(w, http.StatusUnauthorized, "Unauthorized")
        return
    }

    // Usar dados do usuário
    userID := user.ID
    email := user.Email
    roles := user.Roles

    // ... lógica do handler
}
```

### Formato do Token

O middleware espera o seguinte formato no header `Authorization`:

```
Authorization: Bearer <token>
```

### Códigos de Erro

O middleware retorna HTTP 401 (Unauthorized) nos seguintes casos:

- Header `Authorization` ausente
- Formato inválido (não é `Bearer <token>`)
- Token vazio
- Token inválido ou malformado
- Token expirado
- Claims inválidos
- Método de assinatura inválido

**Importante:** Todas as mensagens de erro retornam apenas "Unauthorized" para não expor detalhes de implementação ao cliente. Logs estruturados contêm informações detalhadas para debugging.

### Interface TokenValidator

O middleware usa a interface `TokenValidator` para validar tokens:

```go
type TokenValidator interface {
    Validate(ctx context.Context, token string) (*AuthenticatedUser, error)
}
```

Isso permite:
- Desacoplar o middleware de implementações específicas (JWT, OAuth2, etc.)
- Facilitar testes com mocks
- Trocar implementações sem modificar o middleware

### Tipo AuthenticatedUser

O usuário autenticado é representado por:

```go
type AuthenticatedUser struct {
    ID    string
    Email string
    Roles []string
}
```

### Testes

Para testar handlers que usam autenticação, use `AddUserToContext`:

```go
func TestMyHandler(t *testing.T) {
    // Arrange: criar usuário de teste
    user := auth.NewAuthenticatedUser("user-123", "test@example.com", []string{"admin"})

    // Adicionar usuário ao contexto
    ctx := middlewares.AddUserToContext(context.Background(), user)

    // Criar request com contexto
    req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)

    // Act & Assert
    // ... executar handler e verificar resultados
}
```

### Exemplo de Implementação Completa

```go
package main

import (
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/jailtonjunior94/financial/pkg/api/middlewares"
    "github.com/jailtonjunior94/financial/pkg/auth"
    "github.com/JailtonJunior94/devkit-go/pkg/observability"
    "github.com/JailtonJunior94/devkit-go/pkg/responses"
)

func SetupRouter(jwtAdapter auth.JwtAdapter, o11y observability.Observability) *chi.Mux {
    router := chi.NewRouter()

    // Criar middleware de autenticação
    authMiddleware := middlewares.NewAuthorization(jwtAdapter, o11y)

    // Rotas públicas (sem autenticação)
    router.Post("/api/v1/token", LoginHandler)
    router.Post("/api/v1/users", CreateUserHandler)

    // Rotas protegidas (com autenticação)
    router.Group(func(r chi.Router) {
        r.Use(authMiddleware.Authorization)

        r.Get("/api/v1/profile", GetProfileHandler)
        r.Put("/api/v1/profile", UpdateProfileHandler)
        r.Delete("/api/v1/profile", DeleteProfileHandler)
    })

    return router
}

func GetProfileHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // Recuperar usuário autenticado
    user, err := middlewares.GetUserFromContext(ctx)
    if err != nil {
        responses.Error(w, http.StatusUnauthorized, "Unauthorized")
        return
    }

    // Usar dados do usuário
    profile := map[string]interface{}{
        "id":    user.ID,
        "email": user.Email,
        "roles": user.Roles,
    }

    responses.Success(w, http.StatusOK, profile)
}
```

### Segurança

O middleware segue as melhores práticas de segurança:

1. **Não expõe detalhes**: Mensagens de erro genéricas para o cliente
2. **Logs estruturados**: Informações detalhadas apenas em logs
3. **Sem dados sensíveis**: Nunca loga tokens ou credenciais
4. **Tipos seguros**: Chave de contexto privada para evitar colisões
5. **Validação rigorosa**: Múltiplas camadas de validação

### Observabilidade

O middleware integra-se com o sistema de observabilidade:

- **Logs estruturados**: Todos os erros de autenticação são registrados
- **Traces distribuídos**: Integração automática com OpenTelemetry
- **Métricas**: Pode ser combinado com middleware de métricas

### Limitações

- Suporta apenas tokens no formato `Bearer <token>`
- Case-sensitive: Espera "Bearer" com "B" maiúsculo
- Não faz cache de tokens validados (validação a cada requisição)

### Changelog

#### v2.0.0 (2025-12-30)
- ✨ Nova interface `TokenValidator` para desacoplamento
- ✨ Novo tipo `AuthenticatedUser` com suporte a roles
- ✨ Validação robusta de formato Bearer Token
- ✨ Erros específicos para diferentes cenários
- ✨ Logs estruturados melhorados
- ✨ Cobertura completa de testes
- 🔧 Migração de `auth.User` para `auth.AuthenticatedUser`
- 📚 Documentação completa

#### v1.0.0
- Implementação inicial
