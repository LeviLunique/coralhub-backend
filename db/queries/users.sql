-- name: CreateUser :one
INSERT INTO users (tenant_id, email, full_name)
VALUES ($1, $2, $3)
RETURNING id, tenant_id, email, full_name, active, created_at, updated_at;

-- name: GetUserByID :one
SELECT id, tenant_id, email, full_name, active, created_at, updated_at
FROM users
WHERE tenant_id = $1
  AND id = $2
  AND active = TRUE;

-- name: ListUsersByTenantID :many
SELECT id, tenant_id, email, full_name, active, created_at, updated_at
FROM users
WHERE tenant_id = $1
  AND active = TRUE
ORDER BY full_name ASC, email ASC;
