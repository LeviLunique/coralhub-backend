package notifications

import (
	"context"
	"log/slog"
	"time"

	platformobservability "github.com/LeviLunique/coralhub-backend/internal/platform/observability"
)

type Worker struct {
	logger       *slog.Logger
	service      *Service
	pollInterval time.Duration
	batchSize    int32
	leaseTimeout time.Duration
	retention    time.Duration
}

func NewWorker(logger *slog.Logger, service *Service, pollInterval time.Duration, batchSize int32, leaseTimeout time.Duration, retention time.Duration) *Worker {
	return &Worker{
		logger:       logger,
		service:      service,
		pollInterval: pollInterval,
		batchSize:    batchSize,
		leaseTimeout: leaseTimeout,
		retention:    retention,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	if err := w.runCycle(ctx); err != nil {
		return err
	}

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := w.runCycle(ctx); err != nil {
				return err
			}
		}
	}
}

func (w *Worker) runCycle(ctx context.Context) error {
	platformobservability.DefaultMetrics().IncrementWorkerPoll()

	processed, err := w.service.ProcessDue(ctx, w.batchSize, w.leaseTimeout)
	if err != nil {
		return err
	}
	platformobservability.DefaultMetrics().AddWorkerProcessed(processed)

	if processed > 0 {
		w.logger.InfoContext(ctx, "worker processed scheduled notifications", "count", processed)
	}

	if w.retention > 0 {
		deleted, err := w.service.CleanupExpired(ctx, w.retention)
		if err != nil {
			return err
		}
		platformobservability.DefaultMetrics().AddNotificationCleanupDeleted(deleted)
		if deleted > 0 {
			w.logger.InfoContext(ctx, "worker cleaned expired scheduled notifications", "count", deleted)
		}
	}

	return nil
}
