package files

import (
	"context"
	"errors"
	"testing"

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

func TestServiceCreateRequiresManagerRole(t *testing.T) {
	repository := &stubRepository{}
	service := NewService(repository, &stubVoiceKitReader{
		voiceKit: voicekits.VoiceKit{ID: "kit-1", ChoirID: "choir-1"},
	}, &stubMembershipChecker{
		membership: memberships.Membership{Role: memberships.RoleMember},
	})

	_, err := service.Create(context.Background(), "tenant-1", "kit-1", "actor-1", CreateInput{
		OriginalFilename: "score.pdf",
		StoredFilename:   "stored-score.pdf",
		ContentType:      "application/pdf",
		SizeBytes:        42,
		StorageKey:       "dev/tenants/tenant/files/file-1/stored-score.pdf",
	})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("Create() error = %v, want %v", err, ErrForbidden)
	}
}

func TestServiceCreateValidatesAndTrimsFields(t *testing.T) {
	repository := &stubRepository{
		file: File{ID: "file-1"},
	}
	service := NewService(repository, &stubVoiceKitReader{
		voiceKit: voicekits.VoiceKit{ID: "kit-1", ChoirID: "choir-1"},
	}, &stubMembershipChecker{
		membership: memberships.Membership{Role: memberships.RoleManager},
	})

	_, err := service.Create(context.Background(), "tenant-1", "kit-1", "actor-1", CreateInput{
		OriginalFilename: "  score.pdf  ",
		StoredFilename:   "  01-score.pdf  ",
		ContentType:      "  application/pdf  ",
		SizeBytes:        42,
		StorageKey:       "  dev/tenants/tenant/files/file-1/01-score.pdf  ",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if repository.create.OriginalFilename != "score.pdf" {
		t.Fatalf("repository.create.OriginalFilename = %q", repository.create.OriginalFilename)
	}

	if repository.create.StoredFilename != "01-score.pdf" {
		t.Fatalf("repository.create.StoredFilename = %q", repository.create.StoredFilename)
	}

	if repository.create.ContentType != "application/pdf" {
		t.Fatalf("repository.create.ContentType = %q", repository.create.ContentType)
	}

	if repository.create.StorageKey != "dev/tenants/tenant/files/file-1/01-score.pdf" {
		t.Fatalf("repository.create.StorageKey = %q", repository.create.StorageKey)
	}
}

func TestServiceDeleteRequiresManagerRole(t *testing.T) {
	repository := &stubRepository{
		file: File{ID: "file-1", ChoirID: "choir-1"},
	}
	service := NewService(repository, &stubVoiceKitReader{}, &stubMembershipChecker{
		membership: memberships.Membership{Role: memberships.RoleMember},
	})

	err := service.Delete(context.Background(), "tenant-1", "file-1", "actor-1")
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("Delete() error = %v, want %v", err, ErrForbidden)
	}
}
