# ADR 0003 - Multi-Tenant Data Model

## Status

Accepted

## Date

2026-03-27

## Context

CoralHub is a multi-tenant platform.

The first tenant is:

- `Coral Jovem Asa Norte`

The backend must support future churches and organizations without creating:

- one backend per tenant
- one schema per tenant by default
- one database per tenant by default

The data model must:

- isolate tenant-owned data correctly
- remain operationally simple
- support tenant-specific branding
- preserve room for future growth
- avoid premature infrastructure complexity

The key risk is tenant leakage caused by weak schema discipline or tenant-blind queries.

## Decision

CoralHub will use a **shared-database, shared-schema multi-tenant model** with explicit tenant ownership in application data.

### Core model

Introduce:

- `tenants`
- `tenant_configs`

The `tenants` table represents a church or organization using the platform.

The `tenant_configs` table stores tenant-specific runtime branding and configuration.

### Tenant ownership

Tenant-owned business tables must include `tenant_id` unless there is a strong and explicit reason not to.

Examples of tenant-owned tables:

- `choirs`
- `users` if user identity is tenant-scoped in the product model
- `choir_members`
- `voice_kits`
- `kit_files`
- `events`
- `device_tokens`
- `scheduled_notifications`
- `audit_log`

### Suggested tenant tables

#### `tenants`

Suggested fields:

- `id`
- `slug`
- `display_name`
- `active`
- `created_at`
- `updated_at`

Constraints:

- unique `slug`

#### `tenant_configs`

Suggested fields:

- `tenant_id`
- `logo_url`
- `primary_color`
- `secondary_color`
- `custom_domain`
- `created_at`
- `updated_at`

Constraints:

- unique `tenant_id`
- unique `custom_domain` when present

### Initial tenant seed

Seed the initial tenant as data:

- slug: `coral-jovem-asa-norte`
- display name: `Coral Jovem Asa Norte`

This is seed data, not a hardcoded product branch.

### Query discipline

All tenant-owned reads and writes must be tenant-scoped.

This means:

- inserts carry `tenant_id`
- updates filter by entity identity and `tenant_id`
- deletes or deactivations filter by entity identity and `tenant_id`
- reads filter by `tenant_id`

Never rely only on application memory assumptions for tenant isolation.

### Uniqueness discipline

Where business uniqueness is tenant-local, constraints must include `tenant_id`.

Examples:

- choir names when required by tenant scope
- user email if users are tenant-local
- notification uniqueness

Example notification identity:

- `tenant_id + user_id + event_id + reminder_type + active/pending-state`

### File storage alignment

Storage object keys must include tenant context.

Example shape:

```text
{environment}/tenants/{tenant_slug}/choirs/{choir_id}/voice-kits/{kit_id}/files/{file_id}/{stored_filename}
```

### Audit alignment

Audit history must carry `tenant_id`.

This ensures support and operational review can stay tenant-safe.

## Consequences

### Positive consequences

- simple operational model
- one backend platform for many churches
- lower infrastructure complexity
- predictable schema design
- clear tenant-aware query discipline
- easier support for tenant branding and custom domains

### Negative consequences

- every tenant-owned query must be carefully written
- mistakes in query filtering become security risks
- some constraints become wider because they include tenant context

These trade-offs are accepted because they are still much cheaper and safer than per-tenant infrastructure at this stage.

## Alternatives Considered

### Separate database per tenant

Rejected as the default model.

Reason:

- too much operational complexity
- backup, migration, and deployment overhead grow too quickly

### Separate schema per tenant

Rejected as the default model.

Reason:

- more complexity than value for the current platform stage
- harder operational tooling and migration management

### Single-tenant first, multi-tenant later

Rejected.

Reason:

- likely to create painful redesign later
- tenant isolation affects schema, auth, routes, tests, and branding

### Row-level security in PostgreSQL as the primary isolation mechanism

Rejected as the primary approach for now.

Reason:

- it can be valuable later
- it adds another layer of complexity and operational coupling too early

The current choice is application-enforced tenant scoping with explicit schema discipline.

## Implementation Notes

Implementation must:

- include `tenant_id` in tenant-owned tables early
- include tenant-aware indexes
- include tenant-aware unique constraints
- include tenant-aware test coverage
- seed `Coral Jovem Asa Norte` as the first tenant

Avoid:

- generic multi-tenant frameworks
- magical tenant context propagation
- hidden global tenant state

## Related Documents

- [0001-architecture.md](/Users/levilunique/Workspace/Go/coral/docs/adr/0001-architecture.md)
- [0004-auth-and-tenant-resolution.md](/Users/levilunique/Workspace/Go/coral/docs/adr/0004-auth-and-tenant-resolution.md)
- [AI_IMPLEMENTATION_GUIDE.md](/Users/levilunique/Workspace/Go/coral/docs/AI_IMPLEMENTATION_GUIDE.md)
