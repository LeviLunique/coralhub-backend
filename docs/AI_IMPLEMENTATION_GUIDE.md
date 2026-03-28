# AI Implementation Guide - CoralHub Backend

## 1. Purpose of This Document

This document is the canonical implementation guide for an AI agent building the backend for **CoralHub**.

CoralHub is a multi-tenant choir management platform.

The initial tenant is:

- `Coral Jovem Asa Norte`

The implementation must support future tenants without backend duplication and without hardcoding the first tenant as the product identity.

It defines the architecture, stack, constraints, implementation order, and engineering rules that must be followed.

This guide intentionally does **not** adopt Clean Architecture purism, Hexagonal Architecture as dogma, or DDD-heavy modeling everywhere.

The target is a backend that is:

- production-credible
- simple enough to evolve
- cheap to operate
- efficient in memory and CPU
- fast to onboard
- testable without excessive ceremony
- robust where the domain actually needs robustness
- multi-tenant from the start
- tenant-safe in persistence and authorization

If there is tension between elegance and delivery, prefer the simplest design that preserves correctness, clarity, and operational safety.

---

## 2. Final Recommended Architecture

### 2.1. Chosen Stack

- Language: `Go`
- HTTP router: `chi`
- Database: `PostgreSQL`
- DB access: `pgx + sqlc`
- Object storage: `S3-compatible storage`
- Local storage emulation: `MinIO`
- Push notifications: `FCM`
- Background processing: PostgreSQL-backed worker with polling and `FOR UPDATE SKIP LOCKED`
- Packaging: Docker
- Local orchestration: Docker Compose
- CI: GitHub Actions

### 2.2. Architectural Style

Use a **pragmatic modular monolith**.

This means:

- organize the code by business module
- keep infrastructure and transport separate
- use interfaces only at real external boundaries
- avoid artificial ports for every CRUD operation
- avoid a controller/service/repository dogma
- avoid excessive indirection and mapping layers

### 2.3. What This Project Is Not

Do **not** implement this as:

- Clean Architecture purist with excessive rings
- Hexagonal Architecture with ports for everything
- one use case file per trivial CRUD action
- abstract architecture built for hypothetical future replacements
- event sourcing
- CQRS split into separate deployables
- microservices

This system is a modular monolith with one API process and one worker process, both in the same repository and sharing the same modules.

---

## 3. Core Architectural Principles

The AI must follow these rules:

1. Prefer simple modules over deep architecture layers.
2. Business rules belong in module services and domain types, not in HTTP handlers.
3. SQL is explicit and first-class.
4. Do not use an ORM.
5. Do not introduce generic repositories.
6. Do not create interfaces unless there is a real need for substitution, testing, or external integration.
7. Keep write flows transactional.
8. Keep read flows simple and query-oriented.
9. Treat uploads and notifications as operational concerns with domain implications, not as abstract architecture exercises.
10. Optimize for maintainability by a serious small-to-medium engineering team.
11. Enforce tenant isolation in persistence, authorization, and business flows.

---

## 4. Product and Tenant Model

CoralHub is the platform name.

The backend repository should be generic and reusable:

- repository: `coralhub-backend`
- module path: `github.com/<your-user-or-org>/coralhub-backend`

The product starts with one tenant:

- tenant slug: `coral-jovem-asa-norte`
- tenant display name: `Coral Jovem Asa Norte`

### 4.1. Tenant Rules

The system is multi-tenant from day one.

This means:

- every church or organization is a `tenant`
- tenant-owned business data must carry `tenant_id`
- authorization must be tenant-aware
- files, notifications, and audit history must remain tenant-isolated
- tenant branding must be configurable without branching the backend

### 4.2. Tenant Branding Rules

The backend is shared.
The user-facing brand is tenant-specific.

For example:

- platform: `CoralHub`
- initial tenant shown to end users: `Coral Jovem Asa Norte`

The backend should support a tenant configuration model for:

- display name
- logo URL
- color palette
- custom domain
- optional feature flags

---

## 5. Domain Interpretation

