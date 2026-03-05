package middlewares_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JailtonJunior94/devkit-go/pkg/observability/fake"
	"github.com/go-chi/chi/v5"
	"github.com/jailtonjunior94/financial/pkg/api/httperrors"
	errorsMocks "github.com/jailtonjunior94/financial/pkg/api/httperrors/mocks"
	"github.com/jailtonjunior94/financial/pkg/api/middlewares"
	"github.com/jailtonjunior94/financial/pkg/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupOwnershipMiddleware(t *testing.T, errorHandler httperrors.ErrorHandler) middlewares.ResourceOwnership {
	t.Helper()
	obs := fake.NewProvider()
	return middlewares.NewResourceOwnership(obs, errorHandler)
}

func makeRequestWithChiParam(userID, paramName, paramValue string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/"+paramValue, nil)

	if userID != "" {
		user := auth.NewAuthenticatedUser(userID, "user@example.com", nil)
		ctx := middlewares.AddUserToContext(req.Context(), user)
		req = req.WithContext(ctx)
	}

	// Set chi URL params
	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add(paramName, paramValue)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))

	return req
}

func TestOwnershipMiddleware_AllowsMatchingUser(t *testing.T) {
	errorHandler := httperrors.NewErrorHandler(fake.NewProvider())
	middleware := setupOwnershipMiddleware(t, errorHandler)

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	userID := "550e8400-e29b-41d4-a716-446655440000"
	req := makeRequestWithChiParam(userID, "id", userID)
	w := httptest.NewRecorder()

	middleware.Ownership("id")(next).ServeHTTP(w, req)

	assert.True(t, nextCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestOwnershipMiddleware_ForbidsDifferentUser(t *testing.T) {
	errorHandlerMock := errorsMocks.NewErrorHandler(t)
	errorHandlerMock.EXPECT().
		HandleError(mock.Anything, mock.Anything, mock.Anything).
		Run(func(w http.ResponseWriter, r *http.Request, err error) {
			w.WriteHeader(http.StatusForbidden)
		}).
		Once()

	middleware := setupOwnershipMiddleware(t, errorHandlerMock)

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	userID := "550e8400-e29b-41d4-a716-446655440000"
	paramID := "660f9500-e29b-41d4-a716-446655440001"
	req := makeRequestWithChiParam(userID, "id", paramID)
	w := httptest.NewRecorder()

	middleware.Ownership("id")(next).ServeHTTP(w, req)

	assert.False(t, nextCalled)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestOwnershipMiddleware_UnauthorizedWithoutUser(t *testing.T) {
	errorHandlerMock := errorsMocks.NewErrorHandler(t)
	errorHandlerMock.EXPECT().
		HandleError(mock.Anything, mock.Anything, mock.Anything).
		Run(func(w http.ResponseWriter, r *http.Request, err error) {
			w.WriteHeader(http.StatusUnauthorized)
		}).
		Once()

	middleware := setupOwnershipMiddleware(t, errorHandlerMock)

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	// No user in context
	req := makeRequestWithChiParam("", "id", "550e8400-e29b-41d4-a716-446655440000")
	w := httptest.NewRecorder()

	middleware.Ownership("id")(next).ServeHTTP(w, req)

	assert.False(t, nextCalled)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestOwnershipMiddleware_AllowsMatchingUserWithCustomParam(t *testing.T) {
	errorHandler := httperrors.NewErrorHandler(fake.NewProvider())
	middleware := setupOwnershipMiddleware(t, errorHandler)

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	userID := "550e8400-e29b-41d4-a716-446655440000"
	req := makeRequestWithChiParam(userID, "userID", userID)
	w := httptest.NewRecorder()

	middleware.Ownership("userID")(next).ServeHTTP(w, req)

	assert.True(t, nextCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}
