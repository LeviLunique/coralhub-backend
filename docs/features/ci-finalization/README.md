# CI Finalization and Documentation

This document explains the Stage 11 slice implemented on branch:

- `chore/ci-finalization`

Commit:

- `4358895`

## Goal

Close the roadmap with a practical CI baseline and clearer documentation entrypoints so another engineer can run, understand, and extend the backend with less repo archaeology.

This slice adds:

- stronger GitHub Actions CI coverage
- local `Makefile` targets that mirror the CI checks
- a short architecture summary document
- better README and docs index discoverability

## What Changed

### CI Coverage

The workflow in `.github/workflows/ci.yml` now runs:

- formatting verification
- `sqlc generate` and generated-code drift detection
- `go vet`
- `staticcheck`
- `golangci-lint`
- `govulncheck`
- `go test ./...`
- `go build ./cmd/api ./cmd/worker`
- `gitleaks`
- `Trivy` filesystem scanning

That is the Stage 11 baseline expected by the implementation and infrastructure docs.

### Local Parity Commands

The `Makefile` now exposes local commands for the same quality bar used in CI:

- `make fmt-check`
- `make staticcheck`
- `make lint`
- `make govulncheck`
- `make ci`

This keeps local validation and CI expectations close together.

### Documentation Entry Points

The repository now has faster onboarding entry points:

- `README.md` points clearly to setup, smoke-test, and architecture docs
- `docs/ARCHITECTURE_SUMMARY.md` gives the short system overview
- `docs/INDEX.md` now registers the architecture summary and the feature-docs area

## Repository Result

The repository no longer ignores the Stage 11 documentation files in Git.

That means the documentation, CI, and repo entrypoint updates from this slice can move together through normal review and commit flow.

## File-By-File Explanation

### [.github/workflows/ci.yml](/Users/levilunique/Workspace/Go/coralhub-backend/.github/workflows/ci.yml)

Expands CI from a minimal Go pipeline into a broader quality and security baseline.

It now separates:

- `quality` checks for format, generation, static analysis, tests, and build
- `security` checks for secrets scanning and filesystem vulnerability scanning

### [.golangci.yml](/Users/levilunique/Workspace/Go/coralhub-backend/.golangci.yml)

Adds a small explicit lint configuration so local linting and CI linting use the same baseline.

### [Makefile](/Users/levilunique/Workspace/Go/coralhub-backend/Makefile)

Adds local parity targets that make it easy to run the Stage 11 checks before pushing.

### [README.md](/Users/levilunique/Workspace/Go/coralhub-backend/README.md)

Improves the onboarding path by:

- pointing to the docs entrypoint
- pointing to the architecture summary
- surfacing local setup and smoke-test guides
- listing the new CI-oriented local commands

### [ARCHITECTURE_SUMMARY.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/ARCHITECTURE_SUMMARY.md)

Adds the short architecture overview for faster onboarding.

It is intentionally a summary, not a replacement for the ADRs.

### [INDEX.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/INDEX.md)

Registers the architecture summary and feature-docs structure so documentation discovery stays coherent.

### [IMPLEMENTATION_ORDER.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/features/IMPLEMENTATION_ORDER.md)

Records Stage 11 completion state and the remaining follow-up work after the roadmap is complete.

## What This Slice Does Not Do

This slice does not add:

- deployment workflows
- release automation
- automatic docs publishing

Those can be added later without changing the Stage 11 baseline delivered here.
