-- name: GetTenantBySlug :one
SELECT id, slug, display_name, active, created_at, updated_at
FROM tenants
WHERE slug = $1
  AND active = TRUE;

-- name: GetTenantByCustomDomain :one
SELECT t.id, t.slug, t.display_name, t.active, t.created_at, t.updated_at
FROM tenants AS t
INNER JOIN tenant_configs AS tc ON tc.tenant_id = t.id
WHERE tc.custom_domain = $1
  AND t.active = TRUE;
