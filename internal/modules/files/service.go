package files

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"mime"
	"path/filepath"
	"strings"
	"time"

	"github.com/LeviLunique/coralhub-backend/internal/modules/memberships"
	"github.com/LeviLunique/coralhub-backend/internal/modules/voicekits"
)

var (
	ErrInvalidTenantID         = errors.New("invalid tenant id")
	ErrInvalidTenantSlug       = errors.New("invalid tenant slug")
	ErrInvalidActorID          = errors.New("invalid actor id")
	ErrInvalidVoiceKitID       = errors.New("invalid voice kit id")
	ErrInvalidFileID           = errors.New("invalid file id")
	ErrInvalidOriginalFilename = errors.New("invalid original filename")
	ErrInvalidContentType      = errors.New("invalid content type")
	ErrInvalidSizeBytes        = errors.New("invalid size bytes")
	ErrUnsupportedContentType  = errors.New("unsupported content type")
	ErrFileTooLarge            = errors.New("file too large")
	ErrVoiceKitNotFound        = errors.New("voice kit not found")
	ErrFileNotFound            = errors.New("file not found")
	ErrForbidden               = errors.New("forbidden")
	ErrStorageUnavailable      = errors.New("storage unavailable")
)

const (
	maxUploadSizeBytes = 50 << 20
	defaultPresignTTL  = 15 * time.Minute
)

type voiceKitReader interface {
	GetByIDForMember(ctx context.Context, tenantID string, voiceKitID string, actorUserID string) (voicekits.VoiceKit, error)
}

type membershipChecker interface {
	GetByChoirAndUser(ctx context.Context, tenantID string, choirID string, userID string) (memberships.Membership, error)
}

type Service struct {
	repository  Repository
	storage     Storage
	voiceKits   voiceKitReader
	memberships membershipChecker
	appEnv      string
}

func NewService(repository Repository, storage Storage, voiceKits voiceKitReader, memberships membershipChecker, appEnv string) *Service {
	normalizedEnv := strings.TrimSpace(appEnv)
	if normalizedEnv == "" {
		normalizedEnv = "development"
	}

	return &Service{
		repository:  repository,
		storage:     storage,
		voiceKits:   voiceKits,
		memberships: memberships,
		appEnv:      normalizedEnv,
	}
}

func (s *Service) Upload(ctx context.Context, tenantID string, tenantSlug string, voiceKitID string, actorUserID string, input UploadInput) (File, error) {
	normalizedTenantID := strings.TrimSpace(tenantID)
	if normalizedTenantID == "" {
		return File{}, ErrInvalidTenantID
	}

	normalizedTenantSlug := strings.TrimSpace(tenantSlug)
	if normalizedTenantSlug == "" {
		return File{}, ErrInvalidTenantSlug
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

	contentType, err := normalizeContentType(input.ContentType)
	if err != nil {
		return File{}, err
	}

	if input.SizeBytes <= 0 {
		return File{}, ErrInvalidSizeBytes
	}

	if input.SizeBytes > maxUploadSizeBytes {
		return File{}, ErrFileTooLarge
	}

	if input.Content == nil {
		return File{}, ErrInvalidContentType
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

	fileID, err := newObjectID()
	if err != nil {
		return File{}, ErrStorageUnavailable
	}

	storedFilename := buildStoredFilename(fileID, originalFilename)
	storageKey := buildStorageKey(s.appEnv, normalizedTenantSlug, voiceKit.ChoirID, normalizedVoiceKitID, fileID, storedFilename)

	if err := s.storage.PutObject(ctx, storageKey, input.Content, input.SizeBytes, contentType); err != nil {
		return File{}, ErrStorageUnavailable
	}

	file, err := s.repository.Create(ctx, CreateParams{
		ID:               fileID,
		TenantID:         normalizedTenantID,
		VoiceKitID:       normalizedVoiceKitID,
		OriginalFilename: originalFilename,
		StoredFilename:   storedFilename,
		ContentType:      contentType,
		SizeBytes:        input.SizeBytes,
		StorageKey:       storageKey,
	})
	if err != nil {
		_ = s.storage.DeleteObject(ctx, storageKey)
		return File{}, err
	}

	return file, nil
}

func (s *Service) GetDownloadURL(ctx context.Context, tenantID string, fileID string, actorUserID string) (DownloadURL, error) {
	normalizedTenantID := strings.TrimSpace(tenantID)
	if normalizedTenantID == "" {
		return DownloadURL{}, ErrInvalidTenantID
	}

	normalizedFileID := strings.TrimSpace(fileID)
	if normalizedFileID == "" {
		return DownloadURL{}, ErrInvalidFileID
	}

	normalizedActorID := strings.TrimSpace(actorUserID)
	if normalizedActorID == "" {
		return DownloadURL{}, ErrInvalidActorID
	}

	file, err := s.repository.GetByIDForMember(ctx, normalizedTenantID, normalizedFileID, normalizedActorID)
	if err != nil {
		return DownloadURL{}, err
	}

	url, err := s.storage.PresignGetObject(ctx, file.StorageKey, defaultPresignTTL)
	if err != nil {
		return DownloadURL{}, ErrStorageUnavailable
	}

	return DownloadURL{
		URL:       url,
		ExpiresAt: time.Now().UTC().Add(defaultPresignTTL),
	}, nil
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

	if err := s.storage.DeleteObject(ctx, file.StorageKey); err != nil {
		return ErrStorageUnavailable
	}

	return s.repository.Delete(ctx, normalizedTenantID, normalizedFileID)
}

func normalizeContentType(value string) (string, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return "", ErrInvalidContentType
	}

	mediaType, _, err := mime.ParseMediaType(normalized)
	if err != nil {
		return "", ErrInvalidContentType
	}

	if strings.HasPrefix(mediaType, "audio/") || mediaType == "application/pdf" {
		return mediaType, nil
	}

	return "", ErrUnsupportedContentType
}

func buildStoredFilename(fileID string, originalFilename string) string {
	extension := strings.ToLower(filepath.Ext(strings.TrimSpace(originalFilename)))
	if extension == "" {
		return fileID
	}

	return fileID + extension
}

func buildStorageKey(appEnv string, tenantSlug string, choirID string, voiceKitID string, fileID string, storedFilename string) string {
	return strings.Join([]string{
		appEnv,
		"tenants",
		tenantSlug,
		"choirs",
		choirID,
		"voice-kits",
		voiceKitID,
		"files",
		fileID,
		storedFilename,
	}, "/")
}

func newObjectID() (string, error) {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}

	encoded := hex.EncodeToString(bytes[:])
	return encoded[0:8] + "-" + encoded[8:12] + "-" + encoded[12:16] + "-" + encoded[16:20] + "-" + encoded[20:32], nil
}
