# Codex Prompt Base

Use this prompt as the default starting point when asking Codex to work on this repository.

## Base Prompt

```md
Use $coral-backend for this task.

Before changing code, read these documents and treat them as the source of truth:

- docs/AI_IMPLEMENTATION_GUIDE.md
- docs/INFRASTRUCTURE_BLUEPRINT.md
- docs/adr/0001-architecture.md
- docs/adr/0002-infrastructure.md
- docs/adr/0003-multi-tenant-data-model.md
- docs/adr/0004-auth-and-tenant-resolution.md

Rules:
- use Go + chi + PostgreSQL + pgx + sqlc
- treat this as the CoralHub backend
- keep the backend multi-tenant from the start
- preserve `Coral Jovem Asa Norte` as the initial tenant seed, not as hardcoded product identity
- follow the pragmatic modular monolith architecture
- keep modules organized by business capability
- keep handlers thin
- keep SQL explicit
- do not use ORM
- do not use generic repositories
- do not introduce Clean Architecture purism or Hexagonal Architecture everywhere
- implement in vertical slices
- run relevant tests after changes
- if the task touches an existing feature slice, also read `docs/features/IMPLEMENTATION_ORDER.md` and the relevant `docs/features/<feature>/` docs when they exist
- if the task completes a meaningful feature slice, update the feature docs pattern when relevant:
  - `docs/features/<feature>/README.md`
  - `docs/features/<feature>/TESTING.md`
  - `docs/features/IMPLEMENTATION_ORDER.md`
- if you need to diverge from the documented plan, explain the reason clearly before or while making the change

Task:
<replace this with the concrete task>
```

## Recommended Usage Pattern

Good task examples:

- implement the bootstrap for the API and worker processes
- implement the choirs vertical slice with migration, sqlc queries, service, handler, and tests
- implement event creation with reminder scheduling and integration tests
- add S3 upload support for voice kit files following the project docs
- review this pull request against the ADRs and architecture guide

Bad task examples:

- build the whole backend
- create the entire architecture first
- add abstractions for future flexibility
- improve the architecture without pointing to a specific problem

## Stronger Prompt for Implementation Tasks

```md
Use $coral-backend for this task.

Read first:
- docs/AI_IMPLEMENTATION_GUIDE.md
- docs/INFRASTRUCTURE_BLUEPRINT.md
- docs/adr/0001-architecture.md
- docs/adr/0002-infrastructure.md
- docs/adr/0003-multi-tenant-data-model.md
- docs/adr/0004-auth-and-tenant-resolution.md

Implement only the requested slice.
Do not redesign the project.
Do not invent missing abstractions unless they solve an immediate problem in this slice.
Keep the change small, coherent, and testable.
If the task touches an existing feature, also read `docs/features/IMPLEMENTATION_ORDER.md` and the relevant `docs/features/<feature>/` docs when they exist.
If the task completes a meaningful feature slice, update the feature docs pattern when relevant.
Run the relevant tests at the end and summarize what passed.

Task:
<replace this with the concrete task>
```

## Stronger Prompt for Review Tasks

```md
Use $coral-backend for this review.

Review the change against:
- docs/AI_IMPLEMENTATION_GUIDE.md
- docs/INFRASTRUCTURE_BLUEPRINT.md
- docs/adr/0001-architecture.md
- docs/adr/0002-infrastructure.md
- docs/adr/0003-multi-tenant-data-model.md
- docs/adr/0004-auth-and-tenant-resolution.md

If the change touches an implemented slice, also review against:
- docs/features/IMPLEMENTATION_ORDER.md
- the relevant `docs/features/<feature>/` docs when they exist

Focus on:
- architectural violations
- overengineering
- missing tests
- persistence mistakes
- infra/operability regressions
- SQL, transaction, or worker correctness risks

Do not give a generic summary first.
List findings first, ordered by severity.
```
