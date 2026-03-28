# Application Architecture ADR

## Status

Accepted

## Decision

CoralHub Backend is built as a pragmatic modular monolith in Go.

The repository keeps one codebase and two processes:

- `api` for HTTP traffic
- `worker` for asynchronous notification processing

The code is organized by business capability, with explicit SQL and thin HTTP handlers.

## Why This Decision Was Made

The product has real domain rules, but it does not justify microservices or heavy architectural ceremony.

Most of the system is transactional application logic with a few areas that need extra care:

- tenant isolation
- authorization
- event and reminder scheduling
- file lifecycle
- notification processing

The project therefore chooses the simplest structure that still preserves correctness, testability, and operational credibility.

## What This Means In Practice

Developers should expect:

- modules under `internal/modules/`
- shared operational code under `internal/platform/`
- concrete PostgreSQL implementations under `internal/store/postgres/`
- explicit SQL in `db/queries/`
- generated query bindings from `sqlc`

Handlers should decode requests, call services, and translate results into HTTP responses.
Business rules belong in services and domain-facing module types.

## Preferred Engineering Style

The preferred style is:

- explicit over magical
- small focused interfaces only at real boundaries
- straightforward transactions for write flows
- simple query-oriented reads
- vertical slices that deliver complete behavior

## What To Avoid

Do not add:

- ORM-based persistence
- generic repositories
- abstraction layers that only wrap simple CRUD
- one package per architecture buzzword
- microservices without a concrete operational need

## Consequences

Positive outcomes:

- easier onboarding
- faster delivery
- predictable code navigation
- lower operational complexity

Trade-offs:

- less abstract replaceability
- more deliberate care required around query design and tenant safety

## Guidance For New Contributors

Before introducing a new abstraction, ask:

1. Does this reduce real complexity or just move it?
2. Is this solving a current problem in the codebase?
3. Would a developer new to the repository understand the change faster after it is introduced?

If the answer is no, prefer the simpler implementation.
