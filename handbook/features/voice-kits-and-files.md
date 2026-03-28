# Voice Kits and Files

This guide explains the voice kit and file flows delivered across the metadata and storage slices.

## What It Does

This feature area adds:

- voice kit CRUD
- file metadata
- S3-compatible uploads
- pre-signed download URLs
- manager-only write and delete rules
- membership-aware read visibility

Representative endpoints:

- `POST /api/v1/choirs/{choirID}/voice-kits`
- `GET /api/v1/choirs/{choirID}/voice-kits`
- `POST /api/v1/voice-kits/{voiceKitID}/files`
- `GET /api/v1/files/{fileID}/download-url`

## How It Works

Flow:

1. the actor is resolved in tenant context
2. services check choir membership and role
3. voice kit metadata is stored in PostgreSQL
4. file uploads stream to S3-compatible storage
5. file metadata is stored only after upload succeeds
6. downloads use pre-signed URLs

The feature is split conceptually into metadata and storage, but developers should think of it as one rehearsal-asset workflow.

## Why It Matters

This feature provides the first real binary asset lifecycle in the backend.
It combines tenant-safe metadata, access control, and storage behavior.

## How To Verify

Start dependencies and the API:

```bash
make compose-up
make run-api
```

Create base data:

- users
- a choir
- a voice kit

Upload a file:

```bash
printf 'stage-sample' > /tmp/stage-sample.mp3
curl -s -i \
  -H 'X-Tenant-Slug: coral-jovem-asa-norte' \
  -H 'X-User-Email: ana.stage5@example.com' \
  -F 'file=@/tmp/stage-sample.mp3;type=audio/mpeg' \
  http://127.0.0.1:8080/api/v1/voice-kits/<voice-kit-id>/files
```

Request a download URL:

```bash
curl -s \
  -H 'X-Tenant-Slug: coral-jovem-asa-norte' \
  -H 'X-User-Email: ana.stage5@example.com' \
  http://127.0.0.1:8080/api/v1/files/<file-id>/download-url
```

Expected result:

- upload returns a generated file ID and storage key
- download URL response contains a usable pre-signed URL

Automated validation:

```bash
make test
make vet
make build
```
