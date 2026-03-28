package users

import "context"

type Repository interface {
	Create(ctx context.Context, params CreateParams) (User, error)
	GetByID(ctx context.Context, tenantID string, userID string) (User, error)
	ListByTenantID(ctx context.Context, tenantID string) ([]User, error)
}

type CreateParams struct {
	TenantID string
	Email    string
	FullName string
}
