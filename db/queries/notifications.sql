-- name: ClaimDueScheduledNotifications :many
WITH due AS (
	SELECT id
	FROM scheduled_notifications AS sn
	WHERE (
		sn.status = 'pending'
		AND sn.scheduled_for <= $1
	) OR (
		sn.status = 'processing'
		AND sn.processing_started_at IS NOT NULL
		AND sn.processing_started_at <= $2
	)
	ORDER BY sn.scheduled_for ASC, sn.created_at ASC
	LIMIT $3
	FOR UPDATE SKIP LOCKED
)
UPDATE scheduled_notifications AS sn
SET status = 'processing',
	processing_started_at = $1,
	updated_at = NOW()
FROM due
WHERE sn.id = due.id
RETURNING sn.id, sn.tenant_id, sn.event_id, sn.user_id, sn.reminder_type, sn.scheduled_for, sn.status,
	sn.attempts, sn.last_error, sn.processing_started_at, sn.sent_at, sn.created_at, sn.updated_at;

-- name: MarkScheduledNotificationSent :execrows
UPDATE scheduled_notifications
SET status = 'sent',
	attempts = attempts + 1,
	last_error = NULL,
	processing_started_at = NULL,
	sent_at = $4,
	updated_at = NOW()
WHERE tenant_id = $1
	AND id = $2
	AND status = 'processing'
	AND processing_started_at = $3;

-- name: RetryScheduledNotification :execrows
UPDATE scheduled_notifications
SET status = 'pending',
	attempts = attempts + 1,
	scheduled_for = $4,
	last_error = $5,
	processing_started_at = NULL,
	updated_at = NOW()
WHERE tenant_id = $1
	AND id = $2
	AND status = 'processing'
	AND processing_started_at = $3;

-- name: FailScheduledNotification :execrows
UPDATE scheduled_notifications
SET status = 'failed',
	attempts = attempts + 1,
	last_error = $4,
	processing_started_at = NULL,
	updated_at = NOW()
WHERE tenant_id = $1
	AND id = $2
	AND status = 'processing'
	AND processing_started_at = $3;

-- name: MarkScheduledNotificationInvalidToken :execrows
UPDATE scheduled_notifications
SET status = 'invalid_token',
	attempts = attempts + 1,
	last_error = $4,
	processing_started_at = NULL,
	updated_at = NOW()
WHERE tenant_id = $1
	AND id = $2
	AND status = 'processing'
	AND processing_started_at = $3;
