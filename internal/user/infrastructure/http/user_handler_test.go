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
	"github.com/jailtonjunior94/financial/internal/user/application/dtos"
	"github.com/jailtonjunior94/financial/internal/user/application/usecase"
	userdomain "github.com/jailtonjunior94/financial/internal/user/domain"
	userHttp "github.com/jailtonjunior94/financial/internal/user/infrastructure/http"
	"github.com/jailtonjunior94/financial/pkg/api/httperrors"
	"github.com/jailtonjunior94/financial/pkg/api/middlewares"
	"github.com/jailtonjunior94/financial/pkg/auth"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"

	"github.com/stretchr/testify/assert"
)

// --- stub use cases ---

type stubGetUser struct {
	output *dtos.UserOutput
	err    error
}

func (s *stubGetUser) Execute(_ context.Context, _ string) (*dtos.UserOutput, error) {
	return s.output, s.err
}

type stubListUsers struct {
	output *usecase.ListUsersOutput
	err    error
}

func (s *stubListUsers) Execute(_ context.Context, _ usecase.ListUsersInput) (*usecase.ListUsersOutput, error) {
	return s.output, s.err
}

type stubUpdateUser struct {
	output *dtos.UserOutput
	err    error
}

func (s *stubUpdateUser) Execute(_ context.Context, _ string, _ *dtos.UpdateUserInput) (*dtos.UserOutput, error) {
	return s.output, s.err
}

type stubDeleteUser struct {
	err error
}

func (s *stubDeleteUser) Execute(_ context.Context, _ string) error {
	return s.err
}

type stubCreateUser struct {
	output *dtos.CreateUserOutput
	err    error
}

func (s *stubCreateUser) Execute(_ context.Context, _ *dtos.CreateUserInput) (*dtos.CreateUserOutput, error) {
	return s.output, s.err
}

// --- helpers ---

func newTestHandler(
	createUC usecase.CreateUserUseCase,
	getUC usecase.GetUserUseCase,
	listUC usecase.ListUsersUseCase,
	updateUC usecase.UpdateUserUseCase,
	deleteUC usecase.DeleteUserUseCase,
) *userHttp.UserHandler {
	obs := fake.NewProvider()
	fm := metrics.NewTestFinancialMetrics()
	errorHandler := httperrors.NewErrorHandler(obs, userdomain.ErrorMappings())
	return userHttp.NewUserHandler(userHttp.UserHandlerDeps{
		O11y:              obs,
		FM:                fm,
		ErrorHandler:      errorHandler,
		CreateUserUseCase: createUC,
		GetUserUseCase:    getUC,
		ListUsersUseCase:  listUC,
		UpdateUserUseCase: updateUC,
		DeleteUserUseCase: deleteUC,
	})
}

