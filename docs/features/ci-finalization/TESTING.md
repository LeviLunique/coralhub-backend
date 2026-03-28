# CI Finalization and Documentation Testing

This guide verifies the Stage 11 slice on branch:

- `chore/ci-finalization`

## 1. Run the New Local CI Entry Point

From the repo root:

```bash
make ci
```

Expected:

- formatting check passes
- `sqlc generate` produces no unexpected drift
- `go vet` passes
- `staticcheck` passes
- `golangci-lint` passes
- `govulncheck` completes without reporting reachable vulnerabilities that fail the command
- tests pass
- API and worker binaries build

## 2. Verify the Dedicated Commands Individually

If you want to isolate a failure, rerun the commands separately:

```bash
make fmt-check
```

```bash
make sqlc
git diff --exit-code
```

```bash
make vet
```

```bash
make staticcheck
```

```bash
make lint
```

```bash
make govulncheck
```

```bash
make test
```

```bash
make build
```

Expected:

- each command exits successfully

## 3. Review the CI Workflow Shape

Open:

- `.github/workflows/ci.yml`

Confirm the workflow includes:

- PR-triggered CI
- push to `main`
- quality checks
- security checks
- secret scanning
- filesystem vulnerability scanning

## 4. Check Documentation Entry Points

Open:

- `README.md`
- `docs/INDEX.md`
- `docs/ARCHITECTURE_SUMMARY.md`

Confirm:

- the README links to setup and smoke-test docs
- the docs index includes the architecture summary and feature-docs entrypoint
- the architecture summary gives a short but accurate system overview

## 5. Documentation Tracking Check

Confirm the Stage 11 documentation files appear in Git status and can be committed normally.
