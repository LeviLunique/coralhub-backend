# Tenant Bootstrap

This guide explains the tenant bootstrap feature and how to verify it locally.

## What It Does

This feature exposes a public endpoint that returns basic tenant bootstrap information by slug.

Endpoint:

- `GET /api/v1/public/tenants/{tenantSlug}`

Example:

- `GET /api/v1/public/tenants/coral-jovem-asa-norte`

## How It Works

Flow:

1. the router receives the public tenant request
2. the handler extracts `tenantSlug`
3. the tenant service validates the slug
4. the PostgreSQL repository runs an explicit tenant lookup query
5. the API returns the slug, display name, and branding object

This is intentionally a safe public flow. It does not grant access to tenant-owned protected data.

## Why It Matters

This feature supports:

- branded bootstrap flows
- tenant discovery by slug
- future public branding lookup

## How To Verify

Start dependencies and the API:

```bash
make compose-up
make run-api
```

Call the endpoint:

```bash
curl -s http://127.0.0.1:8080/api/v1/public/tenants/coral-jovem-asa-norte
```

Expected result:

- HTTP `200`
- JSON payload with the seeded tenant slug and display name

Unknown tenant check:

```bash
curl -s -i http://127.0.0.1:8080/api/v1/public/tenants/tenant-that-does-not-exist
```

Expected result:

- HTTP `404`

Automated validation:

```bash
make test
make vet
make build
```
