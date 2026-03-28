# File Storage Integration Testing Guide

This file shows how to verify the Stage 5 slice on branch:

- `feat/file-storage`

## 1. Confirm You Are On The Correct Branch

Run:

```bash
git branch --show-current
```

Expected:

```text
feat/file-storage
```

## 2. Start Local Dependencies

Run:

```bash
make compose-up
```

Expected:

- PostgreSQL is running
- MinIO is running

## 3. Apply The Required Migrations

Run:

```bash
PGPASSWORD='<your DB_PASSWORD from .env>' psql -h localhost -p 5433 -U coralhub -d coralhub -f db/migrations/000002_init_choirs_and_users.up.sql
PGPASSWORD='<your DB_PASSWORD from .env>' psql -h localhost -p 5433 -U coralhub -d coralhub -f db/migrations/000003_init_choir_members.up.sql
PGPASSWORD='<your DB_PASSWORD from .env>' psql -h localhost -p 5433 -U coralhub -d coralhub -f db/migrations/000004_init_voice_kits_and_kit_files.up.sql
```

## 4. Start The API

Run:

```bash
make run-api
```

Leave it running.

## 5. Create The Stage 5 Base Data

Create two users:

```bash
curl -s -H 'X-Tenant-Slug: coral-jovem-asa-norte' -H 'Content-Type: application/json' \
  -d '{"email":"ana.stage5@example.com","full_name":"Ana Stage 5"}' \
  http://127.0.0.1:8080/api/v1/users
```

```bash
curl -s -H 'X-Tenant-Slug: coral-jovem-asa-norte' -H 'Content-Type: application/json' \
  -d '{"email":"maria.stage5@example.com","full_name":"Maria Stage 5"}' \
  http://127.0.0.1:8080/api/v1/users
```

Create a choir:

```bash
curl -s -H 'X-Tenant-Slug: coral-jovem-asa-norte' -H 'X-User-Email: ana.stage5@example.com' \
  -H 'Content-Type: application/json' -d '{"name":"Stage 5 Choir"}' \
  http://127.0.0.1:8080/api/v1/choirs
```

Create a voice kit with the returned choir ID:

```bash
curl -s -H 'X-Tenant-Slug: coral-jovem-asa-norte' -H 'X-User-Email: ana.stage5@example.com' \
  -H 'Content-Type: application/json' -d '{"name":"Stage 5 Warmups"}' \
  http://127.0.0.1:8080/api/v1/choirs/<choir-id>/voice-kits
```

Save the returned `voice_kit_id`.

## 6. Upload A File

Create a local sample file:

```bash
printf 'stage5-audio-sample' > /tmp/stage5-sample.mp3
```

Upload it:

```bash
curl -s -i \
  -H 'X-Tenant-Slug: coral-jovem-asa-norte' \
  -H 'X-User-Email: ana.stage5@example.com' \
  -F 'file=@/tmp/stage5-sample.mp3;type=audio/mpeg' \
  http://127.0.0.1:8080/api/v1/voice-kits/<voice-kit-id>/files
```

Expected:

- HTTP `201`
- JSON body with:
  - generated file `id`
  - generated `stored_filename`
  - non-empty `storage_key`

Save the returned `file_id`.

## 7. List File Metadata

Run:

```bash
curl -s \
  -H 'X-Tenant-Slug: coral-jovem-asa-norte' \
  -H 'X-User-Email: ana.stage5@example.com' \
  http://127.0.0.1:8080/api/v1/voice-kits/<voice-kit-id>/files
```

Expected:

- one item with the uploaded filename metadata

## 8. Request A Download URL

Run:

```bash
curl -s \
  -H 'X-Tenant-Slug: coral-jovem-asa-norte' \
  -H 'X-User-Email: ana.stage5@example.com' \
  http://127.0.0.1:8080/api/v1/files/<file-id>/download-url
```

Expected:

- JSON with:
  - `url`
  - `expires_at`

## 9. Fetch The Uploaded Content Through The Signed URL

Copy the returned `url` and run:

```bash
curl -s '<signed-url>'
```

Expected:

- response body equals:

```text
stage5-audio-sample
```

## 10. Verify Access Control

Try to list the voice kit files as the non-member:

```bash
curl -s -i \
  -H 'X-Tenant-Slug: coral-jovem-asa-norte' \
  -H 'X-User-Email: maria.stage5@example.com' \
  http://127.0.0.1:8080/api/v1/voice-kits/<voice-kit-id>/files
```

Expected:

- HTTP `404` or `403` depending on whether the actor can resolve the voice kit membership path

## 11. Delete The File

Run:

```bash
curl -s -i \
  -H 'X-Tenant-Slug: coral-jovem-asa-norte' \
  -H 'X-User-Email: ana.stage5@example.com' \
  -X DELETE \
  http://127.0.0.1:8080/api/v1/files/<file-id>
```

Expected:

- HTTP `204`

Then list files again and confirm `items` is empty.

## 12. Run Automated Checks

Run:

```bash
env GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./...
env GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go vet ./...
env GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go build ./cmd/api ./cmd/worker
```

Expected:

- all commands succeed
- the package `internal/integrations/storage/s3` passes with MinIO running
