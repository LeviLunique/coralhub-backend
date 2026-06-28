package songs

import "context"

type Repository interface {
	Create(ctx context.Context, params CreateParams) (Song, error)
	Update(ctx context.Context, params UpdateParams) (Song, error)
	GetByIDForMember(ctx context.Context, tenantID string, songID string, userID string) (Song, error)
	ListByChoirID(ctx context.Context, tenantID string, choirID string) ([]Song, error)
	Archive(ctx context.Context, tenantID string, songID string) error
}

type CreateParams struct {
	TenantID string
	ChoirID  string
	Title    string
	Composer *string
	Arranger *string
	SongKey  *string
	Duration *string
	Notes    *string
}

type UpdateParams struct {
	TenantID string
	SongID   string
	Title    string
	Composer *string
	Arranger *string
	SongKey  *string
	Duration *string
	Notes    *string
}
