package tenants

import (
	"context"
	"errors"
	"testing"
)

type stubRepository struct {
	bootstrap Bootstrap
	tenant    Context
	err       error
	slug      string
}

func (s *stubRepository) GetBootstrapBySlug(_ context.Context, slug string) (Bootstrap, error) {
	s.slug = slug
	if s.err != nil {
		return Bootstrap{}, s.err
	}

	return s.bootstrap, nil
}

func (s *stubRepository) GetContextBySlug(_ context.Context, slug string) (Context, error) {
	s.slug = slug
	if s.err != nil {
		return Context{}, s.err
	}

	return s.tenant, nil
}

func TestServiceGetBootstrapTrimsSlug(t *testing.T) {
	repository := &stubRepository{
		bootstrap: Bootstrap{
			Slug:        "coral-jovem-asa-norte",
			DisplayName: "Coral Jovem Asa Norte",
		},
	}

	service := NewService(repository)
	tenant, err := service.GetBootstrap(context.Background(), "  coral-jovem-asa-norte  ")
	if err != nil {
		t.Fatalf("GetBootstrap() error = %v", err)
	}

	if repository.slug != "coral-jovem-asa-norte" {
		t.Fatalf("repository slug = %q, want trimmed slug", repository.slug)
	}

	if tenant.DisplayName != "Coral Jovem Asa Norte" {
		t.Fatalf("tenant.DisplayName = %q", tenant.DisplayName)
	}
}

func TestServiceGetBootstrapRejectsEmptySlug(t *testing.T) {
	service := NewService(&stubRepository{})

	_, err := service.GetBootstrap(context.Background(), "   ")
	if !errors.Is(err, ErrInvalidTenantSlug) {
		t.Fatalf("GetBootstrap() error = %v, want %v", err, ErrInvalidTenantSlug)
	}
}

func TestServiceResolveContextTrimsSlug(t *testing.T) {
	repository := &stubRepository{
		tenant: Context{
			ID:          "6f3c194e-635c-4df4-aa64-e1f95c8f5542",
			Slug:        "coral-jovem-asa-norte",
			DisplayName: "Coral Jovem Asa Norte",
		},
	}

	service := NewService(repository)
	tenant, err := service.ResolveContext(context.Background(), "  coral-jovem-asa-norte  ")
	if err != nil {
		t.Fatalf("ResolveContext() error = %v", err)
	}

	if repository.slug != "coral-jovem-asa-norte" {
		t.Fatalf("repository slug = %q, want trimmed slug", repository.slug)
	}

	if tenant.ID == "" {
		t.Fatal("tenant.ID is empty")
	}
}
