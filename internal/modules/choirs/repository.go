package choirs

import "context"

type Repository interface {
	Create(ctx context.Context, params CreateParams) (Choir, error)
	GetByID(ctx context.Context, tenantID string, choirID string) (Choir, error)
	ListByTenantID(ctx context.Context, tenantID string) ([]Choir, error)
}

type CreateParams struct {
	TenantID    string
	Name        string
	Description *string
}
