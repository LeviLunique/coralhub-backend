package platformhttp

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LeviLunique/coralhub-backend/internal/modules/tenants"
)

func TestNewRouterHealthEndpoints(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	router := NewRouter(logger, nil)

	for _, path := range []string{"/healthz", "/api/v1/healthz"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusOK {
			t.Fatalf("%s returned %d, want %d", path, recorder.Code, http.StatusOK)
		}
	}
}

func TestNewRouterTenantBootstrapEndpoint(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	service := tenants.NewService(&tenantStubRepository{
		bootstrap: tenants.Bootstrap{
			Slug:        "coral-jovem-asa-norte",
			DisplayName: "Coral Jovem Asa Norte",
		},
	})
	router := NewRouter(logger, service)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/public/tenants/coral-jovem-asa-norte", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("tenant bootstrap returned %d, want %d", recorder.Code, http.StatusOK)
	}
}

type tenantStubRepository struct {
	bootstrap tenants.Bootstrap
	err       error
}

func (s *tenantStubRepository) GetBootstrapBySlug(_ context.Context, _ string) (tenants.Bootstrap, error) {
	if s.err != nil {
		return tenants.Bootstrap{}, s.err
	}

	return s.bootstrap, nil
}
