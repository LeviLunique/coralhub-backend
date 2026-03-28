-- name: CreateDeviceToken :one
INSERT INTO device_tokens (tenant_id, user_id, platform, token)
VALUES ($1, $2, $3, $4)
ON CONFLICT (tenant_id, token)
DO UPDATE SET
	user_id = EXCLUDED.user_id,
	platform = EXCLUDED.platform,
	active = TRUE,
	updated_at = NOW()
RETURNING id, tenant_id, user_id, platform, token, active, created_at, updated_at;

-- name: ListActiveDeviceTokensByUserID :many
SELECT id, tenant_id, user_id, platform, token, active, created_at, updated_at
FROM device_tokens
WHERE tenant_id = $1
  AND user_id = $2
  AND active = TRUE
ORDER BY created_at ASC, id ASC;

-- name: DeactivateDeviceTokenByToken :execrows
UPDATE device_tokens
SET active = FALSE,
	updated_at = NOW()
WHERE tenant_id = $1
  AND token = $2
  AND active = TRUE;
