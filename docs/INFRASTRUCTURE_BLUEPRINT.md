# Production Infrastructure Blueprint - CoralHub Backend

## 1. Purpose of This Document

This document defines the recommended production infrastructure for the **CoralHub** backend.

CoralHub is a multi-tenant choir management platform.

The initial tenant is:

- `Coral Jovem Asa Norte`

The infrastructure must support future tenants without backend duplication.

It assumes the following stack choices are fixed unless there is a strong operational reason to change them:

- AWS
- Amazon S3
- Amazon RDS for PostgreSQL
- Amazon ECS Fargate
- GitHub Actions
- GHCR
- AWS Secrets Manager
- CloudWatch
- OpenTelemetry

This document complements `docs/AI_IMPLEMENTATION_GUIDE.md`.

The implementation guide explains how to build the application.
This blueprint explains how to run it safely in production.

---

## 2. Executive Summary

The recommended production topology is:

- one AWS account per environment group or, at minimum, strict environment separation
- one VPC per environment
- public ALB
- private ECS Fargate services for `api` and `worker`
- private RDS PostgreSQL
- private S3 bucket for file storage
- Secrets Manager for sensitive configuration
- CloudWatch Logs, Metrics, Alarms, and Dashboards
- OpenTelemetry instrumentation with traces exported to AWS-compatible backends

The recommended public domain strategy is:

- platform domain: `coralhub.com.br`
- backend API domain: for example `api.coralhub.com.br`
- tenant-facing web domains: platform-managed subdomains or custom domains later

This is a strong default for a serious small-to-medium backend.

It is not the cheapest possible setup, but it is much safer and more maintainable than trying to self-manage VMs or overbuilding Kubernetes too early.

---

## 3. Production Architecture

### 3.1. Core Runtime Topology

Recommended runtime components:

- `api` service on ECS Fargate
- `worker` service on ECS Fargate
- `postgres` on Amazon RDS for PostgreSQL
- `files` on Amazon S3
- `secrets` in AWS Secrets Manager
- `logs/metrics` in CloudWatch
- `traces` through OpenTelemetry

### 3.2. Networking Model

Use:

- one VPC per environment
- at least two availability zones
- public subnets for ALB only
- private subnets for ECS tasks
- private subnets for RDS

Security posture:

- RDS must not be public
- ECS tasks must not have unrestricted inbound access
- only ALB should receive public HTTP/HTTPS traffic
- ECS tasks should accept traffic only from the ALB security group
- worker service should have no public ingress

### 3.3. Recommended Traffic Flow

External client -> ALB -> ECS `api` service -> RDS / S3 / Secrets Manager / external FCM

Worker flow:

ECS `worker` service -> RDS -> FCM -> CloudWatch/OTel

Tenant-facing web and mobile clients should use the same backend, with tenant context resolved by authenticated claims and, where applicable, by domain or subdomain mapping.

### 3.4. NAT and Egress

You need a deliberate decision here.

Options:

- use NAT gateways for private subnet outbound traffic
- reduce NAT dependency with VPC endpoints where possible

Recommended:

- use VPC endpoints for S3, Secrets Manager, CloudWatch Logs, and ECR if applicable
- still expect some outbound internet path for FCM

Be aware:

- NAT can become a meaningful cost line item
- for low traffic systems, NAT cost can look irrationally high compared to app load

If cost pressure becomes significant, reassess the network design carefully instead of blindly accepting NAT overhead.

---

## 4. AWS Service Decisions

### 4.1. ECS Fargate

Use ECS Fargate as the primary compute platform.

Why:

- low operational burden
- no Kubernetes control plane
- easier to onboard than EKS
- production-credible
- good fit for API + worker topology

Run two services:

- `coralhub-api`
- `coralhub-worker`

Do not run the worker inside the API process in production unless scale is tiny and you are intentionally collapsing topology for cost reasons.

### 4.2. RDS PostgreSQL

Use Amazon RDS for PostgreSQL.

Recommended baseline:

- private instance
- automated backups enabled
- Multi-AZ for production if uptime matters
- Performance Insights enabled if budget allows
- encryption at rest enabled
- sensible connection limits and pool sizing in the app

Do not:

- expose RDS publicly
- let app pool sizes exceed what the instance can actually support

### 4.3. Amazon S3

Use private S3 buckets for file storage.

Recommended:

- one bucket per environment or strongly separated prefixes if bucket count must be minimized
- presigned URLs for downloads
- server-side encryption enabled
- lifecycle policies for abandoned or obsolete objects if the product requires cleanup

