-- name: CreateScheduledNotification :one
INSERT INTO scheduled_notifications (tenant_id, event_id, user_id, reminder_type, scheduled_for, status)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, tenant_id, event_id, user_id, reminder_type, scheduled_for, status, attempts, last_error, processing_started_at, sent_at, created_at, updated_at;

-- name: CancelPendingScheduledNotificationsByEventID :execrows
UPDATE scheduled_notifications
SET status = 'canceled',
    updated_at = NOW()
WHERE tenant_id = $1
  AND event_id = $2
  AND status = 'pending';

-- name: ListScheduledNotificationsByEventID :many
SELECT id, tenant_id, event_id, user_id, reminder_type, scheduled_for, status, attempts, last_error, processing_started_at, sent_at, created_at, updated_at
FROM scheduled_notifications
WHERE tenant_id = $1
  AND event_id = $2
ORDER BY scheduled_for ASC, reminder_type ASC, user_id ASC;
