//go:build integration
// +build integration

package http_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/suite"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/jailtonjunior94/financial/configs"
	"github.com/jailtonjunior94/financial/internal/category/application/dtos"
	"github.com/jailtonjunior94/financial/internal/category/application/usecase"
	categoryHttp "github.com/jailtonjunior94/financial/internal/category/infrastructure/http"
	"github.com/jailtonjunior94/financial/internal/category/infrastructure/repositories"
	"github.com/jailtonjunior94/financial/pkg/auth"
	"github.com/jailtonjunior94/financial/pkg/database"

	"github.com/JailtonJunior94/devkit-go/pkg/observability/fake"
)

type CategoryHandlerSuite struct {
	suite.Suite
	ctx        context.Context
	db         *sql.DB
	container  *database.CockroachDBContainer
	handler    *categoryHttp.CategoryHandler
	jwtAdapter auth.JwtAdapter
	authToken  string
	testUserID string
	testEmail  string
}

func TestCategoryHandlerSuite(t *testing.T) {
	suite.Run(t, new(CategoryHandlerSuite))
}

func (s *CategoryHandlerSuite) SetupSuite() {
	s.ctx = context.Background()

	// Arrange: Configurar container CockroachDB
	s.container = database.SetupCockroachDB(s.ctx, s.T())

	// Arrange: Conectar ao banco de dados
	dsn := s.container.DSN(s.ctx, s.T())
	db, err := sql.Open("pgx", dsn)
	s.Require().NoError(err)
	s.db = db

	// Arrange: Executar migrations
	s.container.Migrate(s.T(), s.db, "file://../../../../database/migrations", "financial")

	// Arrange: Configurar JWT
	cfg := &configs.Config{
		AuthConfig: configs.AuthConfig{
			AuthSecretKey:     "test-secret-key",
			AuthTokenDuration: 1,
		},
	}
	obs := fake.NewProvider()
	s.jwtAdapter = auth.NewJwtAdapter(cfg, obs)

	// Arrange: Criar usuário de teste
	s.testUserID = "550e8400-e29b-41d4-a716-446655440000"
	s.testEmail = "test@example.com"
	_, err = s.db.ExecContext(s.ctx, `
		INSERT INTO users (id, name, email, password)
		VALUES ($1, $2, $3, $4)
	`, s.testUserID, "Test User", s.testEmail, "hashed_password")
	s.Require().NoError(err)

	// Arrange: Gerar token JWT
	token, err := s.jwtAdapter.GenerateToken(s.ctx, s.testUserID, s.testEmail)
	s.Require().NoError(err)
	s.authToken = fmt.Sprintf("Bearer %s", token)

	// Arrange: Configurar handler
	categoryRepository := repositories.NewCategoryRepository(s.db, obs)
	findCategoryUseCase := usecase.NewFindCategoryUseCase(obs, categoryRepository)
	createCategoryUseCase := usecase.NewCreateCategoryUseCase(obs, categoryRepository)
	findCategoryByUseCase := usecase.NewFindCategoryByUseCase(obs, categoryRepository)
	updateCategoryUseCase := usecase.NewUpdateCategoryUseCase(obs, categoryRepository)
	removeCategoryUseCase := usecase.NewRemoveCategoryUseCase(obs, categoryRepository)

	s.handler = categoryHttp.NewCategoryHandler(
		obs,
		findCategoryUseCase,
		createCategoryUseCase,
		findCategoryByUseCase,
		updateCategoryUseCase,
		removeCategoryUseCase,
	)
}

func (s *CategoryHandlerSuite) TearDownSuite() {
	if s.db != nil {
		s.db.Close()
	}
	if s.container != nil {
		s.container.Teardown(s.ctx, s.T())
	}
}

func (s *CategoryHandlerSuite) TearDownTest() {
	// Limpar categorias após cada teste
	_, err := s.db.ExecContext(s.ctx, "DELETE FROM categories WHERE user_id = $1", s.testUserID)
	s.Require().NoError(err)
}

