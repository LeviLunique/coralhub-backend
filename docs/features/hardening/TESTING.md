# Hardening Testing

This guide verifies the Stage 10 hardening slice on branch:

- `feat/hardening`

## 1. Run Automated Validation

From the repo root:

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
- `go vet` produces no findings
- both binaries build successfully

## 2. Check the Metrics Endpoint

Start the API:

```bash
make run-api
```

In another terminal:

```bash
curl http://127.0.0.1:8080/metrics
```

Expected output includes lines like:

```text
coralhub_http_requests_total
coralhub_worker_polls_total
coralhub_storage_upload_failures_total
```

## 3. Check Structured Error Responses

Call a protected route without the required tenant header:

```bash
curl -i http://127.0.0.1:8080/api/v1/users
```

Expected:

- status `400`
- JSON error body with:
  - `error.code`
  - `error.message`
  - `error.request_id`

## 4. Check Strict JSON Decoding

Call a JSON route with an unexpected field:

```bash
curl -i \
  -H 'Content-Type: application/json' \
  -H 'X-Tenant-Slug: coral-jovem-asa-norte' \
  -H 'X-User-Email: ana@example.com' \
  -d '{"name":"Sopranos","unexpected":true}' \
  http://127.0.0.1:8080/api/v1/choirs
```

Expected:

- status `400`
- structured error response

## 5. Check Configurable Timeout

Set a custom timeout in `.env`:

```env
HTTP_HANDLER_TIMEOUT=5s
```

Restart the API and confirm it still starts successfully.

This slice does not add a dedicated timeout demo endpoint, so timeout behavior is primarily validated through config loading and middleware wiring rather than a manual long-running request.

## 6. Check Notification Retention Cleanup

Run the worker:

```bash
make run-worker
```

The automated PostgreSQL integration test already validates that terminal notifications older than the retention cutoff are deleted while newer or pending rows remain.

To rerun just that coverage:

```bash
env GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./internal/store/postgres -run 'TestNotificationRepositoryCleanupTerminalBeforeIntegration' -count=1
```

Expected:

- the test passes

## 7. Check the New Config Defaults

The config tests verify the default values for:

- `HTTP_HANDLER_TIMEOUT`
- `WORKER_NOTIFICATION_RETENTION`

To rerun just config coverage:

```bash
env GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./internal/platform/config -count=1
```

Expected:

- the test passes
