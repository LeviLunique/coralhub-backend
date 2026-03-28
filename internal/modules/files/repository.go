package files

import "context"

type Repository interface {
	Create(ctx context.Context, params CreateParams) (File, error)
	GetByIDForMember(ctx context.Context, tenantID string, fileID string, userID string) (File, error)
	ListByVoiceKitID(ctx context.Context, tenantID string, voiceKitID string) ([]File, error)
	Delete(ctx context.Context, tenantID string, fileID string) error
}

type CreateParams struct {
	ID               string
	TenantID         string
	VoiceKitID       string
	OriginalFilename string
	StoredFilename   string
	ContentType      string
	SizeBytes        int64
	StorageKey       string
}
