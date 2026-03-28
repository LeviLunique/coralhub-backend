package users

import (
	"context"
	"errors"
	"testing"
)

type stubRepository struct {
	user   User
	users  []User
	err    error
	params CreateParams
}

func (s *stubRepository) Create(_ context.Context, params CreateParams) (User, error) {
	s.params = params
	if s.err != nil {
		return User{}, s.err
	}

	return s.user, nil
}

func (s *stubRepository) GetByID(_ context.Context, _, _ string) (User, error) {
	if s.err != nil {
		return User{}, s.err
	}

	return s.user, nil
}

func (s *stubRepository) GetByEmail(_ context.Context, _, _ string) (User, error) {
	if s.err != nil {
		return User{}, s.err
	}

	return s.user, nil
}

func (s *stubRepository) ListByTenantID(_ context.Context, _ string) ([]User, error) {
	if s.err != nil {
		return nil, s.err
	}

	return s.users, nil
}

func TestServiceCreateNormalizesEmailAndName(t *testing.T) {
	repository := &stubRepository{
		user: User{ID: "user-1", Email: "ana@example.com"},
	}

	service := NewService(repository)
	_, err := service.Create(context.Background(), "tenant-1", CreateInput{
		Email:    "  ANA@EXAMPLE.COM  ",
		FullName: "  Ana Clara  ",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if repository.params.Email != "ana@example.com" {
		t.Fatalf("repository.params.Email = %q", repository.params.Email)
	}

	if repository.params.FullName != "Ana Clara" {
		t.Fatalf("repository.params.FullName = %q", repository.params.FullName)
	}
}

func TestServiceCreateRejectsInvalidEmail(t *testing.T) {
	service := NewService(&stubRepository{})

	_, err := service.Create(context.Background(), "tenant-1", CreateInput{
		Email:    "invalid",
		FullName: "Ana Clara",
	})
	if !errors.Is(err, ErrInvalidEmail) {
		t.Fatalf("Create() error = %v, want %v", err, ErrInvalidEmail)
	}
}

func TestServiceResolveActorByEmailNormalizesEmail(t *testing.T) {
	repository := &stubRepository{
		user: User{ID: "user-1", Email: "ana@example.com"},
	}

	service := NewService(repository)
	user, err := service.ResolveActorByEmail(context.Background(), "tenant-1", "  ANA@EXAMPLE.COM  ")
	if err != nil {
		t.Fatalf("ResolveActorByEmail() error = %v", err)
	}

	if user.Email != "ana@example.com" {
		t.Fatalf("user.Email = %q", user.Email)
	}
}
