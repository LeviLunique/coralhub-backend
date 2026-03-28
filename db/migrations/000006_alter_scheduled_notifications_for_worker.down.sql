ALTER TABLE scheduled_notifications
	DROP CONSTRAINT IF EXISTS scheduled_notifications_attempts_non_negative_check;

ALTER TABLE scheduled_notifications
	DROP CONSTRAINT IF EXISTS scheduled_notifications_status_check;

ALTER TABLE scheduled_notifications
	ADD CONSTRAINT scheduled_notifications_status_check
	CHECK (status IN ('pending', 'processing', 'sent', 'failed', 'canceled'));

ALTER TABLE scheduled_notifications
	DROP COLUMN IF EXISTS sent_at,
	DROP COLUMN IF EXISTS processing_started_at,
	DROP COLUMN IF EXISTS last_error,
	DROP COLUMN IF EXISTS attempts;
