package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/LeviLunique/coralhub-backend/internal/modules/choirs"
	modulefiles "github.com/LeviLunique/coralhub-backend/internal/modules/files"
	"github.com/LeviLunique/coralhub-backend/internal/modules/memberships"
	moduleusers "github.com/LeviLunique/coralhub-backend/internal/modules/users"
	"github.com/LeviLunique/coralhub-backend/internal/modules/voicekits"
	platformconfig "github.com/LeviLunique/coralhub-backend/internal/platform/config"
	"github.com/LeviLunique/coralhub-backend/internal/store/postgres/sqlc"
	"github.com/jackc/pgx/v5"
)

func TestChoirRepositoryCreateAndListByMemberUserIDIntegration(t *testing.T) {
	ctx, queries, tx := openIntegrationTestQueries(t)
	createTempChoirsTable(t, ctx, tx)
	createTempUsersTable(t, ctx, tx)
	createTempChoirMembersTable(t, ctx, tx)

	tenant := getSeedTenant(t, ctx, queries)
	userRepository := NewUserRepository(queries)
	actor, err := userRepository.Create(ctx, moduleusers.CreateParams{
		TenantID: tenant.ID,
		Email:    "ana@example.com",
		FullName: "Ana Clara",
	})
	if err != nil {
		t.Fatalf("Create actor user error = %v", err)
	}

	repository := NewChoirRepository(tx, queries)

	description := "Main choir"
	created, err := repository.Create(ctx, choirs.CreateParams{
		ActorUserID: actor.ID,
		TenantID:    tenant.ID,
		Name:        "Sopranos",
		Description: &description,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	listed, err := repository.ListByMemberUserID(ctx, tenant.ID, actor.ID)
	if err != nil {
		t.Fatalf("ListByMemberUserID() error = %v", err)
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

func TestMembershipRepositoryCreateAndListByChoirIDIntegration(t *testing.T) {
	ctx, queries, tx := openIntegrationTestQueries(t)
	createTempChoirsTable(t, ctx, tx)
	createTempUsersTable(t, ctx, tx)
	createTempChoirMembersTable(t, ctx, tx)

	tenant := getSeedTenant(t, ctx, queries)
	userRepository := NewUserRepository(queries)
	actor, err := userRepository.Create(ctx, moduleusers.CreateParams{
		TenantID: tenant.ID,
		Email:    "manager@example.com",
		FullName: "Manager",
	})
	if err != nil {
		t.Fatalf("Create manager error = %v", err)
	}

	target, err := userRepository.Create(ctx, moduleusers.CreateParams{
		TenantID: tenant.ID,
		Email:    "member@example.com",
		FullName: "Member",
	})
	if err != nil {
		t.Fatalf("Create member error = %v", err)
	}

	choirRepository := NewChoirRepository(tx, queries)
	choir, err := choirRepository.Create(ctx, choirs.CreateParams{
		ActorUserID: actor.ID,
		TenantID:    tenant.ID,
		Name:        "Altos",
	})
	if err != nil {
		t.Fatalf("Create choir error = %v", err)
	}

	repository := NewMembershipRepository(queries)
	created, err := repository.Create(ctx, memberships.CreateParams{
		TenantID: tenant.ID,
		ChoirID:  choir.ID,
		UserID:   target.ID,
		Role:     memberships.RoleMember,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	listed, err := repository.ListByChoirID(ctx, tenant.ID, choir.ID)
	if err != nil {
		t.Fatalf("ListByChoirID() error = %v", err)
	}

	if len(listed) != 2 {
		t.Fatalf("len(listed) = %d, want 2", len(listed))
	}

	if created.UserID != target.ID {
		t.Fatalf("created.UserID = %q, want %q", created.UserID, target.ID)
	}
}

func TestVoiceKitRepositoryCreateGetListAndDeleteIntegration(t *testing.T) {
	ctx, queries, tx := openIntegrationTestQueries(t)
	createTempChoirsTable(t, ctx, tx)
	createTempUsersTable(t, ctx, tx)
	createTempChoirMembersTable(t, ctx, tx)
	createTempVoiceKitsTable(t, ctx, tx)

	tenant := getSeedTenant(t, ctx, queries)
	userRepository := NewUserRepository(queries)
	actor, err := userRepository.Create(ctx, moduleusers.CreateParams{
		TenantID: tenant.ID,
		Email:    "manager@example.com",
		FullName: "Manager",
	})
	if err != nil {
		t.Fatalf("Create actor user error = %v", err)
	}

	choirRepository := NewChoirRepository(tx, queries)
	choir, err := choirRepository.Create(ctx, choirs.CreateParams{
		ActorUserID: actor.ID,
		TenantID:    tenant.ID,
		Name:        "Altos",
	})
	if err != nil {
		t.Fatalf("Create choir error = %v", err)
	}

	repository := NewVoiceKitRepository(queries)
	description := "Warmup tracks"
	created, err := repository.Create(ctx, voicekits.CreateParams{
		TenantID:    tenant.ID,
		ChoirID:     choir.ID,
		Name:        "Warmups",
		Description: &description,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := repository.GetByIDForMember(ctx, tenant.ID, created.ID, actor.ID)
	if err != nil {
		t.Fatalf("GetByIDForMember() error = %v", err)
	}

	if got.ID != created.ID {
		t.Fatalf("got.ID = %q, want %q", got.ID, created.ID)
	}

	listed, err := repository.ListByChoirID(ctx, tenant.ID, choir.ID)
	if err != nil {
		t.Fatalf("ListByChoirID() error = %v", err)
	}

	if len(listed) != 1 {
		t.Fatalf("len(listed) = %d, want 1", len(listed))
	}

	if err := repository.Delete(ctx, tenant.ID, created.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err = repository.GetByIDForMember(ctx, tenant.ID, created.ID, actor.ID)
	if !errors.Is(err, voicekits.ErrVoiceKitNotFound) {
		t.Fatalf("GetByIDForMember() after delete error = %v, want %v", err, voicekits.ErrVoiceKitNotFound)
	}
}

func TestFileRepositoryCreateListAndDeleteIntegration(t *testing.T) {
	ctx, queries, tx := openIntegrationTestQueries(t)
	createTempChoirsTable(t, ctx, tx)
	createTempUsersTable(t, ctx, tx)
	createTempChoirMembersTable(t, ctx, tx)
	createTempVoiceKitsTable(t, ctx, tx)
	createTempKitFilesTable(t, ctx, tx)

	tenant := getSeedTenant(t, ctx, queries)
	userRepository := NewUserRepository(queries)
	actor, err := userRepository.Create(ctx, moduleusers.CreateParams{
		TenantID: tenant.ID,
		Email:    "manager@example.com",
		FullName: "Manager",
	})
	if err != nil {
		t.Fatalf("Create actor user error = %v", err)
	}

	choirRepository := NewChoirRepository(tx, queries)
	choir, err := choirRepository.Create(ctx, choirs.CreateParams{
		ActorUserID: actor.ID,
		TenantID:    tenant.ID,
		Name:        "Altos",
	})
	if err != nil {
		t.Fatalf("Create choir error = %v", err)
	}

	voiceKitRepository := NewVoiceKitRepository(queries)
	voiceKit, err := voiceKitRepository.Create(ctx, voicekits.CreateParams{
		TenantID: tenant.ID,
		ChoirID:  choir.ID,
		Name:     "Warmups",
	})
	if err != nil {
		t.Fatalf("Create voice kit error = %v", err)
	}

	repository := NewFileRepository(queries)
	created, err := repository.Create(ctx, modulefiles.CreateParams{
		ID:               "8f01f767-68e5-4e99-9cc6-6dfe0fdfd1d7",
		TenantID:         tenant.ID,
		VoiceKitID:       voiceKit.ID,
		OriginalFilename: "score.pdf",
		StoredFilename:   "stored-score.pdf",
		ContentType:      "application/pdf",
		SizeBytes:        128,
		StorageKey:       "dev/tenants/coral-jovem-asa-norte/choirs/" + choir.ID + "/voice-kits/" + voiceKit.ID + "/files/file-1/stored-score.pdf",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := repository.GetByIDForMember(ctx, tenant.ID, created.ID, actor.ID)
	if err != nil {
		t.Fatalf("GetByIDForMember() error = %v", err)
	}

	if got.ID != created.ID {
		t.Fatalf("got.ID = %q, want %q", got.ID, created.ID)
	}

	listed, err := repository.ListByVoiceKitID(ctx, tenant.ID, voiceKit.ID)
	if err != nil {
		t.Fatalf("ListByVoiceKitID() error = %v", err)
	}

	if len(listed) != 1 {
		t.Fatalf("len(listed) = %d, want 1", len(listed))
	}

	if err := repository.Delete(ctx, tenant.ID, created.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err = repository.GetByIDForMember(ctx, tenant.ID, created.ID, actor.ID)
	if !errors.Is(err, modulefiles.ErrFileNotFound) {
		t.Fatalf("GetByIDForMember() after delete error = %v, want %v", err, modulefiles.ErrFileNotFound)
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

func createTempChoirMembersTable(t *testing.T, ctx context.Context, tx pgx.Tx) {
	t.Helper()

	_, err := tx.Exec(ctx, `
		CREATE TEMP TABLE choir_members (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID NOT NULL,
			choir_id UUID NOT NULL,
			user_id UUID NOT NULL,
			role TEXT NOT NULL,
			active BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CONSTRAINT choir_members_role_check CHECK (role IN ('manager', 'member')),
			CONSTRAINT choir_members_tenant_choir_user_unique UNIQUE (tenant_id, choir_id, user_id)
		) ON COMMIT DROP;
	`)
	if err != nil {
		t.Fatalf("creating temp choir_members table: %v", err)
	}
}

func createTempVoiceKitsTable(t *testing.T, ctx context.Context, tx pgx.Tx) {
	t.Helper()

	_, err := tx.Exec(ctx, `
		CREATE TEMP TABLE voice_kits (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID NOT NULL,
			choir_id UUID NOT NULL,
			name TEXT NOT NULL,
			description TEXT,
			active BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CONSTRAINT voice_kits_tenant_choir_name_unique UNIQUE (tenant_id, choir_id, name)
		) ON COMMIT DROP;
	`)
	if err != nil {
		t.Fatalf("creating temp voice_kits table: %v", err)
	}
}

func createTempKitFilesTable(t *testing.T, ctx context.Context, tx pgx.Tx) {
	t.Helper()

	_, err := tx.Exec(ctx, `
		CREATE TEMP TABLE kit_files (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID NOT NULL,
			voice_kit_id UUID NOT NULL,
			original_filename TEXT NOT NULL,
			stored_filename TEXT NOT NULL,
			content_type TEXT NOT NULL,
			size_bytes BIGINT NOT NULL,
			storage_key TEXT NOT NULL,
			active BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CONSTRAINT kit_files_size_bytes_positive CHECK (size_bytes > 0)
		) ON COMMIT DROP;
	`)
	if err != nil {
		t.Fatalf("creating temp kit_files table: %v", err)
	}
}
