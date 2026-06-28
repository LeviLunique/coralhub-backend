package songs

import (
	"context"
	"errors"
	"strings"

	"github.com/LeviLunique/coralhub-backend/internal/modules/memberships"
)

var (
	ErrInvalidTenantID = errors.New("invalid tenant id")
	ErrInvalidChoirID  = errors.New("invalid choir id")
	ErrInvalidActorID  = errors.New("invalid actor id")
	ErrInvalidSongID   = errors.New("invalid song id")
	ErrInvalidTitle    = errors.New("invalid song title")
	ErrSongNotFound    = errors.New("song not found")
	ErrSongTitleTaken  = errors.New("song title already exists")
	ErrForbidden       = errors.New("forbidden")
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

func (s *Service) Create(ctx context.Context, tenantID string, choirID string, actorUserID string, input CreateInput) (Song, error) {
	normalizedTenantID := strings.TrimSpace(tenantID)
	if normalizedTenantID == "" {
		return Song{}, ErrInvalidTenantID
	}

	normalizedChoirID := strings.TrimSpace(choirID)
	if normalizedChoirID == "" {
		return Song{}, ErrInvalidChoirID
	}

	normalizedActorID := strings.TrimSpace(actorUserID)
	if normalizedActorID == "" {
		return Song{}, ErrInvalidActorID
	}

	normalizedTitle := strings.TrimSpace(input.Title)
	if normalizedTitle == "" {
		return Song{}, ErrInvalidTitle
	}

	if err := s.requireManager(ctx, normalizedTenantID, normalizedChoirID, normalizedActorID); err != nil {
		return Song{}, err
	}

	return s.repository.Create(ctx, CreateParams{
		TenantID: normalizedTenantID,
		ChoirID:  normalizedChoirID,
		Title:    normalizedTitle,
		Composer: normalizeOptionalText(input.Composer),
		Arranger: normalizeOptionalText(input.Arranger),
		SongKey:  normalizeOptionalText(input.SongKey),
		Duration: normalizeOptionalText(input.Duration),
		Notes:    normalizeOptionalText(input.Notes),
	})
}

func (s *Service) Update(ctx context.Context, tenantID string, songID string, actorUserID string, input UpdateInput) (Song, error) {
	normalizedTenantID := strings.TrimSpace(tenantID)
	if normalizedTenantID == "" {
		return Song{}, ErrInvalidTenantID
	}

	normalizedSongID := strings.TrimSpace(songID)
	if normalizedSongID == "" {
		return Song{}, ErrInvalidSongID
	}

	normalizedActorID := strings.TrimSpace(actorUserID)
	if normalizedActorID == "" {
		return Song{}, ErrInvalidActorID
	}

	normalizedTitle := strings.TrimSpace(input.Title)
	if normalizedTitle == "" {
		return Song{}, ErrInvalidTitle
	}

	existing, err := s.repository.GetByIDForMember(ctx, normalizedTenantID, normalizedSongID, normalizedActorID)
	if err != nil {
		return Song{}, err
	}

	if err := s.requireManager(ctx, normalizedTenantID, existing.ChoirID, normalizedActorID); err != nil {
		return Song{}, err
	}

	return s.repository.Update(ctx, UpdateParams{
		TenantID: normalizedTenantID,
		SongID:   normalizedSongID,
		Title:    normalizedTitle,
		Composer: normalizeOptionalText(input.Composer),
		Arranger: normalizeOptionalText(input.Arranger),
		SongKey:  normalizeOptionalText(input.SongKey),
		Duration: normalizeOptionalText(input.Duration),
		Notes:    normalizeOptionalText(input.Notes),
	})
}

func (s *Service) Get(ctx context.Context, tenantID string, actorUserID string, songID string) (Song, error) {
	normalizedTenantID := strings.TrimSpace(tenantID)
	if normalizedTenantID == "" {
		return Song{}, ErrInvalidTenantID
	}

	normalizedActorID := strings.TrimSpace(actorUserID)
	if normalizedActorID == "" {
		return Song{}, ErrInvalidActorID
	}

	normalizedSongID := strings.TrimSpace(songID)
	if normalizedSongID == "" {
		return Song{}, ErrInvalidSongID
	}

	return s.repository.GetByIDForMember(ctx, normalizedTenantID, normalizedSongID, normalizedActorID)
}

func (s *Service) ListByChoir(ctx context.Context, tenantID string, choirID string, actorUserID string) ([]Song, error) {
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

func (s *Service) Delete(ctx context.Context, tenantID string, songID string, actorUserID string) error {
	normalizedTenantID := strings.TrimSpace(tenantID)
	if normalizedTenantID == "" {
		return ErrInvalidTenantID
	}

	normalizedSongID := strings.TrimSpace(songID)
	if normalizedSongID == "" {
		return ErrInvalidSongID
	}

	normalizedActorID := strings.TrimSpace(actorUserID)
	if normalizedActorID == "" {
		return ErrInvalidActorID
	}

	existing, err := s.repository.GetByIDForMember(ctx, normalizedTenantID, normalizedSongID, normalizedActorID)
	if err != nil {
		return err
	}

	if err := s.requireManager(ctx, normalizedTenantID, existing.ChoirID, normalizedActorID); err != nil {
		return err
	}

	return s.repository.Archive(ctx, normalizedTenantID, normalizedSongID)
}

func (s *Service) requireManager(ctx context.Context, tenantID string, choirID string, actorUserID string) error {
	member, err := s.memberships.GetByChoirAndUser(ctx, tenantID, choirID, actorUserID)
	if err != nil {
		if errors.Is(err, memberships.ErrMembershipNotFound) {
			return ErrForbidden
		}

		return err
	}

	if member.Role != memberships.RoleManager {
		return ErrForbidden
	}

	return nil
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
