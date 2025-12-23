# Postman Collection - Financial API

## 📦 Importando a Collection

1. Abra o Postman
2. Clique em **Import**
3. Selecione o arquivo `postman_collection.json`
4. A collection "Financial API - Complete Collection" será importada

## 🔧 Configuração

### Variáveis da Collection

A collection possui as seguintes variáveis configuráveis:

- **base_url**: `http://localhost:8080` (padrão)
  - Altere para a URL do seu ambiente (dev, staging, production)
- **token**: Preenchido automaticamente após login
- **category_id**: Preenchido automaticamente após criar categoria

Para editar as variáveis:
1. Clique na collection
2. Vá em **Variables**
3. Edite os valores conforme necessário

## 🚀 Fluxo de Uso Recomendado

### 1️⃣ Criar Usuário
```
POST /api/v1/users
```
Crie um novo usuário fornecendo nome, email e senha.

### 2️⃣ Fazer Login
```
POST /api/v1/token
```
Faça login com as credenciais criadas. O token JWT será salvo automaticamente na variável `{{token}}`.

### 3️⃣ Criar Categorias
```
POST /api/v1/categories
```
Crie categorias para organizar seu orçamento. O ID da categoria criada é salvo automaticamente em `{{category_id}}`.

#### Exemplo: Categoria Raiz
```json
{
  "name": "Transporte",
  "sequence": 1,
  "parent_id": ""
}
```

#### Exemplo: Subcategoria
```json
{
  "name": "Uber",
  "sequence": 1,
  "parent_id": "{{category_id}}"
}
```

### 4️⃣ Listar Categorias
```
GET /api/v1/categories
```
Lista todas as categorias raiz ordenadas por sequence.

### 5️⃣ Buscar Categoria por ID
```
GET /api/v1/categories/{id}
```
Retorna uma categoria específica com suas subcategorias.

### 6️⃣ Atualizar Categoria
```
PUT /api/v1/categories/{id}
```
Permite alterar nome, sequence ou mover para outra categoria pai.

⚠️ **Detecção de Ciclos**: O sistema impede que você crie ciclos na hierarquia (ex: categoria A → B → C → A).

### 7️⃣ Criar Orçamento
```
POST /api/v1/budgets
```
Crie um orçamento distribuindo valores entre categorias.

**Regra importante**: A soma das porcentagens deve ser exatamente **100%**.

```json
{
  "name": "Orçamento Mensal Janeiro 2025",
  "amount": 5000.00,
  "items": [
    {
      "category_id": "<uuid-categoria-1>",
      "percentage": 30.0
    },
    {
      "category_id": "<uuid-categoria-2>",
      "percentage": 70.0
    }
  ]
}
```

### 8️⃣ Deletar Categoria
```
DELETE /api/v1/categories/{id}
```
Soft delete da categoria (marcada como deletada, mas permanece no banco).

## 📋 Endpoints Disponíveis

### Authentication
| Método | Endpoint | Autenticação | Descrição |
|--------|----------|--------------|-----------|
| POST | `/api/v1/token` | Não | Obter JWT token |

### Users
| Método | Endpoint | Autenticação | Descrição |
|--------|----------|--------------|-----------|
| POST | `/api/v1/users` | Não | Criar novo usuário |

### Categories
| Método | Endpoint | Autenticação | Descrição |
|--------|----------|--------------|-----------|
| GET | `/api/v1/categories` | Sim | Listar categorias raiz |
| GET | `/api/v1/categories/{id}` | Sim | Buscar por ID (com children) |
| POST | `/api/v1/categories` | Sim | Criar categoria/subcategoria |
| PUT | `/api/v1/categories/{id}` | Sim | Atualizar categoria |
| DELETE | `/api/v1/categories/{id}` | Sim | Deletar categoria (soft) |

### Budgets
| Método | Endpoint | Autenticação | Descrição |
|--------|----------|--------------|-----------|
| POST | `/api/v1/budgets` | Sim | Criar orçamento |

### Health Check
| Método | Endpoint | Autenticação | Descrição |
|--------|----------|--------------|-----------|
| GET | `/health` | Não | Status da API |

## 🔑 Autenticação

A maioria dos endpoints requer autenticação via **Bearer Token**.

