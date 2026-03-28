package events

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/LeviLunique/coralhub-backend/internal/modules/memberships"
)

type stubRepository struct {
	event    Event
	events   []Event
	err      error
	create   CreateParams
	update   UpdateParams
	canceled CancelParams
}

func (s *stubRepository) Create(_ context.Context, params CreateParams) (Event, error) {
	s.create = params
	if s.err != nil {
		return Event{}, s.err
	}
	return s.event, nil
}

func (s *stubRepository) Update(_ context.Context, params UpdateParams) (Event, error) {
	s.update = params
	if s.err != nil {
		return Event{}, s.err
	}
	return s.event, nil
}

func (s *stubRepository) GetByIDForMember(_ context.Context, _, _, _ string) (Event, error) {
	if s.err != nil {
		return Event{}, s.err
	}
	return s.event, nil
}

func (s *stubRepository) ListByChoirID(_ context.Context, _, _ string) ([]Event, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.events, nil
}

func (s *stubRepository) Cancel(_ context.Context, params CancelParams) error {
	if s.err != nil {
		return s.err
	}
	s.canceled = params
	return nil
}

type stubMembershipReader struct {
	membership  memberships.Membership
	memberships []memberships.Membership
	err         error
}

func (s *stubMembershipReader) GetByChoirAndUser(_ context.Context, _, _, _ string) (memberships.Membership, error) {
	if s.err != nil {
		return memberships.Membership{}, s.err
	}
	return s.membership, nil
}

func (s *stubMembershipReader) ListByChoirID(_ context.Context, _, _ string) ([]memberships.Membership, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.memberships, nil
}

func TestBuildReminderSchedule(t *testing.T) {
	startAt := time.Date(2026, 4, 10, 18, 0, 0, 0, time.UTC)
	now := time.Date(2026, 4, 8, 12, 0, 0, 0, time.UTC)

	items := BuildReminderSchedule(startAt, []string{"user-1"}, now)
	if len(items) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(items))
	}

	if items[0].ReminderType != ReminderTypeDayBefore {
		t.Fatalf("items[0].ReminderType = %q", items[0].ReminderType)
	}

	if items[1].ReminderType != ReminderTypeHourBefore {
		t.Fatalf("items[1].ReminderType = %q", items[1].ReminderType)
	}
}

func TestServiceCreateRequiresManagerRole(t *testing.T) {
	service := NewService(&stubRepository{}, &stubMembershipReader{
		membership: memberships.Membership{Role: memberships.RoleMember},
	})
	_, err := service.Create(context.Background(), "tenant-1", "choir-1", "actor-1", CreateInput{
		Title:     "Rehearsal",
		EventType: EventTypeRehearsal,
		StartAt:   time.Now().Add(48 * time.Hour),
	})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("Create() error = %v, want %v", err, ErrForbidden)
	}
}

func TestServiceCreateBuildsReminders(t *testing.T) {
	repository := &stubRepository{
		event: Event{ID: "event-1"},
	}
	service := NewService(repository, &stubMembershipReader{
		membership: memberships.Membership{Role: memberships.RoleManager},
		memberships: []memberships.Membership{
			{UserID: "user-1"},
			{UserID: "user-2"},
		},
	})
	service.now = func() time.Time {
		return time.Date(2026, 4, 8, 12, 0, 0, 0, time.UTC)
	}

	_, err := service.Create(context.Background(), "tenant-1", "choir-1", "actor-1", CreateInput{
		Title:       "  Main Rehearsal  ",
		Description: stringPointer("  Bring scores  "),
		EventType:   " REHEARSAL ",
		Location:    stringPointer("  Hall A  "),
		StartAt:     time.Date(2026, 4, 10, 18, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if repository.create.Title != "Main Rehearsal" {
		t.Fatalf("repository.create.Title = %q", repository.create.Title)
	}

	if len(repository.create.Reminders) != 4 {
		t.Fatalf("len(repository.create.Reminders) = %d, want 4", len(repository.create.Reminders))
	}
}

func TestServiceCancelRequiresManagerRole(t *testing.T) {
	service := NewService(&stubRepository{
		event: Event{ID: "event-1", ChoirID: "choir-1"},
	}, &stubMembershipReader{
		membership: memberships.Membership{Role: memberships.RoleMember},
	})

	err := service.Cancel(context.Background(), "tenant-1", "event-1", "actor-1")
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("Cancel() error = %v, want %v", err, ErrForbidden)
	}
}

func stringPointer(value string) *string {
	return &value
}
