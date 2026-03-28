package platformhttp

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/LeviLunique/coralhub-backend/internal/modules/choirs"
	"github.com/LeviLunique/coralhub-backend/internal/modules/memberships"
	"github.com/LeviLunique/coralhub-backend/internal/modules/tenants"
	moduleusers "github.com/LeviLunique/coralhub-backend/internal/modules/users"
)

func TestNewRouterHealthEndpoints(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	router := NewRouter(logger, nil, nil, nil, nil)

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
	router := NewRouter(logger, service, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/public/tenants/coral-jovem-asa-norte", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("tenant bootstrap returned %d, want %d", recorder.Code, http.StatusOK)
	}
}

func TestNewRouterChoirCreateEndpointRequiresActorContext(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	tenantService := tenants.NewService(&tenantStubRepository{
		tenant: tenants.Context{
			ID:          "6f3c194e-635c-4df4-aa64-e1f95c8f5542",
			Slug:        "coral-jovem-asa-norte",
			DisplayName: "Coral Jovem Asa Norte",
		},
	})
	userService := moduleusers.NewService(&userStubRepository{
		user: moduleusers.User{
			ID:       "4ab4f4a4-a208-44dc-bf90-7a4e0d65ea7c",
			TenantID: "6f3c194e-635c-4df4-aa64-e1f95c8f5542",
			Email:    "ana@example.com",
			FullName: "Ana Clara",
			Active:   true,
		},
	})
	choirService := choirs.NewService(&choirStubRepository{
		choir: choirs.Choir{
			ID:       "a9eaee4d-e539-488e-90da-c655637ee9b7",
			TenantID: "6f3c194e-635c-4df4-aa64-e1f95c8f5542",
			Name:     "Sopranos",
			Active:   true,
		},
	})
	router := NewRouter(logger, tenantService, choirService, userService, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/choirs", strings.NewReader(`{"name":"Sopranos"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-Slug", "coral-jovem-asa-norte")
	req.Header.Set("X-User-Email", "ana@example.com")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("choir create returned %d, want %d", recorder.Code, http.StatusCreated)
	}
}

func TestNewRouterUserListEndpointRequiresTenantHeader(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	tenantService := tenants.NewService(&tenantStubRepository{
		tenant: tenants.Context{
			ID:          "6f3c194e-635c-4df4-aa64-e1f95c8f5542",
			Slug:        "coral-jovem-asa-norte",
			DisplayName: "Coral Jovem Asa Norte",
		},
	})
	userService := moduleusers.NewService(&userStubRepository{
		users: []moduleusers.User{{ID: "user-1", Email: "ana@example.com", FullName: "Ana Clara", Active: true}},
	})
	router := NewRouter(logger, tenantService, nil, userService, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("user list returned %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestNewRouterMembershipListEndpointRequiresActorHeader(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	tenantService := tenants.NewService(&tenantStubRepository{
		tenant: tenants.Context{
			ID:          "6f3c194e-635c-4df4-aa64-e1f95c8f5542",
			Slug:        "coral-jovem-asa-norte",
			DisplayName: "Coral Jovem Asa Norte",
		},
	})
	userService := moduleusers.NewService(&userStubRepository{})
	membershipService := memberships.NewService(&membershipStubRepository{})
	router := NewRouter(logger, tenantService, nil, userService, membershipService)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/choirs/choir-1/memberships", nil)
	req.Header.Set("X-Tenant-Slug", "coral-jovem-asa-norte")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("membership list returned %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

type tenantStubRepository struct {
	bootstrap tenants.Bootstrap
	tenant    tenants.Context
	err       error
}

func (s *tenantStubRepository) GetBootstrapBySlug(_ context.Context, _ string) (tenants.Bootstrap, error) {
	if s.err != nil {
		return tenants.Bootstrap{}, s.err
	}

	return s.bootstrap, nil
}

func (s *tenantStubRepository) GetContextBySlug(_ context.Context, _ string) (tenants.Context, error) {
	if s.err != nil {
		return tenants.Context{}, s.err
	}

	return s.tenant, nil
}

type choirStubRepository struct {
	choir  choirs.Choir
	choirs []choirs.Choir
	err    error
}

func (s *choirStubRepository) Create(_ context.Context, _ choirs.CreateParams) (choirs.Choir, error) {
	if s.err != nil {
		return choirs.Choir{}, s.err
	}

	return s.choir, nil
}

func (s *choirStubRepository) GetByIDForMember(_ context.Context, _, _, _ string) (choirs.Choir, error) {
	if s.err != nil {
		return choirs.Choir{}, s.err
	}

	return s.choir, nil
}

func (s *choirStubRepository) ListByMemberUserID(_ context.Context, _, _ string) ([]choirs.Choir, error) {
	if s.err != nil {
		return nil, s.err
	}

	return s.choirs, nil
}

type userStubRepository struct {
	user  moduleusers.User
	users []moduleusers.User
	err   error
}

func (s *userStubRepository) Create(_ context.Context, _ moduleusers.CreateParams) (moduleusers.User, error) {
	if s.err != nil {
		return moduleusers.User{}, s.err
	}

	return s.user, nil
}

func (s *userStubRepository) GetByID(_ context.Context, _, _ string) (moduleusers.User, error) {
	if s.err != nil {
		return moduleusers.User{}, s.err
	}

	return s.user, nil
}

func (s *userStubRepository) GetByEmail(_ context.Context, _, _ string) (moduleusers.User, error) {
	if s.err != nil {
		return moduleusers.User{}, s.err
	}

	return s.user, nil
}

func (s *userStubRepository) ListByTenantID(_ context.Context, _ string) ([]moduleusers.User, error) {
	if s.err != nil {
		return nil, s.err
	}

	return s.users, nil
}

type membershipStubRepository struct {
	membership  memberships.Membership
	memberships []memberships.Membership
	err         error
}

func (s *membershipStubRepository) Create(_ context.Context, _ memberships.CreateParams) (memberships.Membership, error) {
	if s.err != nil {
		return memberships.Membership{}, s.err
	}

	return s.membership, nil
}

func (s *membershipStubRepository) GetByChoirAndUser(_ context.Context, _, _, _ string) (memberships.Membership, error) {
	if s.err != nil {
		return memberships.Membership{}, s.err
	}

	return s.membership, nil
}

func (s *membershipStubRepository) ListByChoirID(_ context.Context, _, _ string) ([]memberships.Membership, error) {
	if s.err != nil {
		return nil, s.err
	}

	return s.memberships, nil
}
