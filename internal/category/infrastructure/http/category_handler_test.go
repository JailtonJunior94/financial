package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JailtonJunior94/devkit-go/pkg/observability/fake"
	"github.com/go-chi/chi/v5"
	"github.com/jailtonjunior94/financial/internal/category/application/dtos"
	"github.com/jailtonjunior94/financial/internal/category/application/usecase"
	categorydomain "github.com/jailtonjunior94/financial/internal/category/domain"
	categoryHttp "github.com/jailtonjunior94/financial/internal/category/infrastructure/http"
	"github.com/jailtonjunior94/financial/pkg/api/httperrors"
	"github.com/jailtonjunior94/financial/pkg/api/middlewares"
	"github.com/jailtonjunior94/financial/pkg/auth"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"
	"github.com/stretchr/testify/assert"
)

// --- stub use cases ---

type stubCreateCategory struct {
	output *dtos.CategoryOutput
	err    error
}

func (s *stubCreateCategory) Execute(_ context.Context, _ string, _ *dtos.CategoryInput) (*dtos.CategoryOutput, error) {
	return s.output, s.err
}

type stubFindCategoryPaginated struct {
	output *usecase.FindCategoryPaginatedOutput
	err    error
}

func (s *stubFindCategoryPaginated) Execute(_ context.Context, _ usecase.FindCategoryPaginatedInput) (*usecase.FindCategoryPaginatedOutput, error) {
	return s.output, s.err
}

type stubFindCategoryBy struct {
	output *dtos.CategoryOutput
	err    error
}

func (s *stubFindCategoryBy) Execute(_ context.Context, _, _ string) (*dtos.CategoryOutput, error) {
	return s.output, s.err
}

type stubUpdateCategory struct {
	output *dtos.CategoryOutput
	err    error
}

func (s *stubUpdateCategory) Execute(_ context.Context, _, _ string, _ *dtos.CategoryInput) (*dtos.CategoryOutput, error) {
	return s.output, s.err
}

type stubRemoveCategory struct {
	err error
}

func (s *stubRemoveCategory) Execute(_ context.Context, _, _ string) error {
	return s.err
}

// --- helpers ---

const testUserID = "550e8400-e29b-41d4-a716-446655440000"
const testCategoryID = "660e8400-e29b-41d4-a716-446655440001"

func newCategoryTestHandler(
	createUC usecase.CreateCategoryUseCase,
	findPaginatedUC usecase.FindCategoryPaginatedUseCase,
	findByUC usecase.FindCategoryByUseCase,
	updateUC usecase.UpdateCategoryUseCase,
	removeUC usecase.RemoveCategoryUseCase,
) *categoryHttp.CategoryHandler {
	obs := fake.NewProvider()
	fm := metrics.NewTestFinancialMetrics()
	errorHandler := httperrors.NewErrorHandler(obs, categorydomain.ErrorMappings())
	return categoryHttp.NewCategoryHandler(categoryHttp.CategoryHandlerDeps{
		O11y:                         obs,
		FM:                           fm,
		ErrorHandler:                 errorHandler,
		CreateCategoryUseCase:        createUC,
		FindCategoryPaginatedUseCase: findPaginatedUC,
		FindCategoryByUseCase:        findByUC,
		UpdateCategoryUseCase:        updateUC,
		RemoveCategoryUseCase:        removeUC,
	})
}

func makeCategoryChiRequest(method, path, paramName, paramValue string, body interface{}) *http.Request {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if paramName != "" {
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add(paramName, paramValue)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
	}
	return req
}

func makeAuthenticatedCategoryChiRequest(method, path, paramName, paramValue string, body interface{}, userID string) *http.Request {
	req := makeCategoryChiRequest(method, path, paramName, paramValue, body)
	user := auth.NewAuthenticatedUser(userID, "user@test.com", nil)
	ctx := middlewares.AddUserToContext(req.Context(), user)
	return req.WithContext(ctx)
}

// --- Create tests ---

