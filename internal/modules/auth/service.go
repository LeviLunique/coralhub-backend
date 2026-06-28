package auth

import (
	"context"
	"errors"
	"net/mail"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidEmail        = errors.New("invalid email")
	ErrInvalidPassword     = errors.New("invalid password")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrAmbiguousUserTenant = errors.New("ambiguous user tenant")
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Login(ctx context.Context, input LoginInput) (Session, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))
	if _, err := mail.ParseAddress(email); err != nil {
		return Session{}, ErrInvalidEmail
	}

	password := strings.TrimSpace(input.Password)
	if password == "" {
		return Session{}, ErrInvalidPassword
	}

	identities, err := s.repository.ListLoginIdentitiesByEmail(ctx, email)
	if err != nil {
		return Session{}, err
	}
	if len(identities) == 0 {
		return Session{}, ErrInvalidCredentials
	}
	if len(identities) > 1 {
		return Session{}, ErrAmbiguousUserTenant
	}

	identity := identities[0]
	if identity.PasswordHash == nil || strings.TrimSpace(*identity.PasswordHash) == "" {
		return Session{}, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*identity.PasswordHash), []byte(password)); err != nil {
		return Session{}, ErrInvalidCredentials
	}

	return Session{
		Tenant: Tenant{
			ID:          identity.TenantID,
			Slug:        identity.TenantSlug,
			DisplayName: identity.TenantDisplayName,
		},
		User: User{
			ID:       identity.UserID,
			TenantID: identity.TenantID,
			Email:    identity.Email,
			FullName: identity.FullName,
			Manager:  identity.Manager,
		},
	}, nil
}
