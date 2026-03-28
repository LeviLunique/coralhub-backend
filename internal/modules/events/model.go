package events

import "time"

const (
	EventTypeRehearsal    = "rehearsal"
	EventTypePresentation = "presentation"
	EventTypeOther        = "other"

	ReminderTypeDayBefore  = "day_before"
	ReminderTypeHourBefore = "hour_before"

	NotificationStatusPending  = "pending"
	NotificationStatusCanceled = "canceled"
)

type Event struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	ChoirID     string    `json:"choir_id"`
	Title       string    `json:"title"`
	Description *string   `json:"description,omitempty"`
	EventType   string    `json:"event_type"`
	Location    *string   `json:"location,omitempty"`
	StartAt     time.Time `json:"start_at"`
	Active      bool      `json:"active"`
}

type CreateInput struct {
	Title       string    `json:"title"`
	Description *string   `json:"description,omitempty"`
	EventType   string    `json:"event_type"`
	Location    *string   `json:"location,omitempty"`
	StartAt     time.Time `json:"start_at"`
}

type UpdateInput struct {
	Title       string    `json:"title"`
	Description *string   `json:"description,omitempty"`
	EventType   string    `json:"event_type"`
	Location    *string   `json:"location,omitempty"`
	StartAt     time.Time `json:"start_at"`
}

type ScheduledReminder struct {
	UserID       string
	ReminderType string
	ScheduledFor time.Time
	Status       string
}