The domain is moderately complex, not highly complex.

The complexity is concentrated in:

- tenant isolation and tenant-aware authorization
- event scheduling
- notification generation and reprocessing
- memberships
- file lifecycle

The rest is mostly transactional CRUD with validation.

### 4.1. Richer Modeling Required

Model with more care:

- `events`
- `notifications`
- `memberships`
- `files`

### 4.2. Simpler Transactional Modeling Is Fine

Keep simpler:

- `tenants`
- `choirs`
- `users`
- `voicekits`
- `devices`

Use value objects only where they carry actual validation or prevent bugs:

- `Email`
- `Platform`
- `ReminderType`
- `NotificationStatus`
- `EventType`
- `FileType`

Do not over-model everything into tiny object taxonomies.

---

## 6. Project Structure

Use this directory structure:

```text
cmd/
  api/
    main.go
  worker/
    main.go

internal/
  platform/
    config/
    log/
    observability/
    http/
      router.go
      middleware.go
      errors.go

  modules/
    tenants/
      model.go
      service.go
      repository.go
    choirs/
      model.go
      service.go
      repository.go
      http.go
    users/
      model.go
      service.go
      repository.go
    memberships/
      model.go
      service.go
      repository.go
    voicekits/
      model.go
      service.go
      repository.go
      http.go
    files/
      model.go
      service.go
      repository.go
      storage.go
      http.go
    events/
      model.go
      service.go
      repository.go
      policy.go
      http.go
    devices/
      model.go
      service.go
      repository.go
      http.go
    notifications/
      model.go
      service.go
      repository.go
      worker.go

  interfaces/
    http/
      dto/
      handlers/

  store/
    postgres/
      db.go
      tx.go
      mapping.go
      sqlc/

  integrations/
    storage/
      s3/
    push/
      fcm/

db/
  migrations/
  queries/

build/
  docker/

docs/
  adr/
```

### 6.1. Structure Rules

- `modules/` is the heart of the codebase.
- Each module owns its models, service logic, and persistence contract.
- `interfaces/http` contains HTTP transport-specific DTOs and handlers.
- `store/postgres` contains concrete SQL and DB wiring.
- `integrations/` contains external service adapters.
- `platform/` contains cross-cutting operational concerns.

Do not create top-level folders like:

```text
internal/domain/
internal/application/
internal/adapters/
internal/infrastructure/
```

That structure is intentionally rejected for this project.

---

## 7. Module Design Rules

Each module should follow this pattern:

- `model.go`: domain structs and validation helpers
- `service.go`: application logic for commands and orchestration
- `repository.go`: repository contract used by the service
- `http.go`: module route registration if useful

Optional files:

- `policy.go`: only when there are real decision rules
- `errors.go`: only when module-specific errors are truly needed
- `queries.go`: if module-specific query DTOs become large

### 7.1. Service Responsibilities

Services may:

- validate commands
- enforce business rules
- orchestrate writes
- open DB transactions
- call storage or push integrations through focused interfaces
- resolve and enforce tenant scope when required

Services must not:

- depend on `chi` request types
- depend on generated `sqlc` structs
- parse multipart directly
- build HTTP responses

### 7.2. Repository Responsibilities

Repositories should be small and real.

Good examples:

- `CreateEvent`
- `UpdateEvent`
- `ListChoirEvents`
- `ListPendingNotifications`
- `DeactivateDeviceToken`

Bad examples:

- `GenericRepository[T]`
- `Save(entity any)`
- `FindByCriteria(map[string]any)`

---

## 8. Multi-Tenant Strategy

### 8.1. Tenancy Model

The backend is a single shared multi-tenant system.

Use one shared application and one shared database, with strict row-level tenant scoping enforced in application code and schema design.

Do not create:

- one backend deployment per church by default
- one schema per tenant by default
- one database per tenant by default

### 8.2. Tenant Resolution

The system must resolve tenant context explicitly.

Preferred sources, in order of trust:

1. authenticated tenant claim
2. host or subdomain mapping when relevant
3. explicit tenant slug only for safe public bootstrap flows

