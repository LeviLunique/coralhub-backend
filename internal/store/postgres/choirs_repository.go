package postgres

import (
	"context"
	"errors"

	"github.com/LeviLunique/coralhub-backend/internal/modules/choirs"
	"github.com/LeviLunique/coralhub-backend/internal/store/postgres/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ChoirRepository struct {
	db      txBeginner
	queries *sqlc.Queries
}

type txBeginner interface {
	Begin(context.Context) (pgx.Tx, error)
}

var _ txBeginner = (*pgxpool.Pool)(nil)

func NewChoirRepository(db txBeginner, queries *sqlc.Queries) *ChoirRepository {
	return &ChoirRepository{db: db, queries: queries}
}

func (r *ChoirRepository) Create(ctx context.Context, params choirs.CreateParams) (choirs.Choir, error) {
	tenantID, err := parseUUID(params.TenantID)
	if err != nil {
		return choirs.Choir{}, choirs.ErrInvalidTenantID
	}

	actorUserID, err := parseUUID(params.ActorUserID)
	if err != nil {
		return choirs.Choir{}, choirs.ErrInvalidActorID
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return choirs.Choir{}, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	txQueries := r.queries.WithTx(tx)

	row, err := txQueries.CreateChoir(ctx, sqlc.CreateChoirParams{
		TenantID:    tenantID,
		Name:        params.Name,
		Description: textValue(params.Description),
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return choirs.Choir{}, choirs.ErrChoirNameTaken
		}

		return choirs.Choir{}, err
	}

	if _, err := txQueries.CreateChoirMember(ctx, sqlc.CreateChoirMemberParams{
		TenantID: tenantID,
		ChoirID:  row.ID,
		UserID:   actorUserID,
		Role:     "manager",
	}); err != nil {
		return choirs.Choir{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return choirs.Choir{}, err
	}

	return choirs.Choir{
		ID:          uuidString(row.ID),
		TenantID:    uuidString(row.TenantID),
		Name:        row.Name,
		Description: textPointer(row.Description),
		Active:      row.Active,
	}, nil
}

func (r *ChoirRepository) GetByIDForMember(ctx context.Context, tenantID string, choirID string, userID string) (choirs.Choir, error) {
	tenantUUID, err := parseUUID(tenantID)
	if err != nil {
		return choirs.Choir{}, choirs.ErrInvalidTenantID
	}

	choirUUID, err := parseUUID(choirID)
	if err != nil {
		return choirs.Choir{}, choirs.ErrInvalidChoirID
	}

	userUUID, err := parseUUID(userID)
	if err != nil {
		return choirs.Choir{}, choirs.ErrInvalidActorID
	}

	row, err := r.queries.GetChoirByIDForMember(ctx, sqlc.GetChoirByIDForMemberParams{
		TenantID: tenantUUID,
		ID:       choirUUID,
		UserID:   userUUID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return choirs.Choir{}, choirs.ErrChoirNotFound
		}

		return choirs.Choir{}, err
	}

	return choirs.Choir{
		ID:          uuidString(row.ID),
		TenantID:    uuidString(row.TenantID),
		Name:        row.Name,
		Description: textPointer(row.Description),
		Active:      row.Active,
	}, nil
}

func (r *ChoirRepository) ListByMemberUserID(ctx context.Context, tenantID string, userID string) ([]choirs.Choir, error) {
	tenantUUID, err := parseUUID(tenantID)
	if err != nil {
		return nil, choirs.ErrInvalidTenantID
	}

	userUUID, err := parseUUID(userID)
	if err != nil {
		return nil, choirs.ErrInvalidActorID
	}

	rows, err := r.queries.ListChoirsByMemberUserID(ctx, sqlc.ListChoirsByMemberUserIDParams{
		TenantID: tenantUUID,
		UserID:   userUUID,
	})
	if err != nil {
		return nil, err
	}

	items := make([]choirs.Choir, 0, len(rows))
	for _, row := range rows {
		items = append(items, choirs.Choir{
			ID:          uuidString(row.ID),
			TenantID:    uuidString(row.TenantID),
			Name:        row.Name,
			Description: textPointer(row.Description),
			Active:      row.Active,
		})
	}

	return items, nil
}
