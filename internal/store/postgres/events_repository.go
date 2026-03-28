package postgres

import (
	"context"
	"errors"
	"time"

	moduleaudit "github.com/LeviLunique/coralhub-backend/internal/modules/audit"
	"github.com/LeviLunique/coralhub-backend/internal/modules/events"
	"github.com/LeviLunique/coralhub-backend/internal/store/postgres/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type EventRepository struct {
	db      txBeginner
	queries *sqlc.Queries
}

func NewEventRepository(db txBeginner, queries *sqlc.Queries) *EventRepository {
	return &EventRepository{db: db, queries: queries}
}

func (r *EventRepository) Create(ctx context.Context, params events.CreateParams) (events.Event, error) {
	tenantID, err := parseUUID(params.TenantID)
	if err != nil {
		return events.Event{}, events.ErrInvalidTenantID
	}

	choirID, err := parseUUID(params.ChoirID)
	if err != nil {
		return events.Event{}, events.ErrInvalidChoirID
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return events.Event{}, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	txQueries := r.queries.WithTx(tx)
	row, err := txQueries.CreateEvent(ctx, sqlc.CreateEventParams{
		TenantID:    tenantID,
		ChoirID:     choirID,
		Title:       params.Title,
		Description: textValue(params.Description),
		EventType:   params.EventType,
		Location:    textValue(params.Location),
		StartAt:     timestamptzValue(params.StartAt),
	})
	if err != nil {
		return events.Event{}, err
	}

	if err := insertScheduledNotifications(ctx, txQueries, tenantID, row.ID, params.Reminders); err != nil {
		return events.Event{}, err
	}

	now := time.Now().UTC()
	if _, err := createAuditLog(ctx, txQueries, tenantID, row.ID, moduleaudit.ActionEventCreated, moduleaudit.EntityTypeEvent, stringPointer(params.ActorUserID), now, map[string]any{
		"title":      params.Title,
		"event_type": params.EventType,
		"start_at":   params.StartAt.UTC(),
	}); err != nil {
		return events.Event{}, err
	}
	if _, err := createAuditLog(ctx, txQueries, tenantID, row.ID, moduleaudit.ActionNotificationsGenerated, moduleaudit.EntityTypeEvent, stringPointer(params.ActorUserID), now, map[string]any{
		"event_id":           uuidString(row.ID),
		"generated_count":    len(params.Reminders),
		"reminder_types":     collectReminderTypes(params.Reminders),
		"notification_state": events.NotificationStatusPending,
	}); err != nil {
		return events.Event{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return events.Event{}, err
	}

	return mapEventRow(row), nil
}

func (r *EventRepository) Update(ctx context.Context, params events.UpdateParams) (events.Event, error) {
	tenantID, err := parseUUID(params.TenantID)
	if err != nil {
		return events.Event{}, events.ErrInvalidTenantID
	}

	eventID, err := parseUUID(params.EventID)
	if err != nil {
		return events.Event{}, events.ErrInvalidEventID
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return events.Event{}, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	txQueries := r.queries.WithTx(tx)
	row, err := txQueries.UpdateEvent(ctx, sqlc.UpdateEventParams{
		TenantID:    tenantID,
		ID:          eventID,
		Title:       params.Title,
		Description: textValue(params.Description),
		EventType:   params.EventType,
		Location:    textValue(params.Location),
		StartAt:     timestamptzValue(params.StartAt),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return events.Event{}, events.ErrEventNotFound
		}
		return events.Event{}, err
	}

	if _, err := txQueries.CancelPendingScheduledNotificationsByEventID(ctx, sqlc.CancelPendingScheduledNotificationsByEventIDParams{
		TenantID: tenantID,
		EventID:  eventID,
	}); err != nil {
		return events.Event{}, err
	}

	if err := insertScheduledNotifications(ctx, txQueries, tenantID, row.ID, params.Reminders); err != nil {
		return events.Event{}, err
	}

	now := time.Now().UTC()
	if _, err := createAuditLog(ctx, txQueries, tenantID, row.ID, moduleaudit.ActionEventUpdated, moduleaudit.EntityTypeEvent, stringPointer(params.ActorUserID), now, map[string]any{
		"title":      params.Title,
		"event_type": params.EventType,
		"start_at":   params.StartAt.UTC(),
	}); err != nil {
		return events.Event{}, err
	}
	if _, err := createAuditLog(ctx, txQueries, tenantID, row.ID, moduleaudit.ActionNotificationsGenerated, moduleaudit.EntityTypeEvent, stringPointer(params.ActorUserID), now, map[string]any{
		"event_id":           uuidString(row.ID),
		"generated_count":    len(params.Reminders),
		"reminder_types":     collectReminderTypes(params.Reminders),
		"notification_state": events.NotificationStatusPending,
	}); err != nil {
		return events.Event{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return events.Event{}, err
	}

	return mapEventRow(row), nil
}

func (r *EventRepository) GetByIDForMember(ctx context.Context, tenantID string, eventID string, userID string) (events.Event, error) {
	tenantUUID, err := parseUUID(tenantID)
	if err != nil {
		return events.Event{}, events.ErrInvalidTenantID
	}

	eventUUID, err := parseUUID(eventID)
	if err != nil {
		return events.Event{}, events.ErrInvalidEventID
	}

	userUUID, err := parseUUID(userID)
	if err != nil {
		return events.Event{}, events.ErrInvalidActorID
	}

	row, err := r.queries.GetEventByIDForMember(ctx, sqlc.GetEventByIDForMemberParams{
		TenantID: tenantUUID,
		ID:       eventUUID,
		UserID:   userUUID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return events.Event{}, events.ErrEventNotFound
		}
		return events.Event{}, err
	}

	return mapEventRow(row), nil
}

func (r *EventRepository) ListByChoirID(ctx context.Context, tenantID string, choirID string) ([]events.Event, error) {
	tenantUUID, err := parseUUID(tenantID)
	if err != nil {
		return nil, events.ErrInvalidTenantID
	}

	choirUUID, err := parseUUID(choirID)
	if err != nil {
		return nil, events.ErrInvalidChoirID
	}

	rows, err := r.queries.ListEventsByChoirID(ctx, sqlc.ListEventsByChoirIDParams{
		TenantID: tenantUUID,
		ChoirID:  choirUUID,
	})
	if err != nil {
		return nil, err
	}

	items := make([]events.Event, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapEventRow(row))
	}

	return items, nil
}

func (r *EventRepository) Cancel(ctx context.Context, params events.CancelParams) error {
	tenantUUID, err := parseUUID(params.TenantID)
	if err != nil {
		return events.ErrInvalidTenantID
	}

	eventUUID, err := parseUUID(params.EventID)
	if err != nil {
		return events.ErrInvalidEventID
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	txQueries := r.queries.WithTx(tx)
	affected, err := txQueries.CancelEvent(ctx, sqlc.CancelEventParams{
		TenantID: tenantUUID,
		ID:       eventUUID,
	})
	if err != nil {
		return err
	}
	if affected == 0 {
		return events.ErrEventNotFound
	}

	if _, err := txQueries.CancelPendingScheduledNotificationsByEventID(ctx, sqlc.CancelPendingScheduledNotificationsByEventIDParams{
		TenantID: tenantUUID,
		EventID:  eventUUID,
	}); err != nil {
		return err
	}

	if _, err := createAuditLog(ctx, txQueries, tenantUUID, eventUUID, moduleaudit.ActionEventCanceled, moduleaudit.EntityTypeEvent, stringPointer(params.ActorUserID), time.Now().UTC(), map[string]any{
		"event_id": uuidString(eventUUID),
	}); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func insertScheduledNotifications(ctx context.Context, queries *sqlc.Queries, tenantID pgtype.UUID, eventID pgtype.UUID, reminders []events.ScheduledReminder) error {
	for _, reminder := range reminders {
		userID, err := parseUUID(reminder.UserID)
		if err != nil {
			return events.ErrInvalidActorID
		}

		if _, err := queries.CreateScheduledNotification(ctx, sqlc.CreateScheduledNotificationParams{
			TenantID:     tenantID,
			EventID:      eventID,
			UserID:       userID,
			ReminderType: reminder.ReminderType,
			ScheduledFor: timestamptzValue(reminder.ScheduledFor),
			Status:       reminder.Status,
		}); err != nil {
			return err
		}
	}

	return nil
}

func mapEventRow(row sqlc.Event) events.Event {
	return events.Event{
		ID:          uuidString(row.ID),
		TenantID:    uuidString(row.TenantID),
		ChoirID:     uuidString(row.ChoirID),
		Title:       row.Title,
		Description: textPointer(row.Description),
		EventType:   row.EventType,
		Location:    textPointer(row.Location),
		StartAt:     row.StartAt.Time.UTC(),
		Active:      row.Active,
	}
}

func collectReminderTypes(reminders []events.ScheduledReminder) []string {
	types := make([]string, 0, len(reminders))
	for _, reminder := range reminders {
		types = append(types, reminder.ReminderType)
	}

	return types
}
