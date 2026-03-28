-- name: CreateChoir :one
INSERT INTO choirs (tenant_id, name, description)
VALUES ($1, $2, $3)
RETURNING id, tenant_id, name, description, active, created_at, updated_at;

-- name: GetChoirByID :one
SELECT id, tenant_id, name, description, active, created_at, updated_at
FROM choirs
WHERE tenant_id = $1
  AND id = $2
  AND active = TRUE;

-- name: ListChoirsByTenantID :many
SELECT id, tenant_id, name, description, active, created_at, updated_at
FROM choirs
WHERE tenant_id = $1
  AND active = TRUE
ORDER BY name ASC;
