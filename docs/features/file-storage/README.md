# File Storage Integration

This document explains the Stage 5 slice implemented on branch:

- `feat/file-storage`

Commit:

- `8e6e1c9`

## Goal

Turn the Stage 4 file metadata flow into a real S3-compatible upload and download flow.

This slice adds:

- streamed upload to S3-compatible storage
- MinIO-compatible local behavior through the same adapter
- pre-signed download URL generation
- upload-time file validation
- MinIO-backed integration coverage

## Important Sequencing Note

Stage 4 intentionally stopped at metadata persistence.

Stage 5 evolves the same file endpoints instead of creating a parallel upload path:

- `POST /api/v1/voice-kits/{voiceKitID}/files` now accepts multipart upload
- `GET /api/v1/files/{fileID}/download-url` is now implemented
- delete flows now remove the storage object before deactivating metadata

This is the expected roadmap progression, not an architectural change.

## Runtime Flow

For a file upload:

1. actor context is resolved from `X-Tenant-Slug` and `X-User-Email`
2. the HTTP handler accepts multipart form data and extracts the `file` field
3. the files service validates membership, role, content type, and size
4. the service generates a file ID and tenant-aware storage key
5. the storage adapter streams the payload to S3-compatible storage
6. metadata is persisted in PostgreSQL only after storage succeeds
7. if metadata persistence fails, the uploaded object is deleted best-effort

For a download URL:

1. the service resolves visible file metadata through PostgreSQL
2. the storage adapter generates a pre-signed GET URL
3. the API returns the URL and expiration timestamp

## File-By-File Explanation

### [files.sql](/Users/levilunique/Workspace/Go/coralhub-backend/db/queries/files.sql)

The create query now accepts an explicit `id` so the object key and the database row use the same file identity.

### [storage.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/files/storage.go)

Defines the focused storage boundary used by the files module:

- put object
- delete object
- pre-sign get object

This keeps storage SDK types out of the module.

### [model.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/files/model.go)

Replaces the Stage 4 JSON metadata input with an upload input that carries:

- original filename
- content type
- size
- stream content

It also adds the download URL response model.

### [service.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/files/service.go)

This is the main Stage 5 change.

It now:

- validates upload size
- validates content type
- generates stored filenames and tenant-aware object keys
- uploads the binary stream to storage
- persists metadata only after upload succeeds
- generates pre-signed download URLs
- deletes storage objects during file deletion

Current validation rules:

- max size is `50 MiB`
- accepted types are `audio/*` and `application/pdf`

### [http.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/files/http.go)

The upload route now parses multipart form data and expects the form field:

- `file`

It also adds:

- `GET /api/v1/files/{fileID}/download-url`

### [storage.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/integrations/storage/s3/storage.go)

Implements the concrete S3-compatible adapter using the official AWS SDK.

The adapter works against:

- local MinIO
- future AWS S3-compatible production endpoints

### [storage_integration_test.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/integrations/storage/s3/storage_integration_test.go)

Adds a MinIO-backed integration test that verifies:

- object upload
- pre-signed URL generation
- HTTP retrieval through the signed URL
- object deletion

### [config.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/platform/config/config.go)

Adds a helper for normalizing the storage endpoint into a full URL for the adapter.

### [main.go](/Users/levilunique/Workspace/Go/coralhub-backend/cmd/api/main.go)

Now constructs the S3-compatible storage client and injects it into the files service.

## Request Contract

Upload now uses `multipart/form-data`.

Example:

```text
POST /api/v1/voice-kits/{voiceKitID}/files
X-Tenant-Slug: coral-jovem-asa-norte
X-User-Email: ana@example.com
Content-Type: multipart/form-data
```

Form field:

- `file`

## What This Slice Does Not Yet Do

This stage still does not implement:

- richer file categories stored in the database
- resumable or chunked uploads
- storage cleanup jobs for abandoned metadata
- lifecycle policies beyond the storage platform itself

Those remain later hardening work.
