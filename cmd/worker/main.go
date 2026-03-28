package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	pushfcm "github.com/LeviLunique/coralhub-backend/internal/integrations/push/fcm"
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
	notificationRepository := postgres.NewNotificationRepository(pool, queries)
	deviceRepository := postgres.NewDeviceRepository(queries)

	sender := notifications.Sender(noopSender{})
	if cfg.Firebase.Enabled {
		fcmSender, err := pushfcm.New(ctx, cfg.Firebase, deviceRepository)
		if err != nil {
			logger.Error("failed to initialize fcm sender", "error", err)
			os.Exit(1)
		}
		sender = fcmSender
	} else {
		logger.Warn("firebase disabled; using noop notification sender")
	}

	notificationService := notifications.NewService(
		notificationRepository,
		sender,
		cfg.Worker.MaxAttempts,
		cfg.Worker.RetryBackoff,
	)
	worker := notifications.NewWorker(
		logger,
		notificationService,
		cfg.Worker.PollInterval,
		cfg.Worker.BatchSize,
		cfg.Worker.LeaseTimeout,
		cfg.Worker.NotificationRetention,
	)

	logger.Info(
		"worker starting",
		"poll_interval",
		cfg.Worker.PollInterval.String(),
		"batch_size",
		cfg.Worker.BatchSize,
		"max_attempts",
		cfg.Worker.MaxAttempts,
		"notification_retention",
		cfg.Worker.NotificationRetention.String(),
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
