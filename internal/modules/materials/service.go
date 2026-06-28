package materials

import (
	"context"
	"errors"
	"strings"

	"github.com/LeviLunique/coralhub-backend/internal/modules/instruments"
	"github.com/LeviLunique/coralhub-backend/internal/modules/memberships"
)

var (
	ErrInvalidTenantID        = errors.New("invalid tenant id")
	ErrInvalidChoirID         = errors.New("invalid choir id")
	ErrInvalidActorID         = errors.New("invalid actor id")
	ErrInvalidMaterialID      = errors.New("invalid material id")
	ErrInvalidSongID          = errors.New("invalid song id")
	ErrInvalidName            = errors.New("invalid material name")
	ErrInvalidMaterialType    = errors.New("invalid material type")
	ErrInvalidTargetType      = errors.New("invalid target type")
	ErrInvalidInstrumentID    = errors.New("invalid instrument id")
	ErrMaterialNotFound       = errors.New("material not found")
	ErrInstrumentLinkExists   = errors.New("instrument already linked to material")
	ErrInstrumentLinkNotFound = errors.New("instrument is not linked to material")
	ErrForbidden              = errors.New("forbidden")
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

func (s *Service) Create(ctx context.Context, tenantID string, choirID string, actorUserID string, input CreateInput) (Material, error) {
	normalizedTenantID := strings.TrimSpace(tenantID)
	if normalizedTenantID == "" {
		return Material{}, ErrInvalidTenantID
	}

	normalizedChoirID := strings.TrimSpace(choirID)
	if normalizedChoirID == "" {
		return Material{}, ErrInvalidChoirID
	}

	normalizedActorID := strings.TrimSpace(actorUserID)
	if normalizedActorID == "" {
		return Material{}, ErrInvalidActorID
	}

	normalizedSongID := strings.TrimSpace(input.SongID)
	if normalizedSongID == "" {
		return Material{}, ErrInvalidSongID
	}

	normalizedName := strings.TrimSpace(input.Name)
	if normalizedName == "" {
		return Material{}, ErrInvalidName
	}

	materialType, err := normalizeMaterialType(input.MaterialType)
	if err != nil {
		return Material{}, err
	}

	targetType, err := normalizeTargetType(input.TargetType)
	if err != nil {
		return Material{}, err
	}

	if err := s.requireManager(ctx, normalizedTenantID, normalizedChoirID, normalizedActorID); err != nil {
		return Material{}, err
	}

	return s.repository.Create(ctx, CreateParams{
		TenantID:     normalizedTenantID,
		ChoirID:      normalizedChoirID,
		SongID:       normalizedSongID,
		Name:         normalizedName,
		MaterialType: materialType,
		TargetType:   targetType,
		VoiceType:    normalizeOptionalText(input.VoiceType),
	})
}

func (s *Service) Update(ctx context.Context, tenantID string, materialID string, actorUserID string, input UpdateInput) (Material, error) {
	normalizedTenantID := strings.TrimSpace(tenantID)
	if normalizedTenantID == "" {
		return Material{}, ErrInvalidTenantID
	}

	normalizedMaterialID := strings.TrimSpace(materialID)
	if normalizedMaterialID == "" {
		return Material{}, ErrInvalidMaterialID
	}

	normalizedActorID := strings.TrimSpace(actorUserID)
	if normalizedActorID == "" {
		return Material{}, ErrInvalidActorID
	}

	normalizedName := strings.TrimSpace(input.Name)
	if normalizedName == "" {
		return Material{}, ErrInvalidName
	}

	materialType, err := normalizeMaterialType(input.MaterialType)
	if err != nil {
		return Material{}, err
	}

	targetType, err := normalizeTargetType(input.TargetType)
	if err != nil {
		return Material{}, err
	}

	existing, err := s.repository.GetByIDForMember(ctx, normalizedTenantID, normalizedMaterialID, normalizedActorID)
	if err != nil {
		return Material{}, err
	}

	if err := s.requireManager(ctx, normalizedTenantID, existing.ChoirID, normalizedActorID); err != nil {
		return Material{}, err
	}

	return s.repository.Update(ctx, UpdateParams{
		TenantID:     normalizedTenantID,
		MaterialID:   normalizedMaterialID,
		Name:         normalizedName,
		MaterialType: materialType,
		TargetType:   targetType,
		VoiceType:    normalizeOptionalText(input.VoiceType),
	})
}

func (s *Service) Get(ctx context.Context, tenantID string, actorUserID string, materialID string) (Material, error) {
	normalizedTenantID := strings.TrimSpace(tenantID)
	if normalizedTenantID == "" {
		return Material{}, ErrInvalidTenantID
	}

	normalizedActorID := strings.TrimSpace(actorUserID)
	if normalizedActorID == "" {
		return Material{}, ErrInvalidActorID
	}

	normalizedMaterialID := strings.TrimSpace(materialID)
	if normalizedMaterialID == "" {
		return Material{}, ErrInvalidMaterialID
	}

	return s.repository.GetByIDForMember(ctx, normalizedTenantID, normalizedMaterialID, normalizedActorID)
}

func (s *Service) ListByChoir(ctx context.Context, tenantID string, choirID string, actorUserID string) ([]Material, error) {
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

func (s *Service) ListBySong(ctx context.Context, tenantID string, songID string, actorUserID string) ([]Material, error) {
	normalizedTenantID := strings.TrimSpace(tenantID)
	if normalizedTenantID == "" {
		return nil, ErrInvalidTenantID
	}

	normalizedSongID := strings.TrimSpace(songID)
	if normalizedSongID == "" {
		return nil, ErrInvalidSongID
	}

	normalizedActorID := strings.TrimSpace(actorUserID)
	if normalizedActorID == "" {
		return nil, ErrInvalidActorID
	}

	materials, err := s.repository.ListBySongID(ctx, normalizedTenantID, normalizedSongID)
	if err != nil {
		return nil, err
	}

	if len(materials) == 0 {
		return materials, nil
	}

	if _, err := s.memberships.GetByChoirAndUser(ctx, normalizedTenantID, materials[0].ChoirID, normalizedActorID); err != nil {
		if errors.Is(err, memberships.ErrMembershipNotFound) {
			return nil, ErrForbidden
		}

		return nil, err
	}

	return materials, nil
}

func (s *Service) Delete(ctx context.Context, tenantID string, materialID string, actorUserID string) error {
	normalizedTenantID := strings.TrimSpace(tenantID)
	if normalizedTenantID == "" {
		return ErrInvalidTenantID
	}

	normalizedMaterialID := strings.TrimSpace(materialID)
	if normalizedMaterialID == "" {
		return ErrInvalidMaterialID
	}

	normalizedActorID := strings.TrimSpace(actorUserID)
	if normalizedActorID == "" {
		return ErrInvalidActorID
	}

	existing, err := s.repository.GetByIDForMember(ctx, normalizedTenantID, normalizedMaterialID, normalizedActorID)
	if err != nil {
		return err
	}

	if err := s.requireManager(ctx, normalizedTenantID, existing.ChoirID, normalizedActorID); err != nil {
		return err
	}

	return s.repository.Archive(ctx, normalizedTenantID, normalizedMaterialID)
}

func (s *Service) ListInstruments(ctx context.Context, tenantID string, materialID string, actorUserID string) ([]instruments.Instrument, error) {
	normalizedTenantID := strings.TrimSpace(tenantID)
	if normalizedTenantID == "" {
		return nil, ErrInvalidTenantID
	}

	normalizedMaterialID := strings.TrimSpace(materialID)
	if normalizedMaterialID == "" {
		return nil, ErrInvalidMaterialID
	}

	normalizedActorID := strings.TrimSpace(actorUserID)
	if normalizedActorID == "" {
		return nil, ErrInvalidActorID
	}

	if _, err := s.repository.GetByIDForMember(ctx, normalizedTenantID, normalizedMaterialID, normalizedActorID); err != nil {
		return nil, err
	}

	return s.repository.ListInstruments(ctx, normalizedTenantID, normalizedMaterialID)
}

func (s *Service) AddInstrument(ctx context.Context, tenantID string, materialID string, actorUserID string, input LinkInstrumentInput) error {
	normalizedTenantID := strings.TrimSpace(tenantID)
	if normalizedTenantID == "" {
		return ErrInvalidTenantID
	}

	normalizedMaterialID := strings.TrimSpace(materialID)
	if normalizedMaterialID == "" {
		return ErrInvalidMaterialID
	}

	normalizedActorID := strings.TrimSpace(actorUserID)
	if normalizedActorID == "" {
		return ErrInvalidActorID
	}

	normalizedInstrumentID := strings.TrimSpace(input.InstrumentID)
	if normalizedInstrumentID == "" {
		return ErrInvalidInstrumentID
	}

	existing, err := s.repository.GetByIDForMember(ctx, normalizedTenantID, normalizedMaterialID, normalizedActorID)
	if err != nil {
		return err
	}

	if err := s.requireManager(ctx, normalizedTenantID, existing.ChoirID, normalizedActorID); err != nil {
		return err
	}

	return s.repository.AddInstrument(ctx, normalizedTenantID, normalizedMaterialID, normalizedInstrumentID)
}

func (s *Service) RemoveInstrument(ctx context.Context, tenantID string, materialID string, actorUserID string, instrumentID string) error {
	normalizedTenantID := strings.TrimSpace(tenantID)
	if normalizedTenantID == "" {
		return ErrInvalidTenantID
	}

	normalizedMaterialID := strings.TrimSpace(materialID)
	if normalizedMaterialID == "" {
		return ErrInvalidMaterialID
	}

	normalizedActorID := strings.TrimSpace(actorUserID)
	if normalizedActorID == "" {
		return ErrInvalidActorID
	}

	normalizedInstrumentID := strings.TrimSpace(instrumentID)
	if normalizedInstrumentID == "" {
		return ErrInvalidInstrumentID
	}

	existing, err := s.repository.GetByIDForMember(ctx, normalizedTenantID, normalizedMaterialID, normalizedActorID)
	if err != nil {
		return err
	}

	if err := s.requireManager(ctx, normalizedTenantID, existing.ChoirID, normalizedActorID); err != nil {
		return err
	}

	return s.repository.RemoveInstrument(ctx, normalizedTenantID, normalizedMaterialID, normalizedInstrumentID)
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

func normalizeMaterialType(value *string) (string, error) {
	if value == nil {
		return MaterialTypeOther, nil
	}

	normalized := strings.ToLower(strings.TrimSpace(*value))
	if normalized == "" {
		return MaterialTypeOther, nil
	}

	switch normalized {
	case MaterialTypeAudioGuide, MaterialTypeSheetMusic, MaterialTypePlayback, MaterialTypeLyrics, MaterialTypeOther:
		return normalized, nil
	default:
		return "", ErrInvalidMaterialType
	}
}

func normalizeTargetType(value *string) (string, error) {
	if value == nil {
		return TargetTypeVoice, nil
	}

	normalized := strings.ToLower(strings.TrimSpace(*value))
	if normalized == "" {
		return TargetTypeVoice, nil
	}

	switch normalized {
	case TargetTypeVoice, TargetTypeInstrument:
		return normalized, nil
	default:
		return "", ErrInvalidTargetType
	}
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