Bucket rules:

- deny public object listing
- deny broad public read
- only application roles should have object access
- object keys should include tenant context

### 4.4. Secrets Manager

Use Secrets Manager for:

- database credentials
- Firebase credentials or related secret material
- application secrets
- third-party API secrets

Do not place secrets in:

- task definition plaintext
- committed `.env` files
- GitHub repository variables when a runtime secret source is more appropriate

### 4.5. GHCR

Because this blueprint assumes `GitHub Actions + GHCR`, image publishing should go to GHCR.

That is acceptable.

However, if the infrastructure is fully AWS-centric, `ECR` is operationally cleaner than `GHCR`.

So the honest recommendation is:

- if the requirement is fixed, use `GHCR`
- if it is still negotiable, prefer `ECR`

This blueprint keeps `GHCR` because you explicitly asked for it.

---

## 5. Environment Strategy

At minimum define:

- `dev`
- `staging`
- `prod`

Recommended rules:

- separate secrets per environment
- separate S3 buckets or prefixes per environment
- separate databases per environment
- separate ECS services per environment

Do not share a production database with non-production environments.

Seed the initial tenant through a controlled migration or seed process:

- slug: `coral-jovem-asa-norte`
- display name: `Coral Jovem Asa Norte`

### 5.1. Deployment Promotion

Recommended deployment flow:

1. merge to `main`
2. build and push image
3. deploy automatically to `dev`
4. deploy to `staging` on approval
5. deploy to `prod` on approval

Keep promotion explicit.

Do not auto-promote to production just because a build succeeded.

---

## 6. Compute and Scaling Strategy

### 6.1. API Service

The API service should scale horizontally based on:

- CPU
- memory
- request count or ALB target metrics if needed

Start with conservative scaling:

- minimum 2 tasks in production for availability
- scale out on sustained pressure, not on noise

### 6.2. Worker Service

The worker service should scale based on:

- CPU
- memory
- queue depth proxy metrics if available
- notification lag if exported as a custom metric

Do not autoscale the worker blindly without understanding DB contention.

More workers are not always better if the bottleneck is:

- DB locking
- outbound push rate
- downstream provider throttling

### 6.3. Connection Management

Fargate scaling plus PostgreSQL can become dangerous if every task opens too many DB connections.

Rules:

- keep app connection pools small and explicit
- budget total connections across all API and worker tasks
- consider a connection pooling strategy if concurrency grows materially

If needed later:

- evaluate RDS Proxy

Do not introduce it on day one unless connection churn is already a known issue.

---

## 7. Security Model

### 7.1. IAM

Use separate task roles for:

- API service
- worker service

Grant least privilege only.

Examples:

- S3 read/write only for required bucket and key prefixes
- Secrets Manager read only for required secret ARNs
- CloudWatch Logs write only for the service log groups

### 7.2. TLS

Use HTTPS only at the edge.

Recommended:

- ALB with ACM-managed certificate
- HTTP to HTTPS redirect enabled

### 7.3. Security Groups

Recommended:

- ALB security group allows inbound 80/443 from internet
- API task security group allows inbound only from ALB security group
- worker task security group allows no public inbound
- RDS security group allows inbound only from ECS task security groups

### 7.4. Secret Rotation

If the team can support it operationally, enable periodic rotation for database credentials and sensitive application secrets.

If not, at least:

- define rotation process
- document ownership
- rotate on incidents and staff changes

### 7.5. Auditability

Enable and preserve:

- CloudTrail where appropriate
- application audit logs for critical business actions
- deployment auditability through GitHub Actions history

---

## 8. Database Operations

### 8.1. Migration Strategy

Migrations must be:

- versioned
- forward-only by default
- run automatically in controlled deployment steps

Do not rely on manual schema drift correction.

### 8.2. Backups

For RDS:

- enable automated backups
- set retention according to business needs
- validate restore procedures, not just backup existence

At least once, test:

- point-in-time restore
- full service recovery against a restored database

### 8.3. Maintenance

Define:

- maintenance window
- backup window
- engine upgrade strategy

Do not treat RDS as “managed therefore no operational responsibility”.

### 8.4. Performance

Track:

- slow queries
- lock contention
- connection usage
- storage growth
- CPU and memory pressure

Use:

- PostgreSQL indexes intentionally
- application-level query review
- Performance Insights if available

---

## 9. S3 Operations

### 9.1. Bucket Layout

Recommended object key layout:

