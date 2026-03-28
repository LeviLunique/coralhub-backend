package files

import (
	"context"
	"errors"
	"strings"

	"github.com/LeviLunique/coralhub-backend/internal/modules/memberships"
	"github.com/LeviLunique/coralhub-backend/internal/modules/voicekits"
)

var (
	ErrInvalidTenantID         = errors.New("invalid tenant id")
	ErrInvalidActorID          = errors.New("invalid actor id")
	ErrInvalidVoiceKitID       = errors.New("invalid voice kit id")
	ErrInvalidFileID           = errors.New("invalid file id")
	ErrInvalidOriginalFilename = errors.New("invalid original filename")
	ErrInvalidStoredFilename   = errors.New("invalid stored filename")
	ErrInvalidContentType      = errors.New("invalid content type")
	ErrInvalidStorageKey       = errors.New("invalid storage key")
	ErrInvalidSizeBytes        = errors.New("invalid size bytes")
	ErrVoiceKitNotFound        = errors.New("voice kit not found")
	ErrFileNotFound            = errors.New("file not found")
	ErrForbidden               = errors.New("forbidden")
)

type voiceKitReader interface {
	GetByIDForMember(ctx context.Context, tenantID string, voiceKitID string, actorUserID string) (voicekits.VoiceKit, error)
}

type membershipChecker interface {
	GetByChoirAndUser(ctx context.Context, tenantID string, choirID string, userID string) (memberships.Membership, error)
}

type Service struct {
	repository  Repository
	voiceKits   voiceKitReader
	memberships membershipChecker
}

func NewService(repository Repository, voiceKits voiceKitReader, memberships membershipChecker) *Service {
	return &Service{repository: repository, voiceKits: voiceKits, memberships: memberships}
}

func (s *Service) Create(ctx context.Context, tenantID string, voiceKitID string, actorUserID string, input CreateInput) (File, error) {
	normalizedTenantID := strings.TrimSpace(tenantID)
	if normalizedTenantID == "" {
		return File{}, ErrInvalidTenantID
	}

	normalizedVoiceKitID := strings.TrimSpace(voiceKitID)
	if normalizedVoiceKitID == "" {
		return File{}, ErrInvalidVoiceKitID
	}

	normalizedActorID := strings.TrimSpace(actorUserID)
	if normalizedActorID == "" {
		return File{}, ErrInvalidActorID
	}

	originalFilename := strings.TrimSpace(input.OriginalFilename)
	if originalFilename == "" {
		return File{}, ErrInvalidOriginalFilename
	}

	storedFilename := strings.TrimSpace(input.StoredFilename)
	if storedFilename == "" {
		return File{}, ErrInvalidStoredFilename
	}

	contentType := strings.TrimSpace(input.ContentType)
	if contentType == "" {
		return File{}, ErrInvalidContentType
	}

	storageKey := strings.TrimSpace(input.StorageKey)
	if storageKey == "" {
		return File{}, ErrInvalidStorageKey
	}

	if input.SizeBytes <= 0 {
		return File{}, ErrInvalidSizeBytes
	}

	voiceKit, err := s.voiceKits.GetByIDForMember(ctx, normalizedTenantID, normalizedVoiceKitID, normalizedActorID)
	if err != nil {
		if errors.Is(err, voicekits.ErrVoiceKitNotFound) {
			return File{}, ErrVoiceKitNotFound
		}

		return File{}, err
	}

	member, err := s.memberships.GetByChoirAndUser(ctx, normalizedTenantID, voiceKit.ChoirID, normalizedActorID)
	if err != nil {
		if errors.Is(err, memberships.ErrMembershipNotFound) {
			return File{}, ErrForbidden
		}

		return File{}, err
	}

	if member.Role != memberships.RoleManager {
		return File{}, ErrForbidden
	}

	return s.repository.Create(ctx, CreateParams{
		TenantID:         normalizedTenantID,
		VoiceKitID:       normalizedVoiceKitID,
		OriginalFilename: originalFilename,
		StoredFilename:   storedFilename,
		ContentType:      contentType,
		SizeBytes:        input.SizeBytes,
		StorageKey:       storageKey,
	})
}

func (s *Service) ListByVoiceKit(ctx context.Context, tenantID string, voiceKitID string, actorUserID string) ([]File, error) {
	normalizedTenantID := strings.TrimSpace(tenantID)
	if normalizedTenantID == "" {
		return nil, ErrInvalidTenantID
	}

	normalizedVoiceKitID := strings.TrimSpace(voiceKitID)
	if normalizedVoiceKitID == "" {
		return nil, ErrInvalidVoiceKitID
	}

	normalizedActorID := strings.TrimSpace(actorUserID)
	if normalizedActorID == "" {
		return nil, ErrInvalidActorID
	}

	if _, err := s.voiceKits.GetByIDForMember(ctx, normalizedTenantID, normalizedVoiceKitID, normalizedActorID); err != nil {
		if errors.Is(err, voicekits.ErrVoiceKitNotFound) {
			return nil, ErrVoiceKitNotFound
		}

		return nil, err
	}

	return s.repository.ListByVoiceKitID(ctx, normalizedTenantID, normalizedVoiceKitID)
}

func (s *Service) Delete(ctx context.Context, tenantID string, fileID string, actorUserID string) error {
	normalizedTenantID := strings.TrimSpace(tenantID)
	if normalizedTenantID == "" {
		return ErrInvalidTenantID
	}

	normalizedFileID := strings.TrimSpace(fileID)
	if normalizedFileID == "" {
		return ErrInvalidFileID
	}

	normalizedActorID := strings.TrimSpace(actorUserID)
	if normalizedActorID == "" {
		return ErrInvalidActorID
	}

	file, err := s.repository.GetByIDForMember(ctx, normalizedTenantID, normalizedFileID, normalizedActorID)
	if err != nil {
		return err
	}

	member, err := s.memberships.GetByChoirAndUser(ctx, normalizedTenantID, file.ChoirID, normalizedActorID)
	if err != nil {
		if errors.Is(err, memberships.ErrMembershipNotFound) {
			return ErrForbidden
		}

		return err
	}

	if member.Role != memberships.RoleManager {
		return ErrForbidden
	}

	return s.repository.Delete(ctx, normalizedTenantID, normalizedFileID)
}
