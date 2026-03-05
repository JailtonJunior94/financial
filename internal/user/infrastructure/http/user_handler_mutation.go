package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/jailtonjunior94/financial/internal/user/application/dtos"
	"github.com/jailtonjunior94/financial/pkg/api/middlewares"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/responses"
	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel/trace"
)

// Update godoc
//
//	@Summary		Atualizar usuário
//	@Description	Atualiza dados do próprio usuário (name, email, password).
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"ID do usuário"
//	@Param			request	body		dtos.UpdateUserInput	true	"Dados a atualizar"
//	@Success		200		{object}	dtos.UserOutput			"Usuário atualizado"
//	@Failure		400		{object}	httperrors.ProblemDetail	"Dados inválidos"
//	@Failure		403		{object}	httperrors.ProblemDetail	"Acesso negado"
//	@Failure		404		{object}	httperrors.ProblemDetail	"Usuário não encontrado"
//	@Failure		409		{object}	httperrors.ProblemDetail	"Email já em uso"
//	@Router			/api/v1/users/{id} [put]
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx, span := h.o11y.Tracer().Start(r.Context(), "user_handler.update")
	defer span.End()
	correlationID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()
	id := chi.URLParam(r, "id")
	authUser, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		h.o11y.Logger().Error(ctx, "request_failed",
			observability.String("operation", "UpdateUser"),
			observability.String("layer", "handler"),
			observability.String("entity", "user"),
			observability.String("correlation_id", correlationID),
			observability.String("error_type", "infra"),
			observability.String("error_code", "AUTH_CONTEXT_MISSING"),
			observability.Error(err),
		)
		h.fm.RecordHandlerFailure(ctx, "update_user", "user", "infra", time.Since(start))
		h.errorHandler.HandleError(w, r, err)
		return
	}
	h.o11y.Logger().Info(ctx, "request_received",
		observability.String("operation", "UpdateUser"),
		observability.String("layer", "handler"),
		observability.String("entity", "user"),
		observability.String("correlation_id", correlationID),
		observability.String("user_id", authUser.ID),
	)
	var input *dtos.UpdateUserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.o11y.Logger().Error(ctx, "validation_failed",
			observability.String("operation", "UpdateUser"),
			observability.String("layer", "handler"),
			observability.String("entity", "user"),
			observability.String("correlation_id", correlationID),
			observability.String("user_id", authUser.ID),
			observability.String("error_type", "validation"),
			observability.String("error_code", "DECODE_BODY_FAILED"),
			observability.Error(err),
		)
		h.fm.RecordHandlerFailure(ctx, "update_user", "user", "validation", time.Since(start))
		h.errorHandler.HandleError(w, r, err)
		return
	}
	output, err := h.updateUserUseCase.Execute(ctx, id, input)
	if err != nil {
		h.o11y.Logger().Error(ctx, "request_failed",
			observability.String("operation", "UpdateUser"),
			observability.String("layer", "handler"),
			observability.String("entity", "user"),
			observability.String("correlation_id", correlationID),
			observability.String("user_id", authUser.ID),
			observability.String("error_type", "business"),
			observability.String("error_code", "UPDATE_USER_FAILED"),
			observability.Error(err),
		)
		h.fm.RecordHandlerFailure(ctx, "update_user", "user", "business", time.Since(start))
		h.errorHandler.HandleError(w, r, err)
		return
	}
	h.o11y.Logger().Info(ctx, "request_completed",
		observability.String("operation", "UpdateUser"),
		observability.String("layer", "handler"),
		observability.String("entity", "user"),
		observability.String("correlation_id", correlationID),
		observability.String("user_id", authUser.ID),
	)
	h.fm.RecordHandlerRequest(ctx, "update_user", "user", time.Since(start))
	responses.JSON(w, http.StatusOK, output)
}

// Delete godoc
//
//	@Summary		Desativar usuário
//	@Description	Realiza soft delete do usuário autenticado.
//	@Tags			users
//	@Security		BearerAuth
//	@Param			id	path	string	true	"ID do usuário"
//	@Success		204	"Usuário desativado com sucesso"
//	@Failure		403	{object}	httperrors.ProblemDetail	"Acesso negado"
//	@Failure		404	{object}	httperrors.ProblemDetail	"Usuário não encontrado"
//	@Router			/api/v1/users/{id} [delete]
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx, span := h.o11y.Tracer().Start(r.Context(), "user_handler.delete")
	defer span.End()
	correlationID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()
	id := chi.URLParam(r, "id")
	authUser, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		h.o11y.Logger().Error(ctx, "request_failed",
			observability.String("operation", "DeleteUser"),
			observability.String("layer", "handler"),
			observability.String("entity", "user"),
			observability.String("correlation_id", correlationID),
			observability.String("error_type", "infra"),
			observability.String("error_code", "AUTH_CONTEXT_MISSING"),
			observability.Error(err),
		)
		h.fm.RecordHandlerFailure(ctx, "delete_user", "user", "infra", time.Since(start))
		h.errorHandler.HandleError(w, r, err)
		return
	}
	h.o11y.Logger().Info(ctx, "request_received",
		observability.String("operation", "DeleteUser"),
		observability.String("layer", "handler"),
		observability.String("entity", "user"),
		observability.String("correlation_id", correlationID),
		observability.String("user_id", authUser.ID),
	)
	if err := h.deleteUserUseCase.Execute(ctx, id); err != nil {
		h.o11y.Logger().Error(ctx, "request_failed",
			observability.String("operation", "DeleteUser"),
			observability.String("layer", "handler"),
			observability.String("entity", "user"),
			observability.String("correlation_id", correlationID),
			observability.String("user_id", authUser.ID),
			observability.String("error_type", "business"),
			observability.String("error_code", "DELETE_USER_FAILED"),
			observability.Error(err),
		)
		h.fm.RecordHandlerFailure(ctx, "delete_user", "user", "business", time.Since(start))
		h.errorHandler.HandleError(w, r, err)
		return
	}
	h.o11y.Logger().Info(ctx, "request_completed",
		observability.String("operation", "DeleteUser"),
		observability.String("layer", "handler"),
		observability.String("entity", "user"),
		observability.String("correlation_id", correlationID),
		observability.String("user_id", authUser.ID),
	)
	h.fm.RecordHandlerRequest(ctx, "delete_user", "user", time.Since(start))
	w.WriteHeader(http.StatusNoContent)
}
