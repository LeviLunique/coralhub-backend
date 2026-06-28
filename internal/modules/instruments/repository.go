package instruments

import "context"

type Repository interface {
	Create(ctx context.Context, params CreateParams) (Instrument, error)
	Update(ctx context.Context, params UpdateParams) (Instrument, error)
	GetByIDForMember(ctx context.Context, tenantID string, instrumentID string, userID string) (Instrument, error)
	ListByChoirID(ctx context.Context, tenantID string, choirID string) ([]Instrument, error)
	Archive(ctx context.Context, tenantID string, instrumentID string) error
	AddUser(ctx context.Context, tenantID string, instrumentID string, userID string) error
	RemoveUser(ctx context.Context, tenantID string, instrumentID string, userID string) error
	ListUsers(ctx context.Context, tenantID string, instrumentID string) ([]InstrumentUser, error)
}

type CreateParams struct {
	TenantID    string
	ChoirID     string
	Name        string
	Description *string
	Icon        *string
}

type UpdateParams struct {
	TenantID     string
	InstrumentID string
	Name         string
	Description  *string
	Icon         *string
}