func TestCategoryHandler_Create_Success(t *testing.T) {
	output := &dtos.CategoryOutput{ID: testCategoryID, Name: "Transport", Sequence: 1}
	handler := newCategoryTestHandler(
		&stubCreateCategory{output: output},
		&stubFindCategoryPaginated{},
		&stubFindCategoryBy{},
		&stubUpdateCategory{},
		&stubRemoveCategory{},
	)

	body := map[string]interface{}{"name": "Transport", "sequence": 1}
	req := makeAuthenticatedCategoryChiRequest(http.MethodPost, "/api/v1/categories", "", "", body, testUserID)
	w := httptest.NewRecorder()

	handler.Create(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var result dtos.CategoryOutput
	assert.NoError(t, json.NewDecoder(w.Body).Decode(&result))
	assert.Equal(t, "Transport", result.Name)
}

func TestCategoryHandler_Create_InvalidInput(t *testing.T) {
	handler := newCategoryTestHandler(
		&stubCreateCategory{},
		&stubFindCategoryPaginated{},
		&stubFindCategoryBy{},
		&stubUpdateCategory{},
		&stubRemoveCategory{},
	)

	// missing name → validation fails
	body := map[string]interface{}{"name": "", "sequence": 1}
	req := makeAuthenticatedCategoryChiRequest(http.MethodPost, "/api/v1/categories", "", "", body, testUserID)
	w := httptest.NewRecorder()

	handler.Create(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCategoryHandler_Create_NoUserInContext(t *testing.T) {
	handler := newCategoryTestHandler(
		&stubCreateCategory{},
		&stubFindCategoryPaginated{},
		&stubFindCategoryBy{},
		&stubUpdateCategory{},
		&stubRemoveCategory{},
	)

	body := map[string]interface{}{"name": "Transport", "sequence": 1}
	req := makeCategoryChiRequest(http.MethodPost, "/api/v1/categories", "", "", body)
	w := httptest.NewRecorder()

	handler.Create(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCategoryHandler_Create_UseCaseError(t *testing.T) {
	handler := newCategoryTestHandler(
		&stubCreateCategory{err: errors.New("internal error")},
		&stubFindCategoryPaginated{},
		&stubFindCategoryBy{},
		&stubUpdateCategory{},
		&stubRemoveCategory{},
	)

	body := map[string]interface{}{"name": "Transport", "sequence": 1}
	req := makeAuthenticatedCategoryChiRequest(http.MethodPost, "/api/v1/categories", "", "", body, testUserID)
	w := httptest.NewRecorder()

	handler.Create(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- Find tests ---

func TestCategoryHandler_Find_Success(t *testing.T) {
	output := &usecase.FindCategoryPaginatedOutput{
		Categories: []*dtos.CategoryOutput{{ID: testCategoryID, Name: "Transport", Sequence: 1}},
		NextCursor: nil,
	}
	handler := newCategoryTestHandler(
		&stubCreateCategory{},
		&stubFindCategoryPaginated{output: output},
		&stubFindCategoryBy{},
		&stubUpdateCategory{},
		&stubRemoveCategory{},
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/categories?limit=20", nil)
	user := auth.NewAuthenticatedUser(testUserID, "user@test.com", nil)
	req = req.WithContext(middlewares.AddUserToContext(req.Context(), user))
	w := httptest.NewRecorder()

	handler.Find(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCategoryHandler_Find_NoUserInContext(t *testing.T) {
	handler := newCategoryTestHandler(
		&stubCreateCategory{},
		&stubFindCategoryPaginated{},
		&stubFindCategoryBy{},
		&stubUpdateCategory{},
		&stubRemoveCategory{},
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/categories?limit=20", nil)
	w := httptest.NewRecorder()

	handler.Find(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCategoryHandler_Find_UseCaseError(t *testing.T) {
	handler := newCategoryTestHandler(
		&stubCreateCategory{},
		&stubFindCategoryPaginated{err: errors.New("internal error")},
		&stubFindCategoryBy{},
		&stubUpdateCategory{},
		&stubRemoveCategory{},
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/categories?limit=20", nil)
	user := auth.NewAuthenticatedUser(testUserID, "user@test.com", nil)
	req = req.WithContext(middlewares.AddUserToContext(req.Context(), user))
	w := httptest.NewRecorder()

	handler.Find(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- FindBy tests ---

func TestCategoryHandler_FindBy_Success(t *testing.T) {
	output := &dtos.CategoryOutput{ID: testCategoryID, Name: "Transport", Sequence: 1}
	handler := newCategoryTestHandler(
		&stubCreateCategory{},
		&stubFindCategoryPaginated{},
		&stubFindCategoryBy{output: output},
		&stubUpdateCategory{},
		&stubRemoveCategory{},
	)

	req := makeAuthenticatedCategoryChiRequest(http.MethodGet, "/api/v1/categories/"+testCategoryID, "id", testCategoryID, nil, testUserID)
	w := httptest.NewRecorder()

	handler.FindBy(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var result dtos.CategoryOutput
	assert.NoError(t, json.NewDecoder(w.Body).Decode(&result))
	assert.Equal(t, "Transport", result.Name)
}

func TestCategoryHandler_FindBy_NotFound(t *testing.T) {
	handler := newCategoryTestHandler(
		&stubCreateCategory{},
		&stubFindCategoryPaginated{},
		&stubFindCategoryBy{err: categorydomain.ErrCategoryNotFound},
		&stubUpdateCategory{},
		&stubRemoveCategory{},
	)

	req := makeAuthenticatedCategoryChiRequest(http.MethodGet, "/api/v1/categories/"+testCategoryID, "id", testCategoryID, nil, testUserID)
	w := httptest.NewRecorder()

	handler.FindBy(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCategoryHandler_FindBy_NoUserInContext(t *testing.T) {
	handler := newCategoryTestHandler(
		&stubCreateCategory{},
		&stubFindCategoryPaginated{},
		&stubFindCategoryBy{},
		&stubUpdateCategory{},
		&stubRemoveCategory{},
	)

	req := makeCategoryChiRequest(http.MethodGet, "/api/v1/categories/"+testCategoryID, "id", testCategoryID, nil)
	w := httptest.NewRecorder()

	handler.FindBy(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// --- Update tests ---

func TestCategoryHandler_Update_Success(t *testing.T) {
	output := &dtos.CategoryOutput{ID: testCategoryID, Name: "Transport Updated", Sequence: 2}
	handler := newCategoryTestHandler(
		&stubCreateCategory{},
		&stubFindCategoryPaginated{},
		&stubFindCategoryBy{},
		&stubUpdateCategory{output: output},
		&stubRemoveCategory{},
	)

	body := map[string]interface{}{"name": "Transport Updated", "sequence": 2}
	req := makeAuthenticatedCategoryChiRequest(http.MethodPut, "/api/v1/categories/"+testCategoryID, "id", testCategoryID, body, testUserID)
	w := httptest.NewRecorder()

	handler.Update(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var result dtos.CategoryOutput
	assert.NoError(t, json.NewDecoder(w.Body).Decode(&result))
	assert.Equal(t, "Transport Updated", result.Name)
}

func TestCategoryHandler_Update_InvalidInput(t *testing.T) {
	handler := newCategoryTestHandler(
		&stubCreateCategory{},
		&stubFindCategoryPaginated{},
		&stubFindCategoryBy{},
		&stubUpdateCategory{},
		&stubRemoveCategory{},
	)

	// missing name → validation fails
	body := map[string]interface{}{"name": "", "sequence": 1}
	req := makeAuthenticatedCategoryChiRequest(http.MethodPut, "/api/v1/categories/"+testCategoryID, "id", testCategoryID, body, testUserID)
	w := httptest.NewRecorder()

	handler.Update(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCategoryHandler_Update_NotFound(t *testing.T) {
	handler := newCategoryTestHandler(
		&stubCreateCategory{},
		&stubFindCategoryPaginated{},
		&stubFindCategoryBy{},
		&stubUpdateCategory{err: categorydomain.ErrCategoryNotFound},
		&stubRemoveCategory{},
	)

	body := map[string]interface{}{"name": "Transport Updated", "sequence": 2}
	req := makeAuthenticatedCategoryChiRequest(http.MethodPut, "/api/v1/categories/"+testCategoryID, "id", testCategoryID, body, testUserID)
	w := httptest.NewRecorder()

	handler.Update(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCategoryHandler_Update_NoUserInContext(t *testing.T) {
	handler := newCategoryTestHandler(
		&stubCreateCategory{},
		&stubFindCategoryPaginated{},
		&stubFindCategoryBy{},
		&stubUpdateCategory{},
		&stubRemoveCategory{},
	)

	body := map[string]interface{}{"name": "Transport Updated", "sequence": 2}
	req := makeCategoryChiRequest(http.MethodPut, "/api/v1/categories/"+testCategoryID, "id", testCategoryID, body)
	w := httptest.NewRecorder()

	handler.Update(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// --- Delete tests ---

func TestCategoryHandler_Delete_Success(t *testing.T) {
	handler := newCategoryTestHandler(
		&stubCreateCategory{},
		&stubFindCategoryPaginated{},
		&stubFindCategoryBy{},
		&stubUpdateCategory{},
		&stubRemoveCategory{},
	)

	req := makeAuthenticatedCategoryChiRequest(http.MethodDelete, "/api/v1/categories/"+testCategoryID, "id", testCategoryID, nil, testUserID)
	w := httptest.NewRecorder()

	handler.Delete(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestCategoryHandler_Delete_NotFound(t *testing.T) {
	handler := newCategoryTestHandler(
		&stubCreateCategory{},
		&stubFindCategoryPaginated{},
		&stubFindCategoryBy{},
		&stubUpdateCategory{},
		&stubRemoveCategory{err: categorydomain.ErrCategoryNotFound},
	)

	req := makeAuthenticatedCategoryChiRequest(http.MethodDelete, "/api/v1/categories/"+testCategoryID, "id", testCategoryID, nil, testUserID)
	w := httptest.NewRecorder()

	handler.Delete(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCategoryHandler_Delete_UseCaseError(t *testing.T) {
	handler := newCategoryTestHandler(
		&stubCreateCategory{},
		&stubFindCategoryPaginated{},
		&stubFindCategoryBy{},
		&stubUpdateCategory{},
		&stubRemoveCategory{err: errors.New("internal error")},
	)

	req := makeAuthenticatedCategoryChiRequest(http.MethodDelete, "/api/v1/categories/"+testCategoryID, "id", testCategoryID, nil, testUserID)
	w := httptest.NewRecorder()

	handler.Delete(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCategoryHandler_Delete_NoUserInContext(t *testing.T) {
	handler := newCategoryTestHandler(
		&stubCreateCategory{},
		&stubFindCategoryPaginated{},
		&stubFindCategoryBy{},
		&stubUpdateCategory{},
		&stubRemoveCategory{},
	)

	req := makeCategoryChiRequest(http.MethodDelete, "/api/v1/categories/"+testCategoryID, "id", testCategoryID, nil)
	w := httptest.NewRecorder()

	handler.Delete(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
