package choirs

import (
	"context"
	"errors"
	"strings"
)

var (
	ErrInvalidChoirName = errors.New("invalid choir name")
	ErrInvalidChoirID   = errors.New("invalid choir id")
	ErrInvalidTenantID  = errors.New("invalid tenant id")
	ErrChoirNotFound    = errors.New("choir not found")
	ErrChoirNameTaken   = errors.New("choir name already exists")
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Create(ctx context.Context, tenantID string, input CreateInput) (Choir, error) {
	normalizedTenantID := strings.TrimSpace(tenantID)
	if normalizedTenantID == "" {
		return Choir{}, ErrInvalidTenantID
	}

	normalizedName := strings.TrimSpace(input.Name)
	if normalizedName == "" {
		return Choir{}, ErrInvalidChoirName
	}

	return s.repository.Create(ctx, CreateParams{
		TenantID:    normalizedTenantID,
		Name:        normalizedName,
		Description: normalizeOptionalText(input.Description),
	})
}

func (s *Service) Get(ctx context.Context, tenantID string, choirID string) (Choir, error) {
	normalizedTenantID := strings.TrimSpace(tenantID)
	if normalizedTenantID == "" {
		return Choir{}, ErrInvalidTenantID
	}

	normalizedChoirID := strings.TrimSpace(choirID)
	if normalizedChoirID == "" {
		return Choir{}, ErrInvalidChoirID
	}

	return s.repository.GetByID(ctx, normalizedTenantID, normalizedChoirID)
}

func (s *Service) List(ctx context.Context, tenantID string) ([]Choir, error) {
	normalizedTenantID := strings.TrimSpace(tenantID)
	if normalizedTenantID == "" {
		return nil, ErrInvalidTenantID
	}

	return s.repository.ListByTenantID(ctx, normalizedTenantID)
}

func normalizeOptionalText(value *string) *string {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}
