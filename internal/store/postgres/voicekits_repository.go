package postgres

import (
	"context"
	"errors"

	"github.com/LeviLunique/coralhub-backend/internal/modules/voicekits"
	"github.com/LeviLunique/coralhub-backend/internal/store/postgres/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type VoiceKitRepository struct {
	queries *sqlc.Queries
}

func NewVoiceKitRepository(queries *sqlc.Queries) *VoiceKitRepository {
	return &VoiceKitRepository{queries: queries}
}

func (r *VoiceKitRepository) Create(ctx context.Context, params voicekits.CreateParams) (voicekits.VoiceKit, error) {
	tenantID, err := parseUUID(params.TenantID)
	if err != nil {
		return voicekits.VoiceKit{}, voicekits.ErrInvalidTenantID
	}

	choirID, err := parseUUID(params.ChoirID)
	if err != nil {
		return voicekits.VoiceKit{}, voicekits.ErrInvalidChoirID
	}

	row, err := r.queries.CreateVoiceKit(ctx, sqlc.CreateVoiceKitParams{
		TenantID:    tenantID,
		ChoirID:     choirID,
		Name:        params.Name,
		Description: textValue(params.Description),
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return voicekits.VoiceKit{}, voicekits.ErrVoiceKitNameTaken
		}

		return voicekits.VoiceKit{}, err
	}

	return voicekits.VoiceKit{
		ID:          uuidString(row.ID),
		TenantID:    uuidString(row.TenantID),
		ChoirID:     uuidString(row.ChoirID),
		Name:        row.Name,
		Description: textPointer(row.Description),
		Active:      row.Active,
	}, nil
}

func (r *VoiceKitRepository) GetByIDForMember(ctx context.Context, tenantID string, voiceKitID string, userID string) (voicekits.VoiceKit, error) {
	tenantUUID, err := parseUUID(tenantID)
	if err != nil {
		return voicekits.VoiceKit{}, voicekits.ErrInvalidTenantID
	}

	voiceKitUUID, err := parseUUID(voiceKitID)
	if err != nil {
		return voicekits.VoiceKit{}, voicekits.ErrInvalidVoiceKitID
	}

	userUUID, err := parseUUID(userID)
	if err != nil {
		return voicekits.VoiceKit{}, voicekits.ErrInvalidActorID
	}

	row, err := r.queries.GetVoiceKitByIDForMember(ctx, sqlc.GetVoiceKitByIDForMemberParams{
		TenantID: tenantUUID,
		ID:       voiceKitUUID,
		UserID:   userUUID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return voicekits.VoiceKit{}, voicekits.ErrVoiceKitNotFound
		}

		return voicekits.VoiceKit{}, err
	}

	return voicekits.VoiceKit{
		ID:          uuidString(row.ID),
		TenantID:    uuidString(row.TenantID),
		ChoirID:     uuidString(row.ChoirID),
		Name:        row.Name,
		Description: textPointer(row.Description),
		Active:      row.Active,
	}, nil
}

func (r *VoiceKitRepository) ListByChoirID(ctx context.Context, tenantID string, choirID string) ([]voicekits.VoiceKit, error) {
	tenantUUID, err := parseUUID(tenantID)
	if err != nil {
		return nil, voicekits.ErrInvalidTenantID
	}

	choirUUID, err := parseUUID(choirID)
	if err != nil {
		return nil, voicekits.ErrInvalidChoirID
	}

	rows, err := r.queries.ListVoiceKitsByChoirID(ctx, sqlc.ListVoiceKitsByChoirIDParams{
		TenantID: tenantUUID,
		ChoirID:  choirUUID,
	})
	if err != nil {
		return nil, err
	}

	items := make([]voicekits.VoiceKit, 0, len(rows))
	for _, row := range rows {
		items = append(items, voicekits.VoiceKit{
			ID:          uuidString(row.ID),
			TenantID:    uuidString(row.TenantID),
			ChoirID:     uuidString(row.ChoirID),
			Name:        row.Name,
			Description: textPointer(row.Description),
			Active:      row.Active,
		})
	}

	return items, nil
}

func (r *VoiceKitRepository) Delete(ctx context.Context, tenantID string, voiceKitID string) error {
	tenantUUID, err := parseUUID(tenantID)
	if err != nil {
		return voicekits.ErrInvalidTenantID
	}

	voiceKitUUID, err := parseUUID(voiceKitID)
	if err != nil {
		return voicekits.ErrInvalidVoiceKitID
	}

	affected, err := r.queries.DeactivateVoiceKit(ctx, sqlc.DeactivateVoiceKitParams{
		TenantID: tenantUUID,
		ID:       voiceKitUUID,
	})
	if err != nil {
		return err
	}

	if affected == 0 {
		return voicekits.ErrVoiceKitNotFound
	}

	return nil
}
