CREATE TABLE device_tokens (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id UUID NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
	user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
	platform TEXT NOT NULL,
	token TEXT NOT NULL,
	active BOOLEAN NOT NULL DEFAULT TRUE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT device_tokens_platform_check CHECK (platform IN ('ios', 'android', 'web')),
	CONSTRAINT device_tokens_tenant_token_unique UNIQUE (tenant_id, token)
);

CREATE INDEX device_tokens_user_active_idx ON device_tokens (user_id, active);
