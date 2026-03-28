package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/LeviLunique/coralhub-backend/internal/modules/choirs"
	"github.com/LeviLunique/coralhub-backend/internal/modules/devices"
	"github.com/LeviLunique/coralhub-backend/internal/modules/events"
	modulefiles "github.com/LeviLunique/coralhub-backend/internal/modules/files"
	"github.com/LeviLunique/coralhub-backend/internal/modules/memberships"
	"github.com/LeviLunique/coralhub-backend/internal/modules/notifications"
	moduleusers "github.com/LeviLunique/coralhub-backend/internal/modules/users"
	"github.com/LeviLunique/coralhub-backend/internal/modules/voicekits"
	platformconfig "github.com/LeviLunique/coralhub-backend/internal/platform/config"
	"github.com/LeviLunique/coralhub-backend/internal/store/postgres/sqlc"
	"github.com/jackc/pgx/v5"
)

func TestChoirRepositoryCreateAndListByMemberUserIDIntegration(t *testing.T) {
	ctx, queries, tx := openIntegrationTestQueries(t)
	createTempChoirsTable(t, ctx, tx)
	createTempUsersTable(t, ctx, tx)
	createTempChoirMembersTable(t, ctx, tx)

	tenant := getSeedTenant(t, ctx, queries)
	userRepository := NewUserRepository(queries)
	actor, err := userRepository.Create(ctx, moduleusers.CreateParams{
		TenantID: tenant.ID,
		Email:    "ana@example.com",
		FullName: "Ana Clara",
	})
	if err != nil {
		t.Fatalf("Create actor user error = %v", err)
	}

	repository := NewChoirRepository(tx, queries)

	description := "Main choir"
	created, err := repository.Create(ctx, choirs.CreateParams{
		ActorUserID: actor.ID,
		TenantID:    tenant.ID,
		Name:        "Sopranos",
		Description: &description,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	listed, err := repository.ListByMemberUserID(ctx, tenant.ID, actor.ID)
	if err != nil {
		t.Fatalf("ListByMemberUserID() error = %v", err)
	}

	if len(listed) != 1 {
		t.Fatalf("len(listed) = %d, want 1", len(listed))
	}

	if listed[0].ID != created.ID {
		t.Fatalf("listed[0].ID = %q, want %q", listed[0].ID, created.ID)
	}
}

func TestUserRepositoryCreateAndListByTenantIDIntegration(t *testing.T) {
	ctx, queries, tx := openIntegrationTestQueries(t)
	createTempUsersTable(t, ctx, tx)

	tenant := getSeedTenant(t, ctx, queries)
	repository := NewUserRepository(queries)

	created, err := repository.Create(ctx, moduleusers.CreateParams{
		TenantID: tenant.ID,
		Email:    "ana@example.com",
		FullName: "Ana Clara",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	listed, err := repository.ListByTenantID(ctx, tenant.ID)
	if err != nil {
		t.Fatalf("ListByTenantID() error = %v", err)
	}

	if len(listed) != 1 {
		t.Fatalf("len(listed) = %d, want 1", len(listed))
	}

	if listed[0].ID != created.ID {
		t.Fatalf("listed[0].ID = %q, want %q", listed[0].ID, created.ID)
	}
}

func TestMembershipRepositoryCreateAndListByChoirIDIntegration(t *testing.T) {
	ctx, queries, tx := openIntegrationTestQueries(t)
	createTempChoirsTable(t, ctx, tx)
	createTempUsersTable(t, ctx, tx)
	createTempChoirMembersTable(t, ctx, tx)

	tenant := getSeedTenant(t, ctx, queries)
	userRepository := NewUserRepository(queries)
	actor, err := userRepository.Create(ctx, moduleusers.CreateParams{
		TenantID: tenant.ID,
		Email:    "manager@example.com",
		FullName: "Manager",
	})
	if err != nil {
		t.Fatalf("Create manager error = %v", err)
	}

	target, err := userRepository.Create(ctx, moduleusers.CreateParams{
		TenantID: tenant.ID,
		Email:    "member@example.com",
		FullName: "Member",
	})
	if err != nil {
		t.Fatalf("Create member error = %v", err)
	}

	choirRepository := NewChoirRepository(tx, queries)
	choir, err := choirRepository.Create(ctx, choirs.CreateParams{
		ActorUserID: actor.ID,
		TenantID:    tenant.ID,
		Name:        "Altos",
	})
	if err != nil {
		t.Fatalf("Create choir error = %v", err)
	}

	repository := NewMembershipRepository(queries)
	created, err := repository.Create(ctx, memberships.CreateParams{
		TenantID: tenant.ID,
		ChoirID:  choir.ID,
		UserID:   target.ID,
		Role:     memberships.RoleMember,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	listed, err := repository.ListByChoirID(ctx, tenant.ID, choir.ID)
	if err != nil {
		t.Fatalf("ListByChoirID() error = %v", err)
	}

	if len(listed) != 2 {
		t.Fatalf("len(listed) = %d, want 2", len(listed))
	}

	if created.UserID != target.ID {
		t.Fatalf("created.UserID = %q, want %q", created.UserID, target.ID)
	}
}

func TestDeviceRepositoryCreateListAndDeactivateIntegration(t *testing.T) {
	ctx, queries, tx := openIntegrationTestQueries(t)
	createTempUsersTable(t, ctx, tx)
	createTempDeviceTokensTable(t, ctx, tx)

	tenant := getSeedTenant(t, ctx, queries)
	userRepository := NewUserRepository(queries)
	user, err := userRepository.Create(ctx, moduleusers.CreateParams{
		TenantID: tenant.ID,
		Email:    "device-owner@example.com",
		FullName: "Device Owner",
	})
	if err != nil {
		t.Fatalf("Create user error = %v", err)
	}

	repository := NewDeviceRepository(queries)
	created, err := repository.Create(ctx, devices.CreateParams{
		TenantID: tenant.ID,
		UserID:   user.ID,
		Platform: devices.PlatformAndroid,
		Token:    "token-1",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	listed, err := repository.ListActiveByUserID(ctx, tenant.ID, user.ID)
	if err != nil {
		t.Fatalf("ListActiveByUserID() error = %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("len(listed) = %d, want 1", len(listed))
	}
	if listed[0].ID != created.ID {
		t.Fatalf("listed[0].ID = %q, want %q", listed[0].ID, created.ID)
	}

	if err := repository.DeactivateByToken(ctx, tenant.ID, "token-1"); err != nil {
		t.Fatalf("DeactivateByToken() error = %v", err)
	}

	listed, err = repository.ListActiveByUserID(ctx, tenant.ID, user.ID)
	if err != nil {
		t.Fatalf("ListActiveByUserID() after deactivate error = %v", err)
	}
	if len(listed) != 0 {
		t.Fatalf("len(listed) after deactivate = %d, want 0", len(listed))
	}
}

func TestVoiceKitRepositoryCreateGetListAndDeleteIntegration(t *testing.T) {
	ctx, queries, tx := openIntegrationTestQueries(t)
	createTempChoirsTable(t, ctx, tx)
	createTempUsersTable(t, ctx, tx)
	createTempChoirMembersTable(t, ctx, tx)
	createTempVoiceKitsTable(t, ctx, tx)

	tenant := getSeedTenant(t, ctx, queries)
	userRepository := NewUserRepository(queries)
	actor, err := userRepository.Create(ctx, moduleusers.CreateParams{
		TenantID: tenant.ID,
		Email:    "manager@example.com",
		FullName: "Manager",
	})
	if err != nil {
		t.Fatalf("Create actor user error = %v", err)
	}

	choirRepository := NewChoirRepository(tx, queries)
	choir, err := choirRepository.Create(ctx, choirs.CreateParams{
		ActorUserID: actor.ID,
		TenantID:    tenant.ID,
		Name:        "Altos",
	})
	if err != nil {
		t.Fatalf("Create choir error = %v", err)
	}

	repository := NewVoiceKitRepository(queries)
	description := "Warmup tracks"
	created, err := repository.Create(ctx, voicekits.CreateParams{
		TenantID:    tenant.ID,
		ChoirID:     choir.ID,
		Name:        "Warmups",
		Description: &description,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := repository.GetByIDForMember(ctx, tenant.ID, created.ID, actor.ID)
	if err != nil {
		t.Fatalf("GetByIDForMember() error = %v", err)
	}

	if got.ID != created.ID {
		t.Fatalf("got.ID = %q, want %q", got.ID, created.ID)
	}

	listed, err := repository.ListByChoirID(ctx, tenant.ID, choir.ID)
	if err != nil {
		t.Fatalf("ListByChoirID() error = %v", err)
	}

	if len(listed) != 1 {
		t.Fatalf("len(listed) = %d, want 1", len(listed))
	}

	if err := repository.Delete(ctx, tenant.ID, created.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err = repository.GetByIDForMember(ctx, tenant.ID, created.ID, actor.ID)
	if !errors.Is(err, voicekits.ErrVoiceKitNotFound) {
		t.Fatalf("GetByIDForMember() after delete error = %v, want %v", err, voicekits.ErrVoiceKitNotFound)
	}
}

func TestFileRepositoryCreateListAndDeleteIntegration(t *testing.T) {
	ctx, queries, tx := openIntegrationTestQueries(t)
	createTempChoirsTable(t, ctx, tx)
	createTempUsersTable(t, ctx, tx)
	createTempChoirMembersTable(t, ctx, tx)
	createTempVoiceKitsTable(t, ctx, tx)
	createTempKitFilesTable(t, ctx, tx)

	tenant := getSeedTenant(t, ctx, queries)
	userRepository := NewUserRepository(queries)
	actor, err := userRepository.Create(ctx, moduleusers.CreateParams{
		TenantID: tenant.ID,
		Email:    "manager@example.com",
		FullName: "Manager",
	})
	if err != nil {
		t.Fatalf("Create actor user error = %v", err)
	}

	choirRepository := NewChoirRepository(tx, queries)
	choir, err := choirRepository.Create(ctx, choirs.CreateParams{
		ActorUserID: actor.ID,
		TenantID:    tenant.ID,
		Name:        "Altos",
	})
	if err != nil {
		t.Fatalf("Create choir error = %v", err)
	}

	voiceKitRepository := NewVoiceKitRepository(queries)
	voiceKit, err := voiceKitRepository.Create(ctx, voicekits.CreateParams{
		TenantID: tenant.ID,
		ChoirID:  choir.ID,
		Name:     "Warmups",
	})
	if err != nil {
		t.Fatalf("Create voice kit error = %v", err)
	}

	repository := NewFileRepository(queries)
	created, err := repository.Create(ctx, modulefiles.CreateParams{
		ID:               "8f01f767-68e5-4e99-9cc6-6dfe0fdfd1d7",
		TenantID:         tenant.ID,
		VoiceKitID:       voiceKit.ID,
		OriginalFilename: "score.pdf",
		StoredFilename:   "stored-score.pdf",
		ContentType:      "application/pdf",
		SizeBytes:        128,
		StorageKey:       "dev/tenants/coral-jovem-asa-norte/choirs/" + choir.ID + "/voice-kits/" + voiceKit.ID + "/files/file-1/stored-score.pdf",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := repository.GetByIDForMember(ctx, tenant.ID, created.ID, actor.ID)
	if err != nil {
		t.Fatalf("GetByIDForMember() error = %v", err)
	}

	if got.ID != created.ID {
		t.Fatalf("got.ID = %q, want %q", got.ID, created.ID)
	}

	listed, err := repository.ListByVoiceKitID(ctx, tenant.ID, voiceKit.ID)
	if err != nil {
		t.Fatalf("ListByVoiceKitID() error = %v", err)
	}

	if len(listed) != 1 {
		t.Fatalf("len(listed) = %d, want 1", len(listed))
	}

	if err := repository.Delete(ctx, tenant.ID, created.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err = repository.GetByIDForMember(ctx, tenant.ID, created.ID, actor.ID)
	if !errors.Is(err, modulefiles.ErrFileNotFound) {
		t.Fatalf("GetByIDForMember() after delete error = %v, want %v", err, modulefiles.ErrFileNotFound)
	}
}

func TestEventRepositoryCreateUpdateAndCancelIntegration(t *testing.T) {
	ctx, queries, tx := openIntegrationTestQueries(t)
	createTempChoirsTable(t, ctx, tx)
	createTempUsersTable(t, ctx, tx)
	createTempChoirMembersTable(t, ctx, tx)
	createTempEventsTable(t, ctx, tx)
	createTempScheduledNotificationsTable(t, ctx, tx)

	tenant := getSeedTenant(t, ctx, queries)
	userRepository := NewUserRepository(queries)
	manager, err := userRepository.Create(ctx, moduleusers.CreateParams{
		TenantID: tenant.ID,
		Email:    "manager@example.com",
		FullName: "Manager",
	})
	if err != nil {
		t.Fatalf("Create manager error = %v", err)
	}

	member, err := userRepository.Create(ctx, moduleusers.CreateParams{
		TenantID: tenant.ID,
		Email:    "member@example.com",
		FullName: "Member",
	})
	if err != nil {
		t.Fatalf("Create member error = %v", err)
	}

	choirRepository := NewChoirRepository(tx, queries)
	choir, err := choirRepository.Create(ctx, choirs.CreateParams{
		ActorUserID: manager.ID,
		TenantID:    tenant.ID,
		Name:        "Events Choir",
	})
	if err != nil {
		t.Fatalf("Create choir error = %v", err)
	}

	membershipRepository := NewMembershipRepository(queries)
	if _, err := membershipRepository.Create(ctx, memberships.CreateParams{
		TenantID: tenant.ID,
		ChoirID:  choir.ID,
		UserID:   member.ID,
		Role:     memberships.RoleMember,
	}); err != nil {
		t.Fatalf("Create member membership error = %v", err)
	}

	repository := NewEventRepository(tx, queries)
	startAt := time.Date(2026, 4, 20, 19, 0, 0, 0, time.UTC)
	created, err := repository.Create(ctx, events.CreateParams{
		TenantID:  tenant.ID,
		ChoirID:   choir.ID,
		Title:     "Main Rehearsal",
		EventType: events.EventTypeRehearsal,
		StartAt:   startAt,
		Reminders: []events.ScheduledReminder{
			{UserID: manager.ID, ReminderType: events.ReminderTypeDayBefore, ScheduledFor: startAt.Add(-24 * time.Hour), Status: events.NotificationStatusPending},
			{UserID: manager.ID, ReminderType: events.ReminderTypeHourBefore, ScheduledFor: startAt.Add(-1 * time.Hour), Status: events.NotificationStatusPending},
			{UserID: member.ID, ReminderType: events.ReminderTypeDayBefore, ScheduledFor: startAt.Add(-24 * time.Hour), Status: events.NotificationStatusPending},
			{UserID: member.ID, ReminderType: events.ReminderTypeHourBefore, ScheduledFor: startAt.Add(-1 * time.Hour), Status: events.NotificationStatusPending},
		},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := repository.GetByIDForMember(ctx, tenant.ID, created.ID, manager.ID)
	if err != nil {
		t.Fatalf("GetByIDForMember() error = %v", err)
	}
	if got.ID != created.ID {
		t.Fatalf("got.ID = %q, want %q", got.ID, created.ID)
	}

	tenantUUID, err := parseUUID(tenant.ID)
	if err != nil {
		t.Fatalf("parseUUID(tenant.ID) error = %v", err)
	}
	eventUUID, err := parseUUID(created.ID)
	if err != nil {
		t.Fatalf("parseUUID(created.ID) error = %v", err)
	}

	scheduled, err := queries.ListScheduledNotificationsByEventID(ctx, sqlc.ListScheduledNotificationsByEventIDParams{
		TenantID: tenantUUID,
		EventID:  eventUUID,
	})
	if err != nil {
		t.Fatalf("ListScheduledNotificationsByEventID() after create error = %v", err)
	}
	if len(scheduled) != 4 {
		t.Fatalf("len(scheduled) after create = %d, want 4", len(scheduled))
	}

	updatedStartAt := startAt.Add(48 * time.Hour)
	updated, err := repository.Update(ctx, events.UpdateParams{
		TenantID:  tenant.ID,
		EventID:   created.ID,
		Title:     "Main Rehearsal Updated",
		EventType: events.EventTypePresentation,
		StartAt:   updatedStartAt,
		Reminders: []events.ScheduledReminder{
			{UserID: manager.ID, ReminderType: events.ReminderTypeHourBefore, ScheduledFor: updatedStartAt.Add(-1 * time.Hour), Status: events.NotificationStatusPending},
			{UserID: member.ID, ReminderType: events.ReminderTypeHourBefore, ScheduledFor: updatedStartAt.Add(-1 * time.Hour), Status: events.NotificationStatusPending},
		},
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Title != "Main Rehearsal Updated" {
		t.Fatalf("updated.Title = %q, want %q", updated.Title, "Main Rehearsal Updated")
	}

	scheduled, err = queries.ListScheduledNotificationsByEventID(ctx, sqlc.ListScheduledNotificationsByEventIDParams{
		TenantID: tenantUUID,
		EventID:  eventUUID,
	})
	if err != nil {
		t.Fatalf("ListScheduledNotificationsByEventID() after update error = %v", err)
	}
	if len(scheduled) != 6 {
		t.Fatalf("len(scheduled) after update = %d, want 6", len(scheduled))
	}

	pendingCount := 0
	canceledCount := 0
	for _, item := range scheduled {
		switch item.Status {
		case events.NotificationStatusPending:
			pendingCount++
		case events.NotificationStatusCanceled:
			canceledCount++
		}
	}
	if pendingCount != 2 {
		t.Fatalf("pendingCount after update = %d, want 2", pendingCount)
	}
	if canceledCount != 4 {
		t.Fatalf("canceledCount after update = %d, want 4", canceledCount)
	}

	if err := repository.Cancel(ctx, tenant.ID, created.ID); err != nil {
		t.Fatalf("Cancel() error = %v", err)
	}

	scheduled, err = queries.ListScheduledNotificationsByEventID(ctx, sqlc.ListScheduledNotificationsByEventIDParams{
		TenantID: tenantUUID,
		EventID:  eventUUID,
	})
	if err != nil {
		t.Fatalf("ListScheduledNotificationsByEventID() after cancel error = %v", err)
	}
	for _, item := range scheduled {
		if item.Status != events.NotificationStatusCanceled {
			t.Fatalf("scheduled notification status = %q, want %q", item.Status, events.NotificationStatusCanceled)
		}
	}

	_, err = repository.GetByIDForMember(ctx, tenant.ID, created.ID, manager.ID)
	if !errors.Is(err, events.ErrEventNotFound) {
		t.Fatalf("GetByIDForMember() after cancel error = %v, want %v", err, events.ErrEventNotFound)
	}
}

func TestNotificationRepositoryClaimAndStateTransitionsIntegration(t *testing.T) {
	ctx, queries, tx := openIntegrationTestQueries(t)
	createTempScheduledNotificationsTable(t, ctx, tx)

	tenant := getSeedTenant(t, ctx, queries)
	tenantUUID, err := parseUUID(tenant.ID)
	if err != nil {
		t.Fatalf("parseUUID(tenant.ID) error = %v", err)
	}

	eventID, err := parseUUID("8f01f767-68e5-4e99-9cc6-6dfe0fdfd1d7")
	if err != nil {
		t.Fatalf("parseUUID(eventID) error = %v", err)
	}

	userID, err := parseUUID("4fbc4fb2-cdbe-45d8-a91b-f48862b68ebf")
	if err != nil {
		t.Fatalf("parseUUID(userID) error = %v", err)
	}

	first, err := queries.CreateScheduledNotification(ctx, sqlc.CreateScheduledNotificationParams{
		TenantID:     tenantUUID,
		EventID:      eventID,
		UserID:       userID,
		ReminderType: "day_before",
		ScheduledFor: timestamptzValue(time.Date(2026, 4, 20, 18, 0, 0, 0, time.UTC)),
		Status:       notifications.StatusPending,
	})
	if err != nil {
		t.Fatalf("CreateScheduledNotification() first error = %v", err)
	}

	second, err := queries.CreateScheduledNotification(ctx, sqlc.CreateScheduledNotificationParams{
		TenantID:     tenantUUID,
		EventID:      eventID,
		UserID:       userID,
		ReminderType: "hour_before",
		ScheduledFor: timestamptzValue(time.Date(2026, 4, 20, 19, 0, 0, 0, time.UTC)),
		Status:       notifications.StatusPending,
	})
	if err != nil {
		t.Fatalf("CreateScheduledNotification() second error = %v", err)
	}

	third, err := queries.CreateScheduledNotification(ctx, sqlc.CreateScheduledNotificationParams{
		TenantID:     tenantUUID,
		EventID:      eventID,
		UserID:       userID,
		ReminderType: "custom_retry",
		ScheduledFor: timestamptzValue(time.Date(2026, 4, 20, 19, 30, 0, 0, time.UTC)),
		Status:       notifications.StatusPending,
	})
	if err != nil {
		t.Fatalf("CreateScheduledNotification() third error = %v", err)
	}

	_, err = tx.Exec(ctx, `
		UPDATE scheduled_notifications
		SET status = 'processing',
			processing_started_at = $2
		WHERE id = $1
	`, third.ID, time.Date(2026, 4, 20, 17, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("setting stale processing notification: %v", err)
	}

	repository := NewNotificationRepository(queries)
	claimedAt := time.Date(2026, 4, 20, 20, 0, 0, 0, time.UTC)
	claimed, err := repository.ClaimDue(ctx, notifications.ClaimParams{
		ClaimedAt:   claimedAt,
		StaleBefore: claimedAt.Add(-30 * time.Second),
		Limit:       2,
	})
	if err != nil {
		t.Fatalf("ClaimDue() error = %v", err)
	}

	if len(claimed) != 2 {
		t.Fatalf("len(claimed) = %d, want 2", len(claimed))
	}

	if err := repository.MarkSent(ctx, notifications.FinalizeParams{
		TenantID:            claimed[0].TenantID,
		NotificationID:      claimed[0].ID,
		ProcessingStartedAt: *claimed[0].ProcessingStartedAt,
		At:                  claimedAt,
	}); err != nil {
		t.Fatalf("MarkSent() error = %v", err)
	}

	if err := repository.Retry(ctx, notifications.RetryParams{
		TenantID:            claimed[1].TenantID,
		NotificationID:      claimed[1].ID,
		ProcessingStartedAt: *claimed[1].ProcessingStartedAt,
		NextAttemptAt:       claimedAt.Add(time.Minute),
		LastError:           "temporary failure",
	}); err != nil {
		t.Fatalf("Retry() error = %v", err)
	}

	claimed, err = repository.ClaimDue(ctx, notifications.ClaimParams{
		ClaimedAt:   claimedAt.Add(2 * time.Minute),
		StaleBefore: claimedAt.Add(90 * time.Second),
		Limit:       5,
	})
	if err != nil {
		t.Fatalf("ClaimDue() second error = %v", err)
	}
	if len(claimed) != 2 {
		t.Fatalf("len(claimed) second = %d, want 2", len(claimed))
	}

	claimedByID := map[string]notifications.Notification{}
	for _, item := range claimed {
		claimedByID[item.ID] = item
	}

	retriedNotification, ok := claimedByID[uuidString(second.ID)]
	if !ok {
		t.Fatalf("claimed notifications missing retried row %q", uuidString(second.ID))
	}

	staleNotification, ok := claimedByID[uuidString(third.ID)]
	if !ok {
		t.Fatalf("claimed notifications missing stale row %q", uuidString(third.ID))
	}

	if err := repository.MarkInvalidToken(ctx, notifications.FinalizeParams{
		TenantID:            retriedNotification.TenantID,
		NotificationID:      retriedNotification.ID,
		ProcessingStartedAt: *retriedNotification.ProcessingStartedAt,
		LastError:           "invalid token",
	}); err != nil {
		t.Fatalf("MarkInvalidToken() error = %v", err)
	}

	if err := repository.MarkFailed(ctx, notifications.FinalizeParams{
		TenantID:            staleNotification.TenantID,
		NotificationID:      staleNotification.ID,
		ProcessingStartedAt: *staleNotification.ProcessingStartedAt,
		LastError:           "max attempts reached",
	}); err != nil {
		t.Fatalf("MarkFailed() error = %v", err)
	}

	rows, err := queries.ListScheduledNotificationsByEventID(ctx, sqlc.ListScheduledNotificationsByEventIDParams{
		TenantID: tenantUUID,
		EventID:  eventID,
	})
	if err != nil {
		t.Fatalf("ListScheduledNotificationsByEventID() error = %v", err)
	}
	if len(rows) != 3 {
		t.Fatalf("len(rows) = %d, want 3", len(rows))
	}

	statuses := map[string]string{}
	attempts := map[string]int32{}
	for _, row := range rows {
		statuses[uuidString(row.ID)] = row.Status
		attempts[uuidString(row.ID)] = row.Attempts
	}

	if statuses[uuidString(first.ID)] != notifications.StatusSent {
		t.Fatalf("first status = %q, want %q", statuses[uuidString(first.ID)], notifications.StatusSent)
	}
	if attempts[uuidString(first.ID)] != 1 {
		t.Fatalf("first attempts = %d, want 1", attempts[uuidString(first.ID)])
	}
	if statuses[uuidString(second.ID)] != notifications.StatusInvalidToken {
		t.Fatalf("second status = %q, want %q", statuses[uuidString(second.ID)], notifications.StatusInvalidToken)
	}
	if attempts[uuidString(second.ID)] != 1 {
		t.Fatalf("second attempts = %d, want 1", attempts[uuidString(second.ID)])
	}
	if statuses[uuidString(third.ID)] != notifications.StatusFailed {
		t.Fatalf("third status = %q, want %q", statuses[uuidString(third.ID)], notifications.StatusFailed)
	}
	if attempts[uuidString(third.ID)] != 2 {
		t.Fatalf("third attempts = %d, want 2", attempts[uuidString(third.ID)])
	}
}

func openIntegrationTestQueries(t *testing.T) (context.Context, *sqlc.Queries, pgx.Tx) {
	t.Helper()

	cfg, err := platformconfig.Load()
	if err != nil {
		t.Skipf("integration config unavailable: %v", err)
	}

	ctx := context.Background()
	pool, err := NewPool(ctx, cfg.Database)
	if err != nil {
		t.Skipf("postgres unavailable for integration test: %v", err)
	}
	t.Cleanup(pool.Close)

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("Begin() error = %v", err)
	}

	t.Cleanup(func() {
		_ = tx.Rollback(ctx)
	})

	return ctx, sqlc.New(tx), tx
}

func getSeedTenant(t *testing.T, ctx context.Context, queries *sqlc.Queries) struct{ ID string } {
	t.Helper()

	row, err := queries.GetTenantBySlug(ctx, "coral-jovem-asa-norte")
	if err != nil {
		t.Fatalf("GetTenantBySlug() error = %v", err)
	}

	return struct{ ID string }{ID: uuidString(row.ID)}
}

func createTempChoirsTable(t *testing.T, ctx context.Context, tx pgx.Tx) {
	t.Helper()

	_, err := tx.Exec(ctx, `
		CREATE TEMP TABLE choirs (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID NOT NULL,
			name TEXT NOT NULL,
			description TEXT,
			active BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CONSTRAINT choirs_tenant_name_unique UNIQUE (tenant_id, name)
		) ON COMMIT DROP;
	`)
	if err != nil {
		t.Fatalf("creating temp choirs table: %v", err)
	}
}

func createTempUsersTable(t *testing.T, ctx context.Context, tx pgx.Tx) {
	t.Helper()

	_, err := tx.Exec(ctx, `
		CREATE TEMP TABLE users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID NOT NULL,
			email TEXT NOT NULL,
			full_name TEXT NOT NULL,
			active BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CONSTRAINT users_tenant_email_unique UNIQUE (tenant_id, email)
		) ON COMMIT DROP;
	`)
	if err != nil {
		t.Fatalf("creating temp users table: %v", err)
	}
}

func createTempChoirMembersTable(t *testing.T, ctx context.Context, tx pgx.Tx) {
	t.Helper()

	_, err := tx.Exec(ctx, `
		CREATE TEMP TABLE choir_members (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID NOT NULL,
			choir_id UUID NOT NULL,
			user_id UUID NOT NULL,
			role TEXT NOT NULL,
			active BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CONSTRAINT choir_members_role_check CHECK (role IN ('manager', 'member')),
			CONSTRAINT choir_members_tenant_choir_user_unique UNIQUE (tenant_id, choir_id, user_id)
		) ON COMMIT DROP;
	`)
	if err != nil {
		t.Fatalf("creating temp choir_members table: %v", err)
	}
}

func createTempVoiceKitsTable(t *testing.T, ctx context.Context, tx pgx.Tx) {
	t.Helper()

	_, err := tx.Exec(ctx, `
		CREATE TEMP TABLE voice_kits (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID NOT NULL,
			choir_id UUID NOT NULL,
			name TEXT NOT NULL,
			description TEXT,
			active BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CONSTRAINT voice_kits_tenant_choir_name_unique UNIQUE (tenant_id, choir_id, name)
		) ON COMMIT DROP;
	`)
	if err != nil {
		t.Fatalf("creating temp voice_kits table: %v", err)
	}
}

func createTempKitFilesTable(t *testing.T, ctx context.Context, tx pgx.Tx) {
	t.Helper()

	_, err := tx.Exec(ctx, `
		CREATE TEMP TABLE kit_files (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID NOT NULL,
			voice_kit_id UUID NOT NULL,
			original_filename TEXT NOT NULL,
			stored_filename TEXT NOT NULL,
			content_type TEXT NOT NULL,
			size_bytes BIGINT NOT NULL,
			storage_key TEXT NOT NULL,
			active BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CONSTRAINT kit_files_size_bytes_positive CHECK (size_bytes > 0)
		) ON COMMIT DROP;
	`)
	if err != nil {
		t.Fatalf("creating temp kit_files table: %v", err)
	}
}

func createTempDeviceTokensTable(t *testing.T, ctx context.Context, tx pgx.Tx) {
	t.Helper()

	_, err := tx.Exec(ctx, `
		CREATE TEMP TABLE device_tokens (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID NOT NULL,
			user_id UUID NOT NULL,
			platform TEXT NOT NULL,
			token TEXT NOT NULL,
			active BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CONSTRAINT device_tokens_platform_check CHECK (platform IN ('ios', 'android', 'web')),
			CONSTRAINT device_tokens_tenant_token_unique UNIQUE (tenant_id, token)
		) ON COMMIT DROP;
		CREATE INDEX device_tokens_user_active_idx ON device_tokens (user_id, active);
	`)
	if err != nil {
		t.Fatalf("creating temp device_tokens table: %v", err)
	}
}

func createTempEventsTable(t *testing.T, ctx context.Context, tx pgx.Tx) {
	t.Helper()

	_, err := tx.Exec(ctx, `
		CREATE TEMP TABLE events (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID NOT NULL,
			choir_id UUID NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			event_type TEXT NOT NULL,
			location TEXT,
			start_at TIMESTAMPTZ NOT NULL,
			active BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CONSTRAINT events_event_type_check CHECK (event_type IN ('rehearsal', 'presentation', 'other'))
		) ON COMMIT DROP;
	`)
	if err != nil {
		t.Fatalf("creating temp events table: %v", err)
	}
}

func createTempScheduledNotificationsTable(t *testing.T, ctx context.Context, tx pgx.Tx) {
	t.Helper()

	_, err := tx.Exec(ctx, `
		CREATE TEMP TABLE scheduled_notifications (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID NOT NULL,
			event_id UUID NOT NULL,
			user_id UUID NOT NULL,
			reminder_type TEXT NOT NULL,
			scheduled_for TIMESTAMPTZ NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			attempts INTEGER NOT NULL DEFAULT 0,
			last_error TEXT,
			processing_started_at TIMESTAMPTZ,
			sent_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CONSTRAINT scheduled_notifications_status_check CHECK (status IN ('pending', 'processing', 'sent', 'failed', 'canceled', 'invalid_token')),
			CONSTRAINT scheduled_notifications_attempts_non_negative_check CHECK (attempts >= 0)
		) ON COMMIT DROP;
		CREATE UNIQUE INDEX scheduled_notifications_pending_identity_idx
			ON scheduled_notifications (tenant_id, user_id, event_id, reminder_type)
			WHERE status IN ('pending', 'processing');
	`)
	if err != nil {
		t.Fatalf("creating temp scheduled_notifications table: %v", err)
	}
}
