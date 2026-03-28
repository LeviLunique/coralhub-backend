ALTER TABLE scheduled_notifications
	ADD COLUMN attempts INTEGER NOT NULL DEFAULT 0,
	ADD COLUMN last_error TEXT,
	ADD COLUMN processing_started_at TIMESTAMPTZ,
	ADD COLUMN sent_at TIMESTAMPTZ;

ALTER TABLE scheduled_notifications
	DROP CONSTRAINT scheduled_notifications_status_check;

ALTER TABLE scheduled_notifications
	ADD CONSTRAINT scheduled_notifications_status_check
	CHECK (status IN ('pending', 'processing', 'sent', 'failed', 'canceled', 'invalid_token'));

ALTER TABLE scheduled_notifications
	ADD CONSTRAINT scheduled_notifications_attempts_non_negative_check
	CHECK (attempts >= 0);
