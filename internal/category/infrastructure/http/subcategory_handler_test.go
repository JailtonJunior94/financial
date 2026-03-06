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
	categoryHttp "github.com/jailtonjunior94/financial/internal/category/infrastructure/http"
	"github.com/jailtonjunior94/financial/pkg/api/httperrors"
	"github.com/jailtonjunior94/financial/pkg/api/middlewares"
	"github.com/jailtonjunior94/financial/pkg/auth"
	customerrors "github.com/jailtonjunior94/financial/pkg/custom_errors"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"
	"github.com/stretchr/testify/assert"
)

// --- stub use cases ---

type stubCreateSubcategory struct {
	output *dtos.SubcategoryOutput
	err    error
}

func (s *stubCreateSubcategory) Execute(_ context.Context, _, _ string, _ *dtos.SubcategoryInput) (*dtos.SubcategoryOutput, error) {
	return s.output, s.err
}

type stubFindSubcategoryBy struct {
	output *dtos.SubcategoryOutput
	err    error
}

func (s *stubFindSubcategoryBy) Execute(_ context.Context, _, _, _ string) (*dtos.SubcategoryOutput, error) {
	return s.output, s.err
}

type stubFindSubcategoriesPaginated struct {
	output *dtos.SubcategoryPaginatedOutput
	err    error
}

func (s *stubFindSubcategoriesPaginated) Execute(_ context.Context, _, _ string, _ int, _ string) (*dtos.SubcategoryPaginatedOutput, error) {
	return s.output, s.err
}

type stubUpdateSubcategory struct {
	output *dtos.SubcategoryOutput
	err    error
}

func (s *stubUpdateSubcategory) Execute(_ context.Context, _, _, _ string, _ *dtos.SubcategoryInput) (*dtos.SubcategoryOutput, error) {
	return s.output, s.err
}

type stubRemoveSubcategory struct {
	err error
}

func (s *stubRemoveSubcategory) Execute(_ context.Context, _, _, _ string) error {
	return s.err
}

// --- helpers ---

const testSubcategoryID = "770e8400-e29b-41d4-a716-446655440001"

func newSubcategoryTestHandler(
	createUC usecase.CreateSubcategoryUseCase,
	findByUC usecase.FindSubcategoryByUseCase,
	findPaginatedUC usecase.FindSubcategoriesPaginatedUseCase,
	updateUC usecase.UpdateSubcategoryUseCase,
	removeUC usecase.RemoveSubcategoryUseCase,
) *categoryHttp.SubcategoryHandler {
	obs := fake.NewProvider()
	fm := metrics.NewTestFinancialMetrics()
	errorHandler := httperrors.NewErrorHandler(obs)
	return categoryHttp.NewSubcategoryHandler(categoryHttp.SubcategoryHandlerDeps{
		O11y:                              obs,
		FM:                                fm,
		ErrorHandler:                      errorHandler,
		CreateSubcategoryUseCase:          createUC,
		FindSubcategoryByUseCase:          findByUC,
		FindSubcategoriesPaginatedUseCase: findPaginatedUC,
		UpdateSubcategoryUseCase:          updateUC,
		RemoveSubcategoryUseCase:          removeUC,
	})
}

func makeSubcategoryChiRequest(method, path string, params map[string]string, body interface{}) *http.Request {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if len(params) > 0 {
		chiCtx := chi.NewRouteContext()
		for k, v := range params {
			chiCtx.URLParams.Add(k, v)
		}
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
	}
	return req
}

func makeAuthenticatedSubcategoryChiRequest(method, path string, params map[string]string, body interface{}, userID string) *http.Request {
	req := makeSubcategoryChiRequest(method, path, params, body)
	user := auth.NewAuthenticatedUser(userID, "user@test.com", nil)
	ctx := middlewares.AddUserToContext(req.Context(), user)
	return req.WithContext(ctx)
}

// --- Create tests ---

