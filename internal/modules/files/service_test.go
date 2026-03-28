package files

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/LeviLunique/coralhub-backend/internal/modules/memberships"
	"github.com/LeviLunique/coralhub-backend/internal/modules/voicekits"
)

type stubRepository struct {
	file    File
	files   []File
	err     error
	create  CreateParams
	deleted string
}

func (s *stubRepository) Create(_ context.Context, params CreateParams) (File, error) {
	s.create = params
	if s.err != nil {
		return File{}, s.err
	}

	if s.file.ID == "" {
		s.file.ID = params.ID
	}

	return s.file, nil
}

func (s *stubRepository) GetByIDForMember(_ context.Context, _, _, _ string) (File, error) {
	if s.err != nil {
		return File{}, s.err
	}

	return s.file, nil
}

func (s *stubRepository) ListByVoiceKitID(_ context.Context, _, _ string) ([]File, error) {
	if s.err != nil {
		return nil, s.err
	}

	return s.files, nil
}

func (s *stubRepository) Delete(_ context.Context, _, fileID string) error {
	if s.err != nil {
		return s.err
	}

	s.deleted = fileID
	return nil
}

type stubStorage struct {
	putKey     string
	putBody    []byte
	putType    string
	deleteKey  string
	presignKey string
	presignTTL time.Duration
	presignURL string
	err        error
}

func (s *stubStorage) PutObject(_ context.Context, objectKey string, body io.Reader, _ int64, contentType string) error {
	s.putKey = objectKey
	s.putType = contentType
	if body != nil {
		payload, _ := io.ReadAll(body)
		s.putBody = payload
	}

	return s.err
}

func (s *stubStorage) DeleteObject(_ context.Context, objectKey string) error {
	s.deleteKey = objectKey
	return s.err
}

func (s *stubStorage) PresignGetObject(_ context.Context, objectKey string, expiresIn time.Duration) (string, error) {
	s.presignKey = objectKey
	s.presignTTL = expiresIn
	if s.err != nil {
		return "", s.err
	}

	return s.presignURL, nil
}

type stubVoiceKitReader struct {
	voiceKit voicekits.VoiceKit
	err      error
}

func (s *stubVoiceKitReader) GetByIDForMember(_ context.Context, _, _, _ string) (voicekits.VoiceKit, error) {
	if s.err != nil {
		return voicekits.VoiceKit{}, s.err
	}

	return s.voiceKit, nil
}

type stubMembershipChecker struct {
	membership memberships.Membership
	err        error
}

func (s *stubMembershipChecker) GetByChoirAndUser(_ context.Context, _, _, _ string) (memberships.Membership, error) {
	if s.err != nil {
		return memberships.Membership{}, s.err
	}

	return s.membership, nil
}

func TestServiceUploadRequiresManagerRole(t *testing.T) {
	repository := &stubRepository{}
	storage := &stubStorage{}
	service := NewService(repository, storage, &stubVoiceKitReader{
		voiceKit: voicekits.VoiceKit{ID: "kit-1", ChoirID: "choir-1"},
	}, &stubMembershipChecker{
		membership: memberships.Membership{Role: memberships.RoleMember},
	}, "development")

	_, err := service.Upload(context.Background(), "tenant-1", "tenant-slug", "kit-1", "actor-1", UploadInput{
		OriginalFilename: "score.pdf",
		ContentType:      "application/pdf",
		SizeBytes:        42,
		Content:          bytes.NewBufferString("pdf"),
	})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("Upload() error = %v, want %v", err, ErrForbidden)
	}
}

func TestServiceUploadBuildsStorageMetadata(t *testing.T) {
	repository := &stubRepository{
		file: File{ID: "ignored"},
	}
	storage := &stubStorage{}
	service := NewService(repository, storage, &stubVoiceKitReader{
		voiceKit: voicekits.VoiceKit{ID: "kit-1", ChoirID: "choir-1"},
	}, &stubMembershipChecker{
		membership: memberships.Membership{Role: memberships.RoleManager},
	}, "development")

	file, err := service.Upload(context.Background(), "tenant-1", "coral-jovem-asa-norte", "kit-1", "actor-1", UploadInput{
		OriginalFilename: "  score.pdf  ",
		ContentType:      " application/pdf ",
		SizeBytes:        42,
		Content:          bytes.NewBufferString("pdf"),
	})
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}

	if file.ID == "" {
		t.Fatal("file.ID is empty")
	}

	if repository.create.ID == "" {
		t.Fatal("repository.create.ID is empty")
	}

	if repository.create.StoredFilename == "" {
		t.Fatal("repository.create.StoredFilename is empty")
	}

	if storage.putKey == "" {
		t.Fatal("storage.putKey is empty")
	}

	if repository.create.StorageKey != storage.putKey {
		t.Fatalf("repository.create.StorageKey = %q, want %q", repository.create.StorageKey, storage.putKey)
	}

	if repository.create.ContentType != "application/pdf" {
		t.Fatalf("repository.create.ContentType = %q", repository.create.ContentType)
	}
}

func TestServiceUploadRejectsUnsupportedContentType(t *testing.T) {
	service := NewService(&stubRepository{}, &stubStorage{}, &stubVoiceKitReader{
		voiceKit: voicekits.VoiceKit{ID: "kit-1", ChoirID: "choir-1"},
	}, &stubMembershipChecker{
		membership: memberships.Membership{Role: memberships.RoleManager},
	}, "development")

	_, err := service.Upload(context.Background(), "tenant-1", "tenant-slug", "kit-1", "actor-1", UploadInput{
		OriginalFilename: "notes.txt",
		ContentType:      "text/plain",
		SizeBytes:        5,
		Content:          bytes.NewBufferString("notes"),
	})
	if !errors.Is(err, ErrUnsupportedContentType) {
		t.Fatalf("Upload() error = %v, want %v", err, ErrUnsupportedContentType)
	}
}

func TestServiceGetDownloadURLUsesStorage(t *testing.T) {
	storage := &stubStorage{presignURL: "http://example.com/file"}
	service := NewService(&stubRepository{
		file: File{ID: "file-1", StorageKey: "development/tenants/coral/files/file-1/score.pdf"},
	}, storage, &stubVoiceKitReader{}, &stubMembershipChecker{}, "development")

	result, err := service.GetDownloadURL(context.Background(), "tenant-1", "file-1", "actor-1")
	if err != nil {
		t.Fatalf("GetDownloadURL() error = %v", err)
	}

	if result.URL != "http://example.com/file" {
		t.Fatalf("result.URL = %q", result.URL)
	}

	if storage.presignKey == "" {
		t.Fatal("storage.presignKey is empty")
	}
}

func TestServiceDeleteRequiresManagerRole(t *testing.T) {
	repository := &stubRepository{
		file: File{ID: "file-1", ChoirID: "choir-1"},
	}
	service := NewService(repository, &stubStorage{}, &stubVoiceKitReader{}, &stubMembershipChecker{
		membership: memberships.Membership{Role: memberships.RoleMember},
	}, "development")

	err := service.Delete(context.Background(), "tenant-1", "file-1", "actor-1")
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("Delete() error = %v, want %v", err, ErrForbidden)
	}
}
