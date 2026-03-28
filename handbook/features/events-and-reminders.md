# Events and Reminders

This guide explains the event scheduling feature and its reminder generation behavior.

## What It Does

This feature adds:

- choir event creation, update, list, get, and cancel flows
- reminder scheduling during event writes
- member-visible event reads
- manager-only event writes

Representative endpoints:

- `POST /api/v1/choirs/{choirID}/events`
- `GET /api/v1/choirs/{choirID}/events`
- `GET /api/v1/events/{eventID}`

## How It Works

Flow:

1. the actor is resolved in tenant context
2. the service checks choir membership and manager role where required
3. the event is created or updated transactionally
4. reminder rows are created or regenerated in the same transaction

The current reminder policy supports:

- `day_before`
- `hour_before`

## Why It Matters

This slice establishes one of the core business workflows in CoralHub: scheduled choir coordination with downstream notification generation.

## How To Verify

Start dependencies and the API:

```bash
make compose-up
make run-api
```

Create users, a choir, and a membership. Then create an event:

```bash
curl -s -i \
  -H 'X-Tenant-Slug: coral-jovem-asa-norte' \
  -H 'X-User-Email: ana.stage6@example.com' \
  -H 'Content-Type: application/json' \
  -d '{"title":"Main rehearsal","event_type":"rehearsal","location":"Hall A","start_at":"2026-04-20T19:00:00Z"}' \
  http://127.0.0.1:8080/api/v1/choirs/<choir-id>/events
```

List events as a choir member:

```bash
curl -s \
  -H 'X-Tenant-Slug: coral-jovem-asa-norte' \
  -H 'X-User-Email: maria.stage6@example.com' \
  http://127.0.0.1:8080/api/v1/choirs/<choir-id>/events
```

Expected result:

- the event is visible to members
- reminder rows are created in `scheduled_notifications`

Automated validation:

```bash
make test
make vet
make build
```
