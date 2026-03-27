package tenants

import (
	"context"
	"errors"
	"testing"
)

type stubRepository struct {
	bootstrap Bootstrap
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
