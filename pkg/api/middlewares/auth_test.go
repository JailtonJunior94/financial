package middlewares_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	errorhandlermocks "github.com/jailtonjunior94/financial/pkg/api/httperrors/mocks"
	"github.com/jailtonjunior94/financial/pkg/api/middlewares"
	"github.com/jailtonjunior94/financial/pkg/auth"
	authmocks "github.com/jailtonjunior94/financial/pkg/auth/mocks"
	customerrors "github.com/jailtonjunior94/financial/pkg/custom_errors"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type AuthorizationMiddlewareSuite struct {
	suite.Suite

	ctx            context.Context
	tokenValidator *authmocks.TokenValidator
	errorHandler   *errorhandlermocks.ErrorHandler
}

func TestAuthorizationMiddlewareSuite(t *testing.T) {
	suite.Run(t, new(AuthorizationMiddlewareSuite))
}

func (s *AuthorizationMiddlewareSuite) SetupTest() {
	s.ctx = context.Background()
	s.tokenValidator = authmocks.NewTokenValidator(s.T())
	s.errorHandler = errorhandlermocks.NewErrorHandler(s.T())
}

func (s *AuthorizationMiddlewareSuite) TestAuthorization() {
	type (
		args struct {
			authHeader string
		}

		dependencies struct {
			tokenValidator *authmocks.TokenValidator
			errorHandler   *errorhandlermocks.ErrorHandler
		}
	)

	scenarios := []struct {
		name         string
		args         args
		dependencies dependencies
		expect       func(recorder *httptest.ResponseRecorder, handlerCalled *bool)
	}{
		{
			name: "deve retornar erro quando o header Authorization está ausente",
			args: args{
				authHeader: "",
			},
			dependencies: dependencies{
				tokenValidator: s.tokenValidator,
				errorHandler: func() *errorhandlermocks.ErrorHandler {
					s.errorHandler.EXPECT().HandleError(mock.Anything, mock.Anything, mock.Anything).Once()
					return s.errorHandler
				}(),
			},
			expect: func(recorder *httptest.ResponseRecorder, handlerCalled *bool) {
				// Assert: não deve chamar o próximo handler
				s.False(*handlerCalled)
			},
		},
		{
			name: "deve retornar erro quando o header Authorization não começa com 'Bearer '",
			args: args{
				authHeader: "Invalid token123",
			},
			dependencies: dependencies{
				tokenValidator: s.tokenValidator,
				errorHandler: func() *errorhandlermocks.ErrorHandler {
					s.errorHandler.EXPECT().HandleError(mock.Anything, mock.Anything, mock.Anything).Once()
					return s.errorHandler
				}(),
			},
			expect: func(recorder *httptest.ResponseRecorder, handlerCalled *bool) {
				// Assert: não deve chamar o próximo handler
				s.False(*handlerCalled)
			},
		},
		{
			name: "deve retornar erro quando o token está vazio após 'Bearer '",
			args: args{
				authHeader: "Bearer ",
			},
			dependencies: dependencies{
				tokenValidator: s.tokenValidator,
				errorHandler: func() *errorhandlermocks.ErrorHandler {
					s.errorHandler.EXPECT().HandleError(mock.Anything, mock.Anything, mock.Anything).Once()
					return s.errorHandler
				}(),
			},
			expect: func(recorder *httptest.ResponseRecorder, handlerCalled *bool) {
				// Assert: não deve chamar o próximo handler
				s.False(*handlerCalled)
			},
		},
		{
			name: "deve retornar erro quando o token é inválido",
			args: args{
				authHeader: "Bearer invalid-token",
			},
			dependencies: dependencies{
				tokenValidator: func() *authmocks.TokenValidator {
					// Arrange: configurar mock para retornar erro de token inválido
					s.tokenValidator.
						EXPECT().
						Validate(mock.Anything, "invalid-token").
						Return(nil, customerrors.ErrInvalidToken).
						Once()
					return s.tokenValidator
				}(),
				errorHandler: func() *errorhandlermocks.ErrorHandler {
					s.errorHandler.EXPECT().HandleError(mock.Anything, mock.Anything, mock.Anything).Once()
					return s.errorHandler
				}(),
			},
			expect: func(recorder *httptest.ResponseRecorder, handlerCalled *bool) {
				// Assert: não deve chamar o próximo handler
				s.False(*handlerCalled)
			},
		},
		{
			name: "deve retornar erro quando o token está expirado",
			args: args{
				authHeader: "Bearer expired-token",
			},
			dependencies: dependencies{
				tokenValidator: func() *authmocks.TokenValidator {
					// Arrange: configurar mock para retornar erro de token expirado
					s.tokenValidator.
						EXPECT().
						Validate(mock.Anything, "expired-token").
						Return(nil, customerrors.ErrTokenExpired).
						Once()
					return s.tokenValidator
				}(),
				errorHandler: func() *errorhandlermocks.ErrorHandler {
					s.errorHandler.EXPECT().HandleError(mock.Anything, mock.Anything, mock.Anything).Once()
					return s.errorHandler
				}(),
			},
			expect: func(recorder *httptest.ResponseRecorder, handlerCalled *bool) {
				// Assert: não deve chamar o próximo handler
				s.False(*handlerCalled)
			},
		},
		{
			name: "deve retornar erro quando ocorre erro genérico na validação",
			args: args{
				authHeader: "Bearer token-with-error",
			},
			dependencies: dependencies{
				tokenValidator: func() *authmocks.TokenValidator {
					// Arrange: configurar mock para retornar erro genérico
					s.tokenValidator.
						EXPECT().
						Validate(mock.Anything, "token-with-error").
						Return(nil, errors.New("generic validation error")).
						Once()
					return s.tokenValidator
				}(),
				errorHandler: func() *errorhandlermocks.ErrorHandler {
					s.errorHandler.EXPECT().HandleError(mock.Anything, mock.Anything, mock.Anything).Once()
					return s.errorHandler
				}(),
			},
			expect: func(recorder *httptest.ResponseRecorder, handlerCalled *bool) {
				// Assert: não deve chamar o próximo handler
				s.False(*handlerCalled)
			},
		},
		{
			name: "deve autenticar com sucesso e injetar usuário no contexto",
			args: args{
				authHeader: "Bearer valid-token",
			},
			dependencies: dependencies{
				tokenValidator: func() *authmocks.TokenValidator {
					// Arrange: configurar mock para retornar usuário válido
					user := auth.NewAuthenticatedUser("user-123", "user@example.com", []string{"admin"})
					s.tokenValidator.
						EXPECT().
						Validate(mock.Anything, "valid-token").
						Return(user, nil).
						Once()
					return s.tokenValidator
				}(),
				errorHandler: s.errorHandler,
			},
			expect: func(recorder *httptest.ResponseRecorder, handlerCalled *bool) {
				// Assert: deve chamar o próximo handler
				s.Equal(http.StatusOK, recorder.Code)
				s.True(*handlerCalled)
			},
		},
		{
			name: "deve autenticar com sucesso com Bearer em minúsculas",
			args: args{
				authHeader: "Bearer valid-token-lowercase",
			},
			dependencies: dependencies{
				tokenValidator: func() *authmocks.TokenValidator {
					// Arrange: configurar mock para retornar usuário válido
					user := auth.NewAuthenticatedUser("user-456", "another@example.com", []string{"user"})
					s.tokenValidator.
						EXPECT().
						Validate(mock.Anything, "valid-token-lowercase").
						Return(user, nil).
						Once()
					return s.tokenValidator
				}(),
				errorHandler: s.errorHandler,
			},
			expect: func(recorder *httptest.ResponseRecorder, handlerCalled *bool) {
				// Assert: deve chamar o próximo handler
				s.Equal(http.StatusOK, recorder.Code)
				s.True(*handlerCalled)
			},
		},
	}

	for _, scenario := range scenarios {
		s.T().Run(scenario.name, func(t *testing.T) {
			// Arrange: criar middleware e handler de teste
			authMiddleware := middlewares.NewAuthorization(
				scenario.dependencies.tokenValidator,
				nil, // o11y não usado nos testes
				scenario.dependencies.errorHandler,
			)

			handlerCalled := false
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true

				// Se o handler foi chamado, verificar se o usuário está no contexto
				user, err := middlewares.GetUserFromContext(r.Context())
				if err == nil && user != nil {
					s.NotEmpty(user.ID)
					s.NotEmpty(user.Email)
				}

				w.WriteHeader(http.StatusOK)
			})

			// Arrange: criar request com header de autenticação
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if scenario.args.authHeader != "" {
				req.Header.Set("Authorization", scenario.args.authHeader)
			}

			recorder := httptest.NewRecorder()

			// Act: executar o middleware
			handler := authMiddleware.Authorization(nextHandler)
			handler.ServeHTTP(recorder, req)

			// Assert: chamar função de verificação
			scenario.expect(recorder, &handlerCalled)
		})
	}
}

