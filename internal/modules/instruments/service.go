package instruments

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
	ErrInvalidInstrumentID = errors.New("invalid instrument id")
	ErrInvalidName         = errors.New("invalid instrument name")
	ErrInvalidUserID       = errors.New("invalid user id")
	ErrInstrumentNotFound  = errors.New("instrument not found")
	ErrInstrumentNameTaken = errors.New("instrument name already exists")
	ErrUserLinkExists      = errors.New("user already linked to instrument")
	ErrUserLinkNotFound    = errors.New("user is not linked to instrument")
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

func (s *Service) Create(ctx context.Context, tenantID string, choirID string, actorUserID string, input CreateInput) (Instrument, error) {
	normalizedTenantID := strings.TrimSpace(tenantID)
	if normalizedTenantID == "" {
		return Instrument{}, ErrInvalidTenantID
	}

	normalizedChoirID := strings.TrimSpace(choirID)
	if normalizedChoirID == "" {
		return Instrument{}, ErrInvalidChoirID
	}

	normalizedActorID := strings.TrimSpace(actorUserID)
	if normalizedActorID == "" {
		return Instrument{}, ErrInvalidActorID
	}

	normalizedName := strings.TrimSpace(input.Name)
	if normalizedName == "" {
		return Instrument{}, ErrInvalidName
	}

	if err := s.requireManager(ctx, normalizedTenantID, normalizedChoirID, normalizedActorID); err != nil {
		return Instrument{}, err
	}

	return s.repository.Create(ctx, CreateParams{
		TenantID:    normalizedTenantID,
		ChoirID:     normalizedChoirID,
		Name:        normalizedName,
		Description: normalizeOptionalText(input.Description),
		Icon:        normalizeOptionalText(input.Icon),
	})
}

func (s *Service) Update(ctx context.Context, tenantID string, instrumentID string, actorUserID string, input UpdateInput) (Instrument, error) {
	normalizedTenantID := strings.TrimSpace(tenantID)
	if normalizedTenantID == "" {
		return Instrument{}, ErrInvalidTenantID
	}

	normalizedInstrumentID := strings.TrimSpace(instrumentID)
	if normalizedInstrumentID == "" {
		return Instrument{}, ErrInvalidInstrumentID
	}

	normalizedActorID := strings.TrimSpace(actorUserID)
	if normalizedActorID == "" {
		return Instrument{}, ErrInvalidActorID
	}

	normalizedName := strings.TrimSpace(input.Name)
	if normalizedName == "" {
		return Instrument{}, ErrInvalidName
	}

	existing, err := s.repository.GetByIDForMember(ctx, normalizedTenantID, normalizedInstrumentID, normalizedActorID)
	if err != nil {
		return Instrument{}, err
	}

	if err := s.requireManager(ctx, normalizedTenantID, existing.ChoirID, normalizedActorID); err != nil {
		return Instrument{}, err
	}

	return s.repository.Update(ctx, UpdateParams{
		TenantID:     normalizedTenantID,
		InstrumentID: normalizedInstrumentID,
		Name:         normalizedName,
		Description:  normalizeOptionalText(input.Description),
		Icon:         normalizeOptionalText(input.Icon),
	})
}

func (s *Service) Get(ctx context.Context, tenantID string, actorUserID string, instrumentID string) (Instrument, error) {
	normalizedTenantID := strings.TrimSpace(tenantID)
	if normalizedTenantID == "" {
		return Instrument{}, ErrInvalidTenantID
	}

	normalizedActorID := strings.TrimSpace(actorUserID)
	if normalizedActorID == "" {
		return Instrument{}, ErrInvalidActorID
	}

	normalizedInstrumentID := strings.TrimSpace(instrumentID)
	if normalizedInstrumentID == "" {
		return Instrument{}, ErrInvalidInstrumentID
	}

	return s.repository.GetByIDForMember(ctx, normalizedTenantID, normalizedInstrumentID, normalizedActorID)
}

