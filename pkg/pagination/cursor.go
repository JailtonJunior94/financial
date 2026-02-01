package pagination

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// CursorParams representa parâmetros de paginação extraídos da query string.
type CursorParams struct {
	Limit  int    // Limite de items por página (default: 20, max: 100)
	Cursor string // Cursor opaco (base64 encoded)
}

// CursorResponse representa a resposta paginada genérica.
type CursorResponse[T any] struct {
	Data       []T        `json:"data"`       // Items da página atual
	Pagination Pagination `json:"pagination"` // Metadados de paginação
}

// Pagination contém metadados de paginação.
type Pagination struct {
	Limit      int     `json:"limit"`                 // Limite usado
	HasNext    bool    `json:"has_next"`              // Tem próxima página?
	NextCursor *string `json:"next_cursor,omitempty"` // Cursor para próxima página (null se não houver)
}

// Cursor representa o estado interno do cursor (não exposto diretamente ao cliente).
// Exemplo para cards ordenados por name, id:
// {"f": {"name": "Nubank", "id": "uuid-123"}}.
type Cursor struct {
	Fields map[string]any `json:"f"` // Campos do último item (para keyset pagination)
}

// EncodeCursor codifica o cursor em base64 URL-safe.
func EncodeCursor(c Cursor) (string, error) {
	if len(c.Fields) == 0 {
		return "", nil
	}

	jsonBytes, err := json.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("failed to marshal cursor: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(jsonBytes), nil
}

// DecodeCursor decodifica o cursor de base64.
func DecodeCursor(encoded string) (Cursor, error) {
	if encoded == "" {
		return Cursor{Fields: make(map[string]any)}, nil
	}

	jsonBytes, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return Cursor{}, fmt.Errorf("invalid cursor format: %w", err)
	}

	var cursor Cursor
	if err := json.Unmarshal(jsonBytes, &cursor); err != nil {
		return Cursor{}, fmt.Errorf("invalid cursor data: %w", err)
	}

	if cursor.Fields == nil {
		cursor.Fields = make(map[string]any)
	}

	return cursor, nil
}

// ParseCursorParams extrai parâmetros de paginação da query string.
// Exemplo: ?limit=50&cursor=eyJm...
func ParseCursorParams(r *http.Request, defaultLimit int, maxLimit int) (CursorParams, error) {
	params := CursorParams{
		Limit:  defaultLimit,
		Cursor: "",
	}

	// Parse limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return params, fmt.Errorf("invalid limit parameter: must be a number")
		}

		if limit < 1 {
			return params, fmt.Errorf("limit must be greater than 0")
		}

		if limit > maxLimit {
			limit = maxLimit
		}

		params.Limit = limit
	}

	// Parse cursor
	params.Cursor = r.URL.Query().Get("cursor")

	return params, nil
}

// NewPaginatedResponse cria uma resposta paginada genérica.
// Garante que data nunca é nil (sempre retorna array vazio em JSON).
func NewPaginatedResponse[T any](
	data []T,
	limit int,
	nextCursor *string,
) CursorResponse[T] {
	// Garantir que data nunca é nil (sempre retorna array vazio em JSON)
	if data == nil {
		data = []T{}
	}

	return CursorResponse[T]{
		Data: data,
		Pagination: Pagination{
			Limit:      limit,
			HasNext:    nextCursor != nil,
			NextCursor: nextCursor,
		},
	}
}

// GetString extrai um campo string do cursor de forma segura.
func (c Cursor) GetString(key string) (string, bool) {
	if len(c.Fields) == 0 {
		return "", false
	}

	val, ok := c.Fields[key]
	if !ok {
		return "", false
	}

	str, ok := val.(string)
	return str, ok
}

// GetInt extrai um campo int do cursor de forma segura.
func (c Cursor) GetInt(key string) (int, bool) {
	if len(c.Fields) == 0 {
		return 0, false
	}

	val, ok := c.Fields[key]
	if !ok {
		return 0, false
	}

	// JSON unmarshaling pode retornar float64 para números
	switch v := val.(type) {
	case int:
		return v, true
	case float64:
		return int(v), true
	default:
		return 0, false
	}
}
