-- name: CreateInstrument :one
INSERT INTO instruments (tenant_id, choir_id, name, description, icon)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, tenant_id, choir_id, name, description, icon, created_at, updated_at, archived;

-- name: GetInstrumentByIDForMember :one
SELECT i.id, i.tenant_id, i.choir_id, i.name, i.description, i.icon, i.created_at, i.updated_at, i.archived
FROM instruments AS i
INNER JOIN choir_members AS cm ON cm.choir_id = i.choir_id
WHERE i.tenant_id = $1
  AND i.id = $2
  AND cm.user_id = $3
  AND cm.active = TRUE;

-- name: ListInstrumentsByChoirID :many
SELECT id, tenant_id, choir_id, name, description, icon, created_at, updated_at, archived
FROM instruments
WHERE tenant_id = $1
  AND choir_id = $2
ORDER BY name ASC;

-- name: UpdateInstrument :one
UPDATE instruments
SET name = $3,
    description = $4,
    icon = $5,
    updated_at = NOW()
WHERE tenant_id = $1
  AND id = $2
RETURNING id, tenant_id, choir_id, name, description, icon, created_at, updated_at, archived;

-- name: ArchiveInstrument :execrows
UPDATE instruments
SET archived = TRUE,
    updated_at = NOW()
WHERE tenant_id = $1
  AND id = $2;

-- name: AddInstrumentToUser :one
INSERT INTO user_instruments (tenant_id, user_id, instrument_id)
VALUES ($1, $2, $3)
RETURNING id, tenant_id, user_id, instrument_id, created_at, updated_at;

-- name: RemoveInstrumentFromUser :execrows
DELETE FROM user_instruments
WHERE tenant_id = $1
  AND user_id = $2
  AND instrument_id = $3;

-- name: ListInstrumentsByUser :many
SELECT i.id, i.tenant_id, i.choir_id, i.name, i.description, i.icon, i.created_at, i.updated_at, i.archived
FROM user_instruments AS ui
INNER JOIN instruments AS i ON i.id = ui.instrument_id
WHERE ui.tenant_id = $1
  AND ui.user_id = $2
ORDER BY i.name ASC;

-- name: ListUsersByInstrument :many
SELECT u.id, u.tenant_id, u.email, u.full_name, u.active, u.created_at, u.updated_at
FROM user_instruments AS ui
INNER JOIN users AS u ON u.id = ui.user_id
WHERE ui.tenant_id = $1
  AND ui.instrument_id = $2
  AND u.active = TRUE
ORDER BY u.full_name ASC;
