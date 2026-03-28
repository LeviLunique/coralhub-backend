package main

import (
	"context"
	"errors"
	stdhttp "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	s3storage "github.com/LeviLunique/coralhub-backend/internal/integrations/storage/s3"
	"github.com/LeviLunique/coralhub-backend/internal/modules/choirs"
	"github.com/LeviLunique/coralhub-backend/internal/modules/events"
	modulefiles "github.com/LeviLunique/coralhub-backend/internal/modules/files"
	"github.com/LeviLunique/coralhub-backend/internal/modules/memberships"
	"github.com/LeviLunique/coralhub-backend/internal/modules/tenants"
	moduleusers "github.com/LeviLunique/coralhub-backend/internal/modules/users"
	"github.com/LeviLunique/coralhub-backend/internal/modules/voicekits"
	platformconfig "github.com/LeviLunique/coralhub-backend/internal/platform/config"
	platformhttp "github.com/LeviLunique/coralhub-backend/internal/platform/http"
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
	tenantRepository := postgres.NewTenantRepository(queries)
	tenantService := tenants.NewService(tenantRepository)
	choirRepository := postgres.NewChoirRepository(pool, queries)
	choirService := choirs.NewService(choirRepository)
	userRepository := postgres.NewUserRepository(queries)
	userService := moduleusers.NewService(userRepository)
	membershipRepository := postgres.NewMembershipRepository(pool, queries)
	membershipService := memberships.NewService(membershipRepository)
	voiceKitRepository := postgres.NewVoiceKitRepository(queries)
	voiceKitService := voicekits.NewService(voiceKitRepository, membershipRepository)
	fileRepository := postgres.NewFileRepository(queries)
	storageClient, err := s3storage.New(cfg.Storage)
	if err != nil {
		logger.Error("failed to initialize storage client", "error", err)
		os.Exit(1)
	}
	fileService := modulefiles.NewService(fileRepository, storageClient, voiceKitRepository, membershipRepository, cfg.AppEnv)
	eventRepository := postgres.NewEventRepository(pool, queries)
	eventService := events.NewService(eventRepository, membershipRepository)

	server := &stdhttp.Server{
		Addr:              cfg.HTTP.Addr,
		Handler:           platformhttp.NewRouter(logger, cfg.HTTP.HandlerTimeout, tenantService, choirService, userService, membershipService, voiceKitService, fileService, eventService),
		ReadTimeout:       cfg.HTTP.ReadTimeout,
		WriteTimeout:      cfg.HTTP.WriteTimeout,
		IdleTimeout:       cfg.HTTP.IdleTimeout,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if shutdownErr := server.Shutdown(shutdownCtx); shutdownErr != nil {
			logger.Error("api shutdown failed", "error", shutdownErr)
		}
	}()

	logger.Info("api starting", "addr", cfg.HTTP.Addr, "env", cfg.AppEnv)

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, stdhttp.ErrServerClosed) {
		logger.Error("api stopped unexpectedly", "error", err)
		os.Exit(1)
	}

	logger.Info("api stopped")
}
