package notifications

import (
	"context"
	"time"
)

type Repository interface {
	ClaimDue(ctx context.Context, params ClaimParams) ([]Notification, error)
	MarkSent(ctx context.Context, params FinalizeParams) error
	Retry(ctx context.Context, params RetryParams) error
	MarkFailed(ctx context.Context, params FinalizeParams) error
	MarkInvalidToken(ctx context.Context, params FinalizeParams) error
	CleanupTerminalBefore(ctx context.Context, before time.Time) (int64, error)
}
