package postgres

import (
	"context"
	"errors"

	"github.com/LeviLunique/coralhub-backend/internal/modules/repertoires"
	"github.com/LeviLunique/coralhub-backend/internal/store/postgres/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type RepertoireRepository struct {
	queries *sqlc.Queries
}

func NewRepertoireRepository(queries *sqlc.Queries) *RepertoireRepository {
	return &RepertoireRepository{queries: queries}
}

func (r *RepertoireRepository) Create(ctx context.Context, params repertoires.CreateParams) (repertoires.Repertoire, error) {
	tenantID, err := parseUUID(params.TenantID)
	if err != nil {
		return repertoires.Repertoire{}, repertoires.ErrInvalidTenantID
	}

	choirID, err := parseUUID(params.ChoirID)
	if err != nil {
		return repertoires.Repertoire{}, repertoires.ErrInvalidChoirID
	}

	row, err := r.queries.CreateRepertoire(ctx, sqlc.CreateRepertoireParams{
		TenantID:    tenantID,
		ChoirID:     choirID,
		Name:        params.Name,
		Description: textValue(params.Description),
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return repertoires.Repertoire{}, repertoires.ErrRepertoireNameTaken
		}

		return repertoires.Repertoire{}, err
	}

	return mapRepertoireRow(row), nil
}

func (r *RepertoireRepository) Update(ctx context.Context, params repertoires.UpdateParams) (repertoires.Repertoire, error) {
	tenantID, err := parseUUID(params.TenantID)
	if err != nil {
		return repertoires.Repertoire{}, repertoires.ErrInvalidTenantID
	}

	repertoireID, err := parseUUID(params.RepertoireID)
	if err != nil {
		return repertoires.Repertoire{}, repertoires.ErrInvalidRepertoireID
	}

	row, err := r.queries.UpdateRepertoire(ctx, sqlc.UpdateRepertoireParams{
		TenantID:    tenantID,
		ID:          repertoireID,
		Name:        params.Name,
		Description: textValue(params.Description),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repertoires.Repertoire{}, repertoires.ErrRepertoireNotFound
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return repertoires.Repertoire{}, repertoires.ErrRepertoireNameTaken
		}

		return repertoires.Repertoire{}, err
	}

	return mapRepertoireRow(row), nil
}

func (r *RepertoireRepository) GetByIDForMember(ctx context.Context, tenantID string, repertoireID string, userID string) (repertoires.Repertoire, error) {
	tenantUUID, err := parseUUID(tenantID)
	if err != nil {
		return repertoires.Repertoire{}, repertoires.ErrInvalidTenantID
	}

	repertoireUUID, err := parseUUID(repertoireID)
	if err != nil {
		return repertoires.Repertoire{}, repertoires.ErrInvalidRepertoireID
	}

	userUUID, err := parseUUID(userID)
	if err != nil {
		return repertoires.Repertoire{}, repertoires.ErrInvalidActorID
	}

	row, err := r.queries.GetRepertoireByIDForMember(ctx, sqlc.GetRepertoireByIDForMemberParams{
		TenantID: tenantUUID,
		ID:       repertoireUUID,
		UserID:   userUUID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repertoires.Repertoire{}, repertoires.ErrRepertoireNotFound
		}

		return repertoires.Repertoire{}, err
	}

	return mapRepertoireRow(sqlc.Repertoire(row)), nil
}

func (r *RepertoireRepository) ListByChoirID(ctx context.Context, tenantID string, choirID string) ([]repertoires.Repertoire, error) {
	tenantUUID, err := parseUUID(tenantID)
	if err != nil {
		return nil, repertoires.ErrInvalidTenantID
	}

	choirUUID, err := parseUUID(choirID)
	if err != nil {
		return nil, repertoires.ErrInvalidChoirID
	}

	rows, err := r.queries.ListRepertoiresByChoirID(ctx, sqlc.ListRepertoiresByChoirIDParams{
		TenantID: tenantUUID,
		ChoirID:  choirUUID,
	})
	if err != nil {
		return nil, err
	}

	items := make([]repertoires.Repertoire, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapRepertoireRow(sqlc.Repertoire(row)))
	}

	return items, nil
}

