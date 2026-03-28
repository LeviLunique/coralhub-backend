# ADR 0004 - Auth and Tenant Resolution

## Status

Accepted

## Date

2026-03-27

## Context

CoralHub is a multi-tenant platform.

The backend must enforce:

- authentication
- tenant-aware authorization
- tenant-safe request processing

The initial tenant is:

- `Coral Jovem Asa Norte`

Future tenants may use:

- shared CoralHub domains
- tenant subdomains
- custom domains

The system therefore needs a clear rule for:

- who the user is
- which tenant the request belongs to
- how to authorize operations safely

The key risk is ambiguity:

- a user authenticated in one tenant accessing another tenant's data
- trusting client-supplied tenant IDs directly
- mixing routing identity and authorization identity incorrectly

## Decision

CoralHub will use **authenticated user identity plus explicit tenant context** for authorization.

### Core rule

Every protected request must be evaluated against:

- authenticated user identity
- current tenant identity
- user role or membership inside that tenant

Tenant context is not optional for tenant-owned operations.

### Tenant resolution order

Preferred resolution order:

1. authenticated tenant claim when the auth model supports it
2. trusted host or subdomain mapping
3. explicit tenant slug only for safe bootstrap or public flows

Do not trust arbitrary tenant IDs sent in request bodies for authorization-sensitive operations.

### Recommended request context

The request context should carry:

- `request_id`
- `tenant_id`
- `tenant_slug`
- `user_id`
- user role or relevant membership claims

Services should receive enough context to enforce tenant-aware authorization without depending on HTTP-specific types.

### Authorization model

Use a role-aware and tenant-aware authorization model.

At minimum support:

- tenant-bound user identity
- choir membership checks where required
- role checks for admin or management actions

Examples:

- listing a tenant's choirs requires access to that tenant
- creating an event requires access to the tenant and appropriate choir or management permissions
- registering a device token requires the authenticated user and the current tenant context

### Public vs protected flows

Public or semi-public flows may resolve tenant by:

- domain
- subdomain
- explicit slug

Examples:

- fetching tenant branding
- initial sign-in bootstrap page

Protected flows must not use public tenant hints as the sole authorization source.

### Initial implementation posture

Even if the full identity provider is not finalized yet, the backend must be designed so that:

- tenant context is explicit
- auth middleware can populate request context
- service methods can enforce tenant and role checks

The architecture must not assume anonymous access forever.

### Logging and observability

Where safe and useful, logs and traces should include:

- tenant ID or slug
- user ID
- request ID

This is necessary for support, incident response, and auditability.

## Consequences

### Positive consequences

- reduced risk of cross-tenant data exposure
- clear separation between routing hints and authorization truth
- explicit service-level authorization model
- strong base for branded domains and future tenant growth

### Negative consequences

- more context must flow through the application
- auth and tenant middleware must be designed carefully
- tests need to cover tenant resolution and authorization combinations

These costs are accepted because multi-tenant safety is a core platform requirement.

## Alternatives Considered

### Tenant selected only by client-provided request field

Rejected.

Reason:

- too easy to spoof
- unsafe for privileged operations

### Tenant inferred only from subdomain or custom domain

Rejected as the only source of truth.

Reason:

- useful as routing input
- not sufficient alone for all protected authorization decisions

### Authentication only, tenant inferred later from requested resource

Rejected as the main strategy.

Reason:

- too easy to create ambiguous or inconsistent authorization paths

### Anonymous-first product with auth added later

Rejected.

Reason:

- would push critical authorization design too far downstream

## Implementation Notes

Implementation should include:

- auth middleware boundary in HTTP
- explicit request context extraction
- tenant resolution component or helper
- service methods that accept actor and tenant context explicitly where needed
- tests for:
  - wrong-tenant access
  - missing-tenant access
  - role mismatch
  - domain/subdomain mapping behavior if implemented

Avoid:

- hidden global auth state
- tenant fallback logic that silently broadens access
- mixing branding resolution with authorization truth

## Related Documents

- [0001-architecture.md](/Users/levilunique/Workspace/Go/coral/docs/adr/0001-architecture.md)
- [0003-multi-tenant-data-model.md](/Users/levilunique/Workspace/Go/coral/docs/adr/0003-multi-tenant-data-model.md)
- [AI_IMPLEMENTATION_GUIDE.md](/Users/levilunique/Workspace/Go/coral/docs/AI_IMPLEMENTATION_GUIDE.md)
