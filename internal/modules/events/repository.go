package events

import (
	"context"
	"time"
)

type Repository interface {
	Create(ctx context.Context, params CreateParams) (Event, error)
	Update(ctx context.Context, params UpdateParams) (Event, error)
	GetByIDForMember(ctx context.Context, tenantID string, eventID string, userID string) (Event, error)
	ListByChoirID(ctx context.Context, tenantID string, choirID string) ([]Event, error)
	Cancel(ctx context.Context, tenantID string, eventID string) error
}

type CreateParams struct {
	TenantID    string
	ChoirID     string
	Title       string
	Description *string
	EventType   string
	Location    *string
	StartAt     time.Time
	Reminders   []ScheduledReminder
}

type UpdateParams struct {
	TenantID    string
	EventID     string
	Title       string
	Description *string
	EventType   string
	Location    *string
	StartAt     time.Time
	Reminders   []ScheduledReminder
}
