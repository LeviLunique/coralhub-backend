package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/LeviLunique/coralhub-backend/internal/modules/notifications"
	platformconfig "github.com/LeviLunique/coralhub-backend/internal/platform/config"
	platformlog "github.com/LeviLunique/coralhub-backend/internal/platform/log"
	"github.com/LeviLunique/coralhub-backend/internal/store/postgres"
	"github.com/LeviLunique/coralhub-backend/internal/store/postgres/sqlc"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := platformconfig.Load()
	if err != nil {
		panic(err)
	}

	logger, err := platformlog.New(cfg.Observability.LogLevel)
	if err != nil {
		panic(err)
	}

	pool, err := postgres.NewPool(ctx, cfg.Database)
	if err != nil {
		logger.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	queries := sqlc.New(pool)
	notificationRepository := postgres.NewNotificationRepository(queries)
	notificationService := notifications.NewService(
		notificationRepository,
		noopSender{},
		cfg.Worker.MaxAttempts,
		cfg.Worker.RetryBackoff,
	)
	worker := notifications.NewWorker(
		logger,
		notificationService,
		cfg.Worker.PollInterval,
		cfg.Worker.BatchSize,
		cfg.Worker.LeaseTimeout,
	)

	logger.Info(
		"worker starting",
		"poll_interval",
		cfg.Worker.PollInterval.String(),
		"batch_size",
		cfg.Worker.BatchSize,
		"max_attempts",
		cfg.Worker.MaxAttempts,
		"env",
		cfg.AppEnv,
	)

	if err := worker.Run(ctx); err != nil {
		logger.Error("worker stopped unexpectedly", "error", err)
		os.Exit(1)
	}

	logger.Info("worker stopped")
}

type noopSender struct{}

func (noopSender) Deliver(_ context.Context, _ notifications.Notification) notifications.DeliveryResult {
	return notifications.DeliveryResult{Kind: notifications.DeliverySent}
}
