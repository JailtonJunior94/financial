// Package docs contains the global Swagger/OpenAPI 3.x documentation metadata for the Financial API.
// This file is the entry point for swag init (-g docs/swagger.go).
//
// To regenerate the documentation, run:
//
//	make docs-generate
package docs

//	@title			Financial API
//	@version		1.0.0
//	@description	API para gestão financeira pessoal com suporte a cartões de crédito, transações,
//	@description	orçamentos mensais, categorias e faturas com parcelamento.
//	@description
//	@description	## Valores Monetários
//	@description	Todos os valores monetários são representados como **strings decimais** (ex: `"1234.56"`)
//	@description	para preservar precisão e evitar erros de ponto flutuante.
//	@description
//	@description	## Paginação
//	@description	Endpoints de listagem utilizam **cursor-based pagination**. Parâmetros aceitos:
//	@description	- `limit` (int, default: 20, max: 100): quantidade de itens por página
//	@description	- `cursor` (string, opcional): cursor opaco retornado pelo campo `pagination.next_cursor`
//	@description
//	@description	## Autenticação
//	@description	Endpoints protegidos requerem o header `Authorization: Bearer {token}`.
//	@description	Obtenha o token via `POST /api/v1/token`.
//
//	@contact.name	Jailton Junior
//	@contact.url	https://github.com/jailtonjunior94
//
//	@license.name	MIT
//
//	@host		localhost:8080
//	@BasePath	/
//
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Token JWT no formato: `Bearer {token}`
