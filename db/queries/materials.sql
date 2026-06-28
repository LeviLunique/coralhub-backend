-- name: CreateMaterial :one
INSERT INTO materials (tenant_id, choir_id, song_id, name, material_type, target_type, voice_type)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, tenant_id, choir_id, song_id, name, material_type, target_type, voice_type, archived, created_at, updated_at;

-- name: GetMaterialByIDForMember :one
SELECT m.id, m.tenant_id, m.choir_id, m.song_id, m.name, m.material_type, m.target_type, m.voice_type, m.archived, m.created_at, m.updated_at
FROM materials AS m
INNER JOIN choir_members AS cm ON cm.choir_id = m.choir_id
WHERE m.tenant_id = $1
  AND m.id = $2
  AND cm.user_id = $3
  AND cm.active = TRUE;

-- name: ListMaterialsByChoirID :many
SELECT id, tenant_id, choir_id, song_id, name, material_type, target_type, voice_type, archived, created_at, updated_at
FROM materials
WHERE tenant_id = $1
  AND choir_id = $2
ORDER BY name ASC;

-- name: ListMaterialsBySongID :many
SELECT id, tenant_id, choir_id, song_id, name, material_type, target_type, voice_type, archived, created_at, updated_at
FROM materials
WHERE tenant_id = $1
  AND song_id = $2
ORDER BY name ASC;

-- name: UpdateMaterial :one
UPDATE materials
SET name = $3,
    material_type = $4,
    target_type = $5,
    voice_type = $6,
    updated_at = NOW()
WHERE tenant_id = $1
  AND id = $2
RETURNING id, tenant_id, choir_id, song_id, name, material_type, target_type, voice_type, archived, created_at, updated_at;

-- name: ArchiveMaterial :execrows
UPDATE materials
SET archived = TRUE,
    updated_at = NOW()
WHERE tenant_id = $1
  AND id = $2;

-- name: AddInstrumentToMaterial :one
INSERT INTO material_instruments (tenant_id, material_id, instrument_id)
VALUES ($1, $2, $3)
RETURNING id, tenant_id, material_id, instrument_id, created_at, updated_at;

-- name: RemoveInstrumentFromMaterial :execrows
DELETE FROM material_instruments
WHERE tenant_id = $1
  AND material_id = $2
  AND instrument_id = $3;

-- name: ListInstrumentsByMaterial :many
SELECT i.id, i.tenant_id, i.choir_id, i.name, i.description, i.icon, i.archived, i.created_at, i.updated_at
FROM material_instruments AS mi
INNER JOIN instruments AS i ON i.id = mi.instrument_id
WHERE mi.tenant_id = $1
  AND mi.material_id = $2
ORDER BY i.name ASC;

-- name: ListFilesByMaterial :many
SELECT id, tenant_id, material_id, original_filename, stored_filename, content_type, size_bytes, storage_key, archived, created_at, updated_at
FROM material_files
WHERE tenant_id = $1
  AND material_id = $2
ORDER BY created_at ASC, original_filename ASC;