func TestSubcategoryHandler_Create_Success(t *testing.T) {
	output := &dtos.SubcategoryOutput{ID: testSubcategoryID, CategoryID: testCategoryID, Name: "Uber", Sequence: 1}
	handler := newSubcategoryTestHandler(
		&stubCreateSubcategory{output: output},
		&stubFindSubcategoryBy{},
		&stubFindSubcategoriesPaginated{},
		&stubUpdateSubcategory{},
		&stubRemoveSubcategory{},
	)

	body := map[string]interface{}{"name": "Uber", "sequence": 1}
	params := map[string]string{"categoryId": testCategoryID}
	req := makeAuthenticatedSubcategoryChiRequest(http.MethodPost, "/api/v1/categories/"+testCategoryID+"/subcategories", params, body, testUserID)
	w := httptest.NewRecorder()

	handler.Create(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var result dtos.SubcategoryOutput
	assert.NoError(t, json.NewDecoder(w.Body).Decode(&result))
	assert.Equal(t, "Uber", result.Name)
}

func TestSubcategoryHandler_Create_InvalidInput(t *testing.T) {
	handler := newSubcategoryTestHandler(
		&stubCreateSubcategory{},
		&stubFindSubcategoryBy{},
		&stubFindSubcategoriesPaginated{},
		&stubUpdateSubcategory{},
		&stubRemoveSubcategory{},
	)

	// missing name → validation fails
	body := map[string]interface{}{"name": "", "sequence": 1}
	params := map[string]string{"categoryId": testCategoryID}
	req := makeAuthenticatedSubcategoryChiRequest(http.MethodPost, "/api/v1/categories/"+testCategoryID+"/subcategories", params, body, testUserID)
	w := httptest.NewRecorder()

	handler.Create(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSubcategoryHandler_Create_NoUserInContext(t *testing.T) {
	handler := newSubcategoryTestHandler(
		&stubCreateSubcategory{},
		&stubFindSubcategoryBy{},
		&stubFindSubcategoriesPaginated{},
		&stubUpdateSubcategory{},
		&stubRemoveSubcategory{},
	)

	body := map[string]interface{}{"name": "Uber", "sequence": 1}
	params := map[string]string{"categoryId": testCategoryID}
	req := makeSubcategoryChiRequest(http.MethodPost, "/api/v1/categories/"+testCategoryID+"/subcategories", params, body)
	w := httptest.NewRecorder()

	handler.Create(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSubcategoryHandler_Create_UseCaseError(t *testing.T) {
	handler := newSubcategoryTestHandler(
		&stubCreateSubcategory{err: errors.New("internal error")},
		&stubFindSubcategoryBy{},
		&stubFindSubcategoriesPaginated{},
		&stubUpdateSubcategory{},
		&stubRemoveSubcategory{},
	)

	body := map[string]interface{}{"name": "Uber", "sequence": 1}
	params := map[string]string{"categoryId": testCategoryID}
	req := makeAuthenticatedSubcategoryChiRequest(http.MethodPost, "/api/v1/categories/"+testCategoryID+"/subcategories", params, body, testUserID)
	w := httptest.NewRecorder()

	handler.Create(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- Find (List) tests ---

func TestSubcategoryHandler_Find_Success(t *testing.T) {
	output := &dtos.SubcategoryPaginatedOutput{
		Data: []dtos.SubcategoryOutput{{ID: testSubcategoryID, Name: "Uber", Sequence: 1}},
		Pagination: dtos.CategoryPaginationMeta{Limit: 20, HasNext: false},
	}
	handler := newSubcategoryTestHandler(
		&stubCreateSubcategory{},
		&stubFindSubcategoryBy{},
		&stubFindSubcategoriesPaginated{output: output},
		&stubUpdateSubcategory{},
		&stubRemoveSubcategory{},
	)

	params := map[string]string{"categoryId": testCategoryID}
	req := makeAuthenticatedSubcategoryChiRequest(http.MethodGet, "/api/v1/categories/"+testCategoryID+"/subcategories?limit=20", params, nil, testUserID)
	w := httptest.NewRecorder()

	handler.List(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSubcategoryHandler_Find_NoUserInContext(t *testing.T) {
	handler := newSubcategoryTestHandler(
		&stubCreateSubcategory{},
		&stubFindSubcategoryBy{},
		&stubFindSubcategoriesPaginated{},
		&stubUpdateSubcategory{},
		&stubRemoveSubcategory{},
	)

	params := map[string]string{"categoryId": testCategoryID}
	req := makeSubcategoryChiRequest(http.MethodGet, "/api/v1/categories/"+testCategoryID+"/subcategories?limit=20", params, nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSubcategoryHandler_Find_UseCaseError(t *testing.T) {
	handler := newSubcategoryTestHandler(
		&stubCreateSubcategory{},
		&stubFindSubcategoryBy{},
		&stubFindSubcategoriesPaginated{err: errors.New("internal error")},
		&stubUpdateSubcategory{},
		&stubRemoveSubcategory{},
	)

	params := map[string]string{"categoryId": testCategoryID}
	req := makeAuthenticatedSubcategoryChiRequest(http.MethodGet, "/api/v1/categories/"+testCategoryID+"/subcategories?limit=20", params, nil, testUserID)
	w := httptest.NewRecorder()

	handler.List(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- FindBy tests ---

func TestSubcategoryHandler_FindBy_Success(t *testing.T) {
	output := &dtos.SubcategoryOutput{ID: testSubcategoryID, CategoryID: testCategoryID, Name: "Uber", Sequence: 1}
	handler := newSubcategoryTestHandler(
		&stubCreateSubcategory{},
		&stubFindSubcategoryBy{output: output},
		&stubFindSubcategoriesPaginated{},
		&stubUpdateSubcategory{},
		&stubRemoveSubcategory{},
	)

	params := map[string]string{"categoryId": testCategoryID, "id": testSubcategoryID}
	req := makeAuthenticatedSubcategoryChiRequest(http.MethodGet, "/api/v1/categories/"+testCategoryID+"/subcategories/"+testSubcategoryID, params, nil, testUserID)
	w := httptest.NewRecorder()

	handler.FindBy(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var result dtos.SubcategoryOutput
	assert.NoError(t, json.NewDecoder(w.Body).Decode(&result))
	assert.Equal(t, "Uber", result.Name)
}

func TestSubcategoryHandler_FindBy_NotFound(t *testing.T) {
	handler := newSubcategoryTestHandler(
		&stubCreateSubcategory{},
		&stubFindSubcategoryBy{err: customerrors.ErrSubcategoryNotFound},
		&stubFindSubcategoriesPaginated{},
		&stubUpdateSubcategory{},
		&stubRemoveSubcategory{},
	)

	params := map[string]string{"categoryId": testCategoryID, "id": testSubcategoryID}
	req := makeAuthenticatedSubcategoryChiRequest(http.MethodGet, "/api/v1/categories/"+testCategoryID+"/subcategories/"+testSubcategoryID, params, nil, testUserID)
	w := httptest.NewRecorder()

	handler.FindBy(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestSubcategoryHandler_FindBy_NoUserInContext(t *testing.T) {
	handler := newSubcategoryTestHandler(
		&stubCreateSubcategory{},
		&stubFindSubcategoryBy{},
		&stubFindSubcategoriesPaginated{},
		&stubUpdateSubcategory{},
		&stubRemoveSubcategory{},
	)

	params := map[string]string{"categoryId": testCategoryID, "id": testSubcategoryID}
	req := makeSubcategoryChiRequest(http.MethodGet, "/api/v1/categories/"+testCategoryID+"/subcategories/"+testSubcategoryID, params, nil)
	w := httptest.NewRecorder()

	handler.FindBy(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// --- Update tests ---

func TestSubcategoryHandler_Update_Success(t *testing.T) {
	output := &dtos.SubcategoryOutput{ID: testSubcategoryID, CategoryID: testCategoryID, Name: "Uber Updated", Sequence: 2}
	handler := newSubcategoryTestHandler(
		&stubCreateSubcategory{},
		&stubFindSubcategoryBy{},
		&stubFindSubcategoriesPaginated{},
		&stubUpdateSubcategory{output: output},
		&stubRemoveSubcategory{},
	)

	body := map[string]interface{}{"name": "Uber Updated", "sequence": 2}
	params := map[string]string{"categoryId": testCategoryID, "id": testSubcategoryID}
	req := makeAuthenticatedSubcategoryChiRequest(http.MethodPut, "/api/v1/categories/"+testCategoryID+"/subcategories/"+testSubcategoryID, params, body, testUserID)
	w := httptest.NewRecorder()

	handler.Update(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var result dtos.SubcategoryOutput
	assert.NoError(t, json.NewDecoder(w.Body).Decode(&result))
	assert.Equal(t, "Uber Updated", result.Name)
}

func TestSubcategoryHandler_Update_InvalidInput(t *testing.T) {
	handler := newSubcategoryTestHandler(
		&stubCreateSubcategory{},
		&stubFindSubcategoryBy{},
		&stubFindSubcategoriesPaginated{},
		&stubUpdateSubcategory{},
		&stubRemoveSubcategory{},
	)

	// missing name → validation fails
	body := map[string]interface{}{"name": "", "sequence": 1}
	params := map[string]string{"categoryId": testCategoryID, "id": testSubcategoryID}
	req := makeAuthenticatedSubcategoryChiRequest(http.MethodPut, "/api/v1/categories/"+testCategoryID+"/subcategories/"+testSubcategoryID, params, body, testUserID)
	w := httptest.NewRecorder()

	handler.Update(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSubcategoryHandler_Update_NotFound(t *testing.T) {
	handler := newSubcategoryTestHandler(
		&stubCreateSubcategory{},
		&stubFindSubcategoryBy{},
		&stubFindSubcategoriesPaginated{},
		&stubUpdateSubcategory{err: customerrors.ErrSubcategoryNotFound},
		&stubRemoveSubcategory{},
	)

	body := map[string]interface{}{"name": "Uber Updated", "sequence": 2}
	params := map[string]string{"categoryId": testCategoryID, "id": testSubcategoryID}
	req := makeAuthenticatedSubcategoryChiRequest(http.MethodPut, "/api/v1/categories/"+testCategoryID+"/subcategories/"+testSubcategoryID, params, body, testUserID)
	w := httptest.NewRecorder()

	handler.Update(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestSubcategoryHandler_Update_NoUserInContext(t *testing.T) {
	handler := newSubcategoryTestHandler(
		&stubCreateSubcategory{},
		&stubFindSubcategoryBy{},
		&stubFindSubcategoriesPaginated{},
		&stubUpdateSubcategory{},
		&stubRemoveSubcategory{},
	)

	body := map[string]interface{}{"name": "Uber Updated", "sequence": 2}
	params := map[string]string{"categoryId": testCategoryID, "id": testSubcategoryID}
	req := makeSubcategoryChiRequest(http.MethodPut, "/api/v1/categories/"+testCategoryID+"/subcategories/"+testSubcategoryID, params, body)
	w := httptest.NewRecorder()

	handler.Update(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// --- Delete tests ---

func TestSubcategoryHandler_Delete_Success(t *testing.T) {
	handler := newSubcategoryTestHandler(
		&stubCreateSubcategory{},
		&stubFindSubcategoryBy{},
		&stubFindSubcategoriesPaginated{},
		&stubUpdateSubcategory{},
		&stubRemoveSubcategory{},
	)

	params := map[string]string{"categoryId": testCategoryID, "id": testSubcategoryID}
	req := makeAuthenticatedSubcategoryChiRequest(http.MethodDelete, "/api/v1/categories/"+testCategoryID+"/subcategories/"+testSubcategoryID, params, nil, testUserID)
	w := httptest.NewRecorder()

	handler.Delete(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestSubcategoryHandler_Delete_NotFound(t *testing.T) {
	handler := newSubcategoryTestHandler(
		&stubCreateSubcategory{},
		&stubFindSubcategoryBy{},
		&stubFindSubcategoriesPaginated{},
		&stubUpdateSubcategory{},
		&stubRemoveSubcategory{err: customerrors.ErrSubcategoryNotFound},
	)

	params := map[string]string{"categoryId": testCategoryID, "id": testSubcategoryID}
	req := makeAuthenticatedSubcategoryChiRequest(http.MethodDelete, "/api/v1/categories/"+testCategoryID+"/subcategories/"+testSubcategoryID, params, nil, testUserID)
	w := httptest.NewRecorder()

	handler.Delete(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestSubcategoryHandler_Delete_NoUserInContext(t *testing.T) {
	handler := newSubcategoryTestHandler(
		&stubCreateSubcategory{},
		&stubFindSubcategoryBy{},
		&stubFindSubcategoriesPaginated{},
		&stubUpdateSubcategory{},
		&stubRemoveSubcategory{},
	)

	params := map[string]string{"categoryId": testCategoryID, "id": testSubcategoryID}
	req := makeSubcategoryChiRequest(http.MethodDelete, "/api/v1/categories/"+testCategoryID+"/subcategories/"+testSubcategoryID, params, nil)
	w := httptest.NewRecorder()

	handler.Delete(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
