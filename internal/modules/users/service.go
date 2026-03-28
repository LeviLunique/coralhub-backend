package users

import (
	"context"
	"errors"
	"net/mail"
	"strings"
)

var (
	ErrInvalidUserID   = errors.New("invalid user id")
	ErrInvalidTenantID = errors.New("invalid tenant id")
	ErrInvalidEmail    = errors.New("invalid email")
	ErrInvalidFullName = errors.New("invalid full name")
	ErrUserNotFound    = errors.New("user not found")
	ErrUserEmailTaken  = errors.New("user email already exists")
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Create(ctx context.Context, tenantID string, input CreateInput) (User, error) {
	normalizedTenantID := strings.TrimSpace(tenantID)
	if normalizedTenantID == "" {
		return User{}, ErrInvalidTenantID
	}

	email := strings.ToLower(strings.TrimSpace(input.Email))
	if _, err := mail.ParseAddress(email); err != nil {
		return User{}, ErrInvalidEmail
	}

	fullName := strings.TrimSpace(input.FullName)
	if fullName == "" {
		return User{}, ErrInvalidFullName
	}

	return s.repository.Create(ctx, CreateParams{
		TenantID: normalizedTenantID,
		Email:    email,
		FullName: fullName,
	})
}

func (s *Service) Get(ctx context.Context, tenantID string, userID string) (User, error) {
	normalizedTenantID := strings.TrimSpace(tenantID)
	if normalizedTenantID == "" {
		return User{}, ErrInvalidTenantID
	}

	normalizedUserID := strings.TrimSpace(userID)
	if normalizedUserID == "" {
		return User{}, ErrInvalidUserID
	}

	return s.repository.GetByID(ctx, normalizedTenantID, normalizedUserID)
}

func (s *Service) List(ctx context.Context, tenantID string) ([]User, error) {
	normalizedTenantID := strings.TrimSpace(tenantID)
	if normalizedTenantID == "" {
		return nil, ErrInvalidTenantID
	}

	return s.repository.ListByTenantID(ctx, normalizedTenantID)
}
