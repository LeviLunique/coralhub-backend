package repertoires

import "context"

type Repository interface {
	Create(ctx context.Context, params CreateParams) (Repertoire, error)
	Update(ctx context.Context, params UpdateParams) (Repertoire, error)
	GetByIDForMember(ctx context.Context, tenantID string, repertoireID string, userID string) (Repertoire, error)
	ListByChoirID(ctx context.Context, tenantID string, choirID string) ([]Repertoire, error)
	Archive(ctx context.Context, tenantID string, repertoireID string) error
	AddSong(ctx context.Context, params AddSongParams) error
	RemoveSong(ctx context.Context, tenantID string, repertoireID string, songID string) error
	ListSongs(ctx context.Context, tenantID string, repertoireID string) ([]RepertoireSong, error)
}

type CreateParams struct {
	TenantID    string
	ChoirID     string
	Name        string
	Description *string
}

type UpdateParams struct {
	TenantID     string
	RepertoireID string
	Name         string
	Description  *string
}

type AddSongParams struct {
	TenantID       string
	RepertoireID   string
	SongID         string
	ExecutionOrder int
	Notes          *string
}
