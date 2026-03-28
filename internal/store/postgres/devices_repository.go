package postgres

import (
	"context"
	"strings"

	"github.com/LeviLunique/coralhub-backend/internal/modules/devices"
	"github.com/LeviLunique/coralhub-backend/internal/store/postgres/sqlc"
)

type DeviceRepository struct {
	queries *sqlc.Queries
}

func NewDeviceRepository(queries *sqlc.Queries) *DeviceRepository {
	return &DeviceRepository{queries: queries}
}

func (r *DeviceRepository) Create(ctx context.Context, params devices.CreateParams) (devices.DeviceToken, error) {
	tenantID, err := parseUUID(params.TenantID)
	if err != nil {
		return devices.DeviceToken{}, err
	}

	userID, err := parseUUID(params.UserID)
	if err != nil {
		return devices.DeviceToken{}, err
	}

	row, err := r.queries.CreateDeviceToken(ctx, sqlc.CreateDeviceTokenParams{
		TenantID: tenantID,
		UserID:   userID,
		Platform: strings.TrimSpace(params.Platform),
		Token:    strings.TrimSpace(params.Token),
	})
	if err != nil {
		return devices.DeviceToken{}, err
	}

	return mapDeviceTokenRow(row), nil
}

func (r *DeviceRepository) ListActiveByUserID(ctx context.Context, tenantID string, userID string) ([]devices.DeviceToken, error) {
	tenantUUID, err := parseUUID(tenantID)
	if err != nil {
		return nil, err
	}

	userUUID, err := parseUUID(userID)
	if err != nil {
		return nil, err
	}

	rows, err := r.queries.ListActiveDeviceTokensByUserID(ctx, sqlc.ListActiveDeviceTokensByUserIDParams{
		TenantID: tenantUUID,
		UserID:   userUUID,
	})
	if err != nil {
		return nil, err
	}

	items := make([]devices.DeviceToken, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapDeviceTokenRow(row))
	}

	return items, nil
}

func (r *DeviceRepository) DeactivateByToken(ctx context.Context, tenantID string, token string) error {
	tenantUUID, err := parseUUID(tenantID)
	if err != nil {
		return err
	}

	_, err = r.queries.DeactivateDeviceTokenByToken(ctx, sqlc.DeactivateDeviceTokenByTokenParams{
		TenantID: tenantUUID,
		Token:    strings.TrimSpace(token),
	})
	return err
}

func mapDeviceTokenRow(row sqlc.DeviceToken) devices.DeviceToken {
	return devices.DeviceToken{
		ID:       uuidString(row.ID),
		TenantID: uuidString(row.TenantID),
		UserID:   uuidString(row.UserID),
		Platform: row.Platform,
		Token:    row.Token,
		Active:   row.Active,
	}
}
