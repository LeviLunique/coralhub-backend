package memberships

import (
	"context"
	"errors"
	"testing"
)

type stubRepository struct {
	membership  Membership
	memberships []Membership
	err         error
	create      CreateParams
}

func (s *stubRepository) Create(_ context.Context, params CreateParams) (Membership, error) {
	s.create = params
	if s.err != nil {
		return Membership{}, s.err
	}

	return s.membership, nil
}

func (s *stubRepository) GetByChoirAndUser(_ context.Context, _, _, _ string) (Membership, error) {
	if s.err != nil {
		return Membership{}, s.err
	}

	return s.membership, nil
}

func (s *stubRepository) ListByChoirID(_ context.Context, _, _ string) ([]Membership, error) {
	if s.err != nil {
		return nil, s.err
	}

	return s.memberships, nil
}

func TestServiceAddMemberRequiresManagerRole(t *testing.T) {
	repository := &stubRepository{
		membership: Membership{Role: RoleMember},
	}

	service := NewService(repository)
	_, err := service.AddMember(context.Background(), "tenant-1", "choir-1", "actor-1", CreateInput{
		UserID: "user-2",
		Role:   RoleMember,
	})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("AddMember() error = %v, want %v", err, ErrForbidden)
	}
}

func TestServiceAddMemberNormalizesRole(t *testing.T) {
	repository := &stubRepository{
		membership: Membership{Role: RoleManager},
	}

	service := NewService(repository)
	_, err := service.AddMember(context.Background(), "tenant-1", "choir-1", "actor-1", CreateInput{
		UserID: "user-2",
		Role:   " MEMBER ",
	})
	if err != nil {
		t.Fatalf("AddMember() error = %v", err)
	}

	if repository.create.Role != RoleMember {
		t.Fatalf("repository.create.Role = %q", repository.create.Role)
	}
}
