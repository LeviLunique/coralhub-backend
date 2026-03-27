package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	platformconfig "github.com/LeviLunique/coralhub-backend/internal/platform/config"
	platformlog "github.com/LeviLunique/coralhub-backend/internal/platform/log"
	"github.com/LeviLunique/coralhub-backend/internal/store/postgres"
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

	logger.Info("worker starting", "poll_interval", cfg.Worker.PollInterval.String(), "env", cfg.AppEnv)

	ticker := time.NewTicker(cfg.Worker.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("worker stopped")
			return
		case <-ticker.C:
			logger.Debug("worker heartbeat")
		}
	}
}