Do not trust arbitrary client-supplied tenant IDs for privileged operations.

### 8.3. Initial Seed Tenant

The first seeded tenant should be:

- slug: `coral-jovem-asa-norte`
- display name: `Coral Jovem Asa Norte`

This is seed data, not hardcoded product identity.

---

## 9. HTTP Design

### 7.1. Router

Use `chi`.

Keep HTTP slim:

- route registration
- auth middleware
- request parsing
- external validation
- error translation
- response serialization

### 7.2. API Style

Use REST endpoints under `/api/v1`.

Recommended initial endpoints:

#### Choirs

- `POST /api/v1/choirs`
- `PUT /api/v1/choirs/{choirId}`
- `GET /api/v1/choirs/{choirId}`
- `GET /api/v1/choirs`
- `DELETE /api/v1/choirs/{choirId}`

#### Voice Kits

- `POST /api/v1/choirs/{choirId}/voice-kits`
- `PUT /api/v1/voice-kits/{kitId}`
- `GET /api/v1/voice-kits/{kitId}`
- `GET /api/v1/choirs/{choirId}/voice-kits`
- `DELETE /api/v1/voice-kits/{kitId}`

#### Files

- `POST /api/v1/voice-kits/{kitId}/files`
- `GET /api/v1/voice-kits/{kitId}/files`
- `DELETE /api/v1/files/{fileId}`
- `GET /api/v1/files/{fileId}/download-url`

#### Events

- `POST /api/v1/choirs/{choirId}/events`
- `PUT /api/v1/events/{eventId}`
- `GET /api/v1/events/{eventId}`
- `GET /api/v1/choirs/{choirId}/events`
- `DELETE /api/v1/events/{eventId}`

#### Devices

- `POST /api/v1/devices`
- `DELETE /api/v1/devices/{deviceId}`
- `GET /api/v1/users/{userId}/devices`

#### Notifications

- `GET /api/v1/users/{userId}/notifications`

Do not expose a public endpoint that behaves like a production scheduler.

If a manual trigger is ever added, it must be clearly internal, authenticated, and optional.

### 7.3. HTTP Error Mapping

Translate internal errors into:

- `400` for invalid input
- `404` for not found
- `409` for conflict
- `422` for business rule violations when appropriate
- `500` for internal failures
- `503` for temporary downstream or timeout conditions

---

## 10. Persistence Strategy

### 8.1. Database Choice

Use PostgreSQL.

### 8.2. Access Strategy

Use:

- `pgx` for connection handling
- `sqlc` for generated typed query access
- handwritten SQL in `db/queries`

Do not use:

- GORM
- ent
- query builder as the primary abstraction
- repository generators

### 8.3. Write Model

Use transactional write services.

Examples:

- create event and schedule notifications atomically
- update event and regenerate reminders atomically
- cancel event and cancel pending notifications atomically

### 8.4. Read Model

Keep reads straightforward.

Do not force every read through an aggregate abstraction if a query DTO is enough.

### 8.5. Suggested Tables

Create at least:

- `tenants`
- `tenant_configs`
- `choirs`
- `users`
- `choir_members`
- `voice_kits`
- `kit_files`
- `events`
- `device_tokens`
- `scheduled_notifications`
- `audit_log`

Tenant-owned tables should include `tenant_id` unless there is a strong reason not to.

### 8.6. Important Constraints

Add from the start:

- unique `tenants.slug`
- unique `users.email`
- unique `choir_members (choir_id, user_id)` among active rows if soft delete is used
- uniqueness rules for voice kit names inside a choir if required by business
- unique or equivalently deduplicated device token strategy
- partial unique index on active scheduled notification identity

Example notification uniqueness idea:

- `tenant_id + user_id + event_id + reminder_type + active/pending-state`

### 8.7. Important Indexes

Add early:

- `choirs (tenant_id, name)`
- `users (tenant_id, email)` if users are tenant-scoped
- `events (choir_id, start_at)`
- `voice_kits (choir_id)`
- `kit_files (voice_kit_id, active)`
- `device_tokens (user_id, active)`
- `scheduled_notifications (status, scheduled_for)`
- `scheduled_notifications (event_id)`

