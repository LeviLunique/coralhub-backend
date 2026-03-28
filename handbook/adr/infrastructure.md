# Infrastructure ADR

## Status

Accepted

## Decision

The production target for CoralHub Backend is AWS with:

- ECS Fargate for `api` and `worker`
- RDS PostgreSQL
- S3 for object storage
- Secrets Manager for runtime secrets
- CloudWatch and OpenTelemetry for observability
- GitHub Actions for CI
- GHCR for container images
- Terraform for infrastructure management

## Why This Decision Was Made

The project needs production credibility without taking on unnecessary platform complexity.

The chosen model aims for:

- a manageable operational footprint
- good separation between services and data stores
- safe secret handling
- a path to scaling without moving to Kubernetes too early

## What This Means In Practice

The expected production topology is:

1. a public ALB receives traffic
2. the ALB forwards traffic to the private `api` service
3. the `api` and `worker` services run in private subnets
4. PostgreSQL runs privately in RDS
5. binary files are stored in private S3 buckets

The worker remains a separate service because notification processing is a different operational concern than HTTP serving.

## Operational Rules

Developers and operators should assume:

- RDS is never public
- S3 is private
- secrets are not stored in committed `.env` files
- environments are separated
- deployments use immutable image references

## CI and Delivery Expectations

The baseline CI includes:

- formatting checks
- code generation checks
- static analysis
- vulnerability scanning
- tests
- build verification

This is a quality gate, not full deployment automation.

## What To Avoid

Do not default to:

- self-managed VMs for the main production path
- Kubernetes before the team has a concrete need for it
- console-only infrastructure changes
- public databases or public object storage

## Consequences

Positive outcomes:

- low operational overhead for a serious small team
- strong production baseline
- clean API and worker separation

Trade-offs:

- AWS managed services are not the cheapest possible option
- NAT and networking costs need attention
- GHCR is acceptable but slightly less native to AWS than ECR

## Guidance For New Contributors

When a change affects runtime behavior, ask:

1. Does this require a new secret?
2. Does this change API-only behavior, worker-only behavior, or both?
3. Does this introduce a new operational dependency that must be represented in Terraform and observability?

Infrastructure work should stay explicit and reviewable.