### Como funciona:
1. Faça login em `/api/v1/token`
2. O token JWT é retornado no response
3. O script de teste da collection salva automaticamente em `{{token}}`
4. Todos os endpoints autenticados usam `Authorization: Bearer {{token}}`

### Token Manual
Se precisar configurar o token manualmente:
1. Copie o token do response de login
2. Vá em **Variables** da collection
3. Cole no campo `token`

## 📝 Regras de Validação

### Categories
- **Nome**: 1-255 caracteres, não pode ser vazio
- **Sequence**: Número inteiro > 0 e ≤ 1000
- **Parent ID**: UUID válido ou vazio (para categoria raiz)
- **Hierarquia**: Não permite ciclos (ex: A → B → A)

### Budgets
- **Nome**: Obrigatório
- **Amount**: Valor decimal positivo
- **Items**: Soma das porcentagens deve ser exatamente 100%
- **Category ID**: Deve existir e pertencer ao usuário

## 🧪 Testando a Collection

### Teste Completo (Ordem recomendada):
1. **Health Check** - Verificar se API está online
2. **Create User** - Criar usuário de teste
3. **Login** - Obter token (salvo automaticamente)
4. **Create Category** - Criar categoria raiz (ID salvo automaticamente)
5. **List Categories** - Ver categoria criada
6. **Get Category by ID** - Buscar detalhes
7. **Create Subcategory** - Criar subcategoria usando `{{category_id}}`
8. **Update Category** - Alterar nome/sequence
9. **Create Budget** - Criar orçamento com categorias
10. **Delete Category** - Soft delete

## 🐛 Troubleshooting

### Erro 401 Unauthorized
- Verifique se o token está configurado
- Faça login novamente - o token pode ter expirado
- Verifique se a variável `{{token}}` está preenchida

### Erro 400 Bad Request
- Verifique o formato do JSON
- Confirme que todos os campos obrigatórios estão presentes
- Valide os tipos de dados (string, number, etc.)

### Erro 404 Not Found
- Verifique se o ID da categoria existe
- Confirme que a categoria pertence ao usuário autenticado
- Certifique-se de que a categoria não foi deletada

### Erro "Category Cycle Detected"
- Você está tentando criar um ciclo na hierarquia
- Exemplo: Tentar fazer categoria A ser filha de B, quando B já é filha de A
- Revise a estrutura de parent_id

### Erro "Percentage must equal 100%"
- A soma das porcentagens dos items do budget deve ser exatamente 100
- Não pode ser 99.99 nem 100.01, deve ser exatamente 100.00

## 📚 Estrutura de Dados

### Category Response
```json
{
  "id": "uuid",
  "name": "string",
  "sequence": 1,
  "created_at": "2025-01-01T00:00:00Z",
  "children": [
    {
      "id": "uuid",
      "name": "string",
      "sequence": 1,
      "created_at": "2025-01-01T00:00:00Z"
    }
  ]
}
```

### Budget Request
```json
{
  "name": "string",
  "amount": 5000.00,
  "items": [
    {
      "category_id": "uuid",
      "percentage": 50.0
    }
  ]
}
```

## 🌍 Ambientes

### Local Development
```
base_url: http://localhost:8080
```

### Docker
```
base_url: http://localhost:8080
```

### Production
```
base_url: https://seu-dominio.com
```

Para trocar de ambiente:
1. Vá em **Variables** da collection
2. Altere o valor de `base_url`
3. Ou crie Environments no Postman para cada ambiente

## 🔄 Scripts Automáticos

A collection inclui scripts que automatizam algumas tarefas:

### Login (POST /api/v1/token)
```javascript
// Salva o token automaticamente após login bem-sucedido
if (pm.response.code === 200) {
    const response = pm.response.json();
    pm.collectionVariables.set("token", response.token);
}
```

### Create Category
```javascript
// Salva o ID da categoria criada
if (pm.response.code === 201) {
    const response = pm.response.json();
    pm.collectionVariables.set("category_id", response.id);
}
```

## 📞 Suporte

Para reportar problemas ou sugerir melhorias:
- Abra uma issue no repositório
- Consulte a documentação da API
- Verifique os logs do servidor

---

**Versão da Collection**: 1.0.0
**Última Atualização**: 2025-12-23
**Compatível com**: Postman v10+
