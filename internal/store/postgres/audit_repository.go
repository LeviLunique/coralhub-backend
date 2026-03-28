package postgres

import (
	"context"
	"encoding/json"
	"time"

	moduleaudit "github.com/LeviLunique/coralhub-backend/internal/modules/audit"
	"github.com/LeviLunique/coralhub-backend/internal/store/postgres/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type AuditRepository struct {
	queries *sqlc.Queries
}

func NewAuditRepository(queries *sqlc.Queries) *AuditRepository {
	return &AuditRepository{queries: queries}
}

func (r *AuditRepository) Create(ctx context.Context, params moduleaudit.CreateParams) (moduleaudit.Entry, error) {
	tenantID, err := parseUUID(params.TenantID)
	if err != nil {
		return moduleaudit.Entry{}, err
	}

	entityID, err := parseUUID(params.EntityID)
	if err != nil {
		return moduleaudit.Entry{}, err
	}

	row, err := createAuditLog(ctx, r.queries, tenantID, entityID, params.Action, params.EntityType, params.ActorID, params.OccurredAt, params.Payload)
	if err != nil {
		return moduleaudit.Entry{}, err
	}

	return mapAuditLogRow(row), nil
}

func (r *AuditRepository) ListByTenantID(ctx context.Context, tenantID string, limit int32) ([]moduleaudit.Entry, error) {
	tenantUUID, err := parseUUID(tenantID)
	if err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = 50
	}

	rows, err := r.queries.ListAuditLogByTenantID(ctx, sqlc.ListAuditLogByTenantIDParams{
		TenantID: tenantUUID,
		Limit:    limit,
	})
	if err != nil {
		return nil, err
	}

	items := make([]moduleaudit.Entry, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapAuditLogRow(row))
	}

	return items, nil
}

func createAuditLog(
	ctx context.Context,
	queries *sqlc.Queries,
	tenantID pgtype.UUID,
	entityID pgtype.UUID,
	action string,
	entityType string,
	actorID *string,
	occurredAt time.Time,
	payload any,
) (sqlc.AuditLog, error) {
	encodedPayload, err := json.Marshal(payload)
	if err != nil {
		return sqlc.AuditLog{}, err
	}

	var actorUUID pgtype.UUID
	if actorID != nil {
		actorUUID, err = parseUUID(*actorID)
		if err != nil {
			return sqlc.AuditLog{}, err
		}
	}

	return queries.CreateAuditLog(ctx, sqlc.CreateAuditLogParams{
		TenantID:    tenantID,
		EntityType:  entityType,
		EntityID:    entityID,
		Action:      action,
		ActorID:     actorUUID,
		OccurredAt:  timestamptzValue(occurredAt),
		PayloadJson: encodedPayload,
	})
}

func mapAuditLogRow(row sqlc.AuditLog) moduleaudit.Entry {
	var actorID *string
	if row.ActorID.Valid {
		value := uuidString(row.ActorID)
		actorID = &value
	}

	return moduleaudit.Entry{
		ID:         uuidString(row.ID),
		TenantID:   uuidString(row.TenantID),
		EntityType: row.EntityType,
		EntityID:   uuidString(row.EntityID),
		Action:     row.Action,
		ActorID:    actorID,
		OccurredAt: row.OccurredAt.Time.UTC(),
		Payload:    append([]byte(nil), row.PayloadJson...),
	}
}
