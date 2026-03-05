package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/jailtonjunior94/financial/internal/user/application/dtos"
	"github.com/jailtonjunior94/financial/internal/user/application/usecase"
	"github.com/jailtonjunior94/financial/pkg/api/httperrors"
	"github.com/jailtonjunior94/financial/pkg/api/middlewares"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"
	"github.com/jailtonjunior94/financial/pkg/pagination"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/responses"
	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel/trace"
)

type UserHandlerDeps struct {
	O11y              observability.Observability
	FM                *metrics.FinancialMetrics
	ErrorHandler      httperrors.ErrorHandler
	CreateUserUseCase usecase.CreateUserUseCase
	GetUserUseCase    usecase.GetUserUseCase
	ListUsersUseCase  usecase.ListUsersUseCase
	UpdateUserUseCase usecase.UpdateUserUseCase
	DeleteUserUseCase usecase.DeleteUserUseCase
}

type UserHandler struct {
	o11y              observability.Observability
	fm                *metrics.FinancialMetrics
	errorHandler      httperrors.ErrorHandler
	createUserUseCase usecase.CreateUserUseCase
	getUserUseCase    usecase.GetUserUseCase
	listUsersUseCase  usecase.ListUsersUseCase
	updateUserUseCase usecase.UpdateUserUseCase
	deleteUserUseCase usecase.DeleteUserUseCase
}

func NewUserHandler(deps UserHandlerDeps) *UserHandler {
	return &UserHandler{
		o11y:              deps.O11y,
		fm:                deps.FM,
		errorHandler:      deps.ErrorHandler,
		createUserUseCase: deps.CreateUserUseCase,
		getUserUseCase:    deps.GetUserUseCase,
		listUsersUseCase:  deps.ListUsersUseCase,
		updateUserUseCase: deps.UpdateUserUseCase,
		deleteUserUseCase: deps.DeleteUserUseCase,
	}
}

// Create godoc
//
//	@Summary		Criar usuário
//	@Description	Cria um novo usuário na plataforma. Email deve ser único.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dtos.CreateUserInput		true	"Dados do usuário"
//	@Success		201		{object}	dtos.CreateUserOutput		"Usuário criado com sucesso"
//	@Failure		400		{object}	httperrors.ProblemDetail	"Dados inválidos ou mal-formados"
//	@Failure		409		{object}	httperrors.ProblemDetail	"Email já cadastrado"
//	@Failure		500		{object}	httperrors.ProblemDetail	"Erro interno"
//	@Router			/api/v1/users [post]
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx, span := h.o11y.Tracer().Start(r.Context(), "user_handler.create")
	defer span.End()
	correlationID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()
	h.o11y.Logger().Info(ctx, "request_received",
		observability.String("operation", "CreateUser"),
		observability.String("layer", "handler"),
		observability.String("entity", "user"),
		observability.String("correlation_id", correlationID),
	)
	var input *dtos.CreateUserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.o11y.Logger().Error(ctx, "validation_failed",
			observability.String("operation", "CreateUser"),
			observability.String("layer", "handler"),
			observability.String("entity", "user"),
			observability.String("correlation_id", correlationID),
			observability.String("error_type", "validation"),
			observability.String("error_code", "DECODE_BODY_FAILED"),
			observability.Error(err),
		)
		h.fm.RecordHandlerFailure(ctx, "create_user", "user", "validation", time.Since(start))
		h.errorHandler.HandleError(w, r, err)
		return
	}
	output, err := h.createUserUseCase.Execute(ctx, input)
	if err != nil {
		h.o11y.Logger().Error(ctx, "request_failed",
			observability.String("operation", "CreateUser"),
			observability.String("layer", "handler"),
			observability.String("entity", "user"),
			observability.String("correlation_id", correlationID),
			observability.String("error_type", "business"),
			observability.String("error_code", "CREATE_USER_FAILED"),
			observability.Error(err),
		)
		h.fm.RecordHandlerFailure(ctx, "create_user", "user", "business", time.Since(start))
		h.errorHandler.HandleError(w, r, err)
		return
	}
	h.o11y.Logger().Info(ctx, "request_completed",
		observability.String("operation", "CreateUser"),
		observability.String("layer", "handler"),
		observability.String("entity", "user"),
		observability.String("correlation_id", correlationID),
		observability.String("user_id", output.ID),
	)
	h.fm.RecordHandlerRequest(ctx, "create_user", "user", time.Since(start))
	responses.JSON(w, http.StatusCreated, output)
}

