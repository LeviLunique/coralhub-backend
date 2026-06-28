package postgres

import (
	"context"
	"errors"

	"github.com/LeviLunique/coralhub-backend/internal/modules/instruments"
	"github.com/LeviLunique/coralhub-backend/internal/store/postgres/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type InstrumentRepository struct {
	queries *sqlc.Queries
}

func NewInstrumentRepository(queries *sqlc.Queries) *InstrumentRepository {
	return &InstrumentRepository{queries: queries}
}

func (r *InstrumentRepository) Create(ctx context.Context, params instruments.CreateParams) (instruments.Instrument, error) {
	tenantID, err := parseUUID(params.TenantID)
	if err != nil {
		return instruments.Instrument{}, instruments.ErrInvalidTenantID
	}

	choirID, err := parseUUID(params.ChoirID)
	if err != nil {
		return instruments.Instrument{}, instruments.ErrInvalidChoirID
	}

	row, err := r.queries.CreateInstrument(ctx, sqlc.CreateInstrumentParams{
		TenantID:    tenantID,
		ChoirID:     choirID,
		Name:        params.Name,
		Description: textValue(params.Description),
		Icon:        textValue(params.Icon),
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return instruments.Instrument{}, instruments.ErrInstrumentNameTaken
		}

		return instruments.Instrument{}, err
	}

	return mapInstrumentRow(sqlc.Instrument(row)), nil
}

func (r *InstrumentRepository) Update(ctx context.Context, params instruments.UpdateParams) (instruments.Instrument, error) {
	tenantID, err := parseUUID(params.TenantID)
	if err != nil {
		return instruments.Instrument{}, instruments.ErrInvalidTenantID
	}

	instrumentID, err := parseUUID(params.InstrumentID)
	if err != nil {
		return instruments.Instrument{}, instruments.ErrInvalidInstrumentID
	}

	row, err := r.queries.UpdateInstrument(ctx, sqlc.UpdateInstrumentParams{
		TenantID:    tenantID,
		ID:          instrumentID,
		Name:        params.Name,
		Description: textValue(params.Description),
		Icon:        textValue(params.Icon),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return instruments.Instrument{}, instruments.ErrInstrumentNotFound
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return instruments.Instrument{}, instruments.ErrInstrumentNameTaken
		}

		return instruments.Instrument{}, err
	}

	return mapInstrumentRow(sqlc.Instrument(row)), nil
}

func (r *InstrumentRepository) GetByIDForMember(ctx context.Context, tenantID string, instrumentID string, userID string) (instruments.Instrument, error) {
	tenantUUID, err := parseUUID(tenantID)
	if err != nil {
		return instruments.Instrument{}, instruments.ErrInvalidTenantID
	}

	instrumentUUID, err := parseUUID(instrumentID)
	if err != nil {
		return instruments.Instrument{}, instruments.ErrInvalidInstrumentID
	}

	userUUID, err := parseUUID(userID)
	if err != nil {
		return instruments.Instrument{}, instruments.ErrInvalidActorID
	}

	row, err := r.queries.GetInstrumentByIDForMember(ctx, sqlc.GetInstrumentByIDForMemberParams{
		TenantID: tenantUUID,
		ID:       instrumentUUID,
		UserID:   userUUID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return instruments.Instrument{}, instruments.ErrInstrumentNotFound
		}

		return instruments.Instrument{}, err
	}

	return mapInstrumentRow(sqlc.Instrument(row)), nil
}

func (r *InstrumentRepository) ListByChoirID(ctx context.Context, tenantID string, choirID string) ([]instruments.Instrument, error) {
	tenantUUID, err := parseUUID(tenantID)
	if err != nil {
		return nil, instruments.ErrInvalidTenantID
	}

	choirUUID, err := parseUUID(choirID)
	if err != nil {
		return nil, instruments.ErrInvalidChoirID
	}

	rows, err := r.queries.ListInstrumentsByChoirID(ctx, sqlc.ListInstrumentsByChoirIDParams{
		TenantID: tenantUUID,
		ChoirID:  choirUUID,
	})
	if err != nil {
		return nil, err
	}

	items := make([]instruments.Instrument, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapInstrumentRow(sqlc.Instrument(row)))
	}

	return items, nil
}

