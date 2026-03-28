package voicekits

import (
	"context"
	"errors"
	"testing"

	"github.com/LeviLunique/coralhub-backend/internal/modules/memberships"
)

type stubRepository struct {
	voiceKit  VoiceKit
	voiceKits []VoiceKit
	err       error
	create    CreateParams
	deletedID string
}

func (s *stubRepository) Create(_ context.Context, params CreateParams) (VoiceKit, error) {
	s.create = params
	if s.err != nil {
		return VoiceKit{}, s.err
	}

	return s.voiceKit, nil
}

func (s *stubRepository) GetByIDForMember(_ context.Context, _, voiceKitID, _ string) (VoiceKit, error) {
	if s.err != nil {
		return VoiceKit{}, s.err
	}

	if s.voiceKit.ID == "" {
		s.voiceKit.ID = voiceKitID
	}

	return s.voiceKit, nil
}

func (s *stubRepository) ListByChoirID(_ context.Context, _, _ string) ([]VoiceKit, error) {
	if s.err != nil {
		return nil, s.err
	}

	return s.voiceKits, nil
}

func (s *stubRepository) Delete(_ context.Context, _, voiceKitID string) error {
	if s.err != nil {
		return s.err
	}

	s.deletedID = voiceKitID
	return nil
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
	service := NewService(repository, &stubMembershipChecker{
		membership: memberships.Membership{Role: memberships.RoleMember},
	})

	_, err := service.Create(context.Background(), "tenant-1", "choir-1", "actor-1", CreateInput{Name: "Warmups"})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("Create() error = %v, want %v", err, ErrForbidden)
	}
}

func TestServiceCreateTrimsNameAndDescription(t *testing.T) {
	repository := &stubRepository{
		voiceKit: VoiceKit{ID: "kit-1", Name: "Warmups"},
	}
	service := NewService(repository, &stubMembershipChecker{
		membership: memberships.Membership{Role: memberships.RoleManager},
	})
	description := "  Rehearsal exercises  "

	_, err := service.Create(context.Background(), "tenant-1", "choir-1", "actor-1", CreateInput{
		Name:        "  Warmups  ",
		Description: &description,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if repository.create.Name != "Warmups" {
		t.Fatalf("repository.create.Name = %q", repository.create.Name)
	}

	if repository.create.Description == nil || *repository.create.Description != "Rehearsal exercises" {
		t.Fatalf("repository.create.Description = %#v", repository.create.Description)
	}
}

func TestServiceDeleteRequiresManagerRole(t *testing.T) {
	repository := &stubRepository{
		voiceKit: VoiceKit{ID: "kit-1", ChoirID: "choir-1"},
	}
	service := NewService(repository, &stubMembershipChecker{
		membership: memberships.Membership{Role: memberships.RoleMember},
	})

	err := service.Delete(context.Background(), "tenant-1", "kit-1", "actor-1")
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("Delete() error = %v, want %v", err, ErrForbidden)
	}
}
