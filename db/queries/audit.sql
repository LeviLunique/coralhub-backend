-- name: CreateAuditLog :one
INSERT INTO audit_log (tenant_id, entity_type, entity_id, action, actor_id, occurred_at, payload_json)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, tenant_id, entity_type, entity_id, action, actor_id, occurred_at, payload_json;

-- name: ListAuditLogByTenantID :many
SELECT id, tenant_id, entity_type, entity_id, action, actor_id, occurred_at, payload_json
FROM audit_log
WHERE tenant_id = $1
ORDER BY occurred_at DESC, id DESC
LIMIT $2;
