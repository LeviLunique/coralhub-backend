package notifications

import (
	"context"
	"errors"
	"strings"
	"time"

	platformobservability "github.com/LeviLunique/coralhub-backend/internal/platform/observability"
)

var (
	ErrInvalidClaimLimit     = errors.New("invalid claim limit")
	ErrInvalidLeaseTimeout   = errors.New("invalid lease timeout")
	ErrInvalidRetention      = errors.New("invalid retention period")
	ErrUnknownDeliveryResult = errors.New("unknown delivery result")
	ErrNotificationLeaseLost = errors.New("notification lease lost")
)

type Sender interface {
	Deliver(ctx context.Context, notification Notification) DeliveryResult
}

type Service struct {
	repository   Repository
	sender       Sender
	now          func() time.Time
	maxAttempts  int32
	retryBackoff time.Duration
}

func NewService(repository Repository, sender Sender, maxAttempts int32, retryBackoff time.Duration) *Service {
	if maxAttempts <= 0 {
		maxAttempts = 3
	}
	if retryBackoff <= 0 {
		retryBackoff = time.Minute
	}

	return &Service{
		repository:   repository,
		sender:       sender,
		now:          time.Now,
		maxAttempts:  maxAttempts,
		retryBackoff: retryBackoff,
	}
}

func (s *Service) ProcessDue(ctx context.Context, limit int32, leaseTimeout time.Duration) (int, error) {
	if limit <= 0 {
		return 0, ErrInvalidClaimLimit
	}
	if leaseTimeout <= 0 {
		return 0, ErrInvalidLeaseTimeout
	}

	claimedAt := s.now().UTC()
	notifications, err := s.repository.ClaimDue(ctx, ClaimParams{
		ClaimedAt:   claimedAt,
		StaleBefore: claimedAt.Add(-leaseTimeout),
		Limit:       limit,
	})
	if err != nil {
		return 0, err
	}

	for _, notification := range notifications {
		if err := s.processOne(ctx, notification, claimedAt); err != nil {
			if errors.Is(err, ErrNotificationLeaseLost) {
				continue
			}

			return 0, err
		}
	}

	return len(notifications), nil
}

func (s *Service) CleanupExpired(ctx context.Context, retention time.Duration) (int64, error) {
	if retention <= 0 {
		return 0, ErrInvalidRetention
	}

	return s.repository.CleanupTerminalBefore(ctx, s.now().UTC().Add(-retention))
}

func (s *Service) processOne(ctx context.Context, notification Notification, processedAt time.Time) error {
	if notification.ProcessingStartedAt == nil || notification.ProcessingStartedAt.IsZero() {
		return ErrNotificationLeaseLost
	}

	result := s.sender.Deliver(ctx, notification)
	lastError := normalizeErrorMessage(result.ErrorMessage)

	switch result.Kind {
	case DeliverySent:
		err := s.repository.MarkSent(ctx, FinalizeParams{
			TenantID:            notification.TenantID,
			NotificationID:      notification.ID,
			ProcessingStartedAt: notification.ProcessingStartedAt.UTC(),
			At:                  processedAt,
		})
		if err == nil {
			platformobservability.DefaultMetrics().IncrementNotificationDelivery(string(DeliverySent))
		}
		return err
	case DeliveryInvalidToken:
		if lastError == "" {
			lastError = "invalid token"
		}
		err := s.repository.MarkInvalidToken(ctx, FinalizeParams{
			TenantID:            notification.TenantID,
			NotificationID:      notification.ID,
			ProcessingStartedAt: notification.ProcessingStartedAt.UTC(),
			At:                  processedAt,
			LastError:           lastError,
		})
		if err == nil {
			platformobservability.DefaultMetrics().IncrementNotificationDelivery(string(DeliveryInvalidToken))
		}
		return err
	case DeliveryTransientFailure:
		if lastError == "" {
			lastError = "transient delivery failure"
		}
		if notification.Attempts+1 >= s.maxAttempts {
			err := s.repository.MarkFailed(ctx, FinalizeParams{
				TenantID:            notification.TenantID,
				NotificationID:      notification.ID,
				ProcessingStartedAt: notification.ProcessingStartedAt.UTC(),
				At:                  processedAt,
				LastError:           lastError,
			})
			if err == nil {
				platformobservability.DefaultMetrics().IncrementNotificationDelivery("failed")
			}
			return err
		}

		return s.repository.Retry(ctx, RetryParams{
			TenantID:            notification.TenantID,
			NotificationID:      notification.ID,
			ProcessingStartedAt: notification.ProcessingStartedAt.UTC(),
			NextAttemptAt:       processedAt.Add(s.retryBackoff),
			LastError:           lastError,
		})
	default:
		return ErrUnknownDeliveryResult
	}
}

func normalizeErrorMessage(value string) string {
	return strings.TrimSpace(value)
}
