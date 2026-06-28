-- name: CreateSong :one
INSERT INTO songs (tenant_id, choir_id, title, composer, arranger, song_key, duration, notes)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, tenant_id, choir_id, title, composer, arranger, song_key, duration, notes, created_at, updated_at, archived;

-- name: GetSongByIDForMember :one
SELECT s.id, s.tenant_id, s.choir_id, s.title, s.composer, s.arranger, s.song_key, s.duration, s.notes, s.created_at, s.updated_at, s.archived
FROM songs AS s
INNER JOIN choir_members AS cm ON cm.choir_id = s.choir_id
WHERE s.tenant_id = $1
  AND s.id = $2
  AND cm.user_id = $3
  AND cm.active = TRUE;

-- name: ListSongsByChoirID :many
SELECT id, tenant_id, choir_id, title, composer, arranger, song_key, duration, notes, created_at, updated_at, archived
FROM songs
WHERE tenant_id = $1
  AND choir_id = $2
ORDER BY title ASC;

-- name: UpdateSong :one
UPDATE songs
SET title = $3,
    composer = $4,
    arranger = $5,
    song_key = $6,
    duration = $7,
    notes = $8,
    updated_at = NOW()
WHERE tenant_id = $1
  AND id = $2
RETURNING id, tenant_id, choir_id, title, composer, arranger, song_key, duration, notes, created_at, updated_at, archived;

-- name: ArchiveSong :execrows
UPDATE songs
SET archived = TRUE,
    updated_at = NOW()
WHERE tenant_id = $1
  AND id = $2;
