CREATE TABLE audit_log (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id UUID NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
	entity_type TEXT NOT NULL,
	entity_id UUID NOT NULL,
	action TEXT NOT NULL,
	actor_id UUID REFERENCES users (id) ON DELETE SET NULL,
	occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	payload_json JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX audit_log_tenant_occurred_at_idx ON audit_log (tenant_id, occurred_at DESC);
CREATE INDEX audit_log_tenant_entity_idx ON audit_log (tenant_id, entity_type, entity_id, occurred_at DESC);
