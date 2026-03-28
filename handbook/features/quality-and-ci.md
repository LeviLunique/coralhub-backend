# Quality and CI

This guide explains the current CI and quality baseline.

## What It Does

This slice adds:

- expanded GitHub Actions quality checks
- security scanning
- local parity commands in `Makefile`
- clearer documentation entrypoints

## How It Works

The repository quality gate includes:

- formatting checks
- `sqlc generate`
- generated-code drift detection
- `go vet`
- `staticcheck`
- `golangci-lint`
- `govulncheck`
- `go test ./...`
- `go build ./cmd/api ./cmd/worker`
- `gitleaks`
- `Trivy`

Local parity is centered on:

- `make ci`

## Why It Matters

This slice makes the repository easier to validate consistently across local development and CI.

## How To Verify

Run the full local quality entrypoint:

```bash
make ci
```

If you need to isolate failures, rerun:

```bash
make fmt-check
make sqlc
git diff --exit-code
make vet
make staticcheck
make lint
make govulncheck
make test
make build
```

Expected result:

- each command succeeds

Also review:

- `.github/workflows/ci.yml`
- `Makefile`

to confirm the workflow and local commands remain aligned.
