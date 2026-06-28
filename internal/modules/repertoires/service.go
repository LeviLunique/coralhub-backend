package repertoires

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
	ErrInvalidRepertoireID = errors.New("invalid repertoire id")
	ErrInvalidName         = errors.New("invalid repertoire name")
	ErrInvalidSongID       = errors.New("invalid song id")
	ErrRepertoireNotFound  = errors.New("repertoire not found")
	ErrRepertoireNameTaken = errors.New("repertoire name already exists")
	ErrSongLinkExists      = errors.New("song already linked to repertoire")
	ErrSongLinkNotFound    = errors.New("song is not linked to repertoire")
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

func (s *Service) Create(ctx context.Context, tenantID string, choirID string, actorUserID string, input CreateInput) (Repertoire, error) {
	normalizedTenantID := strings.TrimSpace(tenantID)
	if normalizedTenantID == "" {
		return Repertoire{}, ErrInvalidTenantID
	}

	normalizedChoirID := strings.TrimSpace(choirID)
	if normalizedChoirID == "" {
		return Repertoire{}, ErrInvalidChoirID
	}

	normalizedActorID := strings.TrimSpace(actorUserID)
	if normalizedActorID == "" {
		return Repertoire{}, ErrInvalidActorID
	}

	normalizedName := strings.TrimSpace(input.Name)
	if normalizedName == "" {
		return Repertoire{}, ErrInvalidName
	}

	if err := s.requireManager(ctx, normalizedTenantID, normalizedChoirID, normalizedActorID); err != nil {
		return Repertoire{}, err
	}

	return s.repository.Create(ctx, CreateParams{
		TenantID:    normalizedTenantID,
		ChoirID:     normalizedChoirID,
		Name:        normalizedName,
		Description: normalizeOptionalText(input.Description),
	})
}

func (s *Service) Update(ctx context.Context, tenantID string, repertoireID string, actorUserID string, input UpdateInput) (Repertoire, error) {
	normalizedTenantID := strings.TrimSpace(tenantID)
	if normalizedTenantID == "" {
		return Repertoire{}, ErrInvalidTenantID
	}

	normalizedRepertoireID := strings.TrimSpace(repertoireID)
	if normalizedRepertoireID == "" {
		return Repertoire{}, ErrInvalidRepertoireID
	}

	normalizedActorID := strings.TrimSpace(actorUserID)
	if normalizedActorID == "" {
		return Repertoire{}, ErrInvalidActorID
	}

	normalizedName := strings.TrimSpace(input.Name)
	if normalizedName == "" {
		return Repertoire{}, ErrInvalidName
	}

	existing, err := s.repository.GetByIDForMember(ctx, normalizedTenantID, normalizedRepertoireID, normalizedActorID)
	if err != nil {
		return Repertoire{}, err
	}

	if err := s.requireManager(ctx, normalizedTenantID, existing.ChoirID, normalizedActorID); err != nil {
		return Repertoire{}, err
	}

	return s.repository.Update(ctx, UpdateParams{
		TenantID:     normalizedTenantID,
		RepertoireID: normalizedRepertoireID,
		Name:         normalizedName,
		Description:  normalizeOptionalText(input.Description),
	})
}

func (s *Service) Get(ctx context.Context, tenantID string, actorUserID string, repertoireID string) (Repertoire, error) {
	normalizedTenantID := strings.TrimSpace(tenantID)
	if normalizedTenantID == "" {
		return Repertoire{}, ErrInvalidTenantID
	}

	normalizedActorID := strings.TrimSpace(actorUserID)
	if normalizedActorID == "" {
		return Repertoire{}, ErrInvalidActorID
	}

	normalizedRepertoireID := strings.TrimSpace(repertoireID)
	if normalizedRepertoireID == "" {
		return Repertoire{}, ErrInvalidRepertoireID
	}

	return s.repository.GetByIDForMember(ctx, normalizedTenantID, normalizedRepertoireID, normalizedActorID)
}

