# Voice Kits And File Metadata Testing Guide

This file shows how to verify the Stage 4 slice on branch:

- `feat/voice-kits-file-metadata`

## 1. Confirm You Are On The Correct Branch

Run:

```bash
git branch --show-current
```

Expected:

```text
feat/voice-kits-file-metadata
```

## 2. Start Local Dependencies

Run:

```bash
make compose-up
```

## 3. Apply The Required Migrations

This repository still does not have a migration runner wired into the app.
Apply the Stage 2, Stage 3, and Stage 4 migrations manually:

```bash
PGPASSWORD='<your DB_PASSWORD from .env>' psql -h localhost -p 5433 -U coralhub -d coralhub -f db/migrations/000002_init_choirs_and_users.up.sql
PGPASSWORD='<your DB_PASSWORD from .env>' psql -h localhost -p 5433 -U coralhub -d coralhub -f db/migrations/000003_init_choir_members.up.sql
PGPASSWORD='<your DB_PASSWORD from .env>' psql -h localhost -p 5433 -U coralhub -d coralhub -f db/migrations/000004_init_voice_kits_and_kit_files.up.sql
```

Then confirm:

```bash
PGPASSWORD='<your DB_PASSWORD from .env>' psql -h localhost -p 5433 -U coralhub -d coralhub -c "\dt"
```

Expected:

- `voice_kits` exists
- `kit_files` exists

## 4. Start The API

Run:

```bash
make run-api
```

Leave it running in that terminal.

## 5. Create Two Users

Use addresses that are unlikely to conflict with earlier smoke tests:

```bash
curl -s -H 'X-Tenant-Slug: coral-jovem-asa-norte' -H 'Content-Type: application/json' \
  -d '{"email":"ana.stage4@example.com","full_name":"Ana Stage 4"}' \
  http://127.0.0.1:8080/api/v1/users
```

Then:

```bash
curl -s -H 'X-Tenant-Slug: coral-jovem-asa-norte' -H 'Content-Type: application/json' \
  -d '{"email":"maria.stage4@example.com","full_name":"Maria Stage 4"}' \
  http://127.0.0.1:8080/api/v1/users
```

Save the returned user IDs if you want to inspect the data later.

## 6. Create A Choir As The First Manager

Run:

```bash
curl -s -i \
  -H 'X-Tenant-Slug: coral-jovem-asa-norte' \
  -H 'X-User-Email: ana.stage4@example.com' \
  -H 'Content-Type: application/json' \
  -d '{"name":"Stage 4 Choir"}' \
  http://127.0.0.1:8080/api/v1/choirs
```

Expected:

- HTTP `201`
- JSON body with a generated choir `id`

Save the choir `id`.

## 7. Create A Voice Kit

Run:

```bash
curl -s -i \
  -H 'X-Tenant-Slug: coral-jovem-asa-norte' \
  -H 'X-User-Email: ana.stage4@example.com' \
  -H 'Content-Type: application/json' \
  -d '{"name":"Warmups","description":"Rehearsal exercises"}' \
  http://127.0.0.1:8080/api/v1/choirs/<choir-id>/voice-kits
```

Expected:

- HTTP `201`
- JSON body with:
  - generated voice kit `id`
  - `name = "Warmups"`
  - the saved `choir_id`

Save the voice kit `id`.

## 8. List Voice Kits For The Choir

Run:

```bash
curl -s \
  -H 'X-Tenant-Slug: coral-jovem-asa-norte' \
  -H 'X-User-Email: ana.stage4@example.com' \
  http://127.0.0.1:8080/api/v1/choirs/<choir-id>/voice-kits
```

Expected:

- JSON object with `items`
- one item with `name = "Warmups"`

## 9. Get The Voice Kit By ID

Run:

```bash
curl -s \
  -H 'X-Tenant-Slug: coral-jovem-asa-norte' \
  -H 'X-User-Email: ana.stage4@example.com' \
  http://127.0.0.1:8080/api/v1/voice-kits/<voice-kit-id>
```

Expected:

- JSON object for the saved voice kit

## 10. Verify A Non-Member Cannot See The Choir Voice Kits

Run:

```bash
curl -s -i \
  -H 'X-Tenant-Slug: coral-jovem-asa-norte' \
  -H 'X-User-Email: maria.stage4@example.com' \
  http://127.0.0.1:8080/api/v1/choirs/<choir-id>/voice-kits
```

Expected:

- HTTP `403`

## 11. Create File Metadata For The Voice Kit

Run:

```bash
curl -s -i \
  -H 'X-Tenant-Slug: coral-jovem-asa-norte' \
  -H 'X-User-Email: ana.stage4@example.com' \
  -H 'Content-Type: application/json' \
  -d '{"original_filename":"warmup.mp3","stored_filename":"stored-warmup.mp3","content_type":"audio/mpeg","size_bytes":1024,"storage_key":"dev/tenants/coral-jovem-asa-norte/choirs/<choir-id>/voice-kits/<voice-kit-id>/files/file-1/stored-warmup.mp3"}' \
  http://127.0.0.1:8080/api/v1/voice-kits/<voice-kit-id>/files
```

Expected:

- HTTP `201`
- JSON body with:
  - generated file metadata `id`
  - `voice_kit_id` matching the saved voice kit
  - `original_filename = "warmup.mp3"`

Save the file metadata `id`.

## 12. List File Metadata

Run:

```bash
curl -s \
  -H 'X-Tenant-Slug: coral-jovem-asa-norte' \
  -H 'X-User-Email: ana.stage4@example.com' \
  http://127.0.0.1:8080/api/v1/voice-kits/<voice-kit-id>/files
```

Expected:

- JSON object with `items`
- one item with `original_filename = "warmup.mp3"`

## 13. Delete File Metadata

Run:

```bash
curl -s -i \
  -H 'X-Tenant-Slug: coral-jovem-asa-norte' \
  -H 'X-User-Email: ana.stage4@example.com' \
  -X DELETE \
  http://127.0.0.1:8080/api/v1/files/<file-id>
```

Expected:

- HTTP `204`

Then confirm the list is empty:

```bash
curl -s \
  -H 'X-Tenant-Slug: coral-jovem-asa-norte' \
  -H 'X-User-Email: ana.stage4@example.com' \
  http://127.0.0.1:8080/api/v1/voice-kits/<voice-kit-id>/files
```

Expected:

- `items` is empty

## 14. Delete The Voice Kit

Run:

```bash
curl -s -i \
  -H 'X-Tenant-Slug: coral-jovem-asa-norte' \
  -H 'X-User-Email: ana.stage4@example.com' \
  -X DELETE \
  http://127.0.0.1:8080/api/v1/voice-kits/<voice-kit-id>
```

Expected:

- HTTP `204`

## 15. Run Automated Checks

Run:

```bash
env GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./...
env GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go vet ./...
env GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go build ./cmd/api ./cmd/worker
```

Expected:

- all commands succeed

## 16. Troubleshooting

### Creating users fails with `409`

You probably already created the same email in an earlier manual test.
Use a different email value and retry.

### Voice kit creation returns `403`

Check:

- the choir was created by the same `X-User-Email`
- the request includes both `X-Tenant-Slug` and `X-User-Email`

### File metadata creation returns `404 voice kit not found`

Check:

- the voice kit ID is correct
- the actor is a member of the owning choir
- the voice kit was not already deleted

### File list stays empty after create

Check:

- the create request returned `201`
- you are listing the same `voice-kit-id`
- the file was not deleted immediately after creation