```text
{environment}/tenants/{tenant_slug}/choirs/{choir_id}/voice-kits/{kit_id}/files/{file_id}/{stored_filename}
```

This is explicit, debuggable, and operationally predictable.

### 9.2. Storage Rules

- private bucket only
- no public ACL reliance
- application uses IAM, not hard-coded long-lived access keys where avoidable
- presigned URL expiration should be short

### 9.3. Lifecycle

Define lifecycle only when product rules are clear.

Potential uses:

- delete abandoned multipart uploads
- transition very old rarely-accessed content if business allows
- clean up tombstoned files after retention period

Do not add aggressive tiering before understanding actual access patterns.

---

## 10. Deployment Pipeline

### 10.1. GitHub Actions

Recommended pipeline stages:

1. lint
2. test
3. security scanning
4. build image
5. push image to GHCR
6. deploy to environment

### 10.2. Deploy Strategy

Recommended:

- rolling or blue/green style deployment depending on operational maturity
- one ECS service deployment per application component

For a small team, rolling deployment is usually enough initially.

### 10.3. Image Tagging

Use at least:

- immutable commit SHA tag
- optional semver tag for releases
- human-friendly environment tag only if it does not become the source of truth

Do not deploy mutable-only tags as your sole reference.

### 10.4. Migrations in Deploy

Recommended:

- run DB migrations as an explicit pre-deploy or deploy step
- make deployments fail fast on migration failure

Do not let API and worker boot against an incompatible schema and hope for the best.

---

## 11. Observability

Yes, infrastructure observability belongs here.

This document should own the production observability baseline.

### 11.1. Logs

Use structured JSON logs from the application.

Ship logs to CloudWatch Logs.

Include:

- timestamp
- service name
- environment
- tenant slug or tenant ID where safe and useful
- request ID
- trace ID where available
- user or actor ID where appropriate
- route
- status code
- latency
- error class
- worker job identifiers

### 11.2. Metrics

At minimum track:

- request count
- request latency
- error rate
- DB latency
- DB connection saturation
- worker poll count
- worker processing count
- notification success/failure
- queue lag or oldest pending notification age
- S3 upload failure count

CloudWatch alarms should exist for:

- API 5xx spikes
- high API latency
- ECS task crash/restart behavior
- RDS CPU pressure
- RDS low storage risk
- abnormal worker failure rate
- notification backlog growth

### 11.3. Tracing

Instrument the app with OpenTelemetry.

Recommended:

- propagate trace context across HTTP and worker flows where possible
- capture spans for HTTP handlers, DB queries, S3 operations, and push provider calls

If staying fully on AWS-managed observability, traces can flow into AWS-native backends.

The important point is not the exact exporter.
The important point is to instrument with OTel now so observability remains portable.

### 11.4. Dashboards

Create dashboards for:

- API health
- worker health
- notification pipeline
- database health
- storage error trends

If no one looks at the dashboards, reduce them.
Do not produce decorative observability.

### 11.5. Alerting

Alerts should be actionable.

Good alerts:

- sustained 5xx above threshold
- worker stopped processing
- oldest pending notification too old
- DB unavailable
- abnormal auth failure spike

Bad alerts:

- low-signal noise without operator action
- every transient blip

---

## 12. ECS Task Design

### 12.1. API Task

Task contains:

- application container

Optional:

- OTel sidecar if your tracing/export pattern requires it

### 12.2. Worker Task

Task contains:

- worker container

Optional:

- OTel sidecar

### 12.3. Health Checks

API:

- health endpoint for container health
- ALB target health check

Worker:

- container health check should verify the process is alive
- consider lightweight self-checks, but do not make health dependent on expensive downstream checks every few seconds

### 12.4. Resource Sizing

Start conservative.

Do not overprovision from fear.
Do not underprovision to save pennies while degrading reliability.

Treat initial CPU and memory sizing as empirical and revisit after metrics exist.

---

## 13. Configuration and Secrets Layout

### 13.1. Non-Secret Configuration

Keep in task environment variables or parameterized deploy config:

- app environment
- log level
- HTTP port
- worker poll interval
- feature flags
- S3 bucket name
- AWS region

### 13.2. Secrets

Store in Secrets Manager:

- DB DSN or credentials
- Firebase secret material
- app signing secrets

### 13.3. Config Validation

The app must fail at boot if:

- required secrets are missing
- invalid environment values are supplied
- critical integrations are misconfigured

Fail fast is safer than booting half-broken.

---

## 14. Infrastructure as Code

Use IaC.

Recommended:

- Terraform

