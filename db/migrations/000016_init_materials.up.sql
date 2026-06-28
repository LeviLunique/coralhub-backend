CREATE TABLE materials (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id UUID NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
	choir_id UUID NOT NULL REFERENCES choirs (id) ON DELETE CASCADE,
	song_id UUID NOT NULL REFERENCES songs (id) ON DELETE CASCADE,
	name TEXT NOT NULL,
	material_type TEXT NOT NULL DEFAULT 'other',
	target_type TEXT NOT NULL DEFAULT 'voice',
	voice_type TEXT,
	archived BOOLEAN NOT NULL DEFAULT FALSE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT materials_material_type_check CHECK (material_type IN ('audio_guide', 'sheet_music', 'playback', 'lyrics', 'other')),
	CONSTRAINT materials_target_type_check CHECK (target_type IN ('voice', 'instrument'))
);

CREATE INDEX materials_song_idx ON materials (song_id);

CREATE TABLE material_instruments (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id UUID NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
	material_id UUID NOT NULL REFERENCES materials (id) ON DELETE CASCADE,
	instrument_id UUID NOT NULL REFERENCES instruments (id) ON DELETE CASCADE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT material_instruments_material_instrument_unique UNIQUE (material_id, instrument_id)
);

CREATE INDEX material_instruments_material_idx ON material_instruments (material_id);

CREATE TABLE material_files (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id UUID NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
	material_id UUID NOT NULL REFERENCES materials (id) ON DELETE CASCADE,
	original_filename TEXT NOT NULL,
	stored_filename TEXT NOT NULL,
	content_type TEXT NOT NULL,
	size_bytes BIGINT NOT NULL,
	storage_key TEXT NOT NULL,
	archived BOOLEAN NOT NULL DEFAULT FALSE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT material_files_size_bytes_positive CHECK (size_bytes > 0)
);

CREATE INDEX material_files_material_idx ON material_files (material_id);
