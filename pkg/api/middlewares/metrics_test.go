package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jailtonjunior94/financial/pkg/api/middlewares"

	"github.com/JailtonJunior94/devkit-go/pkg/observability/fake"
	"github.com/go-chi/chi/v5"
)

func TestMetricsMiddleware(t *testing.T) {
	o11y := fake.NewProvider()
	middleware := middlewares.NewMetricsMiddleware(o11y)

	router := chi.NewRouter()
	router.Use(middleware.Handler)
	router.Get("/api/v1/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}
