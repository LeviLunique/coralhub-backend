CREATE TABLE instruments (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id UUID NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
	choir_id UUID NOT NULL REFERENCES choirs (id) ON DELETE CASCADE,
	name TEXT NOT NULL,
	description TEXT,
	icon TEXT,
	active BOOLEAN NOT NULL DEFAULT TRUE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT instruments_tenant_choir_name_unique UNIQUE (tenant_id, choir_id, name)
);

CREATE INDEX instruments_choir_idx ON instruments (choir_id);

CREATE TABLE user_instruments (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id UUID NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
	user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
	instrument_id UUID NOT NULL REFERENCES instruments (id) ON DELETE CASCADE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT user_instruments_user_instrument_unique UNIQUE (user_id, instrument_id)
);

CREATE INDEX user_instruments_instrument_idx ON user_instruments (instrument_id);
