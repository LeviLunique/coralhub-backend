CREATE TABLE repertoires (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id UUID NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
	choir_id UUID NOT NULL REFERENCES choirs (id) ON DELETE CASCADE,
	name TEXT NOT NULL,
	description TEXT,
	status TEXT NOT NULL DEFAULT 'active',
	active BOOLEAN NOT NULL DEFAULT TRUE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT repertoires_tenant_choir_name_unique UNIQUE (tenant_id, choir_id, name)
);

CREATE INDEX repertoires_choir_idx ON repertoires (choir_id);

CREATE TABLE repertoire_songs (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id UUID NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
	repertoire_id UUID NOT NULL REFERENCES repertoires (id) ON DELETE CASCADE,
	song_id UUID NOT NULL REFERENCES songs (id) ON DELETE CASCADE,
	execution_order INT NOT NULL DEFAULT 0,
	notes TEXT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT repertoire_songs_repertoire_song_unique UNIQUE (repertoire_id, song_id)
);

CREATE INDEX repertoire_songs_repertoire_idx ON repertoire_songs (repertoire_id);
