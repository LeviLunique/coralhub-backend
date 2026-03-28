package postgres

import (
	"context"
	"testing"

	"github.com/LeviLunique/coralhub-backend/internal/modules/choirs"
	moduleusers "github.com/LeviLunique/coralhub-backend/internal/modules/users"
	platformconfig "github.com/LeviLunique/coralhub-backend/internal/platform/config"
	"github.com/LeviLunique/coralhub-backend/internal/store/postgres/sqlc"
	"github.com/jackc/pgx/v5"
)

func TestChoirRepositoryCreateAndListByTenantIDIntegration(t *testing.T) {
	ctx, queries, tx := openIntegrationTestQueries(t)
	createTempChoirsTable(t, ctx, tx)

	tenant := getSeedTenant(t, ctx, queries)
	repository := NewChoirRepository(queries)

	description := "Main choir"
	created, err := repository.Create(ctx, choirs.CreateParams{
		TenantID:    tenant.ID,
		Name:        "Sopranos",
		Description: &description,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	listed, err := repository.ListByTenantID(ctx, tenant.ID)
	if err != nil {
		t.Fatalf("ListByTenantID() error = %v", err)
	}

	if len(listed) != 1 {
		t.Fatalf("len(listed) = %d, want 1", len(listed))
	}

	if listed[0].ID != created.ID {
		t.Fatalf("listed[0].ID = %q, want %q", listed[0].ID, created.ID)
	}
}

func TestUserRepositoryCreateAndListByTenantIDIntegration(t *testing.T) {
	ctx, queries, tx := openIntegrationTestQueries(t)
	createTempUsersTable(t, ctx, tx)

	tenant := getSeedTenant(t, ctx, queries)
	repository := NewUserRepository(queries)

	created, err := repository.Create(ctx, moduleusers.CreateParams{
		TenantID: tenant.ID,
		Email:    "ana@example.com",
		FullName: "Ana Clara",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	listed, err := repository.ListByTenantID(ctx, tenant.ID)
	if err != nil {
		t.Fatalf("ListByTenantID() error = %v", err)
	}

	if len(listed) != 1 {
		t.Fatalf("len(listed) = %d, want 1", len(listed))
	}

	if listed[0].ID != created.ID {
		t.Fatalf("listed[0].ID = %q, want %q", listed[0].ID, created.ID)
	}
}

func openIntegrationTestQueries(t *testing.T) (context.Context, *sqlc.Queries, pgx.Tx) {
	t.Helper()

	cfg, err := platformconfig.Load()
	if err != nil {
		t.Skipf("integration config unavailable: %v", err)
	}

	ctx := context.Background()
	pool, err := NewPool(ctx, cfg.Database)
	if err != nil {
		t.Skipf("postgres unavailable for integration test: %v", err)
	}
	t.Cleanup(pool.Close)

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("Begin() error = %v", err)
	}

	t.Cleanup(func() {
		_ = tx.Rollback(ctx)
	})

	return ctx, sqlc.New(tx), tx
}

func getSeedTenant(t *testing.T, ctx context.Context, queries *sqlc.Queries) struct{ ID string } {
	t.Helper()

	row, err := queries.GetTenantBySlug(ctx, "coral-jovem-asa-norte")
	if err != nil {
		t.Fatalf("GetTenantBySlug() error = %v", err)
	}

	return struct{ ID string }{ID: uuidString(row.ID)}
}

func createTempChoirsTable(t *testing.T, ctx context.Context, tx pgx.Tx) {
	t.Helper()

	_, err := tx.Exec(ctx, `
		CREATE TEMP TABLE choirs (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID NOT NULL,
			name TEXT NOT NULL,
			description TEXT,
			active BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CONSTRAINT choirs_tenant_name_unique UNIQUE (tenant_id, name)
		) ON COMMIT DROP;
	`)
	if err != nil {
		t.Fatalf("creating temp choirs table: %v", err)
	}
}

func createTempUsersTable(t *testing.T, ctx context.Context, tx pgx.Tx) {
	t.Helper()

	_, err := tx.Exec(ctx, `
		CREATE TEMP TABLE users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID NOT NULL,
			email TEXT NOT NULL,
			full_name TEXT NOT NULL,
			active BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CONSTRAINT users_tenant_email_unique UNIQUE (tenant_id, email)
		) ON COMMIT DROP;
	`)
	if err != nil {
		t.Fatalf("creating temp users table: %v", err)
	}
}
