# Auth and Tenant Resolution ADR

## Status

Accepted

## Decision

Protected requests in CoralHub Backend must be evaluated using both:

- user identity
- tenant context

Authorization is tenant-aware by design.

## Why This Decision Was Made

In a multi-tenant product, authentication alone is not enough.

The backend must answer three separate questions for protected operations:

1. Who is the user?
2. Which tenant is this request operating inside?
3. What is this user allowed to do in that tenant?

If any of those answers is ambiguous, cross-tenant exposure becomes much more likely.

## What This Means In Practice

Protected flows should work with explicit request context such as:

- `request_id`
- `tenant_id`
- `tenant_slug`
- `user_id`
- role or membership information when needed

Services should receive enough information to enforce authorization without depending on HTTP types.

## Tenant Resolution Rules

Preferred tenant resolution order:

1. authenticated tenant claim when available
2. trusted host or subdomain mapping
3. explicit tenant slug only for safe bootstrap or public flows

Client-supplied tenant identifiers in request bodies are not trustworthy for authorization-sensitive operations.

## Role and Membership Model

Authorization should be based on tenant-aware roles and memberships.

Examples:

- listing a tenant’s choirs requires access to that tenant
- managing a choir requires the correct role in that choir
- registering a device token requires both the user and tenant context

## Public vs Protected Flows

Public flows may use a tenant slug or domain as a routing input.

Examples:

- tenant bootstrap
- branding lookup

Protected flows must not rely on that public hint alone.

## What To Avoid

Do not add:

- hidden auth state
- silent tenant fallback behavior
- authorization decisions based only on public route input
- trust in arbitrary tenant IDs from clients

## Consequences

Positive outcomes:

- stronger tenant safety
- clearer request handling rules
- easier reasoning about authorization decisions

Trade-offs:

- more context needs to flow through the application
- middleware and tests need to stay disciplined

## Guidance For New Contributors

Before merging a protected endpoint, check:

1. How is tenant context resolved?
2. How is actor identity resolved?
3. Where is role or membership enforced?
4. What prevents a user from using one tenant’s identity against another tenant’s data?

If the answers are not explicit in the flow, the authorization design is incomplete.
