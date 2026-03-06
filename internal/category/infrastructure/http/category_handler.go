package http

import (
"encoding/json"
"net/http"

"github.com/jailtonjunior94/financial/internal/category/application/dtos"
"github.com/jailtonjunior94/financial/internal/category/application/usecase"
"github.com/jailtonjunior94/financial/pkg/api/httperrors"
"github.com/jailtonjunior94/financial/pkg/api/middlewares"
"github.com/jailtonjunior94/financial/pkg/observability/metrics"
"github.com/jailtonjunior94/financial/pkg/pagination"

"github.com/JailtonJunior94/devkit-go/pkg/observability"
"github.com/JailtonJunior94/devkit-go/pkg/responses"
"go.opentelemetry.io/otel/trace"

"github.com/go-chi/chi/v5"
)

type CategoryHandlerDeps struct {
O11y                         observability.Observability
FM                           *metrics.FinancialMetrics
ErrorHandler                 httperrors.ErrorHandler
CreateCategoryUseCase        usecase.CreateCategoryUseCase
FindCategoryPaginatedUseCase usecase.FindCategoryPaginatedUseCase
FindCategoryByUseCase        usecase.FindCategoryByUseCase
UpdateCategoryUseCase        usecase.UpdateCategoryUseCase
RemoveCategoryUseCase        usecase.RemoveCategoryUseCase
}

const (
defaultCategoryLimit = 20
maxCategoryLimit     = 100
)

type CategoryHandler struct {
deps CategoryHandlerDeps
}

func NewCategoryHandler(deps CategoryHandlerDeps) *CategoryHandler {
return &CategoryHandler{deps: deps}
}

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
ctx, span := h.deps.O11y.Tracer().Start(r.Context(), "category_handler.create")
defer span.End()

correlationID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()

user, err := middlewares.GetUserFromContext(ctx)
if err != nil {
h.deps.ErrorHandler.HandleError(w, r, err)
return
}

h.deps.O11y.Logger().Info(ctx, "request_received",
observability.String("operation", "CreateCategory"),
observability.String("layer", "handler"),
observability.String("entity", "category"),
observability.String("correlation_id", correlationID),
observability.String("user_id", user.ID),
)

var input *dtos.CategoryInput
if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
h.deps.ErrorHandler.HandleError(w, r, err)
return
}

if validationErrs := input.Validate(); validationErrs.HasErrors() {
h.deps.ErrorHandler.HandleError(w, r, validationErrs)
return
}

output, err := h.deps.CreateCategoryUseCase.Execute(ctx, user.ID, input)
if err != nil {
h.deps.ErrorHandler.HandleError(w, r, err)
return
}

h.deps.O11y.Logger().Info(ctx, "request_completed",
observability.String("operation", "CreateCategory"),
observability.String("layer", "handler"),
observability.String("entity", "category"),
observability.String("correlation_id", correlationID),
observability.String("user_id", user.ID),
observability.String("category_id", output.ID),
)

responses.JSON(w, http.StatusCreated, output)
}

func (h *CategoryHandler) Find(w http.ResponseWriter, r *http.Request) {
ctx, span := h.deps.O11y.Tracer().Start(r.Context(), "category_handler.find")
defer span.End()

correlationID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()

user, err := middlewares.GetUserFromContext(ctx)
if err != nil {
h.deps.ErrorHandler.HandleError(w, r, err)
return
}

h.deps.O11y.Logger().Info(ctx, "request_received",
observability.String("operation", "FindCategories"),
observability.String("layer", "handler"),
observability.String("entity", "category"),
observability.String("correlation_id", correlationID),
observability.String("user_id", user.ID),
)

params, err := pagination.ParseCursorParams(r, defaultCategoryLimit, maxCategoryLimit)
if err != nil {
h.deps.ErrorHandler.HandleError(w, r, err)
return
}

output, err := h.deps.FindCategoryPaginatedUseCase.Execute(ctx, usecase.FindCategoryPaginatedInput{
UserID: user.ID,
Limit:  params.Limit,
Cursor: params.Cursor,
})
if err != nil {
h.deps.ErrorHandler.HandleError(w, r, err)
return
}

h.deps.O11y.Logger().Info(ctx, "request_completed",
observability.String("operation", "FindCategories"),
observability.String("layer", "handler"),
observability.String("entity", "category"),
observability.String("correlation_id", correlationID),
observability.String("user_id", user.ID),
)

response := pagination.NewPaginatedResponse(output.Categories, params.Limit, output.NextCursor)
responses.JSON(w, http.StatusOK, response)
}

func (h *CategoryHandler) FindBy(w http.ResponseWriter, r *http.Request) {
ctx, span := h.deps.O11y.Tracer().Start(r.Context(), "category_handler.find_by")
defer span.End()

correlationID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()

user, err := middlewares.GetUserFromContext(ctx)
if err != nil {
h.deps.ErrorHandler.HandleError(w, r, err)
return
}

categoryID := chi.URLParam(r, "id")

h.deps.O11y.Logger().Info(ctx, "request_received",
observability.String("operation", "FindCategoryBy"),
observability.String("layer", "handler"),
observability.String("entity", "category"),
observability.String("correlation_id", correlationID),
observability.String("user_id", user.ID),
observability.String("category_id", categoryID),
)

output, err := h.deps.FindCategoryByUseCase.Execute(ctx, user.ID, categoryID)
if err != nil {
h.deps.ErrorHandler.HandleError(w, r, err)
return
}

h.deps.O11y.Logger().Info(ctx, "request_completed",
observability.String("operation", "FindCategoryBy"),
observability.String("layer", "handler"),
observability.String("entity", "category"),
observability.String("correlation_id", correlationID),
observability.String("user_id", user.ID),
observability.String("category_id", categoryID),
)

responses.JSON(w, http.StatusOK, output)
}
