-- name: CreateKitFile :one
INSERT INTO kit_files (tenant_id, voice_kit_id, original_filename, stored_filename, content_type, size_bytes, storage_key)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, tenant_id, voice_kit_id, original_filename, stored_filename, content_type, size_bytes, storage_key, active, created_at, updated_at;

-- name: GetKitFileByIDForMember :one
SELECT kf.id, kf.tenant_id, kf.voice_kit_id, kf.original_filename, kf.stored_filename, kf.content_type, kf.size_bytes,
       kf.storage_key, kf.active, kf.created_at, kf.updated_at, vk.choir_id
FROM kit_files AS kf
INNER JOIN voice_kits AS vk ON vk.id = kf.voice_kit_id
INNER JOIN choir_members AS cm ON cm.choir_id = vk.choir_id
WHERE kf.tenant_id = $1
  AND kf.id = $2
  AND cm.user_id = $3
  AND kf.active = TRUE
  AND vk.active = TRUE
  AND cm.active = TRUE;

-- name: ListKitFilesByVoiceKitID :many
SELECT id, tenant_id, voice_kit_id, original_filename, stored_filename, content_type, size_bytes, storage_key, active, created_at, updated_at
FROM kit_files
WHERE tenant_id = $1
  AND voice_kit_id = $2
  AND active = TRUE
ORDER BY created_at ASC, original_filename ASC;

-- name: DeactivateKitFile :execrows
UPDATE kit_files
SET active = FALSE,
    updated_at = NOW()
WHERE tenant_id = $1
  AND id = $2
  AND active = TRUE;
