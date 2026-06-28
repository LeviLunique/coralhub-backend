package postgres

import (
	"context"
	"errors"

	"github.com/LeviLunique/coralhub-backend/internal/modules/instruments"
	"github.com/LeviLunique/coralhub-backend/internal/modules/materials"
	"github.com/LeviLunique/coralhub-backend/internal/store/postgres/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type MaterialRepository struct {
	queries *sqlc.Queries
}

func NewMaterialRepository(queries *sqlc.Queries) *MaterialRepository {
	return &MaterialRepository{queries: queries}
}

func (r *MaterialRepository) Create(ctx context.Context, params materials.CreateParams) (materials.Material, error) {
	tenantID, err := parseUUID(params.TenantID)
	if err != nil {
		return materials.Material{}, materials.ErrInvalidTenantID
	}

	choirID, err := parseUUID(params.ChoirID)
	if err != nil {
		return materials.Material{}, materials.ErrInvalidChoirID
	}

	songID, err := parseUUID(params.SongID)
	if err != nil {
		return materials.Material{}, materials.ErrInvalidSongID
	}

	row, err := r.queries.CreateMaterial(ctx, sqlc.CreateMaterialParams{
		TenantID:     tenantID,
		ChoirID:      choirID,
		SongID:       songID,
		Name:         params.Name,
		MaterialType: params.MaterialType,
		TargetType:   params.TargetType,
		VoiceType:    textValue(params.VoiceType),
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return materials.Material{}, materials.ErrInvalidSongID
		}

		return materials.Material{}, err
	}

	return mapMaterialRow(row), nil
}

func (r *MaterialRepository) Update(ctx context.Context, params materials.UpdateParams) (materials.Material, error) {
	tenantID, err := parseUUID(params.TenantID)
	if err != nil {
		return materials.Material{}, materials.ErrInvalidTenantID
	}

	materialID, err := parseUUID(params.MaterialID)
	if err != nil {
		return materials.Material{}, materials.ErrInvalidMaterialID
	}

	row, err := r.queries.UpdateMaterial(ctx, sqlc.UpdateMaterialParams{
		TenantID:     tenantID,
		ID:           materialID,
		Name:         params.Name,
		MaterialType: params.MaterialType,
		TargetType:   params.TargetType,
		VoiceType:    textValue(params.VoiceType),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return materials.Material{}, materials.ErrMaterialNotFound
		}

		return materials.Material{}, err
	}

	return mapMaterialRow(row), nil
}

func (r *MaterialRepository) GetByIDForMember(ctx context.Context, tenantID string, materialID string, userID string) (materials.Material, error) {
	tenantUUID, err := parseUUID(tenantID)
	if err != nil {
		return materials.Material{}, materials.ErrInvalidTenantID
	}

	materialUUID, err := parseUUID(materialID)
	if err != nil {
		return materials.Material{}, materials.ErrInvalidMaterialID
	}

	userUUID, err := parseUUID(userID)
	if err != nil {
		return materials.Material{}, materials.ErrInvalidActorID
	}

	row, err := r.queries.GetMaterialByIDForMember(ctx, sqlc.GetMaterialByIDForMemberParams{
		TenantID: tenantUUID,
		ID:       materialUUID,
		UserID:   userUUID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return materials.Material{}, materials.ErrMaterialNotFound
		}

		return materials.Material{}, err
	}

	return mapMaterialRow(row), nil
}

func (r *MaterialRepository) ListByChoirID(ctx context.Context, tenantID string, choirID string) ([]materials.Material, error) {
	tenantUUID, err := parseUUID(tenantID)
	if err != nil {
		return nil, materials.ErrInvalidTenantID
	}

	choirUUID, err := parseUUID(choirID)
	if err != nil {
		return nil, materials.ErrInvalidChoirID
	}

	rows, err := r.queries.ListMaterialsByChoirID(ctx, sqlc.ListMaterialsByChoirIDParams{
		TenantID: tenantUUID,
		ChoirID:  choirUUID,
	})
	if err != nil {
		return nil, err
	}

	items := make([]materials.Material, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapMaterialRow(row))
	}

	return items, nil
}

func (r *MaterialRepository) ListBySongID(ctx context.Context, tenantID string, songID string) ([]materials.Material, error) {
	tenantUUID, err := parseUUID(tenantID)
	if err != nil {
		return nil, materials.ErrInvalidTenantID
	}

	songUUID, err := parseUUID(songID)
	if err != nil {
		return nil, materials.ErrInvalidSongID
	}

	rows, err := r.queries.ListMaterialsBySongID(ctx, sqlc.ListMaterialsBySongIDParams{
		TenantID: tenantUUID,
		SongID:   songUUID,
	})
	if err != nil {
		return nil, err
	}

	items := make([]materials.Material, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapMaterialRow(row))
	}

	return items, nil
}

