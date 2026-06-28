-- name: AddRepertoireToEvent :one
INSERT INTO event_repertoires (tenant_id, event_id, repertoire_id)
VALUES ($1, $2, $3)
RETURNING id, tenant_id, event_id, repertoire_id, created_at, updated_at;

-- name: RemoveRepertoireFromEvent :execrows
DELETE FROM event_repertoires
WHERE tenant_id = $1
  AND event_id = $2
  AND repertoire_id = $3;

-- name: ListRepertoiresByEvent :many
SELECT r.id, r.tenant_id, r.choir_id, r.name, r.description, r.created_at, r.updated_at, r.archived
FROM event_repertoires AS er
INNER JOIN repertoires AS r ON r.id = er.repertoire_id
WHERE er.tenant_id = $1
  AND er.event_id = $2
ORDER BY r.name ASC;
