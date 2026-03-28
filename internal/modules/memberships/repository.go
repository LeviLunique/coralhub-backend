package memberships

import "context"

type Repository interface {
	Create(ctx context.Context, params CreateParams) (Membership, error)
	GetByChoirAndUser(ctx context.Context, tenantID string, choirID string, userID string) (Membership, error)
	ListByChoirID(ctx context.Context, tenantID string, choirID string) ([]Membership, error)
}

type CreateParams struct {
	TenantID    string
	ChoirID     string
	UserID      string
	Role        string
	ActorUserID string
}
