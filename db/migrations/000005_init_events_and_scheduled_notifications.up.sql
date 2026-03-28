CREATE TABLE events (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id UUID NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
	choir_id UUID NOT NULL REFERENCES choirs (id) ON DELETE CASCADE,
	title TEXT NOT NULL,
	description TEXT,
	event_type TEXT NOT NULL,
	location TEXT,
	start_at TIMESTAMPTZ NOT NULL,
	active BOOLEAN NOT NULL DEFAULT TRUE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT events_event_type_check CHECK (event_type IN ('rehearsal', 'presentation', 'other'))
);

CREATE INDEX events_choir_start_at_idx ON events (choir_id, start_at);
CREATE INDEX events_tenant_choir_active_idx ON events (tenant_id, choir_id, active);

CREATE TABLE scheduled_notifications (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id UUID NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
	event_id UUID NOT NULL REFERENCES events (id) ON DELETE CASCADE,
	user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
	reminder_type TEXT NOT NULL,
	scheduled_for TIMESTAMPTZ NOT NULL,
	status TEXT NOT NULL DEFAULT 'pending',
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT scheduled_notifications_reminder_type_check CHECK (reminder_type IN ('day_before', 'hour_before')),
	CONSTRAINT scheduled_notifications_status_check CHECK (status IN ('pending', 'processing', 'sent', 'failed', 'canceled'))
);

CREATE INDEX scheduled_notifications_status_scheduled_for_idx ON scheduled_notifications (status, scheduled_for);
CREATE INDEX scheduled_notifications_event_id_idx ON scheduled_notifications (event_id);
CREATE UNIQUE INDEX scheduled_notifications_pending_identity_idx
	ON scheduled_notifications (tenant_id, user_id, event_id, reminder_type)
	WHERE status IN ('pending', 'processing');