func (r *InstrumentRepository) Archive(ctx context.Context, tenantID string, instrumentID string) error {
	tenantUUID, err := parseUUID(tenantID)
	if err != nil {
		return instruments.ErrInvalidTenantID
	}

	instrumentUUID, err := parseUUID(instrumentID)
	if err != nil {
		return instruments.ErrInvalidInstrumentID
	}

	affected, err := r.queries.ArchiveInstrument(ctx, sqlc.ArchiveInstrumentParams{
		TenantID: tenantUUID,
		ID:       instrumentUUID,
	})
	if err != nil {
		return err
	}

	if affected == 0 {
		return instruments.ErrInstrumentNotFound
	}

	return nil
}

func (r *InstrumentRepository) AddUser(ctx context.Context, tenantID string, instrumentID string, userID string) error {
	tenantUUID, err := parseUUID(tenantID)
	if err != nil {
		return instruments.ErrInvalidTenantID
	}

	instrumentUUID, err := parseUUID(instrumentID)
	if err != nil {
		return instruments.ErrInvalidInstrumentID
	}

	userUUID, err := parseUUID(userID)
	if err != nil {
		return instruments.ErrInvalidUserID
	}

	if _, err := r.queries.AddInstrumentToUser(ctx, sqlc.AddInstrumentToUserParams{
		TenantID:     tenantUUID,
		UserID:       userUUID,
		InstrumentID: instrumentUUID,
	}); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return instruments.ErrUserLinkExists
		}

		return err
	}

	return nil
}

func (r *InstrumentRepository) RemoveUser(ctx context.Context, tenantID string, instrumentID string, userID string) error {
	tenantUUID, err := parseUUID(tenantID)
	if err != nil {
		return instruments.ErrInvalidTenantID
	}

	instrumentUUID, err := parseUUID(instrumentID)
	if err != nil {
		return instruments.ErrInvalidInstrumentID
	}

	userUUID, err := parseUUID(userID)
	if err != nil {
		return instruments.ErrInvalidUserID
	}

	affected, err := r.queries.RemoveInstrumentFromUser(ctx, sqlc.RemoveInstrumentFromUserParams{
		TenantID:     tenantUUID,
		UserID:       userUUID,
		InstrumentID: instrumentUUID,
	})
	if err != nil {
		return err
	}

	if affected == 0 {
		return instruments.ErrUserLinkNotFound
	}

	return nil
}

func (r *InstrumentRepository) ListUsers(ctx context.Context, tenantID string, instrumentID string) ([]instruments.InstrumentUser, error) {
	tenantUUID, err := parseUUID(tenantID)
	if err != nil {
		return nil, instruments.ErrInvalidTenantID
	}

	instrumentUUID, err := parseUUID(instrumentID)
	if err != nil {
		return nil, instruments.ErrInvalidInstrumentID
	}

	rows, err := r.queries.ListUsersByInstrument(ctx, sqlc.ListUsersByInstrumentParams{
		TenantID:     tenantUUID,
		InstrumentID: instrumentUUID,
	})
	if err != nil {
		return nil, err
	}

	items := make([]instruments.InstrumentUser, 0, len(rows))
	for _, row := range rows {
		items = append(items, instruments.InstrumentUser{
			ID:       uuidString(row.ID),
			TenantID: uuidString(row.TenantID),
			Email:    row.Email,
			FullName: row.FullName,
			Active:   row.Active,
		})
	}

	return items, nil
}

func mapInstrumentRow(row sqlc.Instrument) instruments.Instrument {
	return instruments.Instrument{
		ID:          uuidString(row.ID),
		TenantID:    uuidString(row.TenantID),
		ChoirID:     uuidString(row.ChoirID),
		Name:        row.Name,
		Description: textPointer(row.Description),
		Icon:        textPointer(row.Icon),
		Archived:    row.Archived,
	}
}
