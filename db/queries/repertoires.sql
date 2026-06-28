-- name: CreateRepertoire :one
INSERT INTO repertoires (tenant_id, choir_id, name, description)
VALUES ($1, $2, $3, $4)
RETURNING id, tenant_id, choir_id, name, description, created_at, updated_at, archived;

-- name: GetRepertoireByIDForMember :one
SELECT r.id, r.tenant_id, r.choir_id, r.name, r.description, r.created_at, r.updated_at, r.archived
FROM repertoires AS r
INNER JOIN choir_members AS cm ON cm.choir_id = r.choir_id
WHERE r.tenant_id = $1
  AND r.id = $2
  AND cm.user_id = $3
  AND cm.active = TRUE;

-- name: ListRepertoiresByChoirID :many
SELECT id, tenant_id, choir_id, name, description, created_at, updated_at, archived
FROM repertoires
WHERE tenant_id = $1
  AND choir_id = $2
ORDER BY name ASC;

-- name: UpdateRepertoire :one
UPDATE repertoires
SET name = $3,
    description = $4,
    updated_at = NOW()
WHERE tenant_id = $1
  AND id = $2
RETURNING id, tenant_id, choir_id, name, description, created_at, updated_at, archived;

-- name: ArchiveRepertoire :execrows
UPDATE repertoires
SET archived = TRUE,
    updated_at = NOW()
WHERE tenant_id = $1
  AND id = $2;

-- name: AddSongToRepertoire :one
INSERT INTO repertoire_songs (tenant_id, repertoire_id, song_id, execution_order, notes)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, tenant_id, repertoire_id, song_id, execution_order, notes, created_at, updated_at;

-- name: RemoveSongFromRepertoire :execrows
DELETE FROM repertoire_songs
WHERE tenant_id = $1
  AND repertoire_id = $2
  AND song_id = $3;

-- name: ListSongsByRepertoire :many
SELECT s.id, s.tenant_id, s.choir_id, s.title, s.composer, s.arranger, s.song_key, s.duration, s.notes, s.created_at, s.updated_at, s.archived, rs.execution_order, rs.notes AS repertoire_notes
FROM repertoire_songs AS rs
INNER JOIN songs AS s ON s.id = rs.song_id
WHERE rs.tenant_id = $1
  AND rs.repertoire_id = $2
ORDER BY rs.execution_order ASC, s.title ASC;

-- name: UpdateRepertoireSongOrder :execrows
UPDATE repertoire_songs
SET execution_order = $4,
    updated_at = NOW()
WHERE tenant_id = $1
  AND repertoire_id = $2
  AND song_id = $3;
