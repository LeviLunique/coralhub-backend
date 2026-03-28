package postgres

import (
	"context"
	"errors"

	"github.com/LeviLunique/coralhub-backend/internal/modules/choirs"
	"github.com/LeviLunique/coralhub-backend/internal/store/postgres/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type ChoirRepository struct {
	queries *sqlc.Queries
}

func NewChoirRepository(queries *sqlc.Queries) *ChoirRepository {
	return &ChoirRepository{queries: queries}
}

func (r *ChoirRepository) Create(ctx context.Context, params choirs.CreateParams) (choirs.Choir, error) {
	tenantID, err := parseUUID(params.TenantID)
	if err != nil {
		return choirs.Choir{}, choirs.ErrInvalidTenantID
	}

	row, err := r.queries.CreateChoir(ctx, sqlc.CreateChoirParams{
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

	return choirs.Choir{
		ID:          uuidString(row.ID),
		TenantID:    uuidString(row.TenantID),
		Name:        row.Name,
		Description: textPointer(row.Description),
		Active:      row.Active,
	}, nil
}

func (r *ChoirRepository) GetByID(ctx context.Context, tenantID string, choirID string) (choirs.Choir, error) {
	tenantUUID, err := parseUUID(tenantID)
	if err != nil {
		return choirs.Choir{}, choirs.ErrInvalidTenantID
	}

	choirUUID, err := parseUUID(choirID)
	if err != nil {
		return choirs.Choir{}, choirs.ErrInvalidChoirID
	}

	row, err := r.queries.GetChoirByID(ctx, sqlc.GetChoirByIDParams{
		TenantID: tenantUUID,
		ID:       choirUUID,
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

func (r *ChoirRepository) ListByTenantID(ctx context.Context, tenantID string) ([]choirs.Choir, error) {
	tenantUUID, err := parseUUID(tenantID)
	if err != nil {
		return nil, choirs.ErrInvalidTenantID
	}

	rows, err := r.queries.ListChoirsByTenantID(ctx, tenantUUID)
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
