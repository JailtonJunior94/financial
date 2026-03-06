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

func (h *SubcategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
ctx, span := h.deps.O11y.Tracer().Start(r.Context(), "subcategory_handler.update")
defer span.End()

correlationID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()

user, err := middlewares.GetUserFromContext(ctx)
if err != nil {
h.deps.ErrorHandler.HandleError(w, r, err)
return
}

categoryID := chi.URLParam(r, "categoryId")
id := chi.URLParam(r, "id")

h.deps.O11y.Logger().Info(ctx, "request_received",
observability.String("operation", "UpdateSubcategory"),
observability.String("layer", "handler"),
observability.String("entity", "subcategory"),
observability.String("correlation_id", correlationID),
observability.String("user_id", user.ID),
observability.String("category_id", categoryID),
observability.String("subcategory_id", id),
)

var input *dtos.SubcategoryInput
if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
h.deps.ErrorHandler.HandleError(w, r, err)
return
}

if validationErrs := input.Validate(); validationErrs.HasErrors() {
h.deps.ErrorHandler.HandleError(w, r, validationErrs)
return
}

output, err := h.deps.UpdateSubcategoryUseCase.Execute(ctx, user.ID, categoryID, id, input)
if err != nil {
h.deps.ErrorHandler.HandleError(w, r, err)
return
}

h.deps.O11y.Logger().Info(ctx, "request_completed",
observability.String("operation", "UpdateSubcategory"),
observability.String("layer", "handler"),
observability.String("entity", "subcategory"),
observability.String("correlation_id", correlationID),
observability.String("user_id", user.ID),
observability.String("subcategory_id", id),
)

responses.JSON(w, http.StatusOK, output)
}

func (h *SubcategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
ctx, span := h.deps.O11y.Tracer().Start(r.Context(), "subcategory_handler.delete")
defer span.End()

correlationID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()

user, err := middlewares.GetUserFromContext(ctx)
if err != nil {
h.deps.ErrorHandler.HandleError(w, r, err)
return
}

categoryID := chi.URLParam(r, "categoryId")
id := chi.URLParam(r, "id")

h.deps.O11y.Logger().Info(ctx, "request_received",
observability.String("operation", "DeleteSubcategory"),
observability.String("layer", "handler"),
observability.String("entity", "subcategory"),
observability.String("correlation_id", correlationID),
observability.String("user_id", user.ID),
observability.String("category_id", categoryID),
observability.String("subcategory_id", id),
)

if err := h.deps.RemoveSubcategoryUseCase.Execute(ctx, user.ID, categoryID, id); err != nil {
h.deps.ErrorHandler.HandleError(w, r, err)
return
}

responses.JSON(w, http.StatusNoContent, nil)
}
