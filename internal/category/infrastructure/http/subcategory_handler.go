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

type SubcategoryHandlerDeps struct {
	O11y                              observability.Observability
	FM                                *metrics.FinancialMetrics
	ErrorHandler                      httperrors.ErrorHandler
	CreateSubcategoryUseCase          usecase.CreateSubcategoryUseCase
	FindSubcategoryByUseCase          usecase.FindSubcategoryByUseCase
	FindSubcategoriesPaginatedUseCase usecase.FindSubcategoriesPaginatedUseCase
	UpdateSubcategoryUseCase          usecase.UpdateSubcategoryUseCase
	RemoveSubcategoryUseCase          usecase.RemoveSubcategoryUseCase
}

const (
	defaultSubcategoryLimitHTTP = 20
	maxSubcategoryLimitHTTP     = 100
)

type SubcategoryHandler struct {
	deps SubcategoryHandlerDeps
}

func NewSubcategoryHandler(deps SubcategoryHandlerDeps) *SubcategoryHandler {
	return &SubcategoryHandler{deps: deps}
}

func (h *SubcategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.deps.O11y.Tracer().Start(r.Context(), "subcategory_handler.create")
	defer span.End()

	correlationID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()

	user, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		h.deps.ErrorHandler.HandleError(w, r, err)
		return
	}

	categoryID := chi.URLParam(r, "categoryId")

	h.deps.O11y.Logger().Info(ctx, "request_received",
		observability.String("operation", "CreateSubcategory"),
		observability.String("layer", "handler"),
		observability.String("entity", "subcategory"),
		observability.String("correlation_id", correlationID),
		observability.String("user_id", user.ID),
		observability.String("category_id", categoryID),
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

	output, err := h.deps.CreateSubcategoryUseCase.Execute(ctx, user.ID, categoryID, input)
	if err != nil {
		h.deps.ErrorHandler.HandleError(w, r, err)
		return
	}

	h.deps.O11y.Logger().Info(ctx, "request_completed",
		observability.String("operation", "CreateSubcategory"),
		observability.String("layer", "handler"),
		observability.String("entity", "subcategory"),
		observability.String("correlation_id", correlationID),
		observability.String("user_id", user.ID),
		observability.String("subcategory_id", output.ID),
	)

	responses.JSON(w, http.StatusCreated, output)
}

func (h *SubcategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.deps.O11y.Tracer().Start(r.Context(), "subcategory_handler.list")
	defer span.End()

	correlationID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()

	user, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		h.deps.ErrorHandler.HandleError(w, r, err)
		return
	}

	categoryID := chi.URLParam(r, "categoryId")

	h.deps.O11y.Logger().Info(ctx, "request_received",
		observability.String("operation", "ListSubcategories"),
		observability.String("layer", "handler"),
		observability.String("entity", "subcategory"),
		observability.String("correlation_id", correlationID),
		observability.String("user_id", user.ID),
		observability.String("category_id", categoryID),
	)

	params, err := pagination.ParseCursorParams(r, defaultSubcategoryLimitHTTP, maxSubcategoryLimitHTTP)
	if err != nil {
		h.deps.ErrorHandler.HandleError(w, r, err)
		return
	}

	output, err := h.deps.FindSubcategoriesPaginatedUseCase.Execute(ctx, user.ID, categoryID, params.Limit, params.Cursor)
	if err != nil {
		h.deps.ErrorHandler.HandleError(w, r, err)
		return
	}

	h.deps.O11y.Logger().Info(ctx, "request_completed",
		observability.String("operation", "ListSubcategories"),
		observability.String("layer", "handler"),
		observability.String("entity", "subcategory"),
		observability.String("correlation_id", correlationID),
		observability.String("user_id", user.ID),
		observability.String("category_id", categoryID),
	)

	responses.JSON(w, http.StatusOK, output)
}

func (h *SubcategoryHandler) FindBy(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.deps.O11y.Tracer().Start(r.Context(), "subcategory_handler.find_by")
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
		observability.String("operation", "FindSubcategoryBy"),
		observability.String("layer", "handler"),
		observability.String("entity", "subcategory"),
		observability.String("correlation_id", correlationID),
		observability.String("user_id", user.ID),
		observability.String("category_id", categoryID),
		observability.String("subcategory_id", id),
	)

	output, err := h.deps.FindSubcategoryByUseCase.Execute(ctx, user.ID, categoryID, id)
	if err != nil {
		h.deps.ErrorHandler.HandleError(w, r, err)
		return
	}

	h.deps.O11y.Logger().Info(ctx, "request_completed",
		observability.String("operation", "FindSubcategoryBy"),
		observability.String("layer", "handler"),
		observability.String("entity", "subcategory"),
		observability.String("correlation_id", correlationID),
		observability.String("user_id", user.ID),
		observability.String("category_id", categoryID),
		observability.String("subcategory_id", id),
	)

	responses.JSON(w, http.StatusOK, output)
}
