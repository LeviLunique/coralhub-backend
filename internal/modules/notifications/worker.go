package notifications

import (
	"context"
	"log/slog"
	"time"
)

type Worker struct {
	logger       *slog.Logger
	service      *Service
	pollInterval time.Duration
	batchSize    int32
	leaseTimeout time.Duration
}

func NewWorker(logger *slog.Logger, service *Service, pollInterval time.Duration, batchSize int32, leaseTimeout time.Duration) *Worker {
	return &Worker{
		logger:       logger,
		service:      service,
		pollInterval: pollInterval,
		batchSize:    batchSize,
		leaseTimeout: leaseTimeout,
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
	processed, err := w.service.ProcessDue(ctx, w.batchSize, w.leaseTimeout)
	if err != nil {
		return err
	}

	if processed > 0 {
		w.logger.InfoContext(ctx, "worker processed scheduled notifications", "count", processed)
	}

	return nil
}