func (s *Service) ListByChoir(ctx context.Context, tenantID string, choirID string, actorUserID string) ([]Instrument, error) {
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

func (s *Service) Delete(ctx context.Context, tenantID string, instrumentID string, actorUserID string) error {
	normalizedTenantID := strings.TrimSpace(tenantID)
	if normalizedTenantID == "" {
		return ErrInvalidTenantID
	}

	normalizedInstrumentID := strings.TrimSpace(instrumentID)
	if normalizedInstrumentID == "" {
		return ErrInvalidInstrumentID
	}

	normalizedActorID := strings.TrimSpace(actorUserID)
	if normalizedActorID == "" {
		return ErrInvalidActorID
	}

	existing, err := s.repository.GetByIDForMember(ctx, normalizedTenantID, normalizedInstrumentID, normalizedActorID)
	if err != nil {
		return err
	}

	if err := s.requireManager(ctx, normalizedTenantID, existing.ChoirID, normalizedActorID); err != nil {
		return err
	}

	return s.repository.Archive(ctx, normalizedTenantID, normalizedInstrumentID)
}

func (s *Service) AddUser(ctx context.Context, tenantID string, instrumentID string, actorUserID string, input LinkUserInput) error {
	normalizedTenantID := strings.TrimSpace(tenantID)
	if normalizedTenantID == "" {
		return ErrInvalidTenantID
	}

	normalizedInstrumentID := strings.TrimSpace(instrumentID)
	if normalizedInstrumentID == "" {
		return ErrInvalidInstrumentID
	}

	normalizedActorID := strings.TrimSpace(actorUserID)
	if normalizedActorID == "" {
		return ErrInvalidActorID
	}

	normalizedUserID := strings.TrimSpace(input.UserID)
	if normalizedUserID == "" {
		return ErrInvalidUserID
	}

	existing, err := s.repository.GetByIDForMember(ctx, normalizedTenantID, normalizedInstrumentID, normalizedActorID)
	if err != nil {
		return err
	}

	if err := s.requireManager(ctx, normalizedTenantID, existing.ChoirID, normalizedActorID); err != nil {
		return err
	}

	return s.repository.AddUser(ctx, normalizedTenantID, normalizedInstrumentID, normalizedUserID)
}

func (s *Service) RemoveUser(ctx context.Context, tenantID string, instrumentID string, actorUserID string, userID string) error {
	normalizedTenantID := strings.TrimSpace(tenantID)
	if normalizedTenantID == "" {
		return ErrInvalidTenantID
	}

	normalizedInstrumentID := strings.TrimSpace(instrumentID)
	if normalizedInstrumentID == "" {
		return ErrInvalidInstrumentID
	}

	normalizedActorID := strings.TrimSpace(actorUserID)
	if normalizedActorID == "" {
		return ErrInvalidActorID
	}

	normalizedUserID := strings.TrimSpace(userID)
	if normalizedUserID == "" {
		return ErrInvalidUserID
	}

	existing, err := s.repository.GetByIDForMember(ctx, normalizedTenantID, normalizedInstrumentID, normalizedActorID)
	if err != nil {
		return err
	}

	if err := s.requireManager(ctx, normalizedTenantID, existing.ChoirID, normalizedActorID); err != nil {
		return err
	}

	return s.repository.RemoveUser(ctx, normalizedTenantID, normalizedInstrumentID, normalizedUserID)
}

func (s *Service) ListUsers(ctx context.Context, tenantID string, instrumentID string, actorUserID string) ([]InstrumentUser, error) {
	normalizedTenantID := strings.TrimSpace(tenantID)
	if normalizedTenantID == "" {
		return nil, ErrInvalidTenantID
	}

	normalizedInstrumentID := strings.TrimSpace(instrumentID)
	if normalizedInstrumentID == "" {
		return nil, ErrInvalidInstrumentID
	}

	normalizedActorID := strings.TrimSpace(actorUserID)
	if normalizedActorID == "" {
		return nil, ErrInvalidActorID
	}

	if _, err := s.repository.GetByIDForMember(ctx, normalizedTenantID, normalizedInstrumentID, normalizedActorID); err != nil {
		return nil, err
	}

	return s.repository.ListUsers(ctx, normalizedTenantID, normalizedInstrumentID)
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
