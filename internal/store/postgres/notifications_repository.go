package postgres

import (
	"context"
	"time"

	"github.com/LeviLunique/coralhub-backend/internal/modules/notifications"
	"github.com/LeviLunique/coralhub-backend/internal/store/postgres/sqlc"
)

type NotificationRepository struct {
	queries *sqlc.Queries
}

func NewNotificationRepository(queries *sqlc.Queries) *NotificationRepository {
	return &NotificationRepository{queries: queries}
}

func (r *NotificationRepository) ClaimDue(ctx context.Context, params notifications.ClaimParams) ([]notifications.Notification, error) {
	rows, err := r.queries.ClaimDueScheduledNotifications(ctx, sqlc.ClaimDueScheduledNotificationsParams{
		ProcessingStartedAt:   timestamptzValue(params.ClaimedAt),
		ProcessingStartedAt_2: timestamptzValue(params.StaleBefore),
		Limit:                 params.Limit,
	})
	if err != nil {
		return nil, err
	}

	items := make([]notifications.Notification, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapScheduledNotificationRow(row))
	}

	return items, nil
}

func (r *NotificationRepository) MarkSent(ctx context.Context, params notifications.FinalizeParams) error {
	tenantID, err := parseUUID(params.TenantID)
	if err != nil {
		return notifications.ErrNotificationLeaseLost
	}

	notificationID, err := parseUUID(params.NotificationID)
	if err != nil {
		return notifications.ErrNotificationLeaseLost
	}

	affected, err := r.queries.MarkScheduledNotificationSent(ctx, sqlc.MarkScheduledNotificationSentParams{
		TenantID:            tenantID,
		ID:                  notificationID,
		ProcessingStartedAt: timestamptzValue(params.ProcessingStartedAt),
		SentAt:              timestamptzValue(params.At),
	})
	if err != nil {
		return err
	}
	if affected == 0 {
		return notifications.ErrNotificationLeaseLost
	}

	return nil
}

func (r *NotificationRepository) Retry(ctx context.Context, params notifications.RetryParams) error {
	tenantID, err := parseUUID(params.TenantID)
	if err != nil {
		return notifications.ErrNotificationLeaseLost
	}

	notificationID, err := parseUUID(params.NotificationID)
	if err != nil {
		return notifications.ErrNotificationLeaseLost
	}

	affected, err := r.queries.RetryScheduledNotification(ctx, sqlc.RetryScheduledNotificationParams{
		TenantID:            tenantID,
		ID:                  notificationID,
		ProcessingStartedAt: timestamptzValue(params.ProcessingStartedAt),
		ScheduledFor:        timestamptzValue(params.NextAttemptAt),
		LastError:           textValue(stringPointer(params.LastError)),
	})
	if err != nil {
		return err
	}
	if affected == 0 {
		return notifications.ErrNotificationLeaseLost
	}

	return nil
}

func (r *NotificationRepository) MarkFailed(ctx context.Context, params notifications.FinalizeParams) error {
	tenantID, err := parseUUID(params.TenantID)
	if err != nil {
		return notifications.ErrNotificationLeaseLost
	}

	notificationID, err := parseUUID(params.NotificationID)
	if err != nil {
		return notifications.ErrNotificationLeaseLost
	}

	affected, err := r.queries.FailScheduledNotification(ctx, sqlc.FailScheduledNotificationParams{
		TenantID:            tenantID,
		ID:                  notificationID,
		ProcessingStartedAt: timestamptzValue(params.ProcessingStartedAt),
		LastError:           textValue(stringPointer(params.LastError)),
	})
	if err != nil {
		return err
	}
	if affected == 0 {
		return notifications.ErrNotificationLeaseLost
	}

	return nil
}

func (r *NotificationRepository) MarkInvalidToken(ctx context.Context, params notifications.FinalizeParams) error {
	tenantID, err := parseUUID(params.TenantID)
	if err != nil {
		return notifications.ErrNotificationLeaseLost
	}

	notificationID, err := parseUUID(params.NotificationID)
	if err != nil {
		return notifications.ErrNotificationLeaseLost
	}

	affected, err := r.queries.MarkScheduledNotificationInvalidToken(ctx, sqlc.MarkScheduledNotificationInvalidTokenParams{
		TenantID:            tenantID,
		ID:                  notificationID,
		ProcessingStartedAt: timestamptzValue(params.ProcessingStartedAt),
		LastError:           textValue(stringPointer(params.LastError)),
	})
	if err != nil {
		return err
	}
	if affected == 0 {
		return notifications.ErrNotificationLeaseLost
	}

	return nil
}

func mapScheduledNotificationRow(row sqlc.ClaimDueScheduledNotificationsRow) notifications.Notification {
	var processingStartedAt *time.Time
	if row.ProcessingStartedAt.Valid {
		value := row.ProcessingStartedAt.Time.UTC()
		processingStartedAt = &value
	}

	var sentAt *time.Time
	if row.SentAt.Valid {
		value := row.SentAt.Time.UTC()
		sentAt = &value
	}

	return notifications.Notification{
		ID:                  uuidString(row.ID),
		TenantID:            uuidString(row.TenantID),
		EventID:             uuidString(row.EventID),
		UserID:              uuidString(row.UserID),
		ReminderType:        row.ReminderType,
		ScheduledFor:        row.ScheduledFor.Time.UTC(),
		Status:              row.Status,
		Attempts:            row.Attempts,
		LastError:           textPointer(row.LastError),
		ProcessingStartedAt: processingStartedAt,
		SentAt:              sentAt,
	}
}

func stringPointer(value string) *string {
	return &value
}
