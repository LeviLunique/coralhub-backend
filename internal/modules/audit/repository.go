package audit

import "context"

type Repository interface {
	Create(ctx context.Context, params CreateParams) (Entry, error)
	ListByTenantID(ctx context.Context, tenantID string, limit int32) ([]Entry, error)
}