func (s *CategoryHandlerSuite) TestCreate() {
	scenarios := []struct {
		name               string
		input              *dtos.CategoryInput
		authToken          string
		expectedStatusCode int
		expectError        bool
	}{
		{
			name: "deve criar categoria com sucesso",
			input: &dtos.CategoryInput{
				Name:     "Transport",
				Sequence: 1,
			},
			authToken:          s.authToken,
			expectedStatusCode: http.StatusCreated,
			expectError:        false,
		},
		{
			name: "deve criar subcategoria com sucesso",
			input: &dtos.CategoryInput{
				Name:     "Uber",
				Sequence: 1,
				ParentID: "",
			},
			authToken:          s.authToken,
			expectedStatusCode: http.StatusCreated,
			expectError:        false,
		},
		{
			name: "deve retornar erro 401 sem autenticação",
			input: &dtos.CategoryInput{
				Name:     "Transport",
				Sequence: 1,
			},
			authToken:          "",
			expectedStatusCode: http.StatusUnauthorized,
			expectError:        true,
		},
	}

	for _, scenario := range scenarios {
		s.Run(scenario.name, func() {
			// Arrange
			body, err := json.Marshal(scenario.input)
			s.Require().NoError(err)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/categories", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			if scenario.authToken != "" {
				req.Header.Set("Authorization", scenario.authToken)
				// Adicionar contexto com usuário
				user := auth.NewUser(s.testUserID, s.testEmail)
				req = test_helpers.AddUserToRequest(req, user)
			}

			rec := httptest.NewRecorder()

			// Act
			err = s.handler.Create(rec, req)

			// Assert
			if scenario.expectError {
				if scenario.authToken == "" {
					// Sem token, espera-se que o middleware retorne 401
					s.Equal(scenario.expectedStatusCode, http.StatusUnauthorized)
				}
			} else {
				s.NoError(err)
				s.Equal(scenario.expectedStatusCode, rec.Code)

				var output dtos.CategoryOutput
				err = json.NewDecoder(rec.Body).Decode(&output)
				s.NoError(err)
				s.Equal(scenario.input.Name, output.Name)
				s.Equal(scenario.input.Sequence, output.Sequence)
				s.NotEmpty(output.ID)
			}
		})
	}
}

func (s *CategoryHandlerSuite) TestFind() {
	// Arrange: Criar categorias de teste
	category1ID := "660e8400-e29b-41d4-a716-446655440001"
	category2ID := "660e8400-e29b-41d4-a716-446655440002"

	_, err := s.db.ExecContext(s.ctx, `
		INSERT INTO categories (id, user_id, name, sequence)
		VALUES ($1, $2, $3, $4), ($5, $6, $7, $8)
	`, category1ID, s.testUserID, "Transport", 1, category2ID, s.testUserID, "Food", 2)
	s.Require().NoError(err)

	scenarios := []struct {
		name               string
		authToken          string
		expectedStatusCode int
		expectedCount      int
		expectError        bool
	}{
		{
			name:               "deve listar todas as categorias com sucesso",
			authToken:          s.authToken,
			expectedStatusCode: http.StatusOK,
			expectedCount:      2,
			expectError:        false,
		},
		{
			name:               "deve retornar erro 401 sem autenticação",
			authToken:          "",
			expectedStatusCode: http.StatusUnauthorized,
			expectError:        true,
		},
	}

	for _, scenario := range scenarios {
		s.Run(scenario.name, func() {
			// Arrange
			req := httptest.NewRequest(http.MethodGet, "/api/v1/categories", nil)
			if scenario.authToken != "" {
				req.Header.Set("Authorization", scenario.authToken)
				user := auth.NewUser(s.testUserID, s.testEmail)
				req = test_helpers.AddUserToRequest(req, user)
			}

			rec := httptest.NewRecorder()

			// Act
			err := s.handler.Find(rec, req)

			// Assert
			if scenario.expectError {
				if scenario.authToken == "" {
					s.Equal(scenario.expectedStatusCode, http.StatusUnauthorized)
				}
			} else {
				s.NoError(err)
				s.Equal(scenario.expectedStatusCode, rec.Code)

				var output []*dtos.CategoryOutput
				err = json.NewDecoder(rec.Body).Decode(&output)
				s.NoError(err)
				s.Len(output, scenario.expectedCount)
			}
		})
	}
}

