package postgres

import (
	"context"
	"errors"
	"time"

	moduleaudit "github.com/LeviLunique/coralhub-backend/internal/modules/audit"
	"github.com/LeviLunique/coralhub-backend/internal/modules/memberships"
	"github.com/LeviLunique/coralhub-backend/internal/store/postgres/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type MembershipRepository struct {
	db      txBeginner
	queries *sqlc.Queries
}

func NewMembershipRepository(db txBeginner, queries *sqlc.Queries) *MembershipRepository {
	return &MembershipRepository{db: db, queries: queries}
}

func (r *MembershipRepository) Create(ctx context.Context, params memberships.CreateParams) (memberships.Membership, error) {
	tenantID, err := parseUUID(params.TenantID)
	if err != nil {
		return memberships.Membership{}, memberships.ErrInvalidTenantID
	}

	choirID, err := parseUUID(params.ChoirID)
	if err != nil {
		return memberships.Membership{}, memberships.ErrInvalidChoirID
	}

	userID, err := parseUUID(params.UserID)
	if err != nil {
		return memberships.Membership{}, memberships.ErrInvalidUserID
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return memberships.Membership{}, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	txQueries := r.queries.WithTx(tx)
	row, err := txQueries.CreateChoirMember(ctx, sqlc.CreateChoirMemberParams{
		TenantID: tenantID,
		ChoirID:  choirID,
		UserID:   userID,
		Role:     params.Role,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return memberships.Membership{}, memberships.ErrMembershipAlreadyExist
		}

		return memberships.Membership{}, err
	}

	if _, err := createAuditLog(ctx, txQueries, tenantID, row.ID, moduleaudit.ActionMembershipAdded, moduleaudit.EntityTypeMembership, stringPointer(params.ActorUserID), time.Now().UTC(), map[string]any{
		"choir_id": uuidString(row.ChoirID),
		"user_id":  uuidString(row.UserID),
		"role":     row.Role,
	}); err != nil {
		return memberships.Membership{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return memberships.Membership{}, err
	}

	return memberships.Membership{
		ID:       uuidString(row.ID),
		TenantID: uuidString(row.TenantID),
		ChoirID:  uuidString(row.ChoirID),
		UserID:   uuidString(row.UserID),
		Role:     row.Role,
		Active:   row.Active,
	}, nil
}

func (r *MembershipRepository) GetByChoirAndUser(ctx context.Context, tenantID string, choirID string, userID string) (memberships.Membership, error) {
	tenantUUID, err := parseUUID(tenantID)
	if err != nil {
		return memberships.Membership{}, memberships.ErrInvalidTenantID
	}

	choirUUID, err := parseUUID(choirID)
	if err != nil {
		return memberships.Membership{}, memberships.ErrInvalidChoirID
	}

	userUUID, err := parseUUID(userID)
	if err != nil {
		return memberships.Membership{}, memberships.ErrInvalidUserID
	}

	row, err := r.queries.GetChoirMemberByChoirAndUser(ctx, sqlc.GetChoirMemberByChoirAndUserParams{
		TenantID: tenantUUID,
		ChoirID:  choirUUID,
		UserID:   userUUID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return memberships.Membership{}, memberships.ErrMembershipNotFound
		}

		return memberships.Membership{}, err
	}

	return memberships.Membership{
		ID:       uuidString(row.ID),
		TenantID: uuidString(row.TenantID),
		ChoirID:  uuidString(row.ChoirID),
		UserID:   uuidString(row.UserID),
		Role:     row.Role,
		Active:   row.Active,
	}, nil
}

func (r *MembershipRepository) ListByChoirID(ctx context.Context, tenantID string, choirID string) ([]memberships.Membership, error) {
	tenantUUID, err := parseUUID(tenantID)
	if err != nil {
		return nil, memberships.ErrInvalidTenantID
	}

	choirUUID, err := parseUUID(choirID)
	if err != nil {
		return nil, memberships.ErrInvalidChoirID
	}

	rows, err := r.queries.ListChoirMembersByChoirID(ctx, sqlc.ListChoirMembersByChoirIDParams{
		TenantID: tenantUUID,
		ChoirID:  choirUUID,
	})
	if err != nil {
		return nil, err
	}

	items := make([]memberships.Membership, 0, len(rows))
	for _, row := range rows {
		items = append(items, memberships.Membership{
			ID:       uuidString(row.ID),
			TenantID: uuidString(row.TenantID),
			ChoirID:  uuidString(row.ChoirID),
			UserID:   uuidString(row.UserID),
			Email:    row.Email,
			FullName: row.FullName,
			Role:     row.Role,
			Active:   row.Active,
		})
	}

	return items, nil
}
