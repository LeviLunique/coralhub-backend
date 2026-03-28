package tenants

import (
	"context"
	"errors"
	"strings"
)

var (
	ErrTenantNotFound    = errors.New("tenant not found")
	ErrInvalidTenantSlug = errors.New("invalid tenant slug")
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) GetBootstrap(ctx context.Context, slug string) (Bootstrap, error) {
	normalizedSlug := strings.TrimSpace(slug)
	if normalizedSlug == "" {
		return Bootstrap{}, ErrInvalidTenantSlug
	}

	tenant, err := s.repository.GetBootstrapBySlug(ctx, normalizedSlug)
	if err != nil {
		return Bootstrap{}, err
	}

	return tenant, nil
}

func (s *Service) ResolveContext(ctx context.Context, slug string) (Context, error) {
	normalizedSlug := strings.TrimSpace(slug)
	if normalizedSlug == "" {
		return Context{}, ErrInvalidTenantSlug
	}

	tenant, err := s.repository.GetContextBySlug(ctx, normalizedSlug)
	if err != nil {
		return Context{}, err
	}

	return tenant, nil
}