func (s *CategoryHandlerSuite) TestFindBy() {
	// Arrange: Criar categoria de teste
	categoryID := "660e8400-e29b-41d4-a716-446655440001"
	_, err := s.db.ExecContext(s.ctx, `
		INSERT INTO categories (id, user_id, name, sequence)
		VALUES ($1, $2, $3, $4)
	`, categoryID, s.testUserID, "Transport", 1)
	s.Require().NoError(err)

	scenarios := []struct {
		name               string
		categoryID         string
		authToken          string
		expectedStatusCode int
		expectError        bool
	}{
		{
			name:               "deve buscar categoria por ID com sucesso",
			categoryID:         categoryID,
			authToken:          s.authToken,
			expectedStatusCode: http.StatusOK,
			expectError:        false,
		},
		{
			name:               "deve retornar erro ao buscar categoria inexistente",
			categoryID:         "770e8400-e29b-41d4-a716-446655440000",
			authToken:          s.authToken,
			expectedStatusCode: http.StatusNotFound,
			expectError:        true,
		},
		{
			name:               "deve retornar erro 401 sem autenticação",
			categoryID:         categoryID,
			authToken:          "",
			expectedStatusCode: http.StatusUnauthorized,
			expectError:        true,
		},
	}

	for _, scenario := range scenarios {
		s.Run(scenario.name, func() {
			// Arrange
			req := httptest.NewRequest(http.MethodGet, "/api/v1/categories/"+scenario.categoryID, nil)
			if scenario.authToken != "" {
				req.Header.Set("Authorization", scenario.authToken)
				user := auth.NewUser(s.testUserID, s.testEmail)
				req = test_helpers.AddUserToRequest(req, user)
			}

			// Adicionar chi context para URL params
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", scenario.categoryID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rec := httptest.NewRecorder()

			// Act
			err := s.handler.FindBy(rec, req)

			// Assert
			if scenario.expectError {
				if scenario.authToken == "" {
					s.Equal(scenario.expectedStatusCode, http.StatusUnauthorized)
				} else {
					s.Error(err)
				}
			} else {
				s.NoError(err)
				s.Equal(scenario.expectedStatusCode, rec.Code)

				var output dtos.CategoryOutput
				err = json.NewDecoder(rec.Body).Decode(&output)
				s.NoError(err)
				s.Equal("Transport", output.Name)
				s.Equal(categoryID, output.ID)
			}
		})
	}
}