func (r *RepertoireRepository) Archive(ctx context.Context, tenantID string, repertoireID string) error {
	tenantUUID, err := parseUUID(tenantID)
	if err != nil {
		return repertoires.ErrInvalidTenantID
	}

	repertoireUUID, err := parseUUID(repertoireID)
	if err != nil {
		return repertoires.ErrInvalidRepertoireID
	}

	affected, err := r.queries.ArchiveRepertoire(ctx, sqlc.ArchiveRepertoireParams{
		TenantID: tenantUUID,
		ID:       repertoireUUID,
	})
	if err != nil {
		return err
	}

	if affected == 0 {
		return repertoires.ErrRepertoireNotFound
	}

	return nil
}

func (r *RepertoireRepository) AddSong(ctx context.Context, params repertoires.AddSongParams) error {
	tenantUUID, err := parseUUID(params.TenantID)
	if err != nil {
		return repertoires.ErrInvalidTenantID
	}

	repertoireUUID, err := parseUUID(params.RepertoireID)
	if err != nil {
		return repertoires.ErrInvalidRepertoireID
	}

	songUUID, err := parseUUID(params.SongID)
	if err != nil {
		return repertoires.ErrInvalidSongID
	}

	if _, err := r.queries.AddSongToRepertoire(ctx, sqlc.AddSongToRepertoireParams{
		TenantID:       tenantUUID,
		RepertoireID:   repertoireUUID,
		SongID:         songUUID,
		ExecutionOrder: int32(params.ExecutionOrder),
		Notes:          textValue(params.Notes),
	}); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return repertoires.ErrSongLinkExists
		}

		return err
	}

	return nil
}

func (r *RepertoireRepository) RemoveSong(ctx context.Context, tenantID string, repertoireID string, songID string) error {
	tenantUUID, err := parseUUID(tenantID)
	if err != nil {
		return repertoires.ErrInvalidTenantID
	}

	repertoireUUID, err := parseUUID(repertoireID)
	if err != nil {
		return repertoires.ErrInvalidRepertoireID
	}

	songUUID, err := parseUUID(songID)
	if err != nil {
		return repertoires.ErrInvalidSongID
	}

	affected, err := r.queries.RemoveSongFromRepertoire(ctx, sqlc.RemoveSongFromRepertoireParams{
		TenantID:     tenantUUID,
		RepertoireID: repertoireUUID,
		SongID:       songUUID,
	})
	if err != nil {
		return err
	}

	if affected == 0 {
		return repertoires.ErrSongLinkNotFound
	}

	return nil
}

func (r *RepertoireRepository) ListSongs(ctx context.Context, tenantID string, repertoireID string) ([]repertoires.RepertoireSong, error) {
	tenantUUID, err := parseUUID(tenantID)
	if err != nil {
		return nil, repertoires.ErrInvalidTenantID
	}

	repertoireUUID, err := parseUUID(repertoireID)
	if err != nil {
		return nil, repertoires.ErrInvalidRepertoireID
	}

	rows, err := r.queries.ListSongsByRepertoire(ctx, sqlc.ListSongsByRepertoireParams{
		TenantID:     tenantUUID,
		RepertoireID: repertoireUUID,
	})
	if err != nil {
		return nil, err
	}

	items := make([]repertoires.RepertoireSong, 0, len(rows))
	for _, row := range rows {
		items = append(items, repertoires.RepertoireSong{
			ID:             uuidString(row.ID),
			TenantID:       uuidString(row.TenantID),
			ChoirID:        uuidString(row.ChoirID),
			Title:          row.Title,
			Composer:       textPointer(row.Composer),
			Arranger:       textPointer(row.Arranger),
			SongKey:        textPointer(row.SongKey),
			Duration:       textPointer(row.Duration),
			Notes:          textPointer(row.Notes),
			Archived:       row.Archived,
			ExecutionOrder: int(row.ExecutionOrder),
			RepertoireNote: textPointer(row.RepertoireNotes),
		})
	}

	return items, nil
}

func mapRepertoireRow(row sqlc.Repertoire) repertoires.Repertoire {
	return repertoires.Repertoire{
		ID:          uuidString(row.ID),
		TenantID:    uuidString(row.TenantID),
		ChoirID:     uuidString(row.ChoirID),
		Name:        row.Name,
		Description: textPointer(row.Description),
		Archived:    row.Archived,
	}
}
