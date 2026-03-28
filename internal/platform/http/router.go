package platformhttp

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/LeviLunique/coralhub-backend/internal/modules/choirs"
	"github.com/LeviLunique/coralhub-backend/internal/modules/events"
	modulefiles "github.com/LeviLunique/coralhub-backend/internal/modules/files"
	"github.com/LeviLunique/coralhub-backend/internal/modules/memberships"
	"github.com/LeviLunique/coralhub-backend/internal/modules/tenants"
	moduleusers "github.com/LeviLunique/coralhub-backend/internal/modules/users"
	"github.com/LeviLunique/coralhub-backend/internal/modules/voicekits"
	platformobservability "github.com/LeviLunique/coralhub-backend/internal/platform/observability"
	platformweb "github.com/LeviLunique/coralhub-backend/internal/platform/web"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func NewRouter(
	logger *slog.Logger,
	handlerTimeout time.Duration,
	tenantService *tenants.Service,
	choirService *choirs.Service,
	userService *moduleusers.Service,
	membershipService *memberships.Service,
	voiceKitService *voicekits.Service,
	fileService *modulefiles.Service,
	eventService *events.Service,
) http.Handler {
	router := chi.NewRouter()

	router.Use(chimiddleware.RequestID)
	router.Use(chimiddleware.RealIP)
	router.Use(chimiddleware.Recoverer)
	router.Use(Timeout(handlerTimeout))
	router.Use(RequestLogger(logger))

	router.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		platformweb.WriteJSON(w, http.StatusOK, map[string]string{
			"service": "coralhub-api",
			"status":  "ok",
		})
	})
	router.Handle("/metrics", platformobservability.DefaultMetrics().Handler())

	router.Route("/api/v1", func(r chi.Router) {
		r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
			platformweb.WriteJSON(w, http.StatusOK, map[string]string{
				"service": "coralhub-api",
				"status":  "ok",
			})
		})

		r.Route("/public", func(public chi.Router) {
			if tenantService != nil {
				tenants.RegisterPublicRoutes(public, tenantService)
			}
		})

		if tenantService != nil && userService != nil {
			r.Group(func(protected chi.Router) {
				protected.Use(RequireTenantContext(tenantService))
				moduleusers.RegisterRoutes(protected, userService)
			})
		}

		if tenantService != nil && userService != nil && (choirService != nil || membershipService != nil || voiceKitService != nil || fileService != nil || eventService != nil) {
			r.Group(func(protected chi.Router) {
				protected.Use(RequireActorContext(tenantService, userService))

				if choirService != nil {
					choirs.RegisterRoutes(protected, choirService)
				}

				if membershipService != nil {
					memberships.RegisterRoutes(protected, membershipService)
				}

				if voiceKitService != nil {
					voicekits.RegisterRoutes(protected, voiceKitService)
				}

				if fileService != nil {
					modulefiles.RegisterRoutes(protected, fileService)
				}

				if eventService != nil {
					events.RegisterRoutes(protected, eventService)
				}
			})
		}
	})

	return router
}
