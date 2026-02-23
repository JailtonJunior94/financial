package http

import (
	"net/http"

	"github.com/jailtonjunior94/financial/internal/user/application/dtos"
	"github.com/jailtonjunior94/financial/internal/user/application/usecase"
	"github.com/jailtonjunior94/financial/pkg/api/httperrors"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/responses"
	"go.opentelemetry.io/otel/trace"
)

type AuthHandler struct {
	o11y         observability.Observability
	errorHandler httperrors.ErrorHandler
	tokenUseCase usecase.TokenUseCase
}

func NewAuthHandler(
	o11y observability.Observability,
	errorHandler httperrors.ErrorHandler,
	tokenUseCase usecase.TokenUseCase,
) *AuthHandler {
	return &AuthHandler{
		o11y:         o11y,
		errorHandler: errorHandler,
		tokenUseCase: tokenUseCase,
	}
}

// Token godoc
//
//	@Summary		Gerar token de autenticação
//	@Description	Autentica o usuário com email e senha e retorna um JWT Bearer Token.
//	@Description	O token gerado deve ser enviado no header `Authorization: Bearer {token}` nas requisições protegidas.
//	@Tags			auth
//	@Accept			x-www-form-urlencoded
//	@Produce		json
//	@Param			email		formData	string	true	"Email do usuário"		example(usuario@email.com)
//	@Param			password	formData	string	true	"Senha do usuário"		example(senha123)
//	@Success		200			{object}	dtos.AuthOutput			"Token gerado com sucesso"
//	@Failure		400			{object}	httperrors.ProblemDetail	"Requisição inválida"
//	@Failure		401			{object}	httperrors.ProblemDetail	"Credenciais inválidas"
//	@Failure		500			{object}	httperrors.ProblemDetail	"Erro interno"
//	@Router			/api/v1/token [post]
func (h *AuthHandler) Token(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "auth_handler.token")
	defer span.End()

	correlationID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()

	h.o11y.Logger().Info(ctx, "request_received",
		observability.String("operation", "Token"),
		observability.String("layer", "handler"),
		observability.String("entity", "user"),
		observability.String("correlation_id", correlationID),
	)

	if err := r.ParseForm(); err != nil {
		h.o11y.Logger().Error(ctx, "request_failed",
			observability.String("operation", "Token"),
			observability.String("layer", "handler"),
			observability.String("entity", "user"),
			observability.String("correlation_id", correlationID),
			observability.String("error_type", "validation"),
			observability.String("error_code", "PARSE_FORM_FAILED"),
			observability.Error(err),
		)
		h.errorHandler.HandleError(w, r, err)
		return
	}

	input := &dtos.AuthInput{
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}

	output, err := h.tokenUseCase.Execute(ctx, input)
	if err != nil {
		h.o11y.Logger().Error(ctx, "request_failed",
			observability.String("operation", "Token"),
			observability.String("layer", "handler"),
			observability.String("entity", "user"),
			observability.String("correlation_id", correlationID),
			observability.String("error_type", "business"),
			observability.String("error_code", "TOKEN_GENERATION_FAILED"),
			observability.Error(err),
		)
		h.errorHandler.HandleError(w, r, err)
		return
	}

	h.o11y.Logger().Info(ctx, "request_completed",
		observability.String("operation", "Token"),
		observability.String("layer", "handler"),
		observability.String("entity", "user"),
		observability.String("correlation_id", correlationID),
	)

	responses.JSON(w, http.StatusOK, output)
}
