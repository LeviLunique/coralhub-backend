package choirs

import (
	"context"
	"errors"
	"strings"
)

var (
	ErrInvalidChoirName = errors.New("invalid choir name")
	ErrInvalidChoirID   = errors.New("invalid choir id")
	ErrInvalidActorID   = errors.New("invalid actor id")
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

func (s *Service) Create(ctx context.Context, tenantID string, actorUserID string, input CreateInput) (Choir, error) {
	normalizedTenantID := strings.TrimSpace(tenantID)
	if normalizedTenantID == "" {
		return Choir{}, ErrInvalidTenantID
	}

	normalizedActorUserID := strings.TrimSpace(actorUserID)
	if normalizedActorUserID == "" {
		return Choir{}, ErrInvalidActorID
	}

	normalizedName := strings.TrimSpace(input.Name)
	if normalizedName == "" {
		return Choir{}, ErrInvalidChoirName
	}

	return s.repository.Create(ctx, CreateParams{
		ActorUserID: normalizedActorUserID,
		TenantID:    normalizedTenantID,
		Name:        normalizedName,
		Description: normalizeOptionalText(input.Description),
	})
}

func (s *Service) Get(ctx context.Context, tenantID string, actorUserID string, choirID string) (Choir, error) {
	normalizedTenantID := strings.TrimSpace(tenantID)
	if normalizedTenantID == "" {
		return Choir{}, ErrInvalidTenantID
	}

	normalizedActorUserID := strings.TrimSpace(actorUserID)
	if normalizedActorUserID == "" {
		return Choir{}, ErrInvalidActorID
	}

	normalizedChoirID := strings.TrimSpace(choirID)
	if normalizedChoirID == "" {
		return Choir{}, ErrInvalidChoirID
	}

	return s.repository.GetByIDForMember(ctx, normalizedTenantID, normalizedChoirID, normalizedActorUserID)
}

func (s *Service) List(ctx context.Context, tenantID string, actorUserID string) ([]Choir, error) {
	normalizedTenantID := strings.TrimSpace(tenantID)
	if normalizedTenantID == "" {
		return nil, ErrInvalidTenantID
	}

	normalizedActorUserID := strings.TrimSpace(actorUserID)
	if normalizedActorUserID == "" {
		return nil, ErrInvalidActorID
	}

	return s.repository.ListByMemberUserID(ctx, normalizedTenantID, normalizedActorUserID)
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