func (r *MaterialRepository) Archive(ctx context.Context, tenantID string, materialID string) error {
	tenantUUID, err := parseUUID(tenantID)
	if err != nil {
		return materials.ErrInvalidTenantID
	}

	materialUUID, err := parseUUID(materialID)
	if err != nil {
		return materials.ErrInvalidMaterialID
	}

	affected, err := r.queries.ArchiveMaterial(ctx, sqlc.ArchiveMaterialParams{
		TenantID: tenantUUID,
		ID:       materialUUID,
	})
	if err != nil {
		return err
	}

	if affected == 0 {
		return materials.ErrMaterialNotFound
	}

	return nil
}

func (r *MaterialRepository) AddInstrument(ctx context.Context, tenantID string, materialID string, instrumentID string) error {
	tenantUUID, err := parseUUID(tenantID)
	if err != nil {
		return materials.ErrInvalidTenantID
	}

	materialUUID, err := parseUUID(materialID)
	if err != nil {
		return materials.ErrInvalidMaterialID
	}

	instrumentUUID, err := parseUUID(instrumentID)
	if err != nil {
		return materials.ErrInvalidInstrumentID
	}

	if _, err := r.queries.AddInstrumentToMaterial(ctx, sqlc.AddInstrumentToMaterialParams{
		TenantID:     tenantUUID,
		MaterialID:   materialUUID,
		InstrumentID: instrumentUUID,
	}); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return materials.ErrInstrumentLinkExists
		}

		return err
	}

	return nil
}

func (r *MaterialRepository) RemoveInstrument(ctx context.Context, tenantID string, materialID string, instrumentID string) error {
	tenantUUID, err := parseUUID(tenantID)
	if err != nil {
		return materials.ErrInvalidTenantID
	}

	materialUUID, err := parseUUID(materialID)
	if err != nil {
		return materials.ErrInvalidMaterialID
	}

	instrumentUUID, err := parseUUID(instrumentID)
	if err != nil {
		return materials.ErrInvalidInstrumentID
	}

	affected, err := r.queries.RemoveInstrumentFromMaterial(ctx, sqlc.RemoveInstrumentFromMaterialParams{
		TenantID:     tenantUUID,
		MaterialID:   materialUUID,
		InstrumentID: instrumentUUID,
	})
	if err != nil {
		return err
	}

	if affected == 0 {
		return materials.ErrInstrumentLinkNotFound
	}

	return nil
}

func (r *MaterialRepository) ListInstruments(ctx context.Context, tenantID string, materialID string) ([]instruments.Instrument, error) {
	tenantUUID, err := parseUUID(tenantID)
	if err != nil {
		return nil, materials.ErrInvalidTenantID
	}

	materialUUID, err := parseUUID(materialID)
	if err != nil {
		return nil, materials.ErrInvalidMaterialID
	}

	rows, err := r.queries.ListInstrumentsByMaterial(ctx, sqlc.ListInstrumentsByMaterialParams{
		TenantID:   tenantUUID,
		MaterialID: materialUUID,
	})
	if err != nil {
		return nil, err
	}

	items := make([]instruments.Instrument, 0, len(rows))
	for _, row := range rows {
		items = append(items, instruments.Instrument{
			ID:          uuidString(row.ID),
			TenantID:    uuidString(row.TenantID),
			ChoirID:     uuidString(row.ChoirID),
			Name:        row.Name,
			Description: textPointer(row.Description),
			Icon:        textPointer(row.Icon),
			Archived:    row.Archived,
		})
	}

	return items, nil
}

func mapMaterialRow(row sqlc.Material) materials.Material {
	return materials.Material{
		ID:           uuidString(row.ID),
		TenantID:     uuidString(row.TenantID),
		ChoirID:      uuidString(row.ChoirID),
		SongID:       uuidString(row.SongID),
		Name:         row.Name,
		MaterialType: row.MaterialType,
		TargetType:   row.TargetType,
		VoiceType:    textPointer(row.VoiceType),
		Archived:     row.Archived,
		CreatedAt:    row.CreatedAt.Time.UTC(),
		UpdatedAt:    row.UpdatedAt.Time.UTC(),
	}
}