### 8.8. Soft Delete Guidance

Do not apply soft delete everywhere by reflex.

Use soft delete where history matters:

- events
- files
- memberships
- device tokens

Avoid unnecessary soft delete for everything else unless product requirements demand it.

---

## 11. File Storage Strategy

### 9.1. Decision

Store binary payloads in S3-compatible storage.
Store metadata in PostgreSQL.

### 9.2. Rules

- do not store file blobs in PostgreSQL
- stream uploads to storage
- do not load entire files into memory
- persist metadata only after storage succeeds
- handle cleanup for partial failure cases
- include tenant context in object key design

### 9.3. Download Strategy

Return pre-signed URLs instead of proxying file downloads through the API unless there is a security reason not to.

### 9.4. File Validation

Validate:

- size limits
- allowed content types
- file type category
- ownership relation to voice kit

Do not let multipart or storage SDK types leak into module models.

---

## 12. Notification Strategy

### 10.1. Decision

Use a PostgreSQL-backed durable queue in `scheduled_notifications`.

This is correct for MVP and solid V1.

### 10.2. Processing Model

Worker flow:

1. poll pending jobs due for execution
2. lock rows using `FOR UPDATE SKIP LOCKED`
3. send push notification through FCM
4. mark as sent, failed, or invalid token
5. increment attempts and apply retry/backoff policy for temporary failures

### 10.3. Rules

- sending push must not happen inline during event creation
- event creation schedules notifications
- worker sends notifications later
- worker must be idempotent
- duplicate sends must be prevented by state + constraints
- notification processing must stay inside tenant boundaries

### 10.4. Retry Policy

Implement a simple retry policy:

- retry transient failures
- stop retrying after configured max attempts
- mark token inactive on invalid token response
- record last error

Do not introduce Kafka, RabbitMQ, SQS, or distributed schedulers initially.

---

## 13. History and Audit Strategy

Implement **selective audit history**, not event sourcing.

### 11.1. Store History For

- event creation/update/cancelation
- membership add/remove
- file upload/remove
- notification generated/sent/failed

### 11.2. Avoid

- full historical reconstruction for all entities
- generic framework-first audit systems
- field-by-field deep object diffing everywhere

### 11.3. Suggested Audit Table

`audit_log`:

- `id`
- `tenant_id`
- `entity_type`
- `entity_id`
- `action`
- `actor_id`
- `occurred_at`
- `payload_json`

The payload should be pragmatic and useful for support and tracing, not a dumping ground.

---

## 14. Auth and Authorization

This must not be ignored.

Even if the full auth mechanism is not implemented first, the architecture must leave room for:

- authenticated users
- role-aware access
- choir-scoped authorization
- tenant-scoped authorization
- actor identity flowing into services where relevant

The AI must not design the system as if all endpoints are anonymous forever.

At minimum, define:

- request context tenant identity
- request context user identity
- role model
- choir membership checks for protected operations

---

## 15. Observability

Observability is mandatory from the beginning.

### 13.1. Logging

Use structured logs.

Include:

- request ID
- tenant ID or tenant slug where safe
- route
- method
- status code
- latency
- error class
- worker job identifiers where relevant

### 13.2. Metrics

Track at least:

- HTTP request count
- HTTP latency
- DB query latency
- worker poll count
- notification send success/failure counts
- storage upload failure count

### 13.3. Tracing

Adopt OpenTelemetry-compatible instrumentation if feasible.

Tracing does not need to be perfect on day one, but the app should not block later instrumentation.

---

## 16. Configuration Strategy

Use environment variables only.

Required deliverables:

- `.env.example`
- validated config loader
- sensible defaults for local development where safe
- seed configuration for the initial tenant where needed

Config groups:

- app
- http
- database
- storage
- firebase
- worker
- observability

Fail fast on missing required config.

---

## 17. Docker and Local Development

