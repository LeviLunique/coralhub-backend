-- name: CreateEvent :one
INSERT INTO events (tenant_id, choir_id, title, description, event_type, location, start_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, tenant_id, choir_id, title, description, event_type, location, start_at, active, created_at, updated_at;

-- name: UpdateEvent :one
UPDATE events
SET title = $3,
    description = $4,
    event_type = $5,
    location = $6,
    start_at = $7,
    updated_at = NOW()
WHERE tenant_id = $1
  AND id = $2
  AND active = TRUE
RETURNING id, tenant_id, choir_id, title, description, event_type, location, start_at, active, created_at, updated_at;

-- name: GetEventByIDForMember :one
SELECT e.id, e.tenant_id, e.choir_id, e.title, e.description, e.event_type, e.location, e.start_at, e.active, e.created_at, e.updated_at
FROM events AS e
INNER JOIN choir_members AS cm ON cm.choir_id = e.choir_id
WHERE e.tenant_id = $1
  AND e.id = $2
  AND cm.user_id = $3
  AND e.active = TRUE
  AND cm.active = TRUE;

-- name: ListEventsByChoirID :many
SELECT id, tenant_id, choir_id, title, description, event_type, location, start_at, active, created_at, updated_at
FROM events
WHERE tenant_id = $1
  AND choir_id = $2
  AND active = TRUE
ORDER BY start_at ASC, title ASC;

-- name: CancelEvent :execrows
UPDATE events
SET active = FALSE,
    updated_at = NOW()
WHERE tenant_id = $1
  AND id = $2
  AND active = TRUE;

-- name: ListActiveChoirMemberUserIDs :many
SELECT cm.user_id
FROM choir_members AS cm
INNER JOIN users AS u ON u.id = cm.user_id
WHERE cm.tenant_id = $1
  AND cm.choir_id = $2
  AND cm.active = TRUE
  AND u.active = TRUE
ORDER BY cm.user_id;
