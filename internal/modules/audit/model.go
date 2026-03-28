package audit

import (
	"encoding/json"
	"time"
)

const (
	EntityTypeEvent        = "event"
	EntityTypeMembership   = "membership"
	EntityTypeNotification = "scheduled_notification"
)

const (
	ActionEventCreated           = "event.created"
	ActionEventUpdated           = "event.updated"
	ActionEventCanceled          = "event.canceled"
	ActionNotificationsGenerated = "notification.generated"
	ActionMembershipAdded        = "membership.added"
	ActionNotificationSent       = "notification.sent"
	ActionNotificationFailed     = "notification.failed"
	ActionNotificationInvalid    = "notification.invalid_token"
)

type Entry struct {
	ID         string
	TenantID   string
	EntityType string
	EntityID   string
	Action     string
	ActorID    *string
	OccurredAt time.Time
	Payload    json.RawMessage
}

type CreateParams struct {
	TenantID   string
	EntityType string
	EntityID   string
	Action     string
	ActorID    *string
	OccurredAt time.Time
	Payload    any
}