Why:

- broad team familiarity
- mature AWS support
- standard for serious infrastructure work

Define at least:

- VPC
- subnets
- security groups
- ALB
- ECS cluster
- ECS services
- task definitions
- RDS
- S3 bucket
- Secrets Manager secrets
- CloudWatch log groups
- alarms
- IAM roles and policies

Do not build production infrastructure manually in the console as the primary process.

---

## 15. Repository Governance

Yes, branch and commit policy should be documented, but they are secondary to application and infrastructure design.

They belong here as delivery governance, not as “infra runtime”.

### 15.1. Branch Strategy

Use trunk-based development with short-lived branches.

Recommended naming:

- `feat/<short-description>`
- `fix/<short-description>`
- `refactor/<short-description>`
- `chore/<short-description>`
- `docs/<short-description>`
- `ci/<short-description>`
- `infra/<short-description>`

Recommended repository set for the platform:

- `coralhub-backend`
- `coralhub-web`
- `coralhub-mobile`

Protect `main`:

- no direct pushes
- PR required
- green CI required
- at least one review required

### 15.2. Commit Strategy

Use Conventional Commits.

Examples:

- `feat(events): schedule reminders on create`
- `fix(worker): prevent duplicate retry update`
- `infra(ecs): add worker service autoscaling policy`
- `ci(actions): add container security scan`

Do not mix unrelated infrastructure, refactor, and feature work in one large commit if it makes rollback or review unclear.

### 15.3. Pull Request Policy

Each PR should include:

- objective
- risk summary
- test evidence
- migration impact if any
- rollback considerations if relevant

---

## 16. Static Analysis and Quality Tooling

### 16.1. SonarQube

SonarQube is **not required** for this project to be production-serious.

For a Go backend of this size, I would not make SonarQube mandatory unless:

- your organization already standardizes on it
- compliance or management reporting requires it
- you want organization-wide dashboarding across many repos

Why I would not force it here:

- it adds operational and workflow overhead
- many findings overlap with better native or language-specific tooling
- small teams often get less value from it than they expect

### 16.2. Recommended Quality Tooling Instead

Use:

- `go vet`
- `staticcheck`
- `golangci-lint`
- `govulncheck`
- container image scanning such as `Trivy`
- secret scanning such as `gitleaks`

Optional:

- `Semgrep` if the team wants broader policy-based scanning
- `CodeQL` if you want GitHub-native security analysis

### 16.3. Recommended Default Tool Stack

My recommendation:

- mandatory: `go test`, `go vet`, `staticcheck`, `golangci-lint`, `govulncheck`, `Trivy`, `gitleaks`
- optional later: `CodeQL`
- only if organization requires it: `SonarQube` or `SonarCloud`

This is a better cost-benefit profile than forcing SonarQube by default.

---

## 17. Security and Delivery Checks in CI

Recommended CI checks:

- formatting check
- lint
- unit tests
- integration tests
- `sqlc generate` consistency check
- vulnerability scan for Go dependencies
- container image scan
- secret scan

Recommended release checks:

- migration compatibility
- image push success
- deploy approval gates for staging and production

---

## 18. Disaster Recovery and Operational Readiness

Define before production launch:

- RTO target
- RPO target
- restore playbook
- database rollback policy
- failed deployment rollback policy
- incident ownership

At minimum rehearse:

- service rollback
- DB restore into non-production
- secret rotation procedure

If these are undocumented, the infrastructure is not truly production-ready.

---

## 19. Cost Control

Main cost drivers to watch:

- NAT
- Fargate always-on tasks
- RDS instance sizing
- S3 storage and egress
- observability volume

Practical guidance:

- do not flood logs with high-cardinality noise
- keep worker count justified by backlog
- right-size RDS after real usage
- keep S3 objects private and use short-lived presigned URLs
- revisit GHCR vs ECR if AWS integration friction appears

Cost optimization should be deliberate, not reactive panic after the first bill.

---

## 20. Final Recommendation

This infrastructure blueprint is appropriate for:

- MVP that must already look professional
- solid V1
- early growth without replatforming

It avoids:

- Kubernetes overhead
- microservice sprawl
- VM snowflakes
- fake simplicity that collapses in production

If the system grows materially, the first evolutions should be:

- tighter autoscaling and DB pooling strategy
- stronger notification backlog metrics
- refined security controls
- deeper performance analysis

Not:

- a premature migration to microservices
- a premature move to Kafka
- a premature move to EKS

This is the right kind of boring infrastructure for this backend.
