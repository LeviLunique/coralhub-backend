# Hardening

This guide explains the current operational hardening slice.

## What It Does

This feature adds:

- structured JSON error envelopes
- strict JSON decoding
- configurable HTTP handler timeout
- a `/metrics` endpoint
- notification retention cleanup

## How It Works

The hardening work improves the existing slices without redesigning them.

Key behaviors:

- handlers return stable error shapes
- malformed or unexpected JSON is rejected early
- timeout behavior is configuration-driven
- metrics expose basic API, worker, notification, and storage signals
- old terminal notification rows are cleaned up by the worker

## Why It Matters

This slice raises the operational credibility of the backend and makes failures easier to observe and debug.

## How To Verify

Start the API:

```bash
make run-api
```

Check the metrics endpoint:

```bash
curl http://127.0.0.1:8080/metrics
```

Expected result:

- Prometheus-style metrics output

Check structured errors:

```bash
curl -i http://127.0.0.1:8080/api/v1/users
```

Expected result:

- HTTP `400`
- structured JSON error body

Automated validation:

```bash
make test
make vet
make build
```
