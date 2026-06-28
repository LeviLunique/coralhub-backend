CREATE TABLE songs (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id UUID NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
	choir_id UUID NOT NULL REFERENCES choirs (id) ON DELETE CASCADE,
	title TEXT NOT NULL,
	composer TEXT,
	arranger TEXT,
	song_key TEXT,
	duration TEXT,
	notes TEXT,
	status TEXT NOT NULL DEFAULT 'active',
	active BOOLEAN NOT NULL DEFAULT TRUE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT songs_tenant_choir_title_unique UNIQUE (tenant_id, choir_id, title)
);

CREATE INDEX songs_choir_idx ON songs (choir_id);
