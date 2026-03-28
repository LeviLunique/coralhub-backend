# Voice Kits And File Metadata

This document explains the Stage 4 slice implemented on branch:

- `feat/voice-kits-file-metadata`

Commit:

- `fc3d738`

## Goal

Add the first tenant-scoped metadata flows for choir rehearsal assets without pulling in real S3 or MinIO upload logic yet.

This slice adds:

- `voice_kits` persistence and HTTP flows
- `kit_files` metadata persistence and HTTP flows
- membership-aware access to voice kits
- manager-only write and delete operations for voice kits and file metadata

## Important Sequencing Note

The implementation guide separates Stage 4 metadata from Stage 5 storage integration.

This slice keeps that boundary explicit:

- Stage 4 stores only metadata in PostgreSQL
- Stage 4 does not upload binaries
- Stage 4 does not generate download URLs
- Stage 5 should plug real S3 or MinIO operations into the metadata already created here

This keeps the current change small and aligned with the roadmap.

## Runtime Flow

For actor-protected voice kit and file routes:

1. middleware resolves tenant context from `X-Tenant-Slug`
2. middleware resolves the actor inside that tenant from `X-User-Email`
3. handlers read tenant and actor from request context
4. services validate IDs and input fields
5. services enforce membership and manager-role rules
6. repositories execute explicit tenant-scoped SQL through `sqlc`

Authorization in this slice is intentionally simple:

- choir members can list voice kits
- choir members can read a voice kit
- choir members can list file metadata for a voice kit
- only choir managers can create voice kits
- only choir managers can create file metadata
- only choir managers can delete voice kits
- only choir managers can delete file metadata

## File-By-File Explanation

### [000004_init_voice_kits_and_kit_files.up.sql](/Users/levilunique/Workspace/Go/coralhub-backend/db/migrations/000004_init_voice_kits_and_kit_files.up.sql)

Adds:

- `voice_kits`
- `kit_files`

Important schema choices:

- both tables include `tenant_id`
- `voice_kits` is scoped to a `choir_id`
- voice kit names are unique inside a choir through `(tenant_id, choir_id, name)`
- `kit_files` stores metadata only
- `kit_files` uses `active` for soft-delete behavior
- `size_bytes` must be positive

### [voice_kits.sql](/Users/levilunique/Workspace/Go/coralhub-backend/db/queries/voice_kits.sql)

Adds explicit SQL for:

- create voice kit
- get a voice kit only if the actor is a choir member
- list voice kits for a choir
- soft-delete a voice kit

### [files.sql](/Users/levilunique/Workspace/Go/coralhub-backend/db/queries/files.sql)

Adds explicit SQL for:

- create file metadata
- get file metadata only if the actor is a choir member
- list file metadata for a voice kit
- soft-delete file metadata

### [model.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/voicekits/model.go)

Defines the public voice kit model and create input.

### [service.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/voicekits/service.go)

Implements voice kit validation and authorization.

Key rules:

- tenant, choir, actor, and voice kit IDs must be present where required
- voice kit name must be non-blank
- description is trimmed and optional
- create and delete require `manager` role in the choir
- get and list require only choir membership

### [repository.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/voicekits/repository.go)

Defines the focused repository contract used by the voice kit service.

### [http.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/voicekits/http.go)

Registers the new voice kit routes:

- `POST /api/v1/choirs/{choirID}/voice-kits`
- `GET /api/v1/choirs/{choirID}/voice-kits`
- `GET /api/v1/voice-kits/{voiceKitID}`
- `DELETE /api/v1/voice-kits/{voiceKitID}`

Handlers stay thin and translate service errors into HTTP responses.

### [model.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/files/model.go)

Defines the file metadata model and create input.

### [service.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/files/service.go)

Implements file metadata validation and authorization.

Key rules:

- voice kit, actor, and tenant IDs must be present
- original filename, stored filename, content type, and storage key must be non-blank
- `size_bytes` must be greater than zero
- create requires a visible voice kit plus manager role in the owning choir
- list requires only visibility of the voice kit
- delete requires manager role in the owning choir

### [repository.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/files/repository.go)

Defines the focused repository contract for file metadata.

### [http.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/files/http.go)

Registers the new file metadata routes:

- `POST /api/v1/voice-kits/{voiceKitID}/files`
- `GET /api/v1/voice-kits/{voiceKitID}/files`
- `DELETE /api/v1/files/{fileID}`

### [voicekits_repository.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/store/postgres/voicekits_repository.go)

Implements PostgreSQL persistence for voice kits.

Important mappings:

- duplicate voice kit names become `ErrVoiceKitNameTaken`
- member-scoped lookup translates no rows into `ErrVoiceKitNotFound`
- delete is implemented as a soft delete

### [files_repository.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/store/postgres/files_repository.go)

Implements PostgreSQL persistence for file metadata.

Important mappings:

- file visibility is member-scoped through the owning voice kit and choir
- delete is implemented as a soft delete
- metadata is stored without any storage SDK dependency

### [router.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/platform/http/router.go)

Now wires the actor-protected Stage 4 routes into the existing authenticated group.

### [main.go](/Users/levilunique/Workspace/Go/coralhub-backend/cmd/api/main.go)

Composes the new repositories and services into the API process.

### [repositories_integration_test.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/store/postgres/repositories_integration_test.go)

Adds PostgreSQL-backed integration tests for:

- voice kit create, get, list, and delete
- file metadata create, get, list, and delete

## New Protected Request Contract

Voice kit and file metadata routes require:

```text
X-Tenant-Slug: coral-jovem-asa-norte
X-User-Email: ana@example.com
```

## What This Slice Does Not Yet Do

This stage still does not implement:

- binary file upload to MinIO or S3
- pre-signed download URLs
- content-type allowlists or size-limit policy beyond positive byte count
- cleanup of storage objects during delete flows
- richer voice kit updates

Those remain Stage 5 or later work.
