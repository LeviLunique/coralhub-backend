package events

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/LeviLunique/coralhub-backend/internal/modules/memberships"
)

var (
	ErrInvalidTenantID  = errors.New("invalid tenant id")
	ErrInvalidChoirID   = errors.New("invalid choir id")
	ErrInvalidEventID   = errors.New("invalid event id")
	ErrInvalidActorID   = errors.New("invalid actor id")
	ErrInvalidTitle     = errors.New("invalid title")
	ErrInvalidEventType = errors.New("invalid event type")
	ErrInvalidStartAt   = errors.New("invalid start at")
	ErrEventNotFound    = errors.New("event not found")
	ErrForbidden        = errors.New("forbidden")
)

type membershipReader interface {
	GetByChoirAndUser(ctx context.Context, tenantID string, choirID string, userID string) (memberships.Membership, error)
	ListByChoirID(ctx context.Context, tenantID string, choirID string) ([]memberships.Membership, error)
}

type Service struct {
	repository  Repository
	memberships membershipReader
	now         func() time.Time
}

func NewService(repository Repository, memberships membershipReader) *Service {
	return &Service{
		repository:  repository,
		memberships: memberships,
		now:         time.Now,
	}
}

func (s *Service) Create(ctx context.Context, tenantID string, choirID string, actorUserID string, input CreateInput) (Event, error) {
	normalizedTenantID, normalizedChoirID, normalizedActorID, normalizedTitle, eventType, description, location, startAt, err := normalizeCreateInput(tenantID, choirID, actorUserID, input)
	if err != nil {
		return Event{}, err
	}

	member, err := s.memberships.GetByChoirAndUser(ctx, normalizedTenantID, normalizedChoirID, normalizedActorID)
	if err != nil {
		if errors.Is(err, memberships.ErrMembershipNotFound) {
			return Event{}, ErrForbidden
		}
		return Event{}, err
	}
	if member.Role != memberships.RoleManager {
		return Event{}, ErrForbidden
	}

	items, err := s.memberships.ListByChoirID(ctx, normalizedTenantID, normalizedChoirID)
	if err != nil {
		return Event{}, err
	}

	return s.repository.Create(ctx, CreateParams{
		TenantID:    normalizedTenantID,
		ChoirID:     normalizedChoirID,
		ActorUserID: normalizedActorID,
		Title:       normalizedTitle,
		Description: description,
		EventType:   eventType,
		Location:    location,
		StartAt:     startAt,
		Reminders:   BuildReminderSchedule(startAt, membershipUserIDs(items), s.now().UTC()),
	})
}

func (s *Service) Update(ctx context.Context, tenantID string, eventID string, actorUserID string, input UpdateInput) (Event, error) {
	normalizedTenantID := strings.TrimSpace(tenantID)
	if normalizedTenantID == "" {
		return Event{}, ErrInvalidTenantID
	}

	normalizedEventID := strings.TrimSpace(eventID)
	if normalizedEventID == "" {
		return Event{}, ErrInvalidEventID
	}

	normalizedActorID := strings.TrimSpace(actorUserID)
	if normalizedActorID == "" {
		return Event{}, ErrInvalidActorID
	}

	existing, err := s.repository.GetByIDForMember(ctx, normalizedTenantID, normalizedEventID, normalizedActorID)
	if err != nil {
		return Event{}, err
	}

	member, err := s.memberships.GetByChoirAndUser(ctx, normalizedTenantID, existing.ChoirID, normalizedActorID)
	if err != nil {
		if errors.Is(err, memberships.ErrMembershipNotFound) {
			return Event{}, ErrForbidden
		}
		return Event{}, err
	}
	if member.Role != memberships.RoleManager {
		return Event{}, ErrForbidden
	}

	normalizedTitle, eventType, description, location, startAt, err := normalizeUpdateInput(input)
	if err != nil {
		return Event{}, err
	}

	items, err := s.memberships.ListByChoirID(ctx, normalizedTenantID, existing.ChoirID)
	if err != nil {
		return Event{}, err
	}

	return s.repository.Update(ctx, UpdateParams{
		TenantID:    normalizedTenantID,
		EventID:     normalizedEventID,
		ActorUserID: normalizedActorID,
		Title:       normalizedTitle,
		Description: description,
		EventType:   eventType,
		Location:    location,
		StartAt:     startAt,
		Reminders:   BuildReminderSchedule(startAt, membershipUserIDs(items), s.now().UTC()),
	})
}

func (s *Service) Get(ctx context.Context, tenantID string, actorUserID string, eventID string) (Event, error) {
	normalizedTenantID := strings.TrimSpace(tenantID)
	if normalizedTenantID == "" {
		return Event{}, ErrInvalidTenantID
	}

	normalizedActorID := strings.TrimSpace(actorUserID)
	if normalizedActorID == "" {
		return Event{}, ErrInvalidActorID
	}

	normalizedEventID := strings.TrimSpace(eventID)
	if normalizedEventID == "" {
		return Event{}, ErrInvalidEventID
	}

	return s.repository.GetByIDForMember(ctx, normalizedTenantID, normalizedEventID, normalizedActorID)
}

