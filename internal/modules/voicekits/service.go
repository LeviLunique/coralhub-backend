package voicekits

import (
	"context"
	"errors"
	"strings"

	"github.com/LeviLunique/coralhub-backend/internal/modules/memberships"
)

var (
	ErrInvalidTenantID     = errors.New("invalid tenant id")
	ErrInvalidChoirID      = errors.New("invalid choir id")
	ErrInvalidActorID      = errors.New("invalid actor id")
	ErrInvalidVoiceKitID   = errors.New("invalid voice kit id")
	ErrInvalidVoiceKitName = errors.New("invalid voice kit name")
	ErrVoiceKitNotFound    = errors.New("voice kit not found")
	ErrVoiceKitNameTaken   = errors.New("voice kit name already exists")
	ErrForbidden           = errors.New("forbidden")
)

type membershipChecker interface {
	GetByChoirAndUser(ctx context.Context, tenantID string, choirID string, userID string) (memberships.Membership, error)
}

type Service struct {
	repository  Repository
	memberships membershipChecker
}

func NewService(repository Repository, memberships membershipChecker) *Service {
	return &Service{repository: repository, memberships: memberships}
}

func (s *Service) Create(ctx context.Context, tenantID string, choirID string, actorUserID string, input CreateInput) (VoiceKit, error) {
	normalizedTenantID := strings.TrimSpace(tenantID)
	if normalizedTenantID == "" {
		return VoiceKit{}, ErrInvalidTenantID
	}

	normalizedChoirID := strings.TrimSpace(choirID)
	if normalizedChoirID == "" {
		return VoiceKit{}, ErrInvalidChoirID
	}

	normalizedActorID := strings.TrimSpace(actorUserID)
	if normalizedActorID == "" {
		return VoiceKit{}, ErrInvalidActorID
	}

	normalizedName := strings.TrimSpace(input.Name)
	if normalizedName == "" {
		return VoiceKit{}, ErrInvalidVoiceKitName
	}

	member, err := s.memberships.GetByChoirAndUser(ctx, normalizedTenantID, normalizedChoirID, normalizedActorID)
	if err != nil {
		if errors.Is(err, memberships.ErrMembershipNotFound) {
			return VoiceKit{}, ErrForbidden
		}

		return VoiceKit{}, err
	}

	if member.Role != memberships.RoleManager {
		return VoiceKit{}, ErrForbidden
	}

	return s.repository.Create(ctx, CreateParams{
		TenantID:    normalizedTenantID,
		ChoirID:     normalizedChoirID,
		Name:        normalizedName,
		Description: normalizeOptionalText(input.Description),
	})
}

func (s *Service) Get(ctx context.Context, tenantID string, actorUserID string, voiceKitID string) (VoiceKit, error) {
	normalizedTenantID := strings.TrimSpace(tenantID)
	if normalizedTenantID == "" {
		return VoiceKit{}, ErrInvalidTenantID
	}

	normalizedActorID := strings.TrimSpace(actorUserID)
	if normalizedActorID == "" {
		return VoiceKit{}, ErrInvalidActorID
	}

	normalizedVoiceKitID := strings.TrimSpace(voiceKitID)
	if normalizedVoiceKitID == "" {
		return VoiceKit{}, ErrInvalidVoiceKitID
	}

	return s.repository.GetByIDForMember(ctx, normalizedTenantID, normalizedVoiceKitID, normalizedActorID)
}

func (s *Service) ListByChoir(ctx context.Context, tenantID string, choirID string, actorUserID string) ([]VoiceKit, error) {
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
		return nil, ErrInvalidActorID
	}

	if _, err := s.memberships.GetByChoirAndUser(ctx, normalizedTenantID, normalizedChoirID, normalizedActorID); err != nil {
		if errors.Is(err, memberships.ErrMembershipNotFound) {
			return nil, ErrForbidden
		}

		return nil, err
	}

	return s.repository.ListByChoirID(ctx, normalizedTenantID, normalizedChoirID)
}

func (s *Service) Delete(ctx context.Context, tenantID string, voiceKitID string, actorUserID string) error {
	normalizedTenantID := strings.TrimSpace(tenantID)
	if normalizedTenantID == "" {
		return ErrInvalidTenantID
	}

	normalizedVoiceKitID := strings.TrimSpace(voiceKitID)
	if normalizedVoiceKitID == "" {
		return ErrInvalidVoiceKitID
	}

	normalizedActorID := strings.TrimSpace(actorUserID)
	if normalizedActorID == "" {
		return ErrInvalidActorID
	}

	voiceKit, err := s.repository.GetByIDForMember(ctx, normalizedTenantID, normalizedVoiceKitID, normalizedActorID)
	if err != nil {
		return err
	}

	member, err := s.memberships.GetByChoirAndUser(ctx, normalizedTenantID, voiceKit.ChoirID, normalizedActorID)
	if err != nil {
		if errors.Is(err, memberships.ErrMembershipNotFound) {
			return ErrForbidden
		}

		return err
	}

	if member.Role != memberships.RoleManager {
		return ErrForbidden
	}

	return s.repository.Delete(ctx, normalizedTenantID, normalizedVoiceKitID)
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
