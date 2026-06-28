CREATE TABLE event_repertoires (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id UUID NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
	event_id UUID NOT NULL REFERENCES events (id) ON DELETE CASCADE,
	repertoire_id UUID NOT NULL REFERENCES repertoires (id) ON DELETE CASCADE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT event_repertoires_event_repertoire_unique UNIQUE (event_id, repertoire_id)
);

CREATE INDEX event_repertoires_event_idx ON event_repertoires (event_id);
