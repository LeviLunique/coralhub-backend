# CoralHub Documentation Index

## 1. Purpose of This Index

This file is the central entry point for the CoralHub backend documentation.

Use it to understand:

- what the project is
- which documents are authoritative
- what order they should be read in
- which documents are primarily for humans
- which documents Codex must follow before implementing

CoralHub is:

- a multi-tenant platform
- initially launched for `Coral Jovem Asa Norte`
- designed to support future churches without backend duplication

The backend repository is:

- `coralhub-backend`

---

## 2. Source of Truth

The main source of truth is the repository documentation itself.

Primary decision documents:

1. [AI_IMPLEMENTATION_GUIDE.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/AI_IMPLEMENTATION_GUIDE.md)
2. [INFRASTRUCTURE_BLUEPRINT.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/INFRASTRUCTURE_BLUEPRINT.md)
3. [0001-architecture.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/adr/0001-architecture.md)
4. [0002-infrastructure.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/adr/0002-infrastructure.md)
5. [0003-multi-tenant-data-model.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/adr/0003-multi-tenant-data-model.md)
6. [0004-auth-and-tenant-resolution.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/adr/0004-auth-and-tenant-resolution.md)

Codex support documents:

7. [CODEX_PROMPT_BASE.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/CODEX_PROMPT_BASE.md)
8. [coral-backend skill](/Users/levilunique/Workspace/Go/coralhub-backend/.codex/skills/coral-backend/SKILL.md)

Architecture summary:

9. [ARCHITECTURE_SUMMARY.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/ARCHITECTURE_SUMMARY.md)

Repository bootstrap guide:

10. [REPOSITORY_BOOTSTRAP_CORALHUB.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/REPOSITORY_BOOTSTRAP_CORALHUB.md)

Feature documentation entrypoint:

11. [IMPLEMENTATION_ORDER.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/features/IMPLEMENTATION_ORDER.md)

If two documents ever conflict:

- ADRs and implementation/infrastructure guides win over convenience documents
- repository documents win over ad hoc prompts
- explicit ADR decisions win over vague summaries

---

## 3. Recommended Reading Order for You

If you are the project owner or lead, read in this order:

1. [0001-architecture.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/adr/0001-architecture.md)
   This gives the application architecture decision.
2. [0003-multi-tenant-data-model.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/adr/0003-multi-tenant-data-model.md)
   This defines tenancy and schema ownership rules.
3. [0004-auth-and-tenant-resolution.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/adr/0004-auth-and-tenant-resolution.md)
   This defines auth and tenant-safety rules.
4. [0002-infrastructure.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/adr/0002-infrastructure.md)
   This defines production platform choices.
5. [AI_IMPLEMENTATION_GUIDE.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/AI_IMPLEMENTATION_GUIDE.md)
   This explains how the code should be built.
6. [INFRASTRUCTURE_BLUEPRINT.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/INFRASTRUCTURE_BLUEPRINT.md)
   This explains how production should be run.
7. [REPOSITORY_BOOTSTRAP_CORALHUB.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/REPOSITORY_BOOTSTRAP_CORALHUB.md)
   This is the operational bootstrap guide for you.
8. [ARCHITECTURE_SUMMARY.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/ARCHITECTURE_SUMMARY.md)
   This is the compact architecture overview.
9. [CODEX_PROMPT_BASE.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/CODEX_PROMPT_BASE.md)
   This is the default prompt pattern for using Codex.
10. [IMPLEMENTATION_ORDER.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/features/IMPLEMENTATION_ORDER.md)
   This tracks completed feature slices and where to find their docs.

---

## 4. Required Reading Order for Codex

Before implementing any non-trivial change, Codex should read in this order:

1. [AI_IMPLEMENTATION_GUIDE.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/AI_IMPLEMENTATION_GUIDE.md)
2. [0001-architecture.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/adr/0001-architecture.md)
3. [0003-multi-tenant-data-model.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/adr/0003-multi-tenant-data-model.md)
4. [0004-auth-and-tenant-resolution.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/adr/0004-auth-and-tenant-resolution.md)
5. [0002-infrastructure.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/adr/0002-infrastructure.md)
6. [INFRASTRUCTURE_BLUEPRINT.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/INFRASTRUCTURE_BLUEPRINT.md)
7. [coral-backend skill](/Users/levilunique/Workspace/Go/coralhub-backend/.codex/skills/coral-backend/SKILL.md)
8. [CODEX_PROMPT_BASE.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/CODEX_PROMPT_BASE.md)
9. [ARCHITECTURE_SUMMARY.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/ARCHITECTURE_SUMMARY.md)
10. [IMPLEMENTATION_ORDER.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/features/IMPLEMENTATION_ORDER.md)

