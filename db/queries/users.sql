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

-- name: GetUserByEmail :one
SELECT id, tenant_id, email, full_name, active, created_at, updated_at
FROM users
WHERE tenant_id = $1
  AND email = $2
  AND active = TRUE;

-- name: ListLoginIdentitiesByEmail :many
SELECT
  u.id,
  u.tenant_id,
  t.slug AS tenant_slug,
  t.display_name AS tenant_display_name,
  u.email,
  u.full_name,
  u.active,
  u.password_hash,
  EXISTS (
    SELECT 1
    FROM choir_members AS cm
    WHERE cm.tenant_id = u.tenant_id
      AND cm.user_id = u.id
      AND cm.role = 'manager'
      AND cm.active = TRUE
  ) AS manager
FROM users AS u
INNER JOIN tenants AS t ON t.id = u.tenant_id
WHERE lower(u.email) = lower($1)
  AND u.active = TRUE
ORDER BY t.slug ASC;

-- name: ListUsersByTenantID :many
SELECT id, tenant_id, email, full_name, active, created_at, updated_at
FROM users
WHERE tenant_id = $1
  AND active = TRUE
ORDER BY full_name ASC, email ASC;
