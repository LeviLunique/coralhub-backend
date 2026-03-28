CREATE TABLE choirs (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id UUID NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
	name TEXT NOT NULL,
	description TEXT,
	active BOOLEAN NOT NULL DEFAULT TRUE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT choirs_tenant_name_unique UNIQUE (tenant_id, name)
);

CREATE INDEX choirs_tenant_id_idx ON choirs (tenant_id);

CREATE TABLE users (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id UUID NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
	email TEXT NOT NULL,
	full_name TEXT NOT NULL,
	active BOOLEAN NOT NULL DEFAULT TRUE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT users_tenant_email_unique UNIQUE (tenant_id, email)
);

CREATE INDEX users_tenant_id_idx ON users (tenant_id);
