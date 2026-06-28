package platformhttp

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/LeviLunique/coralhub-backend/internal/modules/choirs"
	"github.com/LeviLunique/coralhub-backend/internal/modules/events"
	"github.com/LeviLunique/coralhub-backend/internal/modules/instruments"
	"github.com/LeviLunique/coralhub-backend/internal/modules/materials"
	"github.com/LeviLunique/coralhub-backend/internal/modules/memberships"
	"github.com/LeviLunique/coralhub-backend/internal/modules/tenants"
	moduleusers "github.com/LeviLunique/coralhub-backend/internal/modules/users"
	platformweb "github.com/LeviLunique/coralhub-backend/internal/platform/web"
)

func TestNewRouterHealthEndpoints(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	router := NewRouter(logger, 30*time.Second, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	for _, path := range []string{"/healthz", "/api/v1/healthz"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusOK {
			t.Fatalf("%s returned %d, want %d", path, recorder.Code, http.StatusOK)
		}
	}
}

func TestNewRouterMetricsEndpoint(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	router := NewRouter(logger, 30*time.Second, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("metrics returned %d, want %d", recorder.Code, http.StatusOK)
	}
	if !strings.Contains(recorder.Body.String(), "coralhub_http_requests_total") {
		t.Fatalf("metrics body missing http counter: %s", recorder.Body.String())
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
	router := NewRouter(logger, 30*time.Second, nil, service, nil, nil, nil, nil, nil, nil, nil, nil)

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
	router := NewRouter(logger, 30*time.Second, nil, tenantService, choirService, userService, nil, nil, nil, nil, nil, nil)

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
	router := NewRouter(logger, 30*time.Second, nil, tenantService, nil, userService, nil, nil, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("user list returned %d, want %d", recorder.Code, http.StatusBadRequest)
	}

	var payload platformweb.ErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if payload.Error.Code != "tenant_header_required" {
		t.Fatalf("payload.Error.Code = %q, want %q", payload.Error.Code, "tenant_header_required")
	}
	if payload.Error.RequestID == "" {
		t.Fatal("payload.Error.RequestID is empty")
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
	router := NewRouter(logger, 30*time.Second, nil, tenantService, nil, userService, membershipService, nil, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/choirs/choir-1/memberships", nil)
	req.Header.Set("X-Tenant-Slug", "coral-jovem-asa-norte")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("membership list returned %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestNewRouterMaterialCreateEndpointRequiresManagerActorContext(t *testing.T) {
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
	membershipService := memberships.NewService(&membershipStubRepository{
		membership: memberships.Membership{Role: memberships.RoleManager},
	})
	materialService := materials.NewService(&materialStubRepository{
		material: materials.Material{
			ID:           "material-1",
			TenantID:     "6f3c194e-635c-4df4-aa64-e1f95c8f5542",
			ChoirID:      "choir-1",
			SongID:       "song-1",
			Name:         "Soprano guide",
			MaterialType: materials.MaterialTypeAudioGuide,
			TargetType:   materials.TargetTypeVoice,
		},
	}, &membershipStubRepository{
		membership: memberships.Membership{Role: memberships.RoleManager},
	})
	router := NewRouter(logger, 30*time.Second, nil, tenantService, nil, userService, membershipService, nil, nil, nil, nil, materialService)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/choirs/choir-1/materials", strings.NewReader(`{"song_id":"song-1","name":"Soprano guide","material_type":"audio_guide"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-Slug", "coral-jovem-asa-norte")
	req.Header.Set("X-User-Email", "ana@example.com")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("material create returned %d, want %d", recorder.Code, http.StatusCreated)
	}
}

func TestNewRouterEventCreateEndpointRequiresManagerActorContext(t *testing.T) {
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
	membershipRepository := &membershipStubRepository{
		membership: memberships.Membership{Role: memberships.RoleManager},
		memberships: []memberships.Membership{
			{UserID: "4ab4f4a4-a208-44dc-bf90-7a4e0d65ea7c"},
		},
	}
	eventService := events.NewService(&eventStubRepository{
		event: events.Event{
			ID:        "event-1",
			TenantID:  "6f3c194e-635c-4df4-aa64-e1f95c8f5542",
			ChoirID:   "choir-1",
			Title:     "Main rehearsal",
			EventType: events.EventTypeRehearsal,
			StartAt:   time.Date(2026, 4, 20, 19, 0, 0, 0, time.UTC),
			Active:    true,
		},
	}, membershipRepository)
	router := NewRouter(logger, 30*time.Second, nil, tenantService, nil, userService, nil, eventService, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/choirs/choir-1/events", strings.NewReader(`{"title":"Main rehearsal","event_type":"rehearsal","start_at":"2026-04-20T19:00:00Z"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-Slug", "coral-jovem-asa-norte")
	req.Header.Set("X-User-Email", "ana@example.com")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("event create returned %d, want %d", recorder.Code, http.StatusCreated)
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

type materialStubRepository struct {
	material  materials.Material
	materials []materials.Material
	err       error
}

func (s *materialStubRepository) Create(_ context.Context, _ materials.CreateParams) (materials.Material, error) {
	if s.err != nil {
		return materials.Material{}, s.err
	}

	return s.material, nil
}

func (s *materialStubRepository) Update(_ context.Context, _ materials.UpdateParams) (materials.Material, error) {
	if s.err != nil {
		return materials.Material{}, s.err
	}

	return s.material, nil
}

func (s *materialStubRepository) GetByIDForMember(_ context.Context, _, _, _ string) (materials.Material, error) {
	if s.err != nil {
		return materials.Material{}, s.err
	}

	return s.material, nil
}

func (s *materialStubRepository) ListByChoirID(_ context.Context, _, _ string) ([]materials.Material, error) {
	if s.err != nil {
		return nil, s.err
	}

	return s.materials, nil
}

func (s *materialStubRepository) ListBySongID(_ context.Context, _, _ string) ([]materials.Material, error) {
	if s.err != nil {
		return nil, s.err
	}

	return s.materials, nil
}

func (s *materialStubRepository) Archive(_ context.Context, _, _ string) error {
	return s.err
}

func (s *materialStubRepository) AddInstrument(_ context.Context, _, _, _ string) error {
	return s.err
}

func (s *materialStubRepository) RemoveInstrument(_ context.Context, _, _, _ string) error {
	return s.err
}

func (s *materialStubRepository) ListInstruments(_ context.Context, _, _ string) ([]instruments.Instrument, error) {
	if s.err != nil {
		return nil, s.err
	}

	return nil, nil
}

type eventStubRepository struct {
	event    events.Event
	events   []events.Event
	err      error
	create   events.CreateParams
	update   events.UpdateParams
	canceled events.CancelParams
}

func (s *eventStubRepository) Create(_ context.Context, params events.CreateParams) (events.Event, error) {
	s.create = params
	if s.err != nil {
		return events.Event{}, s.err
	}

	return s.event, nil
}

func (s *eventStubRepository) Update(_ context.Context, params events.UpdateParams) (events.Event, error) {
	s.update = params
	if s.err != nil {
		return events.Event{}, s.err
	}

	return s.event, nil
}

func (s *eventStubRepository) GetByIDForMember(_ context.Context, _, _, _ string) (events.Event, error) {
	if s.err != nil {
		return events.Event{}, s.err
	}

	return s.event, nil
}

func (s *eventStubRepository) ListByChoirID(_ context.Context, _, _ string) ([]events.Event, error) {
	if s.err != nil {
		return nil, s.err
	}

	return s.events, nil
}

func (s *eventStubRepository) Cancel(_ context.Context, params events.CancelParams) error {
	if s.err != nil {
		return s.err
	}

	s.canceled = params
	return nil
}

func (s *eventStubRepository) AddRepertoire(_ context.Context, _, _, _ string) error {
	return s.err
}

func (s *eventStubRepository) RemoveRepertoire(_ context.Context, _, _, _ string) error {
	return s.err
}

func (s *eventStubRepository) ListRepertoires(_ context.Context, _, _ string) ([]events.Repertoire, error) {
	if s.err != nil {
		return nil, s.err
	}

	return nil, nil
}
