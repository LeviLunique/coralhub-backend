# Production Blueprint

This document explains the intended production operating model for CoralHub Backend in practical terms.

Use it when you need more operational detail than the summary in [INFRASTRUCTURE.md](INFRASTRUCTURE.md).

## Scope

This blueprint focuses on:

- runtime topology
- network boundaries
- environment separation
- storage and secret handling
- deployment expectations
- observability and security baseline

## Production Topology

The expected production shape is:

- one public application load balancer
- one private `api` service on ECS Fargate
- one private `worker` service on ECS Fargate
- one private PostgreSQL database on RDS
- one private S3 bucket per environment, or strong prefix separation if needed

Traffic flow:

1. clients call the public API endpoint
2. the load balancer forwards requests to the `api` service
3. the API reads and writes PostgreSQL data, accesses S3 objects, and loads secrets
4. the worker polls PostgreSQL for due notification jobs and calls FCM when delivery is enabled

## Network Model

Recommended network layout:

- one VPC per environment
- at least two availability zones
- public subnets only for the load balancer
- private subnets for ECS tasks
- private subnets for RDS

Security expectations:

- RDS is never public
- the worker exposes no public ingress
- ECS services accept traffic only from approved security groups
- egress is controlled deliberately, especially for FCM access

## Environment Separation

Minimum environments:

- `dev`
- `staging`
- `prod`

Recommended rules:

- separate secrets for each environment
- separate databases for each environment
- separate S3 buckets or clearly separated prefixes
- explicit deployment promotion rather than automatic production rollout

The initial tenant seed stays the same across environments as seed data:

- slug: `coral-jovem-asa-norte`
- display name: `Coral Jovem Asa Norte`

## Storage Model

Binary files are stored in private S3-compatible object storage.

Operational rules:

- no public object listing
- no broad public read access
- application roles own bucket access
- object keys include tenant context
- server-side encryption is enabled

Downloads should be exposed through pre-signed URLs rather than public objects.

## Secret Management

Secrets belong in AWS Secrets Manager.

Examples:

- database credentials
- Firebase credentials
- signing secrets
- third-party integration secrets

Do not rely on committed `.env` files for production secrets.

## Deployment Model

The repository currently provides a CI baseline, not a full deployment pipeline.

Current baseline:

- test and build verification in GitHub Actions
- static analysis and vulnerability scanning
- image-oriented delivery model through GHCR

Expected deployment direction:

1. build immutable container images
2. push images to GHCR
3. deploy by environment with explicit promotion steps
4. keep infrastructure managed through Terraform

## Observability

The operational baseline should include:

- structured application logs
- CloudWatch log aggregation
- request and worker metrics
- OpenTelemetry traces
- actionable alarms and dashboards

The application already exposes a basic `/metrics` endpoint and runtime counters from the hardening slice.

## CI and Quality Baseline

The quality gate currently includes:

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
- `Trivy`

This is the minimum baseline expected before deployment work grows further.

## What This Blueprint Optimizes For

The blueprint optimizes for:

- operational simplicity
- tenant-safe infrastructure defaults
- predictable production behavior
- small-team maintainability

It does not optimize for:

- the absolute cheapest hosting option
- early Kubernetes adoption
- infrastructure novelty

## Follow-Up Work Still Expected

The current production direction is credible, but not finished.

Likely next operational slices:

- deployment workflows per environment
- rollout and rollback procedures
- richer dashboards and alarms
- secret rotation procedures
- more explicit runbooks for incidents and restores

## Related Handbook Documents

- [Infrastructure Summary](INFRASTRUCTURE.md)
- [Infrastructure ADR](adr/infrastructure.md)
- [Technical Overview](TECHNICAL_OVERVIEW.md)
