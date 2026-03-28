-- name: CreateChoirMember :one
INSERT INTO choir_members (tenant_id, choir_id, user_id, role)
VALUES ($1, $2, $3, $4)
RETURNING id, tenant_id, choir_id, user_id, role, active, created_at, updated_at;

-- name: GetChoirMemberByChoirAndUser :one
SELECT id, tenant_id, choir_id, user_id, role, active, created_at, updated_at
FROM choir_members
WHERE tenant_id = $1
  AND choir_id = $2
  AND user_id = $3
  AND active = TRUE;

-- name: ListChoirMembersByChoirID :many
SELECT cm.id, cm.tenant_id, cm.choir_id, cm.user_id, cm.role, cm.active, cm.created_at, cm.updated_at,
       u.email, u.full_name
FROM choir_members AS cm
INNER JOIN users AS u ON u.id = cm.user_id
WHERE cm.tenant_id = $1
  AND cm.choir_id = $2
  AND cm.active = TRUE
  AND u.active = TRUE
ORDER BY u.full_name ASC, u.email ASC;
