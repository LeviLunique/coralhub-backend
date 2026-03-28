# Events and Reminder Scheduling Testing Guide

This file shows how to verify the Stage 6 slice on branch:

- `feat/events-reminders`

## 1. Confirm You Are On The Correct Branch

Run:

```bash
git branch --show-current
```

Expected:

```text
feat/events-reminders
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
```

## 4. Start The API

Run:

```bash
make run-api
```

Leave it running.

## 5. Create The Stage 6 Base Data

Create two users:

```bash
curl -s -H 'X-Tenant-Slug: coral-jovem-asa-norte' -H 'Content-Type: application/json' \
  -d '{"email":"ana.stage6@example.com","full_name":"Ana Stage 6"}' \
  http://127.0.0.1:8080/api/v1/users
```

```bash
curl -s -H 'X-Tenant-Slug: coral-jovem-asa-norte' -H 'Content-Type: application/json' \
  -d '{"email":"maria.stage6@example.com","full_name":"Maria Stage 6"}' \
  http://127.0.0.1:8080/api/v1/users
```

Create a choir as Ana:

```bash
curl -s -H 'X-Tenant-Slug: coral-jovem-asa-norte' -H 'X-User-Email: ana.stage6@example.com' \
  -H 'Content-Type: application/json' -d '{"name":"Stage 6 Choir"}' \
  http://127.0.0.1:8080/api/v1/choirs
```

Save the returned `choir_id`.

Add Maria as a member:

```bash
curl -s -H 'X-Tenant-Slug: coral-jovem-asa-norte' -H 'X-User-Email: ana.stage6@example.com' \
  -H 'Content-Type: application/json' \
  -d '{"user_id":"<maria-user-id>","role":"member"}' \
  http://127.0.0.1:8080/api/v1/choirs/<choir-id>/memberships
```

## 6. Create An Event

Run:

```bash
curl -s -i \
  -H 'X-Tenant-Slug: coral-jovem-asa-norte' \
  -H 'X-User-Email: ana.stage6@example.com' \
  -H 'Content-Type: application/json' \
  -d '{"title":"Main rehearsal","event_type":"rehearsal","location":"Hall A","start_at":"2026-04-20T19:00:00Z"}' \
  http://127.0.0.1:8080/api/v1/choirs/<choir-id>/events
```

Expected:

- HTTP `201`
- JSON body with:
  - `id`
  - `choir_id`
  - `event_type`
  - `start_at`
  - `active: true`

Save the returned `event_id`.

## 7. List Choir Events

Run:

```bash
curl -s \
  -H 'X-Tenant-Slug: coral-jovem-asa-norte' \
  -H 'X-User-Email: maria.stage6@example.com' \
  http://127.0.0.1:8080/api/v1/choirs/<choir-id>/events
```

Expected:

- one event item visible to Maria because she is a choir member

## 8. Get The Event Directly

Run:

```bash
curl -s \
  -H 'X-Tenant-Slug: coral-jovem-asa-norte' \
  -H 'X-User-Email: maria.stage6@example.com' \
  http://127.0.0.1:8080/api/v1/events/<event-id>
```

Expected:

- the created event payload

## 9. Verify Scheduled Reminders In PostgreSQL

Connect with `psql`:

```bash
PGPASSWORD='<your DB_PASSWORD from .env>' psql -h localhost -p 5433 -U coralhub -d coralhub
```

Then run these separately:

```sql
select reminder_type, status, count(*) from scheduled_notifications group by reminder_type, status order by reminder_type, status;
```

```sql
select event_id, user_id, reminder_type, scheduled_for, status from scheduled_notifications where event_id = '<event-id>' order by scheduled_for, reminder_type, user_id;
```

Expected:

- four pending reminders
- two reminder types per user:
  - `day_before`
  - `hour_before`

Exit with:

```sql
\q
```

## 10. Update The Event

Run:

```bash
curl -s -i \
  -H 'X-Tenant-Slug: coral-jovem-asa-norte' \
  -H 'X-User-Email: ana.stage6@example.com' \
  -H 'Content-Type: application/json' \
  -X PUT \
  -d '{"title":"Main rehearsal updated","event_type":"presentation","location":"Main sanctuary","start_at":"2026-04-22T19:00:00Z"}' \
  http://127.0.0.1:8080/api/v1/events/<event-id>
```

Expected:

- HTTP `200`
- updated title and event type

Then re-run the SQL from Step 9.

Expected after the update:

- older pending reminders were marked `canceled`
- replacement pending reminders exist for the new schedule

## 11. Verify Authorization

Try to create an event as Maria:

```bash
curl -s -i \
  -H 'X-Tenant-Slug: coral-jovem-asa-norte' \
  -H 'X-User-Email: maria.stage6@example.com' \
  -H 'Content-Type: application/json' \
  -d '{"title":"Unauthorized event","event_type":"other","start_at":"2026-04-25T18:00:00Z"}' \
  http://127.0.0.1:8080/api/v1/choirs/<choir-id>/events
```

Expected:

- HTTP `403`

## 12. Cancel The Event

Run:

```bash
curl -s -i \
  -H 'X-Tenant-Slug: coral-jovem-asa-norte' \
  -H 'X-User-Email: ana.stage6@example.com' \
  -X DELETE \
  http://127.0.0.1:8080/api/v1/events/<event-id>
```

Expected:

- HTTP `204`

Then re-run:

```sql
select event_id, user_id, reminder_type, scheduled_for, status from scheduled_notifications where event_id = '<event-id>' order by scheduled_for, reminder_type, user_id;
```

Expected:

- all reminders for the event are `canceled`

Also verify the event is no longer visible:

```bash
curl -s -i \
  -H 'X-Tenant-Slug: coral-jovem-asa-norte' \
  -H 'X-User-Email: ana.stage6@example.com' \
  http://127.0.0.1:8080/api/v1/events/<event-id>
```

Expected:

- HTTP `404`

## 13. Run Automated Checks

Run:

```bash
env GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./...
env GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go vet ./...
env GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go build ./cmd/api ./cmd/worker
```

Expected:

- all commands succeed
- `internal/modules/events` passes
- `internal/store/postgres` passes, including the event repository integration test
