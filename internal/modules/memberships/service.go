package memberships

import (
	"context"
	"errors"
	"strings"
)

var (
	ErrInvalidTenantID        = errors.New("invalid tenant id")
	ErrInvalidChoirID         = errors.New("invalid choir id")
	ErrInvalidUserID          = errors.New("invalid user id")
	ErrInvalidRole            = errors.New("invalid role")
	ErrMembershipNotFound     = errors.New("membership not found")
	ErrMembershipAlreadyExist = errors.New("membership already exists")
	ErrForbidden              = errors.New("forbidden")
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) AddMember(ctx context.Context, tenantID string, choirID string, actorUserID string, input CreateInput) (Membership, error) {
	normalizedTenantID := strings.TrimSpace(tenantID)
	if normalizedTenantID == "" {
		return Membership{}, ErrInvalidTenantID
	}

	normalizedChoirID := strings.TrimSpace(choirID)
	if normalizedChoirID == "" {
		return Membership{}, ErrInvalidChoirID
	}

	normalizedActorID := strings.TrimSpace(actorUserID)
	if normalizedActorID == "" {
		return Membership{}, ErrInvalidUserID
	}

	managerMembership, err := s.repository.GetByChoirAndUser(ctx, normalizedTenantID, normalizedChoirID, normalizedActorID)
	if err != nil {
		return Membership{}, err
	}

	if managerMembership.Role != RoleManager {
		return Membership{}, ErrForbidden
	}

	normalizedUserID := strings.TrimSpace(input.UserID)
	if normalizedUserID == "" {
		return Membership{}, ErrInvalidUserID
	}

	role := strings.TrimSpace(strings.ToLower(input.Role))
	if role != RoleManager && role != RoleMember {
		return Membership{}, ErrInvalidRole
	}

	return s.repository.Create(ctx, CreateParams{
		TenantID:    normalizedTenantID,
		ChoirID:     normalizedChoirID,
		UserID:      normalizedUserID,
		Role:        role,
		ActorUserID: normalizedActorID,
	})
}

func (s *Service) ListByChoir(ctx context.Context, tenantID string, choirID string, actorUserID string) ([]Membership, error) {
	normalizedTenantID := strings.TrimSpace(tenantID)
	if normalizedTenantID == "" {
		return nil, ErrInvalidTenantID
	}

	normalizedChoirID := strings.TrimSpace(choirID)
	if normalizedChoirID == "" {
		return nil, ErrInvalidChoirID
	}

	normalizedActorID := strings.TrimSpace(actorUserID)
	if normalizedActorID == "" {
		return nil, ErrInvalidUserID
	}

	if _, err := s.repository.GetByChoirAndUser(ctx, normalizedTenantID, normalizedChoirID, normalizedActorID); err != nil {
		return nil, err
	}

	return s.repository.ListByChoirID(ctx, normalizedTenantID, normalizedChoirID)
}