func makeChiRequest(method, path, paramName, paramValue string, body interface{}) *http.Request {
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

func makeAuthenticatedChiRequest(method, path, paramName, paramValue string, body interface{}, userID string) *http.Request {
	req := makeChiRequest(method, path, paramName, paramValue, body)
	user := auth.NewAuthenticatedUser(userID, "user@test.com", nil)
	ctx := middlewares.AddUserToContext(req.Context(), user)
	return req.WithContext(ctx)
}

// --- GetByID tests ---

func TestUserHandler_GetByID_Success(t *testing.T) {
	output := &dtos.UserOutput{ID: "user-1", Name: "John", Email: "john@test.com"}
	handler := newTestHandler(
		&stubCreateUser{},
		&stubGetUser{output: output},
		&stubListUsers{},
		&stubUpdateUser{},
		&stubDeleteUser{},
	)

	req := makeAuthenticatedChiRequest(http.MethodGet, "/api/v1/users/user-1", "id", "user-1", nil, "user-1")
	w := httptest.NewRecorder()

	handler.GetByID(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var result dtos.UserOutput
	_ = json.NewDecoder(w.Body).Decode(&result)
	assert.Equal(t, "user-1", result.ID)
}

func TestUserHandler_GetByID_NotFound(t *testing.T) {
	handler := newTestHandler(
		&stubCreateUser{},
		&stubGetUser{err: userdomain.ErrUserNotFound},
		&stubListUsers{},
		&stubUpdateUser{},
		&stubDeleteUser{},
	)

	req := makeAuthenticatedChiRequest(http.MethodGet, "/api/v1/users/not-found", "id", "not-found", nil, "not-found")
	w := httptest.NewRecorder()

	handler.GetByID(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUserHandler_GetByID_NoUserInContext(t *testing.T) {
	handler := newTestHandler(
		&stubCreateUser{},
		&stubGetUser{},
		&stubListUsers{},
		&stubUpdateUser{},
		&stubDeleteUser{},
	)

	req := makeChiRequest(http.MethodGet, "/api/v1/users/user-1", "id", "user-1", nil)
	w := httptest.NewRecorder()

	handler.GetByID(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// --- List tests ---

func TestUserHandler_List_Success(t *testing.T) {
	users := []*dtos.UserOutput{{ID: "u1", Name: "Alice", Email: "alice@test.com"}}
	handler := newTestHandler(
		&stubCreateUser{},
		&stubGetUser{},
		&stubListUsers{output: &usecase.ListUsersOutput{Users: users}},
		&stubUpdateUser{},
		&stubDeleteUser{},
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users?limit=20", nil)
	user := auth.NewAuthenticatedUser("user-1", "user@test.com", nil)
	req = req.WithContext(middlewares.AddUserToContext(req.Context(), user))
	w := httptest.NewRecorder()

	handler.List(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUserHandler_List_InvalidLimit(t *testing.T) {
	handler := newTestHandler(
		&stubCreateUser{},
		&stubGetUser{},
		&stubListUsers{},
		&stubUpdateUser{},
		&stubDeleteUser{},
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users?limit=abc", nil)
	user := auth.NewAuthenticatedUser("user-1", "user@test.com", nil)
	req = req.WithContext(middlewares.AddUserToContext(req.Context(), user))
	w := httptest.NewRecorder()

	handler.List(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_List_NoUserInContext(t *testing.T) {
	handler := newTestHandler(
		&stubCreateUser{},
		&stubGetUser{},
		&stubListUsers{},
		&stubUpdateUser{},
		&stubDeleteUser{},
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users?limit=20", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// --- Update tests ---

func TestUserHandler_Update_Success(t *testing.T) {
	name := "New Name"
	output := &dtos.UserOutput{ID: "user-1", Name: name, Email: "john@test.com"}
	handler := newTestHandler(
		&stubCreateUser{},
		&stubGetUser{},
		&stubListUsers{},
		&stubUpdateUser{output: output},
		&stubDeleteUser{},
	)

	body := map[string]string{"name": name}
	req := makeAuthenticatedChiRequest(http.MethodPut, "/api/v1/users/user-1", "id", "user-1", body, "user-1")
	w := httptest.NewRecorder()

	handler.Update(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUserHandler_Update_DecodeError(t *testing.T) {
	handler := newTestHandler(
		&stubCreateUser{},
		&stubGetUser{},
		&stubListUsers{},
		&stubUpdateUser{},
		&stubDeleteUser{},
	)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/users/user-1", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add("id", "user-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
	user := auth.NewAuthenticatedUser("user-1", "user@test.com", nil)
	req = req.WithContext(middlewares.AddUserToContext(req.Context(), user))
	w := httptest.NewRecorder()

	handler.Update(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_Update_Conflict(t *testing.T) {
	handler := newTestHandler(
		&stubCreateUser{},
		&stubGetUser{},
		&stubListUsers{},
		&stubUpdateUser{err: userdomain.ErrEmailAlreadyExists},
		&stubDeleteUser{},
	)

	body := map[string]string{"email": "conflict@test.com"}
	req := makeAuthenticatedChiRequest(http.MethodPut, "/api/v1/users/user-1", "id", "user-1", body, "user-1")
	w := httptest.NewRecorder()

	handler.Update(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestUserHandler_Update_NoUserInContext(t *testing.T) {
	handler := newTestHandler(
		&stubCreateUser{},
		&stubGetUser{},
		&stubListUsers{},
		&stubUpdateUser{},
		&stubDeleteUser{},
	)

	body := map[string]string{"name": "New Name"}
	req := makeChiRequest(http.MethodPut, "/api/v1/users/user-1", "id", "user-1", body)
	w := httptest.NewRecorder()

	handler.Update(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// --- Delete tests ---

func TestUserHandler_Delete_Success(t *testing.T) {
	handler := newTestHandler(
		&stubCreateUser{},
		&stubGetUser{},
		&stubListUsers{},
		&stubUpdateUser{},
		&stubDeleteUser{},
	)

	req := makeAuthenticatedChiRequest(http.MethodDelete, "/api/v1/users/user-1", "id", "user-1", nil, "user-1")
	w := httptest.NewRecorder()

	handler.Delete(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestUserHandler_Delete_NotFound(t *testing.T) {
	handler := newTestHandler(
		&stubCreateUser{},
		&stubGetUser{},
		&stubListUsers{},
		&stubUpdateUser{},
		&stubDeleteUser{err: userdomain.ErrUserNotFound},
	)

	req := makeAuthenticatedChiRequest(http.MethodDelete, "/api/v1/users/user-1", "id", "user-1", nil, "user-1")
	w := httptest.NewRecorder()

	handler.Delete(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUserHandler_Delete_NoUserInContext(t *testing.T) {
	handler := newTestHandler(
		&stubCreateUser{},
		&stubGetUser{},
		&stubListUsers{},
		&stubUpdateUser{},
		&stubDeleteUser{},
	)

	req := makeChiRequest(http.MethodDelete, "/api/v1/users/user-1", "id", "user-1", nil)
	w := httptest.NewRecorder()

	handler.Delete(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// --- Create tests ---

func TestUserHandler_Create_Success(t *testing.T) {
	output := &dtos.CreateUserOutput{ID: "new-user", Name: "Jane", Email: "jane@test.com"}
	handler := newTestHandler(
		&stubCreateUser{output: output},
		&stubGetUser{},
		&stubListUsers{},
		&stubUpdateUser{},
		&stubDeleteUser{},
	)

	body := map[string]string{"name": "Jane", "email": "jane@test.com", "password": "secret123"}
	req := makeChiRequest(http.MethodPost, "/api/v1/users", "", "", body)
	w := httptest.NewRecorder()

	handler.Create(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestUserHandler_Create_DecodeError(t *testing.T) {
	handler := newTestHandler(
		&stubCreateUser{},
		&stubGetUser{},
		&stubListUsers{},
		&stubUpdateUser{},
		&stubDeleteUser{},
	)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Create(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_Create_BusinessError(t *testing.T) {
	handler := newTestHandler(
		&stubCreateUser{err: errors.New("email already exists")},
		&stubGetUser{},
		&stubListUsers{},
		&stubUpdateUser{},
		&stubDeleteUser{},
	)

	body := map[string]string{"name": "Jane", "email": "jane@test.com", "password": "secret123"}
	req := makeChiRequest(http.MethodPost, "/api/v1/users", "", "", body)
	w := httptest.NewRecorder()

	handler.Create(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