func (s *CategoryHandlerSuite) TestUpdate() {
	// Arrange: Criar categoria de teste
	categoryID := "660e8400-e29b-41d4-a716-446655440001"
	_, err := s.db.ExecContext(s.ctx, `
		INSERT INTO categories (id, user_id, name, sequence)
		VALUES ($1, $2, $3, $4)
	`, categoryID, s.testUserID, "Transport", 1)
	s.Require().NoError(err)

	scenarios := []struct {
		name               string
		categoryID         string
		input              *dtos.CategoryInput
		authToken          string
		expectedStatusCode int
		expectError        bool
	}{
		{
			name:       "deve atualizar categoria com sucesso",
			categoryID: categoryID,
			input: &dtos.CategoryInput{
				Name:     "Updated Transport",
				Sequence: 2,
			},
			authToken:          s.authToken,
			expectedStatusCode: http.StatusOK,
			expectError:        false,
		},
		{
			name:       "deve retornar erro ao atualizar categoria inexistente",
			categoryID: "770e8400-e29b-41d4-a716-446655440000",
			input: &dtos.CategoryInput{
				Name:     "Updated Transport",
				Sequence: 2,
			},
			authToken:          s.authToken,
			expectedStatusCode: http.StatusNotFound,
			expectError:        true,
		},
		{
			name:       "deve retornar erro 401 sem autenticação",
			categoryID: categoryID,
			input: &dtos.CategoryInput{
				Name:     "Updated Transport",
				Sequence: 2,
			},
			authToken:          "",
			expectedStatusCode: http.StatusUnauthorized,
			expectError:        true,
		},
	}

	for _, scenario := range scenarios {
		s.Run(scenario.name, func() {
			// Arrange
			body, err := json.Marshal(scenario.input)
			s.Require().NoError(err)

			req := httptest.NewRequest(http.MethodPut, "/api/v1/categories/"+scenario.categoryID, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			if scenario.authToken != "" {
				req.Header.Set("Authorization", scenario.authToken)
				user := auth.NewUser(s.testUserID, s.testEmail)
				req = test_helpers.AddUserToRequest(req, user)
			}

			// Adicionar chi context para URL params
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", scenario.categoryID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rec := httptest.NewRecorder()

			// Act
			err = s.handler.Update(rec, req)

			// Assert
			if scenario.expectError {
				if scenario.authToken == "" {
					s.Equal(scenario.expectedStatusCode, http.StatusUnauthorized)
				} else {
					s.Error(err)
				}
			} else {
				s.NoError(err)
				s.Equal(scenario.expectedStatusCode, rec.Code)

				var output dtos.CategoryOutput
				err = json.NewDecoder(rec.Body).Decode(&output)
				s.NoError(err)
				s.Equal(scenario.input.Name, output.Name)
				s.Equal(scenario.input.Sequence, output.Sequence)
			}
		})
	}
}

func (s *CategoryHandlerSuite) TestDelete() {
	scenarios := []struct {
		name               string
		setupCategory      bool
		categoryID         string
		authToken          string
		expectedStatusCode int
		expectError        bool
	}{
		{
			name:               "deve deletar categoria com sucesso",
			setupCategory:      true,
			categoryID:         "660e8400-e29b-41d4-a716-446655440001",
			authToken:          s.authToken,
			expectedStatusCode: http.StatusNoContent,
			expectError:        false,
		},
		{
			name:               "deve retornar erro ao deletar categoria inexistente",
			setupCategory:      false,
			categoryID:         "770e8400-e29b-41d4-a716-446655440000",
			authToken:          s.authToken,
			expectedStatusCode: http.StatusNotFound,
			expectError:        true,
		},
		{
			name:               "deve retornar erro 401 sem autenticação",
			setupCategory:      true,
			categoryID:         "660e8400-e29b-41d4-a716-446655440002",
			authToken:          "",
			expectedStatusCode: http.StatusUnauthorized,
			expectError:        true,
		},
	}

	for _, scenario := range scenarios {
		s.Run(scenario.name, func() {
			// Arrange
			if scenario.setupCategory {
				_, err := s.db.ExecContext(s.ctx, `
					INSERT INTO categories (id, user_id, name, sequence)
					VALUES ($1, $2, $3, $4)
				`, scenario.categoryID, s.testUserID, "Transport", 1)
				s.Require().NoError(err)
			}

			req := httptest.NewRequest(http.MethodDelete, "/api/v1/categories/"+scenario.categoryID, nil)
			if scenario.authToken != "" {
				req.Header.Set("Authorization", scenario.authToken)
				user := auth.NewUser(s.testUserID, s.testEmail)
				req = test_helpers.AddUserToRequest(req, user)
			}

			// Adicionar chi context para URL params
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", scenario.categoryID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rec := httptest.NewRecorder()

			// Act
			err := s.handler.Delete(rec, req)

			// Assert
			if scenario.expectError {
				if scenario.authToken == "" {
					s.Equal(scenario.expectedStatusCode, http.StatusUnauthorized)
				} else {
					s.Error(err)
				}
			} else {
				s.NoError(err)
				s.Equal(scenario.expectedStatusCode, rec.Code)

				// Verificar se categoria foi soft deleted
				var deletedAt sql.NullTime
				err = s.db.QueryRowContext(s.ctx, "SELECT deleted_at FROM categories WHERE id = $1 AND deleted_at IS NOT NULL", scenario.categoryID).Scan(&deletedAt)
				s.NoError(err)
				s.True(deletedAt.Valid)
				s.NotNil(deletedAt.Time)
			}
		})
	}
}

func (s *CategoryHandlerSuite) TestCreateWithFormUrlEncoded() {
	// Teste adicional para verificar que apenas JSON é aceito
	s.Run("deve rejeitar form-urlencoded", func() {
		// Arrange
		formData := url.Values{}
		formData.Set("name", "Transport")
		formData.Set("sequence", "1")

		req := httptest.NewRequest(http.MethodPost, "/api/v1/categories", bytes.NewBufferString(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Authorization", s.authToken)
		user := auth.NewUser(s.testUserID, s.testEmail)
		ctx := context.WithValue(req.Context(), "user", user)
		req = req.WithContext(ctx)

		rec := httptest.NewRecorder()

		// Act
		err := s.handler.Create(rec, req)

		// Assert
		s.Error(err)
	})
}
