# Audit History

This guide explains the audit history feature and how it is verified.

## What It Does

This feature adds tenant-scoped audit history for critical business transitions.

Current audited areas include:

- membership changes
- event lifecycle changes
- reminder generation
- notification delivery outcomes

## How It Works

Flow:

1. a transactional repository changes business state
2. an audit row is written in the same transaction
3. the transaction commits with business state and audit history aligned

The backend stores audit history at the repository layer without exposing a public audit endpoint.

## Why It Matters

This gives the system selective business history without adopting event sourcing.
It preserves useful operational traceability while keeping the architecture simple.

## How To Verify

Start dependencies:

```bash
make compose-up
```

Confirm the table exists:

```bash
PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME"
```

Inside `psql`:

```sql
\dt audit_log
\d audit_log
```

Expected result:

- `audit_log` exists
- the table includes tenant and entity context, action, actor, timestamp, and payload fields

Automated validation:

```bash
make test
make vet
make build
```

Useful focused checks:

- membership repository integration tests
- event repository integration tests
- notification repository integration tests