func (s *Service) ListByChoir(ctx context.Context, tenantID string, choirID string, actorUserID string) ([]Repertoire, error) {
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

func (s *Service) Delete(ctx context.Context, tenantID string, repertoireID string, actorUserID string) error {
	normalizedTenantID := strings.TrimSpace(tenantID)
	if normalizedTenantID == "" {
		return ErrInvalidTenantID
	}

	normalizedRepertoireID := strings.TrimSpace(repertoireID)
	if normalizedRepertoireID == "" {
		return ErrInvalidRepertoireID
	}

	normalizedActorID := strings.TrimSpace(actorUserID)
	if normalizedActorID == "" {
		return ErrInvalidActorID
	}

	existing, err := s.repository.GetByIDForMember(ctx, normalizedTenantID, normalizedRepertoireID, normalizedActorID)
	if err != nil {
		return err
	}

	if err := s.requireManager(ctx, normalizedTenantID, existing.ChoirID, normalizedActorID); err != nil {
		return err
	}

	return s.repository.Archive(ctx, normalizedTenantID, normalizedRepertoireID)
}

func (s *Service) ListSongs(ctx context.Context, tenantID string, repertoireID string, actorUserID string) ([]RepertoireSong, error) {
	normalizedTenantID := strings.TrimSpace(tenantID)
	if normalizedTenantID == "" {
		return nil, ErrInvalidTenantID
	}

	normalizedRepertoireID := strings.TrimSpace(repertoireID)
	if normalizedRepertoireID == "" {
		return nil, ErrInvalidRepertoireID
	}

	normalizedActorID := strings.TrimSpace(actorUserID)
	if normalizedActorID == "" {
		return nil, ErrInvalidActorID
	}

	if _, err := s.repository.GetByIDForMember(ctx, normalizedTenantID, normalizedRepertoireID, normalizedActorID); err != nil {
		return nil, err
	}

	return s.repository.ListSongs(ctx, normalizedTenantID, normalizedRepertoireID)
}

func (s *Service) AddSong(ctx context.Context, tenantID string, repertoireID string, actorUserID string, input AddSongInput) error {
	normalizedTenantID := strings.TrimSpace(tenantID)
	if normalizedTenantID == "" {
		return ErrInvalidTenantID
	}

	normalizedRepertoireID := strings.TrimSpace(repertoireID)
	if normalizedRepertoireID == "" {
		return ErrInvalidRepertoireID
	}

	normalizedActorID := strings.TrimSpace(actorUserID)
	if normalizedActorID == "" {
		return ErrInvalidActorID
	}

	normalizedSongID := strings.TrimSpace(input.SongID)
	if normalizedSongID == "" {
		return ErrInvalidSongID
	}

	existing, err := s.repository.GetByIDForMember(ctx, normalizedTenantID, normalizedRepertoireID, normalizedActorID)
	if err != nil {
		return err
	}

	if err := s.requireManager(ctx, normalizedTenantID, existing.ChoirID, normalizedActorID); err != nil {
		return err
	}

	return s.repository.AddSong(ctx, AddSongParams{
		TenantID:       normalizedTenantID,
		RepertoireID:   normalizedRepertoireID,
		SongID:         normalizedSongID,
		ExecutionOrder: input.ExecutionOrder,
		Notes:          normalizeOptionalText(input.Notes),
	})
}

func (s *Service) RemoveSong(ctx context.Context, tenantID string, repertoireID string, actorUserID string, songID string) error {
	normalizedTenantID := strings.TrimSpace(tenantID)
	if normalizedTenantID == "" {
		return ErrInvalidTenantID
	}

	normalizedRepertoireID := strings.TrimSpace(repertoireID)
	if normalizedRepertoireID == "" {
		return ErrInvalidRepertoireID
	}

	normalizedActorID := strings.TrimSpace(actorUserID)
	if normalizedActorID == "" {
		return ErrInvalidActorID
	}

	normalizedSongID := strings.TrimSpace(songID)
	if normalizedSongID == "" {
		return ErrInvalidSongID
	}

	existing, err := s.repository.GetByIDForMember(ctx, normalizedTenantID, normalizedRepertoireID, normalizedActorID)
	if err != nil {
		return err
	}

	if err := s.requireManager(ctx, normalizedTenantID, existing.ChoirID, normalizedActorID); err != nil {
		return err
	}

	return s.repository.RemoveSong(ctx, normalizedTenantID, normalizedRepertoireID, normalizedSongID)
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