// GetByID godoc
//
//	@Summary		Buscar usuário por ID
//	@Description	Retorna os dados do usuário autenticado pelo ID.
//	@Tags			users
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string					true	"ID do usuário"
//	@Success		200	{object}	dtos.UserOutput			"Usuário encontrado"
//	@Failure		403	{object}	httperrors.ProblemDetail	"Acesso negado"
//	@Failure		404	{object}	httperrors.ProblemDetail	"Usuário não encontrado"
//	@Failure		401	{object}	httperrors.ProblemDetail	"Não autenticado"
//	@Router			/api/v1/users/{id} [get]
func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx, span := h.o11y.Tracer().Start(r.Context(), "user_handler.get_by_id")
	defer span.End()
	correlationID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()
	id := chi.URLParam(r, "id")
	authUser, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		h.o11y.Logger().Error(ctx, "request_failed",
			observability.String("operation", "GetUser"),
			observability.String("layer", "handler"),
			observability.String("entity", "user"),
			observability.String("correlation_id", correlationID),
			observability.String("error_type", "infra"),
			observability.String("error_code", "AUTH_CONTEXT_MISSING"),
			observability.Error(err),
		)
		h.fm.RecordHandlerFailure(ctx, "get_user", "user", "infra", time.Since(start))
		h.errorHandler.HandleError(w, r, err)
		return
	}
	h.o11y.Logger().Info(ctx, "request_received",
		observability.String("operation", "GetUser"),
		observability.String("layer", "handler"),
		observability.String("entity", "user"),
		observability.String("correlation_id", correlationID),
		observability.String("user_id", authUser.ID),
	)
	output, err := h.getUserUseCase.Execute(ctx, id)
	if err != nil {
		h.o11y.Logger().Error(ctx, "request_failed",
			observability.String("operation", "GetUser"),
			observability.String("layer", "handler"),
			observability.String("entity", "user"),
			observability.String("correlation_id", correlationID),
			observability.String("error_type", "business"),
			observability.String("error_code", "GET_USER_FAILED"),
			observability.String("user_id", authUser.ID),
			observability.Error(err),
		)
		h.fm.RecordHandlerFailure(ctx, "get_user", "user", "business", time.Since(start))
		h.errorHandler.HandleError(w, r, err)
		return
	}
	h.o11y.Logger().Info(ctx, "request_completed",
		observability.String("operation", "GetUser"),
		observability.String("layer", "handler"),
		observability.String("entity", "user"),
		observability.String("correlation_id", correlationID),
		observability.String("user_id", authUser.ID),
	)
	h.fm.RecordHandlerRequest(ctx, "get_user", "user", time.Since(start))
	responses.JSON(w, http.StatusOK, output)
}

// List godoc
//
//	@Summary		Listar usuários
//	@Description	Lista usuários com paginação cursor-based.
//	@Tags			users
//	@Produce		json
//	@Security		BearerAuth
//	@Param			limit	query		int		false	"Limite de resultados (padrão: 20, máx: 100)"
//	@Param			cursor	query		string	false	"Cursor para próxima página"
//	@Success		200		{object}	pagination.CursorResponse[dtos.UserOutput]	"Lista de usuários"
//	@Failure		401		{object}	httperrors.ProblemDetail					"Não autenticado"
//	@Router			/api/v1/users [get]
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx, span := h.o11y.Tracer().Start(r.Context(), "user_handler.list")
	defer span.End()
	correlationID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()
	authUser, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		h.o11y.Logger().Error(ctx, "request_failed",
			observability.String("operation", "ListUsers"),
			observability.String("layer", "handler"),
			observability.String("entity", "user"),
			observability.String("correlation_id", correlationID),
			observability.String("error_type", "infra"),
			observability.String("error_code", "AUTH_CONTEXT_MISSING"),
			observability.Error(err),
		)
		h.fm.RecordHandlerFailure(ctx, "list_users", "user", "infra", time.Since(start))
		h.errorHandler.HandleError(w, r, err)
		return
	}
	h.o11y.Logger().Info(ctx, "request_received",
		observability.String("operation", "ListUsers"),
		observability.String("layer", "handler"),
		observability.String("entity", "user"),
		observability.String("correlation_id", correlationID),
		observability.String("user_id", authUser.ID),
	)
	params, err := pagination.ParseCursorParams(r, 20, 100)
	if err != nil {
		h.o11y.Logger().Error(ctx, "request_failed",
			observability.String("operation", "ListUsers"),
			observability.String("layer", "handler"),
			observability.String("entity", "user"),
			observability.String("correlation_id", correlationID),
			observability.String("user_id", authUser.ID),
			observability.String("error_type", "validation"),
			observability.Error(err),
		)
		h.fm.RecordHandlerFailure(ctx, "list_users", "user", "validation", time.Since(start))
		h.errorHandler.HandleError(w, r, err)
		return
	}
	output, err := h.listUsersUseCase.Execute(ctx, usecase.ListUsersInput{
		Limit:  params.Limit,
		Cursor: params.Cursor,
	})
	if err != nil {
		h.o11y.Logger().Error(ctx, "request_failed",
			observability.String("operation", "ListUsers"),
			observability.String("layer", "handler"),
			observability.String("entity", "user"),
			observability.String("correlation_id", correlationID),
			observability.String("user_id", authUser.ID),
			observability.String("error_type", "business"),
			observability.String("error_code", "LIST_USERS_FAILED"),
			observability.Error(err),
		)
		h.fm.RecordHandlerFailure(ctx, "list_users", "user", "business", time.Since(start))
		h.errorHandler.HandleError(w, r, err)
		return
	}
	h.o11y.Logger().Info(ctx, "request_completed",
		observability.String("operation", "ListUsers"),
		observability.String("layer", "handler"),
		observability.String("entity", "user"),
		observability.String("correlation_id", correlationID),
		observability.String("user_id", authUser.ID),
	)
	h.fm.RecordHandlerRequest(ctx, "list_users", "user", time.Since(start))
	response := pagination.NewPaginatedResponse(output.Users, params.Limit, output.NextCursor)
	responses.JSON(w, http.StatusOK, response)
}
