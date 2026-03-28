CREATE TABLE voice_kits (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id UUID NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
	choir_id UUID NOT NULL REFERENCES choirs (id) ON DELETE CASCADE,
	name TEXT NOT NULL,
	description TEXT,
	active BOOLEAN NOT NULL DEFAULT TRUE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT voice_kits_tenant_choir_name_unique UNIQUE (tenant_id, choir_id, name)
);

CREATE INDEX voice_kits_choir_idx ON voice_kits (choir_id);
CREATE INDEX voice_kits_tenant_choir_active_idx ON voice_kits (tenant_id, choir_id, active);

CREATE TABLE kit_files (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id UUID NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
	voice_kit_id UUID NOT NULL REFERENCES voice_kits (id) ON DELETE CASCADE,
	original_filename TEXT NOT NULL,
	stored_filename TEXT NOT NULL,
	content_type TEXT NOT NULL,
	size_bytes BIGINT NOT NULL,
	storage_key TEXT NOT NULL,
	active BOOLEAN NOT NULL DEFAULT TRUE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT kit_files_size_bytes_positive CHECK (size_bytes > 0)
);

CREATE INDEX kit_files_voice_kit_active_idx ON kit_files (voice_kit_id, active);
CREATE INDEX kit_files_tenant_voice_kit_idx ON kit_files (tenant_id, voice_kit_id);
