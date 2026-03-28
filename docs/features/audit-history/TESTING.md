# Audit History Testing

This guide verifies the Stage 9 audit history slice on branch:

- `feat/audit-history`

## 1. Start Local Dependencies

From the repo root:

```bash
make compose-up
```

Expected:

- PostgreSQL is healthy
- MinIO is up

## 2. Run the Automated Validation

From the repo root, run:

```bash
sqlc generate
```

```bash
gofmt -w ./cmd ./internal
```

```bash
env GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./...
```

```bash
env GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go vet ./...
```

```bash
env GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go build ./cmd/api ./cmd/worker
```

Expected:

- all tests pass
- `go vet` prints no issues
- both binaries build successfully

## 3. Apply Migrations

If your local database is already running, apply the latest migration set with the project migration flow you are using locally.

The important Stage 9 result is that `audit_log` exists.

## 4. Confirm the Audit Table Exists

Connect to PostgreSQL:

```bash
PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME"
```

Inside `psql`, run each command on its own line:

```sql
\dt audit_log
```

```sql
\d audit_log
```

Expected:

- the `audit_log` table exists
- it includes `tenant_id`, `entity_type`, `entity_id`, `action`, `actor_id`, `occurred_at`, and `payload_json`

## 5. Verify Membership Audit Through Tests

The repository integration tests already validate that a membership add writes:

- `membership.added`

If you want to inspect that specific test run again:

```bash
env GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./internal/store/postgres -run 'TestMembershipRepositoryCreateAndListByChoirIDIntegration' -count=1
```

Expected:

- the test passes

## 6. Verify Event Audit Through Tests

Run:

```bash
env GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./internal/store/postgres -run 'TestEventRepositoryCreateUpdateAndCancelIntegration' -count=1
```

Expected:

- the test passes
- the flow proves audit rows for:
  - `event.created`
  - `event.updated`
  - `event.canceled`
  - `notification.generated`

## 7. Verify Notification Audit Through Tests

Run:

```bash
env GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./internal/store/postgres -run 'TestNotificationRepositoryClaimAndStateTransitionsIntegration' -count=1
```

Expected:

- the test passes
- the flow proves audit rows for:
  - `notification.sent`
  - `notification.failed`
  - `notification.invalid_token`

## 8. Verify Repository-Level Audit Retrieval

Run:

```bash
env GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./internal/store/postgres -run 'TestAuditRepositoryCreateAndListByTenantIDIntegration' -count=1
```

Expected:

- the test passes
- the audit repository can create and list tenant-scoped audit rows

## 9. Manual Spot Check In PostgreSQL

After exercising audited flows in the running application, inspect the latest tenant history:

```sql
select entity_type, action, actor_id, occurred_at
from audit_log
where tenant_id = (select id from tenants where slug = 'coral-jovem-asa-norte')
order by occurred_at desc
limit 20;
```

Expected:

- you see recent actions for memberships, events, or notifications
- rows are tenant-scoped

Exit `psql` with:

```sql
\q
```

## 10. Known Stage 9 Limitations

These are not failures for this slice:

- there is no audit HTTP endpoint yet
- file upload/remove audit is not implemented yet
- membership removal audit is not implemented yet
