package notifications

import (
	"context"
	"errors"
	"testing"
	"time"
)

type stubRepository struct {
	claimed       []Notification
	claimErr      error
	sent          []FinalizeParams
	retried       []RetryParams
	failed        []FinalizeParams
	invalid       []FinalizeParams
	cleanupBefore []time.Time
	cleanupCount  int64
	finalizeErr   error
	retryErr      error
}

func (s *stubRepository) ClaimDue(_ context.Context, _ ClaimParams) ([]Notification, error) {
	if s.claimErr != nil {
		return nil, s.claimErr
	}
	return s.claimed, nil
}

func (s *stubRepository) MarkSent(_ context.Context, params FinalizeParams) error {
	if s.finalizeErr != nil {
		return s.finalizeErr
	}
	s.sent = append(s.sent, params)
	return nil
}

func (s *stubRepository) Retry(_ context.Context, params RetryParams) error {
	if s.retryErr != nil {
		return s.retryErr
	}
	s.retried = append(s.retried, params)
	return nil
}

func (s *stubRepository) MarkFailed(_ context.Context, params FinalizeParams) error {
	if s.finalizeErr != nil {
		return s.finalizeErr
	}
	s.failed = append(s.failed, params)
	return nil
}

func (s *stubRepository) MarkInvalidToken(_ context.Context, params FinalizeParams) error {
	if s.finalizeErr != nil {
		return s.finalizeErr
	}
	s.invalid = append(s.invalid, params)
	return nil
}

func (s *stubRepository) CleanupTerminalBefore(_ context.Context, before time.Time) (int64, error) {
	s.cleanupBefore = append(s.cleanupBefore, before)
	return s.cleanupCount, nil
}

type stubSender struct {
	results []DeliveryResult
}

func (s *stubSender) Deliver(_ context.Context, _ Notification) DeliveryResult {
	result := s.results[0]
	s.results = s.results[1:]
	return result
}

func TestServiceProcessDueMarksSent(t *testing.T) {
	claimedAt := time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)
	repository := &stubRepository{
		claimed: []Notification{{
			ID:                  "notification-1",
			TenantID:            "tenant-1",
			Attempts:            0,
			ProcessingStartedAt: timePointer(claimedAt),
		}},
	}
	service := NewService(repository, &stubSender{
		results: []DeliveryResult{{Kind: DeliverySent}},
	}, 3, time.Minute)
	service.now = func() time.Time { return claimedAt }

	processed, err := service.ProcessDue(context.Background(), 10, 30*time.Second)
	if err != nil {
		t.Fatalf("ProcessDue() error = %v", err)
	}
	if processed != 1 {
		t.Fatalf("processed = %d, want 1", processed)
	}
	if len(repository.sent) != 1 {
		t.Fatalf("len(repository.sent) = %d, want 1", len(repository.sent))
	}
}

func TestServiceProcessDueRetriesTransientFailure(t *testing.T) {
	claimedAt := time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)
	repository := &stubRepository{
		claimed: []Notification{{
			ID:                  "notification-1",
			TenantID:            "tenant-1",
			Attempts:            0,
			ProcessingStartedAt: timePointer(claimedAt),
		}},
	}
	service := NewService(repository, &stubSender{
		results: []DeliveryResult{{Kind: DeliveryTransientFailure, ErrorMessage: "temporary failure"}},
	}, 3, 2*time.Minute)
	service.now = func() time.Time { return claimedAt }

	_, err := service.ProcessDue(context.Background(), 10, 30*time.Second)
	if err != nil {
		t.Fatalf("ProcessDue() error = %v", err)
	}
	if len(repository.retried) != 1 {
		t.Fatalf("len(repository.retried) = %d, want 1", len(repository.retried))
	}
	if repository.retried[0].NextAttemptAt != claimedAt.Add(2*time.Minute) {
		t.Fatalf("NextAttemptAt = %v, want %v", repository.retried[0].NextAttemptAt, claimedAt.Add(2*time.Minute))
	}
}

func TestServiceProcessDueMarksFailedAtMaxAttempts(t *testing.T) {
	claimedAt := time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)
	repository := &stubRepository{
		claimed: []Notification{{
			ID:                  "notification-1",
			TenantID:            "tenant-1",
			Attempts:            2,
			ProcessingStartedAt: timePointer(claimedAt),
		}},
	}
	service := NewService(repository, &stubSender{
		results: []DeliveryResult{{Kind: DeliveryTransientFailure, ErrorMessage: "temporary failure"}},
	}, 3, time.Minute)
	service.now = func() time.Time { return claimedAt }

	_, err := service.ProcessDue(context.Background(), 10, 30*time.Second)
	if err != nil {
		t.Fatalf("ProcessDue() error = %v", err)
	}
	if len(repository.failed) != 1 {
		t.Fatalf("len(repository.failed) = %d, want 1", len(repository.failed))
	}
}

func TestServiceProcessDueMarksInvalidToken(t *testing.T) {
	claimedAt := time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)
	repository := &stubRepository{
		claimed: []Notification{{
			ID:                  "notification-1",
			TenantID:            "tenant-1",
			ProcessingStartedAt: timePointer(claimedAt),
		}},
	}
	service := NewService(repository, &stubSender{
		results: []DeliveryResult{{Kind: DeliveryInvalidToken, ErrorMessage: "token expired"}},
	}, 3, time.Minute)
	service.now = func() time.Time { return claimedAt }

	_, err := service.ProcessDue(context.Background(), 10, 30*time.Second)
	if err != nil {
		t.Fatalf("ProcessDue() error = %v", err)
	}
	if len(repository.invalid) != 1 {
		t.Fatalf("len(repository.invalid) = %d, want 1", len(repository.invalid))
	}
}

func TestServiceProcessDueRejectsUnknownDeliveryResult(t *testing.T) {
	claimedAt := time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)
	repository := &stubRepository{
		claimed: []Notification{{
			ID:                  "notification-1",
			TenantID:            "tenant-1",
			ProcessingStartedAt: timePointer(claimedAt),
		}},
	}
	service := NewService(repository, &stubSender{
		results: []DeliveryResult{{Kind: "mystery"}},
	}, 3, time.Minute)
	service.now = func() time.Time { return claimedAt }

	_, err := service.ProcessDue(context.Background(), 10, 30*time.Second)
	if !errors.Is(err, ErrUnknownDeliveryResult) {
		t.Fatalf("ProcessDue() error = %v, want %v", err, ErrUnknownDeliveryResult)
	}
}

func TestServiceCleanupExpiredUsesRetentionCutoff(t *testing.T) {
	now := time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)
	repository := &stubRepository{cleanupCount: 4}
	service := NewService(repository, &stubSender{}, 3, time.Minute)
	service.now = func() time.Time { return now }

	deleted, err := service.CleanupExpired(context.Background(), 24*time.Hour)
	if err != nil {
		t.Fatalf("CleanupExpired() error = %v", err)
	}
	if deleted != 4 {
		t.Fatalf("deleted = %d, want 4", deleted)
	}
	if len(repository.cleanupBefore) != 1 {
		t.Fatalf("len(repository.cleanupBefore) = %d, want 1", len(repository.cleanupBefore))
	}
	if repository.cleanupBefore[0] != now.Add(-24*time.Hour) {
		t.Fatalf("cleanupBefore = %v, want %v", repository.cleanupBefore[0], now.Add(-24*time.Hour))
	}
}

func timePointer(value time.Time) *time.Time {
	return &value
}
