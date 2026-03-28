# Notification Worker Testing Guide

This file shows how to verify the Stage 7 slice on branch:

- `feat/notification-worker`

## 1. Confirm You Are On The Correct Branch

Run:

```bash
git branch --show-current
```

Expected:

```text
feat/notification-worker
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
```

## 4. Start The API

Run:

```bash
make run-api
```

Leave it running.

## 5. Create The Stage 7 Base Data

Follow the Stage 6 setup flow to create:

- one manager user
- one member user
- one choir
- one choir membership
- one event

If you already have a Stage 6 event, you can reuse it.

## 6. Force Notifications To Become Due

Open `psql`:

```bash
PGPASSWORD='<your DB_PASSWORD from .env>' psql -h localhost -p 5433 -U coralhub -d coralhub
```

Then run:

```sql
update scheduled_notifications
set scheduled_for = now() - interval '1 minute',
    status = 'pending',
    processing_started_at = null,
    last_error = null
where status = 'pending';
```

Check the rows:

```sql
select id, event_id, user_id, reminder_type, status, attempts, scheduled_for from scheduled_notifications order by scheduled_for, reminder_type, user_id;
```

Expected:

- due rows with `status = 'pending'`
- `attempts = 0`

Exit with:

```sql
\q
```

## 7. Start The Worker

Run in another terminal:

```bash
make run-worker
```

Expected:

- the worker starts
- after one poll cycle it logs that notifications were processed

Important:

- Stage 7 uses a temporary no-op sender
- due notifications are therefore marked as delivered without calling FCM

Stop the worker with `Ctrl+C` after one or two poll cycles.

## 8. Verify Notification State In PostgreSQL

Open `psql` again:

```bash
PGPASSWORD='<your DB_PASSWORD from .env>' psql -h localhost -p 5433 -U coralhub -d coralhub
```

Run:

```sql
select status, attempts, count(*) from scheduled_notifications group by status, attempts order by status, attempts;
```

```sql
select id, reminder_type, status, attempts, processing_started_at, sent_at, last_error from scheduled_notifications order by updated_at desc, id;
```

Expected:

- due rows moved to `sent`
- `attempts = 1`
- `processing_started_at` is `null`
- `sent_at` is not `null`
- `last_error` is `null`

Exit with:

```sql
\q
```

## 9. Run Automated Checks

Run:

```bash
env GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./...
env GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go vet ./...
env GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go build ./cmd/api ./cmd/worker
```

Expected:

- all commands succeed
- `internal/modules/notifications` passes
- `internal/store/postgres` passes, including the worker-state integration test
