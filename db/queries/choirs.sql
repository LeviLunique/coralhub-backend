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

-- name: GetChoirByIDForMember :one
SELECT c.id, c.tenant_id, c.name, c.description, c.active, c.created_at, c.updated_at
FROM choirs AS c
INNER JOIN choir_members AS cm ON cm.choir_id = c.id
WHERE c.tenant_id = $1
  AND c.id = $2
  AND cm.user_id = $3
  AND c.active = TRUE
  AND cm.active = TRUE;

-- name: ListChoirsByTenantID :many
SELECT id, tenant_id, name, description, active, created_at, updated_at
FROM choirs
WHERE tenant_id = $1
  AND active = TRUE
ORDER BY name ASC;

-- name: ListChoirsByMemberUserID :many
SELECT c.id, c.tenant_id, c.name, c.description, c.active, c.created_at, c.updated_at
FROM choirs AS c
INNER JOIN choir_members AS cm ON cm.choir_id = c.id
WHERE c.tenant_id = $1
  AND cm.user_id = $2
  AND c.active = TRUE
  AND cm.active = TRUE
ORDER BY c.name ASC;
