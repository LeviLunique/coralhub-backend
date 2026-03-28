# FCM Integration Testing Guide

This file shows how to verify the Stage 8 slice on branch:

- `feat/fcm-integration`

## 1. Confirm You Are On The Correct Branch

Run:

```bash
git branch --show-current
```

Expected:

```text
feat/fcm-integration
```

## 2. Start Local Dependencies

Run:

```bash
make compose-up
```

## 3. Apply The Required Migrations

Run:

```bash
PGPASSWORD='<your DB_PASSWORD from .env>' psql -h localhost -p 5433 -U coralhub -d coralhub -f db/migrations/000002_init_choirs_and_users.up.sql
PGPASSWORD='<your DB_PASSWORD from .env>' psql -h localhost -p 5433 -U coralhub -d coralhub -f db/migrations/000003_init_choir_members.up.sql
PGPASSWORD='<your DB_PASSWORD from .env>' psql -h localhost -p 5433 -U coralhub -d coralhub -f db/migrations/000004_init_voice_kits_and_kit_files.up.sql
PGPASSWORD='<your DB_PASSWORD from .env>' psql -h localhost -p 5433 -U coralhub -d coralhub -f db/migrations/000005_init_events_and_scheduled_notifications.up.sql
PGPASSWORD='<your DB_PASSWORD from .env>' psql -h localhost -p 5433 -U coralhub -d coralhub -f db/migrations/000006_alter_scheduled_notifications_for_worker.up.sql
PGPASSWORD='<your DB_PASSWORD from .env>' psql -h localhost -p 5433 -U coralhub -d coralhub -f db/migrations/000007_init_device_tokens.up.sql
```

## 4. Choose Your Verification Mode

You have two useful ways to verify Stage 8:

- automated tests only
- real worker startup with Firebase disabled

Real end-to-end push delivery through FCM requires a Firebase service account JSON.

## 5. Verify Device Token Persistence In PostgreSQL

Open `psql`:

```bash
PGPASSWORD='<your DB_PASSWORD from .env>' psql -h localhost -p 5433 -U coralhub -d coralhub
```

Then insert a token manually:

```sql
insert into device_tokens (tenant_id, user_id, platform, token)
select t.id, u.id, 'android', 'stage8-test-token'
from tenants t
join users u on u.tenant_id = t.id
where t.slug = 'coral-jovem-asa-norte'
limit 1;
```

Check it:

```sql
select user_id, platform, token, active from device_tokens order by created_at desc;
```

Expected:

- one active token row

Exit with:

```sql
\q
```

## 6. Start The Worker With Firebase Disabled

Make sure your `.env` contains:

```env
FIREBASE_ENABLED=false
```

Then run:

```bash
make run-worker
```

Expected:

- the worker starts successfully
- it logs that Firebase is disabled and the no-op sender is being used

This confirms the fallback path still works for local development.

## 7. Real FCM Verification

To exercise the real FCM sender:

1. obtain a Firebase service account JSON
2. set in `.env`:

```env
FIREBASE_ENABLED=true
FIREBASE_CREDENTIALS_FILE=/absolute/path/to/service-account.json
```

3. ensure there is at least one active `device_tokens` row for the target user
4. create a due scheduled notification for that user
5. run `make run-worker`

Expected:

- the worker starts without falling back to the no-op sender
- the FCM sender attempts delivery through Firebase
- invalid tokens are deactivated
- successful sends move notifications to `sent`

## 8. Run Automated Checks

Run:

```bash
env GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./...
env GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go vet ./...
env GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go build ./cmd/api ./cmd/worker
```

Expected:

- all commands succeed
- `internal/integrations/push/fcm` passes
- `internal/store/postgres` passes, including the device token repository test
