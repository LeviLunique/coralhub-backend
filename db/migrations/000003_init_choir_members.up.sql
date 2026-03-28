CREATE TABLE choir_members (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id UUID NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
	choir_id UUID NOT NULL REFERENCES choirs (id) ON DELETE CASCADE,
	user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
	role TEXT NOT NULL,
	active BOOLEAN NOT NULL DEFAULT TRUE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT choir_members_role_check CHECK (role IN ('manager', 'member')),
	CONSTRAINT choir_members_tenant_choir_user_unique UNIQUE (tenant_id, choir_id, user_id)
);

CREATE INDEX choir_members_tenant_choir_idx ON choir_members (tenant_id, choir_id);
CREATE INDEX choir_members_tenant_user_idx ON choir_members (tenant_id, user_id);
