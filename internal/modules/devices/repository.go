package devices

import "context"

type Repository interface {
	Create(ctx context.Context, params CreateParams) (DeviceToken, error)
	ListActiveByUserID(ctx context.Context, tenantID string, userID string) ([]DeviceToken, error)
	DeactivateByToken(ctx context.Context, tenantID string, token string) error
}
