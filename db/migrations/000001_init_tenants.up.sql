CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE tenants (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	slug TEXT NOT NULL,
	display_name TEXT NOT NULL,
	active BOOLEAN NOT NULL DEFAULT TRUE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT tenants_slug_unique UNIQUE (slug)
);

CREATE TABLE tenant_configs (
	tenant_id UUID PRIMARY KEY REFERENCES tenants (id) ON DELETE CASCADE,
	logo_url TEXT,
	primary_color TEXT,
	secondary_color TEXT,
	custom_domain TEXT UNIQUE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO tenants (slug, display_name)
VALUES ('coral-jovem-asa-norte', 'Coral Jovem Asa Norte');