func (s *Service) ListByChoir(ctx context.Context, tenantID string, choirID string, actorUserID string) ([]Event, error) {
	normalizedTenantID := strings.TrimSpace(tenantID)
	if normalizedTenantID == "" {
		return nil, ErrInvalidTenantID
	}

	normalizedChoirID := strings.TrimSpace(choirID)
	if normalizedChoirID == "" {
		return nil, ErrInvalidChoirID
	}

	normalizedActorID := strings.TrimSpace(actorUserID)
	if normalizedActorID == "" {
		return nil, ErrInvalidActorID
	}

	if _, err := s.memberships.GetByChoirAndUser(ctx, normalizedTenantID, normalizedChoirID, normalizedActorID); err != nil {
		if errors.Is(err, memberships.ErrMembershipNotFound) {
			return nil, ErrForbidden
		}
		return nil, err
	}

	return s.repository.ListByChoirID(ctx, normalizedTenantID, normalizedChoirID)
}

func (s *Service) Cancel(ctx context.Context, tenantID string, eventID string, actorUserID string) error {
	normalizedTenantID := strings.TrimSpace(tenantID)
	if normalizedTenantID == "" {
		return ErrInvalidTenantID
	}

	normalizedEventID := strings.TrimSpace(eventID)
	if normalizedEventID == "" {
		return ErrInvalidEventID
	}

	normalizedActorID := strings.TrimSpace(actorUserID)
	if normalizedActorID == "" {
		return ErrInvalidActorID
	}

	existing, err := s.repository.GetByIDForMember(ctx, normalizedTenantID, normalizedEventID, normalizedActorID)
	if err != nil {
		return err
	}

	member, err := s.memberships.GetByChoirAndUser(ctx, normalizedTenantID, existing.ChoirID, normalizedActorID)
	if err != nil {
		if errors.Is(err, memberships.ErrMembershipNotFound) {
			return ErrForbidden
		}
		return err
	}
	if member.Role != memberships.RoleManager {
		return ErrForbidden
	}

	return s.repository.Cancel(ctx, CancelParams{
		TenantID:    normalizedTenantID,
		EventID:     normalizedEventID,
		ActorUserID: normalizedActorID,
	})
}

func normalizeCreateInput(tenantID string, choirID string, actorUserID string, input CreateInput) (string, string, string, string, string, *string, *string, time.Time, error) {
	normalizedTenantID := strings.TrimSpace(tenantID)
	if normalizedTenantID == "" {
		return "", "", "", "", "", nil, nil, time.Time{}, ErrInvalidTenantID
	}

	normalizedChoirID := strings.TrimSpace(choirID)
	if normalizedChoirID == "" {
		return "", "", "", "", "", nil, nil, time.Time{}, ErrInvalidChoirID
	}

	normalizedActorID := strings.TrimSpace(actorUserID)
	if normalizedActorID == "" {
		return "", "", "", "", "", nil, nil, time.Time{}, ErrInvalidActorID
	}

	title, eventType, description, location, startAt, err := normalizeEventPayload(input.Title, input.EventType, input.Description, input.Location, input.StartAt)
	if err != nil {
		return "", "", "", "", "", nil, nil, time.Time{}, err
	}

	return normalizedTenantID, normalizedChoirID, normalizedActorID, title, eventType, description, location, startAt, nil
}

func normalizeUpdateInput(input UpdateInput) (string, string, *string, *string, time.Time, error) {
	return normalizeEventPayload(input.Title, input.EventType, input.Description, input.Location, input.StartAt)
}

func normalizeEventPayload(title string, eventType string, description *string, location *string, startAt time.Time) (string, string, *string, *string, time.Time, error) {
	normalizedTitle := strings.TrimSpace(title)
	if normalizedTitle == "" {
		return "", "", nil, nil, time.Time{}, ErrInvalidTitle
	}

	normalizedEventType := strings.ToLower(strings.TrimSpace(eventType))
	switch normalizedEventType {
	case EventTypeRehearsal, EventTypePresentation, EventTypeOther:
	default:
		return "", "", nil, nil, time.Time{}, ErrInvalidEventType
	}

	if startAt.IsZero() {
		return "", "", nil, nil, time.Time{}, ErrInvalidStartAt
	}

	return normalizedTitle, normalizedEventType, normalizeOptionalText(description), normalizeOptionalText(location), startAt.UTC(), nil
}

func normalizeOptionalText(value *string) *string {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}

func membershipUserIDs(items []memberships.Membership) []string {
	userIDs := make([]string, 0, len(items))
	for _, item := range items {
		if strings.TrimSpace(item.UserID) == "" {
			continue
		}
		userIDs = append(userIDs, item.UserID)
	}

	return userIDs
}