Codex must treat these as binding project context unless a new approved ADR changes the decision.

---

## 5. What Each Document Is For

### Architecture

- [ARCHITECTURE_SUMMARY.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/ARCHITECTURE_SUMMARY.md)
  Use for the short architecture overview.

- [0001-architecture.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/adr/0001-architecture.md)
  Use for application architecture direction.

- [AI_IMPLEMENTATION_GUIDE.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/AI_IMPLEMENTATION_GUIDE.md)
  Use for implementation workflow, coding rules, testing strategy, and roadmap.

### Multi-tenant and auth

- [0003-multi-tenant-data-model.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/adr/0003-multi-tenant-data-model.md)
  Use for data ownership, `tenant_id`, schema rules, and tenant-safe persistence.

- [0004-auth-and-tenant-resolution.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/adr/0004-auth-and-tenant-resolution.md)
  Use for auth context, tenant resolution, and authorization design.

### Infrastructure

- [0002-infrastructure.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/adr/0002-infrastructure.md)
  Use for the high-level production decision.

- [INFRASTRUCTURE_BLUEPRINT.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/INFRASTRUCTURE_BLUEPRINT.md)
  Use for concrete production infrastructure planning.

### Codex usage

- [CODEX_PROMPT_BASE.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/CODEX_PROMPT_BASE.md)
  Use as the default task prompt template.

- [coral-backend skill](/Users/levilunique/Workspace/Go/coralhub-backend/.codex/skills/coral-backend/SKILL.md)
  Use as the project skill for Codex.

### Human repository setup

- [REPOSITORY_BOOTSTRAP_CORALHUB.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/REPOSITORY_BOOTSTRAP_CORALHUB.md)
  Use for local and remote repository creation, machine setup, and GitHub governance.

- [LOCAL_ENV_SETUP.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/LOCAL_ENV_SETUP.md)
  Use for local environment configuration.

- [LOCAL_SMOKE_TEST.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/LOCAL_SMOKE_TEST.md)
  Use for local verification after setup.

### Feature documentation

- [IMPLEMENTATION_ORDER.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/features/IMPLEMENTATION_ORDER.md)
  Use as the entrypoint for implemented feature slices and branch/commit history.

Per-feature docs should follow this pattern when a feature is substantial enough to document:

- `docs/features/<feature>/README.md`
- `docs/features/<feature>/TESTING.md`

---

## 6. Recommended Working Flow

### For you

1. Read the ADRs first.
2. Confirm the business and platform decisions.
3. Use the repository bootstrap guide to create or finalize the repo.
4. Use the implementation guide and infrastructure blueprint as execution references.
5. Use the Codex prompt base when assigning tasks.
6. Use feature docs from `docs/features/` when you need implementation or testing guidance for a specific slice.

### For Codex

1. Read the implementation guide.
2. Read ADRs `0001` to `0004`.
3. Read the infrastructure blueprint if the task touches infrastructure, storage, deployment, auth integration, or observability.
4. Implement one vertical slice at a time.
5. Run relevant validations.
6. Do not silently diverge from the ADRs.
7. If relevant, read `docs/features/IMPLEMENTATION_ORDER.md` and the feature-specific docs for the slice being changed.

---

## 7. Current Platform Decisions Summary

Current approved direction:

- platform name: `CoralHub`
- public domain: `coralhub.com.br`
- backend repository: `coralhub-backend`
- architecture: pragmatic modular monolith
- language: `Go`
- HTTP: `chi`
- persistence: `PostgreSQL + pgx + sqlc`
- object storage: `S3`
- push: `FCM`
- runtime: `AWS + ECS Fargate + RDS + S3 + Secrets Manager + CloudWatch + OTel`
- tenancy: shared backend, shared database, explicit `tenant_id`
- initial tenant: `Coral Jovem Asa Norte`

---

## 8. Final Rule

When in doubt:

- check this index
- read the ADRs
- follow the implementation guide
- prefer explicit, tenant-safe, operationally simple decisions

If the documentation becomes inconsistent in the future, create a new ADR or update the existing ADRs first, then update the derived documents.
