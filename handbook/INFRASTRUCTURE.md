# Infrastructure and Operations

This document is the short operational summary for CoralHub Backend.

Use it to understand the production target quickly, then continue with [PRODUCTION_BLUEPRINT.md](PRODUCTION_BLUEPRINT.md) for the detailed operating model.

## Production Target

The backend is designed to run on AWS with:

- ECS Fargate for compute
- RDS PostgreSQL for the database
- S3 for object storage
- Secrets Manager for runtime secrets
- CloudWatch for logs, metrics, alarms, and dashboards
- OpenTelemetry for traces
- GitHub Actions for CI
- GHCR for container images
- Terraform for infrastructure management

## Runtime Summary

Production keeps two separate services:

- `api` for HTTP traffic
- `worker` for asynchronous processing

The intended topology is a public load balancer in front of a private API service, with the worker and RDS remaining private.

## Environment Strategy

Expected environments:

- `dev`
- `staging`
- `prod`

Recommended rules:

- separate secrets per environment
- separate databases per environment
- separate buckets or strong prefix separation per environment
- explicit promotion between environments

## Network and Security Posture

Security expectations:

- RDS is never public
- the worker has no public ingress
- application secrets are not committed to the repository
- S3 buckets remain private
- object keys include tenant context

## Deployment and CI

The Stage 11 CI baseline includes:

- formatting verification
- `sqlc generate`
- generated-code drift detection
- `go vet`
- `staticcheck`
- `golangci-lint`
- `govulncheck`
- `go test ./...`
- `go build ./cmd/api ./cmd/worker`
- `gitleaks`
- `Trivy` filesystem scanning

Local parity commands are exposed through `Makefile`, especially:

- `make ci`
- `make test`
- `make build`

## What Is Already In Place

Current operational baseline:

- API and worker separation
- configurable HTTP timeout handling
- metrics endpoint
- notification retention cleanup
- CI quality and security checks

Observability and deployment details are documented in [PRODUCTION_BLUEPRINT.md](PRODUCTION_BLUEPRINT.md).

## What Is Still Follow-Up Work

This repository has a production-oriented baseline, but it does not yet include full deployment automation.

Likely next operational work:

- environment-specific deployment workflows
- release automation
- broader pagination and additional hardening follow-up
- more production dashboards and alarms

## Reference Material

For the detailed operating model and decision rationale, see:

- [Infrastructure ADR](adr/infrastructure.md)
- [Production Blueprint](PRODUCTION_BLUEPRINT.md)
