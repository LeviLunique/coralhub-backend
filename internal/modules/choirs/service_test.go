package choirs

import (
	"context"
	"errors"
	"testing"
)

type stubRepository struct {
	choir  Choir
	choirs []Choir
	err    error
	params CreateParams
}

func (s *stubRepository) Create(_ context.Context, params CreateParams) (Choir, error) {
	s.params = params
	if s.err != nil {
		return Choir{}, s.err
	}

	return s.choir, nil
}

func (s *stubRepository) GetByIDForMember(_ context.Context, _, _, _ string) (Choir, error) {
	if s.err != nil {
		return Choir{}, s.err
	}

	return s.choir, nil
}

func (s *stubRepository) ListByMemberUserID(_ context.Context, _, _ string) ([]Choir, error) {
	if s.err != nil {
		return nil, s.err
	}

	return s.choirs, nil
}

func TestServiceCreateTrimsNameAndDescription(t *testing.T) {
	repository := &stubRepository{
		choir: Choir{ID: "choir-1", Name: "Sopranos"},
	}

	service := NewService(repository)
	description := "  Main choir  "

	_, err := service.Create(context.Background(), "tenant-1", "user-1", CreateInput{
		Name:        "  Sopranos  ",
		Description: &description,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if repository.params.Name != "Sopranos" {
		t.Fatalf("repository.params.Name = %q", repository.params.Name)
	}

	if repository.params.Description == nil || *repository.params.Description != "Main choir" {
		t.Fatalf("repository.params.Description = %#v", repository.params.Description)
	}
}

func TestServiceCreateRejectsBlankName(t *testing.T) {
	service := NewService(&stubRepository{})

	_, err := service.Create(context.Background(), "tenant-1", "user-1", CreateInput{Name: "   "})
	if !errors.Is(err, ErrInvalidChoirName) {
		t.Fatalf("Create() error = %v, want %v", err, ErrInvalidChoirName)
	}
}
