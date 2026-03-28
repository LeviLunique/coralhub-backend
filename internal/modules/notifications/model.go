package notifications

import "time"

const (
	StatusPending      = "pending"
	StatusProcessing   = "processing"
	StatusSent         = "sent"
	StatusFailed       = "failed"
	StatusCanceled     = "canceled"
	StatusInvalidToken = "invalid_token"

	DeliverySent             = "sent"
	DeliveryTransientFailure = "transient_failure"
	DeliveryInvalidToken     = "invalid_token"
)

type Notification struct {
	ID                  string
	TenantID            string
	EventID             string
	UserID              string
	ReminderType        string
	ScheduledFor        time.Time
	Status              string
	Attempts            int32
	LastError           *string
	ProcessingStartedAt *time.Time
	SentAt              *time.Time
}

type ClaimParams struct {
	ClaimedAt   time.Time
	StaleBefore time.Time
	Limit       int32
}

type RetryParams struct {
	TenantID            string
	NotificationID      string
	ProcessingStartedAt time.Time
	NextAttemptAt       time.Time
	LastError           string
}

type FinalizeParams struct {
	TenantID            string
	NotificationID      string
	ProcessingStartedAt time.Time
	At                  time.Time
	LastError           string
}

type DeliveryResult struct {
	Kind         string
	ErrorMessage string
}
