package voicekits

import "context"

type Repository interface {
	Create(ctx context.Context, params CreateParams) (VoiceKit, error)
	GetByIDForMember(ctx context.Context, tenantID string, voiceKitID string, userID string) (VoiceKit, error)
	ListByChoirID(ctx context.Context, tenantID string, choirID string) ([]VoiceKit, error)
	Delete(ctx context.Context, tenantID string, voiceKitID string) error
}

type CreateParams struct {
	TenantID    string
	ChoirID     string
	Name        string
	Description *string
}
