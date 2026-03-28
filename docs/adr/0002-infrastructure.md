# ADR 0002 - CoralHub Production Infrastructure

## Status

Accepted

## Date

2026-03-27

## Context

The backend requires a production infrastructure that is:

- reliable
- operationally simple
- secure
- scalable enough for early growth
- observable
- maintainable by a small-to-medium engineering team

The infrastructure must support:

- one API service
- one background worker service
- PostgreSQL
- object storage
- mobile push integration
- safe deployments
- environment separation

The platform is:

- `CoralHub`
- public domain: `coralhub.com.br`
- initial tenant: `Coral Jovem Asa Norte`

The infrastructure must support future tenants without per-tenant backend duplication.

The goal is to avoid both extremes:

- underengineered VM-based setups that become brittle
- overengineered Kubernetes or microservice platforms that are unjustified for the product stage

## Decision

The production infrastructure will use:

- AWS as the cloud provider
- Amazon ECS Fargate as the compute platform
- Amazon RDS for PostgreSQL as the relational database
- Amazon S3 as object storage
- AWS Secrets Manager for runtime secrets
- CloudWatch for logs, metrics, alarms, and dashboards
- OpenTelemetry for application instrumentation
- GitHub Actions for CI/CD
- GHCR for container image registry

The primary backend repository is:

- `coralhub-backend`

### Runtime topology

Deploy two services:

- `api`
- `worker`

Both run as separate ECS Fargate services.

The database runs in RDS.
Files live in private S3 buckets.

The runtime is shared across tenants.
Tenant context is resolved at the application layer and may be influenced by authenticated claims, domains, or subdomains.

### Network topology

Use:

- one VPC per environment
- at least two availability zones
- public subnets for ALB
- private subnets for ECS tasks
- private subnets for RDS

The API is exposed through an ALB.
The worker has no public ingress.
RDS is private only.

Recommended public entrypoint shape:

- `api.coralhub.com.br` for the backend API

### Secret strategy

Runtime secrets must be stored in Secrets Manager.

Secrets include:

- database credentials
- Firebase credentials or equivalent secret material
- application signing secrets

### Storage strategy

Use private S3 buckets with:

- server-side encryption
- presigned URLs for downloads
- no public read access
- tenant-aware object key design

### Database strategy

Use RDS PostgreSQL with:

- automated backups
- encryption at rest
- private access only
- backup retention configured

Enable Multi-AZ in production if uptime expectations justify the cost.

### Deployment strategy

Use GitHub Actions to:

- run tests and checks
- build container images
- push images to GHCR
- deploy to AWS environments

Images must be referenced by immutable tags, such as commit SHA.

### Infrastructure as code

Infrastructure must be managed with Terraform.

Console-only infrastructure management is rejected.

### Observability strategy

Use:

- structured application logs
- CloudWatch Logs
- CloudWatch Metrics and Alarms
- OpenTelemetry instrumentation for traces

Dashboards and alerts must focus on actionable signals, not vanity telemetry.

### Quality and security tooling

The mandatory baseline is:

- `go vet`
- `staticcheck`
- `golangci-lint`
- `govulncheck`
- `Trivy`
- `gitleaks`

SonarQube is not mandatory.

It may be added later only if organizational reporting or compliance requires it.

### Repository governance

Adopt:

- trunk-based development
- short-lived branches
- PR-based merge flow
- branch protection on `main`
- Conventional Commits

Recommended companion repositories:

- `coralhub-web`
- `coralhub-mobile`

## Consequences

### Positive consequences

- low operational complexity compared to Kubernetes
- safer and more repeatable deployments than ad hoc VM setups
- strong production baseline for a small team
- clear runtime separation between API and worker
- manageable security and observability model
- one backend platform can serve multiple churches cleanly

### Negative consequences

- AWS managed services increase baseline cost versus the cheapest possible deployment
- NAT and AWS networking can become a noticeable cost factor
- GHCR is slightly less operationally aligned with AWS than ECR

These trade-offs are accepted because they produce a more reliable and maintainable production setup.

## Alternatives Considered

### Kubernetes / EKS

Rejected.

Reason:

- too much operational overhead for this project stage

### EC2 / self-managed VMs

Rejected.

Reason:

- more operational toil
- less predictable deployment and scaling model

### AWS Lambda

Rejected.

Reason:

- poor fit for a continuously running worker and transactional backend with DB-heavy behavior

### Railway / Fly.io / Render

Rejected as primary production recommendation.

Reason:

- good platforms for smaller products
- less aligned with the requested AWS-based production model

### ECR instead of GHCR

Operationally preferred, but not selected here because GHCR was part of the chosen requirements.

If registry choice becomes negotiable later, ECR should be reconsidered.

### SonarQube mandatory in CI

Rejected.

Reason:

- not necessary for this Go backend
- worse cost-benefit than native and focused analysis tooling

## Operational Notes

Before production launch, the team must define:

- restore procedures
- deployment rollback procedures
- secret rotation ownership
- RTO and RPO expectations
- alarm ownership and response flow

Without these, the platform is not production-ready even if the infrastructure exists.

## Related Documents

- [INFRASTRUCTURE_BLUEPRINT.md](/Users/levilunique/Workspace/Go/coral/docs/INFRASTRUCTURE_BLUEPRINT.md)
- [0001-architecture.md](/Users/levilunique/Workspace/Go/coral/docs/adr/0001-architecture.md)
