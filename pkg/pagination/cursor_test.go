package pagination

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeCursor(t *testing.T) {
	tests := []struct {
		name    string
		cursor  Cursor
		wantErr bool
	}{
		{
			name: "valid cursor with string and id",
			cursor: Cursor{
				Fields: map[string]any{
					"name": "Nubank",
					"id":   "uuid-123",
				},
			},
			wantErr: false,
		},
		{
			name: "empty cursor",
			cursor: Cursor{
				Fields: map[string]any{},
			},
			wantErr: false,
		},
		{
			name: "nil fields",
			cursor: Cursor{
				Fields: nil,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := EncodeCursor(tt.cursor)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			if len(tt.cursor.Fields) == 0 {
				assert.Empty(t, encoded)
				return
			}

			// Decode e verificar se é igual
			decoded, err := DecodeCursor(encoded)
			assert.NoError(t, err)
			assert.Equal(t, tt.cursor.Fields, decoded.Fields)
		})
	}
}

func TestDecodeCursor(t *testing.T) {
	tests := []struct {
		name    string
		encoded string
		want    Cursor
		wantErr bool
	}{
		{
			name:    "empty string",
			encoded: "",
			want: Cursor{
				Fields: map[string]any{},
			},
			wantErr: false,
		},
		{
			name:    "invalid base64",
			encoded: "invalid!!!",
			wantErr: true,
		},
		{
			name:    "invalid json",
			encoded: "aW52YWxpZCBqc29u", // "invalid json" in base64
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeCursor(tt.encoded)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want.Fields, got.Fields)
		})
	}
}

func TestParseCursorParams(t *testing.T) {
	tests := []struct {
		name         string
		queryString  string
		defaultLimit int
		maxLimit     int
		want         CursorParams
		wantErr      bool
	}{
		{
			name:         "no params - use defaults",
			queryString:  "",
			defaultLimit: 20,
			maxLimit:     100,
			want: CursorParams{
				Limit:  20,
				Cursor: "",
			},
			wantErr: false,
		},
		{
			name:         "valid limit and cursor",
			queryString:  "limit=50&cursor=eyJmIjp7fX0",
			defaultLimit: 20,
			maxLimit:     100,
			want: CursorParams{
				Limit:  50,
				Cursor: "eyJmIjp7fX0",
			},
			wantErr: false,
		},
		{
			name:         "limit exceeds max - cap to max",
			queryString:  "limit=200",
			defaultLimit: 20,
			maxLimit:     100,
			want: CursorParams{
				Limit:  100,
				Cursor: "",
			},
			wantErr: false,
		},
		{
			name:         "invalid limit - not a number",
			queryString:  "limit=abc",
			defaultLimit: 20,
			maxLimit:     100,
			wantErr:      true,
		},
		{
			name:         "limit less than 1",
			queryString:  "limit=0",
			defaultLimit: 20,
			maxLimit:     100,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/?"+tt.queryString, nil)

			got, err := ParseCursorParams(req, tt.defaultLimit, tt.maxLimit)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want.Limit, got.Limit)
			assert.Equal(t, tt.want.Cursor, got.Cursor)
		})
	}
}

func TestNewPaginatedResponse(t *testing.T) {
	t.Run("nil data - return empty array", func(t *testing.T) {
		var data []string
		cursor := "next-cursor"

		response := NewPaginatedResponse(data, 20, &cursor)

		assert.NotNil(t, response.Data)
		assert.Equal(t, 0, len(response.Data))
		assert.Equal(t, 20, response.Pagination.Limit)
		assert.True(t, response.Pagination.HasNext)
		assert.NotNil(t, response.Pagination.NextCursor)
		assert.Equal(t, cursor, *response.Pagination.NextCursor)
	})

	t.Run("valid data with next cursor", func(t *testing.T) {
		data := []string{"item1", "item2"}
		cursor := "next-cursor"

		response := NewPaginatedResponse(data, 20, &cursor)

		assert.Equal(t, data, response.Data)
		assert.Equal(t, 20, response.Pagination.Limit)
		assert.True(t, response.Pagination.HasNext)
		assert.Equal(t, cursor, *response.Pagination.NextCursor)
	})

	t.Run("last page - no next cursor", func(t *testing.T) {
		data := []string{"item1", "item2"}

		response := NewPaginatedResponse(data, 20, nil)

		assert.Equal(t, data, response.Data)
		assert.Equal(t, 20, response.Pagination.Limit)
		assert.False(t, response.Pagination.HasNext)
		assert.Nil(t, response.Pagination.NextCursor)
	})
}

func TestCursor_GetString(t *testing.T) {
	tests := []struct {
		name   string
		cursor Cursor
		key    string
		want   string
		wantOk bool
	}{
		{
			name: "existing string field",
			cursor: Cursor{
				Fields: map[string]any{
					"name": "Nubank",
				},
			},
			key:    "name",
			want:   "Nubank",
			wantOk: true,
		},
		{
			name: "non-existing field",
			cursor: Cursor{
				Fields: map[string]any{
					"name": "Nubank",
				},
			},
			key:    "id",
			want:   "",
			wantOk: false,
		},
		{
			name: "nil fields",
			cursor: Cursor{
				Fields: nil,
			},
			key:    "name",
			want:   "",
			wantOk: false,
		},
		{
			name: "wrong type",
			cursor: Cursor{
				Fields: map[string]any{
					"count": 123,
				},
			},
			key:    "count",
			want:   "",
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := tt.cursor.GetString(tt.key)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantOk, ok)
		})
	}
}

func TestCursor_GetInt(t *testing.T) {
	tests := []struct {
		name   string
		cursor Cursor
		key    string
		want   int
		wantOk bool
	}{
		{
			name: "existing int field",
			cursor: Cursor{
				Fields: map[string]any{
					"sequence": 10,
				},
			},
			key:    "sequence",
			want:   10,
			wantOk: true,
		},
		{
			name: "existing float64 field (from JSON)",
			cursor: Cursor{
				Fields: map[string]any{
					"sequence": float64(10),
				},
			},
			key:    "sequence",
			want:   10,
			wantOk: true,
		},
		{
			name: "non-existing field",
			cursor: Cursor{
				Fields: map[string]any{
					"sequence": 10,
				},
			},
			key:    "other",
			want:   0,
			wantOk: false,
		},
		{
			name: "wrong type",
			cursor: Cursor{
				Fields: map[string]any{
					"name": "text",
				},
			},
			key:    "name",
			want:   0,
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := tt.cursor.GetInt(tt.key)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantOk, ok)
		})
	}
}