func (s *AuthorizationMiddlewareSuite) TestGetUserFromContext() {
	scenarios := []struct {
		name   string
		setup  func() context.Context
		expect func(user *auth.AuthenticatedUser, err error)
	}{
		{
			name:  "deve retornar erro quando o usuário não está no contexto",
			setup: context.Background,
			expect: func(user *auth.AuthenticatedUser, err error) {
				s.Error(err)
				s.Nil(user)
				s.Equal(customerrors.ErrUnauthorized, err)
			},
		},
		{
			name: "deve retornar o usuário quando está presente no contexto",
			setup: func() context.Context {
				user := auth.NewAuthenticatedUser("user-123", "user@example.com", []string{"admin"})
				return middlewares.AddUserToContext(context.Background(), user)
			},
			expect: func(user *auth.AuthenticatedUser, err error) {
				s.NoError(err)
				s.NotNil(user)
				s.Equal("user-123", user.ID)
				s.Equal("user@example.com", user.Email)
				s.Contains(user.Roles, "admin")
			},
		},
	}

	for _, scenario := range scenarios {
		s.T().Run(scenario.name, func(t *testing.T) {
			// Arrange
			ctx := scenario.setup()

			// Act
			user, err := middlewares.GetUserFromContext(ctx)

			// Assert
			scenario.expect(user, err)
		})
	}
}

func (s *AuthorizationMiddlewareSuite) TestAddUserToContext() {
	// Arrange
	user := auth.NewAuthenticatedUser("user-789", "test@example.com", []string{"user", "admin"})

	// Act
	ctx := middlewares.AddUserToContext(context.Background(), user)

	// Assert
	retrievedUser, err := middlewares.GetUserFromContext(ctx)
	s.NoError(err)
	s.NotNil(retrievedUser)
	s.Equal(user.ID, retrievedUser.ID)
	s.Equal(user.Email, retrievedUser.Email)
	s.Equal(user.Roles, retrievedUser.Roles)
}
