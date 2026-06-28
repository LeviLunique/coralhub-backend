package materials

import (
	"context"

	"github.com/LeviLunique/coralhub-backend/internal/modules/instruments"
)

type Repository interface {
	Create(ctx context.Context, params CreateParams) (Material, error)
	Update(ctx context.Context, params UpdateParams) (Material, error)
	GetByIDForMember(ctx context.Context, tenantID string, materialID string, userID string) (Material, error)
	ListByChoirID(ctx context.Context, tenantID string, choirID string) ([]Material, error)
	ListBySongID(ctx context.Context, tenantID string, songID string) ([]Material, error)
	Archive(ctx context.Context, tenantID string, materialID string) error
	AddInstrument(ctx context.Context, tenantID string, materialID string, instrumentID string) error
	RemoveInstrument(ctx context.Context, tenantID string, materialID string, instrumentID string) error
	ListInstruments(ctx context.Context, tenantID string, materialID string) ([]instruments.Instrument, error)
}

type CreateParams struct {
	TenantID     string
	ChoirID      string
	SongID       string
	Name         string
	MaterialType string
	TargetType   string
	VoiceType    *string
}

type UpdateParams struct {
	TenantID     string
	MaterialID   string
	Name         string
	MaterialType string
	TargetType   string
	VoiceType    *string
}
