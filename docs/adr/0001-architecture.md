# ADR 0001 - CoralHub Application Architecture

## Status

Accepted

## Date

2026-03-27

## Context

The CoralHub backend must support:

- multi-tenant platform behavior
- tenant-specific branding
- choir management
- user and membership management
- voice kits and files
- events
- mobile device registration
- scheduled push notifications
- selective business audit history

The system must be:

- production-credible
- cheap to operate
- easy to maintain
- efficient in memory and CPU
- straightforward for a serious engineering team to understand

The initial tenant is:

- `Coral Jovem Asa Norte`

The backend must support future churches and organizations without redesign.

The initial proposal leaned toward Clean Architecture purism, Hexagonal Architecture everywhere, many ports, many use-case files, and a deeply layered structure.

That approach was rejected because the domain is only moderately complex and most of the system is transactional CRUD plus one meaningful asynchronous workflow.

The main risk was overengineering:

- too many abstractions
- too many mapping layers
- too much package ceremony
- poor cost-benefit for maintenance and onboarding

## Decision

The application will be built as a **pragmatic modular monolith in Go**.

### Chosen stack

- language: `Go`
- HTTP router: `chi`
- database: `PostgreSQL`
- DB access: `pgx + sqlc`
- object storage: `S3-compatible storage`
- push provider: `FCM`
- background processing: PostgreSQL-backed worker with polling and `FOR UPDATE SKIP LOCKED`

### Chosen architectural style

Use a **modular monolith** with:

- code organized by business module
- simple, explicit package boundaries
- services holding application logic
- domain types used where they add real validation or invariants
- interfaces only at real boundaries such as repositories, storage, and push integrations
- tenant-aware data and authorization boundaries

The architecture will **not** use:

- Clean Architecture purism
- Hexagonal Architecture as a governing dogma
- one use case type per trivial CRUD operation
- generic repositories
- ORM-driven persistence
- event sourcing
- microservices

### Module organization

The primary code organization is:

- `internal/modules` for business modules
- `internal/platform` for cross-cutting operational concerns
- `internal/interfaces/http` for transport-specific code
- `internal/store/postgres` for concrete persistence implementation
- `internal/integrations` for external service adapters

The backend repository name is:

- `coralhub-backend`

### Multi-tenant decision

The backend is multi-tenant from day one.

This means:

- a `tenant` model exists explicitly
- tenant-owned rows carry `tenant_id`
- tenant configuration exists for runtime branding and tenant metadata
- authorization always considers tenant scope
- the first tenant, `Coral Jovem Asa Norte`, is seed data rather than hardcoded product identity

### Domain modeling level

The system will use richer modeling only where justified:

- events
- notifications
- memberships
- files

The rest will remain simpler and more transactional:

- choirs
- users
- voice kits
- devices

### Persistence strategy

Persistence is SQL-first:

- explicit SQL in `db/queries`
- generated access via `sqlc`
- concrete PostgreSQL implementation using `pgx`

The system will avoid ORM abstractions and generic repository patterns.

### File strategy

Binary files are stored in object storage.
Metadata is stored in PostgreSQL.

Uploads must stream to storage.
Downloads should use presigned URLs.

### Notification strategy

Push notifications are asynchronous.

The system will:

- create scheduled notification records during event workflows
- process them in a worker
- retry transient failures
- deactivate invalid tokens

No external broker is introduced initially.

### History strategy

The system will implement **selective audit history**, not full event sourcing.

Critical business actions will be written to an `audit_log` style table.

## Consequences

### Positive consequences

- lower architectural ceremony
- faster delivery
- better onboarding for engineers familiar with Go
- simpler navigation
- clear operational model
- explicit SQL and predictable persistence behavior
- easier cost control
- safer long-term expansion to other churches without replatforming

### Negative consequences

- less theoretical replaceability of infrastructure
- less “architectural purity”
- some direct coupling to PostgreSQL and S3-compatible assumptions
- fewer abstraction seams than a purist layered model
- stricter care required in query design to avoid tenant leakage

These trade-offs are accepted because they improve the cost-benefit ratio for this product.

## Alternatives Considered

### Clean Architecture purist with ports and adapters everywhere

Rejected.

Reason:

- too much complexity for the actual domain
- too many files and interfaces
- poor ergonomics in Go for this project

### Hexagonal Architecture as the primary organizing principle

Rejected as the main structure.

Reason:

- useful as a boundary idea
- excessive as the dominant code organization strategy here

### ASP.NET Core instead of Go

Considered and not chosen.

Reason:

- ASP.NET Core is an excellent backend platform
- Go remains the better fit for this specific project because of runtime simplicity, footprint, and low operational overhead

### Java or Kotlin with Spring

Rejected.

Reason:

- strong enterprise platforms
- heavier stack than needed for this backend

### Node.js with TypeScript

Rejected.

Reason:

- strong productivity
- weaker cost-performance profile for this kind of backend

### Rust

Rejected.

Reason:

- technically strong
- unjustified development complexity for this domain

## Implementation Notes

The implementation must follow these rules:

- prefer vertical slices over layer-first delivery
- keep handlers thin
- keep business logic out of transport code
- keep transactions explicit
- keep interfaces focused and minimal
- avoid a generic `utils` package
- avoid package trees that exist only to imitate architecture diagrams
- treat tenant scope as a first-class invariant

## Related Documents

- [AI_IMPLEMENTATION_GUIDE.md](/Users/levilunique/Workspace/Go/coral/docs/AI_IMPLEMENTATION_GUIDE.md)
- [0002-infrastructure.md](/Users/levilunique/Workspace/Go/coral/docs/adr/0002-infrastructure.md)
