# Multi-Tenant Data Model ADR

## Status

Accepted

## Decision

CoralHub Backend uses a shared-database, shared-schema multi-tenant model with explicit tenant ownership in application data.

The first tenant is seeded as data:

- slug: `coral-jovem-asa-norte`
- display name: `Coral Jovem Asa Norte`

That tenant is not the product identity.

## Why This Decision Was Made

The project needs multi-tenant safety from the start, but separate databases or schemas per tenant would add too much operational complexity for the current stage.

The simpler and intended model is:

- one backend
- one shared schema
- explicit tenant scoping in data access and business rules

## What This Means In Practice

Tenant-owned tables should include `tenant_id`.

That applies to areas such as:

- choirs
- users
- memberships
- voice kits
- files
- events
- device tokens
- scheduled notifications
- audit history

Queries must stay tenant-aware for:

- inserts
- updates
- deletes
- reads

## Schema Discipline

When a uniqueness rule is tenant-local, it must include `tenant_id`.

Examples:

- choir names
- tenant-scoped user emails
- notification uniqueness

This is one of the main ways the schema helps prevent cross-tenant leakage.

## Storage Alignment

Object storage keys must also include tenant context.

That keeps binary assets aligned with the same isolation model used in PostgreSQL.

## What To Avoid

Do not add:

- tenant-blind queries
- hidden global tenant state
- assumptions that a record can be looked up safely without tenant scope

Do not treat multi-tenancy as an afterthought.
It affects schema, queries, authorization, testing, and operations.

## Consequences

Positive outcomes:

- simpler operations than per-tenant infrastructure
- consistent tenant model across modules
- easier future tenant onboarding

Trade-offs:

- every tenant-owned query must be written carefully
- mistakes in tenant filtering become security bugs

## Guidance For New Contributors

Before shipping a new feature, verify:

1. Does every tenant-owned table include `tenant_id`?
2. Do all reads and writes filter by tenant correctly?
3. Are unique constraints tenant-aware where they should be?
4. Do tests cover wrong-tenant access or tenant leakage risks?

If any of these answers is no, the feature is not complete yet.
