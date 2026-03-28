package postgres

import (
	"context"
	"errors"

	modulefiles "github.com/LeviLunique/coralhub-backend/internal/modules/files"
	"github.com/LeviLunique/coralhub-backend/internal/store/postgres/sqlc"
	"github.com/jackc/pgx/v5"
)

type FileRepository struct {
	queries *sqlc.Queries
}

func NewFileRepository(queries *sqlc.Queries) *FileRepository {
	return &FileRepository{queries: queries}
}

func (r *FileRepository) Create(ctx context.Context, params modulefiles.CreateParams) (modulefiles.File, error) {
	tenantID, err := parseUUID(params.TenantID)
	if err != nil {
		return modulefiles.File{}, modulefiles.ErrInvalidTenantID
	}

	voiceKitID, err := parseUUID(params.VoiceKitID)
	if err != nil {
		return modulefiles.File{}, modulefiles.ErrInvalidVoiceKitID
	}

	row, err := r.queries.CreateKitFile(ctx, sqlc.CreateKitFileParams{
		TenantID:         tenantID,
		VoiceKitID:       voiceKitID,
		OriginalFilename: params.OriginalFilename,
		StoredFilename:   params.StoredFilename,
		ContentType:      params.ContentType,
		SizeBytes:        params.SizeBytes,
		StorageKey:       params.StorageKey,
	})
	if err != nil {
		return modulefiles.File{}, err
	}

	return modulefiles.File{
		ID:               uuidString(row.ID),
		TenantID:         uuidString(row.TenantID),
		VoiceKitID:       uuidString(row.VoiceKitID),
		OriginalFilename: row.OriginalFilename,
		StoredFilename:   row.StoredFilename,
		ContentType:      row.ContentType,
		SizeBytes:        row.SizeBytes,
		StorageKey:       row.StorageKey,
		Active:           row.Active,
	}, nil
}

func (r *FileRepository) GetByIDForMember(ctx context.Context, tenantID string, fileID string, userID string) (modulefiles.File, error) {
	tenantUUID, err := parseUUID(tenantID)
	if err != nil {
		return modulefiles.File{}, modulefiles.ErrInvalidTenantID
	}

	fileUUID, err := parseUUID(fileID)
	if err != nil {
		return modulefiles.File{}, modulefiles.ErrInvalidFileID
	}

	userUUID, err := parseUUID(userID)
	if err != nil {
		return modulefiles.File{}, modulefiles.ErrInvalidActorID
	}

	row, err := r.queries.GetKitFileByIDForMember(ctx, sqlc.GetKitFileByIDForMemberParams{
		TenantID: tenantUUID,
		ID:       fileUUID,
		UserID:   userUUID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return modulefiles.File{}, modulefiles.ErrFileNotFound
		}

		return modulefiles.File{}, err
	}

	return modulefiles.File{
		ID:               uuidString(row.ID),
		TenantID:         uuidString(row.TenantID),
		ChoirID:          uuidString(row.ChoirID),
		VoiceKitID:       uuidString(row.VoiceKitID),
		OriginalFilename: row.OriginalFilename,
		StoredFilename:   row.StoredFilename,
		ContentType:      row.ContentType,
		SizeBytes:        row.SizeBytes,
		StorageKey:       row.StorageKey,
		Active:           row.Active,
	}, nil
}

func (r *FileRepository) ListByVoiceKitID(ctx context.Context, tenantID string, voiceKitID string) ([]modulefiles.File, error) {
	tenantUUID, err := parseUUID(tenantID)
	if err != nil {
		return nil, modulefiles.ErrInvalidTenantID
	}

	voiceKitUUID, err := parseUUID(voiceKitID)
	if err != nil {
		return nil, modulefiles.ErrInvalidVoiceKitID
	}

	rows, err := r.queries.ListKitFilesByVoiceKitID(ctx, sqlc.ListKitFilesByVoiceKitIDParams{
		TenantID:   tenantUUID,
		VoiceKitID: voiceKitUUID,
	})
	if err != nil {
		return nil, err
	}

	items := make([]modulefiles.File, 0, len(rows))
	for _, row := range rows {
		items = append(items, modulefiles.File{
			ID:               uuidString(row.ID),
			TenantID:         uuidString(row.TenantID),
			VoiceKitID:       uuidString(row.VoiceKitID),
			OriginalFilename: row.OriginalFilename,
			StoredFilename:   row.StoredFilename,
			ContentType:      row.ContentType,
			SizeBytes:        row.SizeBytes,
			StorageKey:       row.StorageKey,
			Active:           row.Active,
		})
	}

	return items, nil
}

func (r *FileRepository) Delete(ctx context.Context, tenantID string, fileID string) error {
	tenantUUID, err := parseUUID(tenantID)
	if err != nil {
		return modulefiles.ErrInvalidTenantID
	}

	fileUUID, err := parseUUID(fileID)
	if err != nil {
		return modulefiles.ErrInvalidFileID
	}

	affected, err := r.queries.DeactivateKitFile(ctx, sqlc.DeactivateKitFileParams{
		TenantID: tenantUUID,
		ID:       fileUUID,
	})
	if err != nil {
		return err
	}

	if affected == 0 {
		return modulefiles.ErrFileNotFound
	}

	return nil
}