### 15.1. Local Development Model

Use Docker Compose for dependencies:

- PostgreSQL
- MinIO

API and worker may run either:

- inside containers
- locally from the developer machine

Do not force every local dev loop through a full rebuild cycle inside containers.

### 15.2. Dockerfile

Use multi-stage build:

- build stage with Go toolchain
- small runtime stage

### 15.3. Compose Deliverables

Provide:

- `docker-compose.yml`
- optional `docker-compose.override.yml`
- bucket bootstrap support if needed

---

## 18. Testing Strategy

Use a practical test pyramid.

### 16.1. Unit Tests

Focus on:

- tenant resolution and tenant isolation rules
- event validation and invariants
- reminder generation rules
- file validation rules
- notification retry and invalid token handling
- module service logic with small fakes where useful

### 16.2. Integration Tests

Cover:

- PostgreSQL persistence
- tenant-scoped queries
- transaction boundaries
- notification selection with locking
- S3/MinIO upload and pre-signed URLs

### 16.3. HTTP Tests

Cover only key flows:

- tenant bootstrap or tenant-config retrieval if implemented
- create choir
- list choirs
- create voice kit
- upload file
- create event
- list user notifications

### 16.4. Avoid

- testing every trivial mapper
- mocking everything
- building a giant fake infrastructure layer
- duplicate tests across all layers for the same behavior

---

## 19. CI/CD Strategy

### 17.1. CI on Pull Requests

Run:

1. checkout
2. setup Go
3. cache modules
4. `go mod tidy` verification
5. formatting verification
6. `go vet`
7. unit tests
8. integration tests
9. `sqlc generate` + dirty diff check
10. binary build
11. Docker build

### 17.2. Release Flow

On `main` or tags:

- build versioned image
- push to container registry
- keep deploy target decoupled until infrastructure is chosen

Do not over-design CD before the hosting platform is defined.

---

## 20. Implementation Roadmap

This project must be implemented in **vertical slices**, not as architecture-first theater.

### Stage 0. Repository Bootstrap

Deliver:

- `go.mod`
- `.gitignore`
- `.editorconfig`
- `README.md`
- `Makefile`
- GitHub Actions placeholder
- `docs/adr/0001-architecture.md`

Definition of done:

- repository is ready for disciplined implementation

### Stage 1. Platform Bootstrap

Deliver:

- config loader
- logger
- HTTP server bootstrap
- health endpoint
- DB connection bootstrap
- Dockerfile
- local compose for PostgreSQL and MinIO

Definition of done:

- API process starts
- worker process starts
- health endpoint works

### Stage 2. First Vertical Slice: Choirs and Users

Deliver:

- `tenants` and `tenant_configs` schema
- DB migrations for `choirs` and `users`
- `sqlc` queries
- tenant module
- choir and user modules
- basic HTTP endpoints
- integration tests
- initial seed for `Coral Jovem Asa Norte`

Definition of done:

- first real feature works end to end with tenant context

This stage is critical because it validates the project structure.

### Stage 3. Memberships and Authorization Baseline

Deliver:

- `choir_members` table
- membership module
- role checks or authorization stubs wired into request context
- membership-aware operations

Definition of done:

- access model is no longer ignored
- tenant context is no longer optional

### Stage 4. Voice Kits and File Metadata

Deliver:

- `voice_kits` and `kit_files` schema
- voice kit module
- file metadata module logic
- list/create/delete flows without storage integration finalized yet

Definition of done:

- metadata flows compile and persist correctly

### Stage 5. S3/MinIO File Upload Integration

Deliver:

- S3 storage adapter
- streaming upload
- pre-signed download URL generation
- file validation rules
- integration tests with MinIO

Definition of done:

- upload and download URL flow works end to end

### Stage 6. Events and Reminder Scheduling

Deliver:

- event schema
- scheduled notification schema
- event module
- reminder policy rules
- create/update/cancel event flows
- atomic scheduling logic

Definition of done:

- event changes generate or cancel reminders correctly

This is the most important domain stage.

### Stage 7. Notification Worker

