package http

import (
"encoding/json"
"net/http"

"github.com/jailtonjunior94/financial/internal/category/application/dtos"
"github.com/jailtonjunior94/financial/pkg/api/middlewares"

"github.com/JailtonJunior94/devkit-go/pkg/observability"
"github.com/JailtonJunior94/devkit-go/pkg/responses"
"go.opentelemetry.io/otel/trace"

"github.com/go-chi/chi/v5"
)

func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
ctx, span := h.deps.O11y.Tracer().Start(r.Context(), "category_handler.update")
defer span.End()

correlationID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()

user, err := middlewares.GetUserFromContext(ctx)
if err != nil {
h.deps.ErrorHandler.HandleError(w, r, err)
return
}

categoryID := chi.URLParam(r, "id")

h.deps.O11y.Logger().Info(ctx, "request_received",
observability.String("operation", "UpdateCategory"),
observability.String("layer", "handler"),
observability.String("entity", "category"),
observability.String("correlation_id", correlationID),
observability.String("user_id", user.ID),
observability.String("category_id", categoryID),
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

output, err := h.deps.UpdateCategoryUseCase.Execute(ctx, user.ID, categoryID, input)
if err != nil {
h.deps.ErrorHandler.HandleError(w, r, err)
return
}

h.deps.O11y.Logger().Info(ctx, "request_completed",
observability.String("operation", "UpdateCategory"),
observability.String("layer", "handler"),
observability.String("entity", "category"),
observability.String("correlation_id", correlationID),
observability.String("user_id", user.ID),
observability.String("category_id", categoryID),
)

responses.JSON(w, http.StatusOK, output)
}

func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
ctx, span := h.deps.O11y.Tracer().Start(r.Context(), "category_handler.delete")
defer span.End()

correlationID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()

user, err := middlewares.GetUserFromContext(ctx)
if err != nil {
h.deps.ErrorHandler.HandleError(w, r, err)
return
}

categoryID := chi.URLParam(r, "id")

h.deps.O11y.Logger().Info(ctx, "request_received",
observability.String("operation", "DeleteCategory"),
observability.String("layer", "handler"),
observability.String("entity", "category"),
observability.String("correlation_id", correlationID),
observability.String("user_id", user.ID),
observability.String("category_id", categoryID),
)

if err := h.deps.RemoveCategoryUseCase.Execute(ctx, user.ID, categoryID); err != nil {
h.deps.ErrorHandler.HandleError(w, r, err)
return
}

h.deps.O11y.Logger().Info(ctx, "request_completed",
observability.String("operation", "DeleteCategory"),
observability.String("layer", "handler"),
observability.String("entity", "category"),
observability.String("correlation_id", correlationID),
observability.String("user_id", user.ID),
observability.String("category_id", categoryID),
)

responses.JSON(w, http.StatusNoContent, nil)
}
