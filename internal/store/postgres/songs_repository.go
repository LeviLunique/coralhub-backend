package postgres

import (
	"context"
	"errors"

	"github.com/LeviLunique/coralhub-backend/internal/modules/songs"
	"github.com/LeviLunique/coralhub-backend/internal/store/postgres/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type SongRepository struct {
	queries *sqlc.Queries
}

func NewSongRepository(queries *sqlc.Queries) *SongRepository {
	return &SongRepository{queries: queries}
}

func (r *SongRepository) Create(ctx context.Context, params songs.CreateParams) (songs.Song, error) {
	tenantID, err := parseUUID(params.TenantID)
	if err != nil {
		return songs.Song{}, songs.ErrInvalidTenantID
	}

	choirID, err := parseUUID(params.ChoirID)
	if err != nil {
		return songs.Song{}, songs.ErrInvalidChoirID
	}

	row, err := r.queries.CreateSong(ctx, sqlc.CreateSongParams{
		TenantID: tenantID,
		ChoirID:  choirID,
		Title:    params.Title,
		Composer: textValue(params.Composer),
		Arranger: textValue(params.Arranger),
		SongKey:  textValue(params.SongKey),
		Duration: textValue(params.Duration),
		Notes:    textValue(params.Notes),
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return songs.Song{}, songs.ErrSongTitleTaken
		}

		return songs.Song{}, err
	}

	return mapSongRow(row), nil
}

func (r *SongRepository) Update(ctx context.Context, params songs.UpdateParams) (songs.Song, error) {
	tenantID, err := parseUUID(params.TenantID)
	if err != nil {
		return songs.Song{}, songs.ErrInvalidTenantID
	}

	songID, err := parseUUID(params.SongID)
	if err != nil {
		return songs.Song{}, songs.ErrInvalidSongID
	}

	row, err := r.queries.UpdateSong(ctx, sqlc.UpdateSongParams{
		TenantID: tenantID,
		ID:       songID,
		Title:    params.Title,
		Composer: textValue(params.Composer),
		Arranger: textValue(params.Arranger),
		SongKey:  textValue(params.SongKey),
		Duration: textValue(params.Duration),
		Notes:    textValue(params.Notes),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return songs.Song{}, songs.ErrSongNotFound
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return songs.Song{}, songs.ErrSongTitleTaken
		}

		return songs.Song{}, err
	}

	return mapSongRow(row), nil
}

func (r *SongRepository) GetByIDForMember(ctx context.Context, tenantID string, songID string, userID string) (songs.Song, error) {
	tenantUUID, err := parseUUID(tenantID)
	if err != nil {
		return songs.Song{}, songs.ErrInvalidTenantID
	}

	songUUID, err := parseUUID(songID)
	if err != nil {
		return songs.Song{}, songs.ErrInvalidSongID
	}

	userUUID, err := parseUUID(userID)
	if err != nil {
		return songs.Song{}, songs.ErrInvalidActorID
	}

	row, err := r.queries.GetSongByIDForMember(ctx, sqlc.GetSongByIDForMemberParams{
		TenantID: tenantUUID,
		ID:       songUUID,
		UserID:   userUUID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return songs.Song{}, songs.ErrSongNotFound
		}

		return songs.Song{}, err
	}

	return mapSongRow(sqlc.Song(row)), nil
}

func (r *SongRepository) ListByChoirID(ctx context.Context, tenantID string, choirID string) ([]songs.Song, error) {
	tenantUUID, err := parseUUID(tenantID)
	if err != nil {
		return nil, songs.ErrInvalidTenantID
	}

	choirUUID, err := parseUUID(choirID)
	if err != nil {
		return nil, songs.ErrInvalidChoirID
	}

	rows, err := r.queries.ListSongsByChoirID(ctx, sqlc.ListSongsByChoirIDParams{
		TenantID: tenantUUID,
		ChoirID:  choirUUID,
	})
	if err != nil {
		return nil, err
	}

	items := make([]songs.Song, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapSongRow(sqlc.Song(row)))
	}

	return items, nil
}

func (r *SongRepository) Archive(ctx context.Context, tenantID string, songID string) error {
	tenantUUID, err := parseUUID(tenantID)
	if err != nil {
		return songs.ErrInvalidTenantID
	}

	songUUID, err := parseUUID(songID)
	if err != nil {
		return songs.ErrInvalidSongID
	}

	affected, err := r.queries.ArchiveSong(ctx, sqlc.ArchiveSongParams{
		TenantID: tenantUUID,
		ID:       songUUID,
	})
	if err != nil {
		return err
	}

	if affected == 0 {
		return songs.ErrSongNotFound
	}

	return nil
}

func mapSongRow(row sqlc.Song) songs.Song {
	return songs.Song{
		ID:       uuidString(row.ID),
		TenantID: uuidString(row.TenantID),
		ChoirID:  uuidString(row.ChoirID),
		Title:    row.Title,
		Composer: textPointer(row.Composer),
		Arranger: textPointer(row.Arranger),
		SongKey:  textPointer(row.SongKey),
		Duration: textPointer(row.Duration),
		Notes:    textPointer(row.Notes),
		Archived: row.Archived,
	}
}