Deliver:

- notification repository queries for due jobs
- worker polling loop
- locking with `FOR UPDATE SKIP LOCKED`
- retry policy
- invalid token behavior

Definition of done:

- pending reminders are processed safely and idempotently

### Stage 8. FCM Integration

Deliver:

- FCM adapter
- push provider interface
- error classification for transient vs invalid token
- worker integration

Definition of done:

- push delivery path is functional

### Stage 9. Audit History

Deliver:

- `audit_log` schema
- history writes for key actions
- basic retrieval if product needs it

Definition of done:

- support and traceability are possible for critical flows

### Stage 10. Hardening

Deliver:

- pagination
- request validation hardening
- timeout handling
- metrics
- structured error responses
- cleanup and retention rules for notifications

Definition of done:

- system is operationally credible

### Stage 11. CI Finalization and Documentation

Deliver:

- complete CI
- `.env.example`
- local run instructions
- ADR set
- architecture summary

Definition of done:

- another engineer can run, understand, and extend the system

---

## 21. Coding Rules for the AI

The AI must obey these implementation rules:

1. Write code in English.
2. Keep package names short and idiomatic.
3. Prefer explicitness over framework magic.
4. Prefer concrete types over unnecessary interfaces.
5. Use interfaces for:
   - repositories used by services
   - storage integration
   - push provider integration
   - time and ID generation only if they materially help tests
6. Do not create interfaces for handlers, routers, or trivial services.
7. Keep DTOs out of module models.
8. Keep generated `sqlc` structs inside persistence mapping code.
9. Keep business logic out of SQL except for filtering, locking, and data access concerns.
10. Keep transactions explicit and narrow.
11. Never forget tenant scope in queries, writes, or authorization-sensitive operations.

---

## 22. Things the AI Must Explicitly Avoid

Do not introduce:

- Echo
- Fiber
- GORM
- generic service layers
- generic repositories
- mediator frameworks
- a base model for all entities
- a shared `utils` dump package
- an internal event bus unless a real need appears
- Clean Architecture package theater
- one package per layer with no module ownership
- tenant-blind queries on tenant-owned data

Do not hide complexity under abstractions.
Remove complexity instead.

---

## 23. Minimum ADRs to Create

The AI should create these ADRs:

1. `0001-architecture.md` for application architecture
2. `0002-infrastructure.md` for production infrastructure
3. `0003-multi-tenant-data-model.md` for multi-tenant schema and tenant ownership rules
4. `0004-auth-and-tenant-resolution.md` for auth, tenant context, and authorization rules
5. language and HTTP stack if later split into a dedicated ADR
6. persistence strategy if later split into a dedicated ADR
7. file storage strategy if later split into a dedicated ADR
8. notification scheduling strategy if later split into a dedicated ADR
9. audit history strategy if later split into a dedicated ADR
10. observability baseline if later split into a dedicated ADR

---

## 24. Final Instruction to the Implementing AI

Build this system as a serious, pragmatic Go backend for the CoralHub platform.

Before implementing any non-trivial feature, the AI must read and follow:

- `docs/adr/0001-architecture.md`
- `docs/adr/0002-infrastructure.md`
- `docs/adr/0003-multi-tenant-data-model.md`
- `docs/adr/0004-auth-and-tenant-resolution.md`

If implementation pressure conflicts with those ADRs, the AI should not silently diverge.
It should either follow them or explicitly justify the deviation.

The implementation should look like it was designed by engineers who care about:

- operational simplicity
- correctness
- cost
- testability
- long-term maintainability

It should **not** look like a demonstration of architectural ideology.

When in doubt:

- prefer fewer layers
- prefer explicit SQL
- prefer vertical slices
- prefer real constraints in PostgreSQL
- prefer small, concrete module services
- prefer operational clarity over theoretical purity
- prefer tenant-safe defaults over convenience shortcuts

This project succeeds if another experienced engineer can open the repository and think:

"This is disciplined, clear, and practical."

Not:

"This is architecturally impressive but unnecessarily complex."
