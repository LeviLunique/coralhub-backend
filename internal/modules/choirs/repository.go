package choirs

import "context"

type Repository interface {
	Create(ctx context.Context, params CreateParams) (Choir, error)
	GetByIDForMember(ctx context.Context, tenantID string, choirID string, userID string) (Choir, error)
	ListByMemberUserID(ctx context.Context, tenantID string, userID string) ([]Choir, error)
}

type CreateParams struct {
	ActorUserID string
	TenantID    string
	Name        string
	Description *string
}
