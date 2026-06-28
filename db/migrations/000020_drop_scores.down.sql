CREATE TABLE scores (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id UUID NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
	choir_id UUID NOT NULL REFERENCES choirs (id) ON DELETE CASCADE,
	song_id UUID NOT NULL REFERENCES songs (id) ON DELETE CASCADE,
	name TEXT NOT NULL,
	target_type TEXT NOT NULL DEFAULT 'voice',
	voice_type TEXT,
	status TEXT NOT NULL DEFAULT 'active',
	active BOOLEAN NOT NULL DEFAULT TRUE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT scores_target_type_check CHECK (target_type IN ('voice', 'instrument'))
);

CREATE INDEX scores_song_idx ON scores (song_id);

CREATE TABLE score_instruments (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id UUID NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
	score_id UUID NOT NULL REFERENCES scores (id) ON DELETE CASCADE,
	instrument_id UUID NOT NULL REFERENCES instruments (id) ON DELETE CASCADE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT score_instruments_score_instrument_unique UNIQUE (score_id, instrument_id)
);

CREATE INDEX score_instruments_score_idx ON score_instruments (score_id);
