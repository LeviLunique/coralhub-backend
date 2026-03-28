-- name: CreateVoiceKit :one
INSERT INTO voice_kits (tenant_id, choir_id, name, description)
VALUES ($1, $2, $3, $4)
RETURNING id, tenant_id, choir_id, name, description, active, created_at, updated_at;

-- name: GetVoiceKitByIDForMember :one
SELECT vk.id, vk.tenant_id, vk.choir_id, vk.name, vk.description, vk.active, vk.created_at, vk.updated_at
FROM voice_kits AS vk
INNER JOIN choir_members AS cm ON cm.choir_id = vk.choir_id
WHERE vk.tenant_id = $1
  AND vk.id = $2
  AND cm.user_id = $3
  AND vk.active = TRUE
  AND cm.active = TRUE;

-- name: ListVoiceKitsByChoirID :many
SELECT id, tenant_id, choir_id, name, description, active, created_at, updated_at
FROM voice_kits
WHERE tenant_id = $1
  AND choir_id = $2
  AND active = TRUE
ORDER BY name ASC;

-- name: DeactivateVoiceKit :execrows
UPDATE voice_kits
SET active = FALSE,
    updated_at = NOW()
WHERE tenant_id = $1
  AND id = $2
  AND active = TRUE;
