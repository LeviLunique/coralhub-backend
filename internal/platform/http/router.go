package platformhttp

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/LeviLunique/coralhub-backend/internal/modules/auth"
	"github.com/LeviLunique/coralhub-backend/internal/modules/choirs"
	"github.com/LeviLunique/coralhub-backend/internal/modules/events"
	"github.com/LeviLunique/coralhub-backend/internal/modules/instruments"
	"github.com/LeviLunique/coralhub-backend/internal/modules/materials"
	"github.com/LeviLunique/coralhub-backend/internal/modules/memberships"
	"github.com/LeviLunique/coralhub-backend/internal/modules/repertoires"
	"github.com/LeviLunique/coralhub-backend/internal/modules/songs"
	"github.com/LeviLunique/coralhub-backend/internal/modules/tenants"
	moduleusers "github.com/LeviLunique/coralhub-backend/internal/modules/users"
	platformobservability "github.com/LeviLunique/coralhub-backend/internal/platform/observability"
	platformweb "github.com/LeviLunique/coralhub-backend/internal/platform/web"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func NewRouter(
	logger *slog.Logger,
	handlerTimeout time.Duration,
	authService *auth.Service,
	tenantService *tenants.Service,
	choirService *choirs.Service,
	userService *moduleusers.Service,
	membershipService *memberships.Service,
	eventService *events.Service,
	songService *songs.Service,
	repertoireService *repertoires.Service,
	instrumentService *instruments.Service,
	materialService *materials.Service,
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

		if authService != nil {
			auth.RegisterRoutes(r, authService)
		}

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

		if tenantService != nil && userService != nil && (choirService != nil || membershipService != nil || eventService != nil || songService != nil || repertoireService != nil || instrumentService != nil || materialService != nil) {
			r.Group(func(protected chi.Router) {
				protected.Use(RequireActorContext(tenantService, userService))

				if choirService != nil {
					choirs.RegisterRoutes(protected, choirService)
				}

				if membershipService != nil {
					memberships.RegisterRoutes(protected, membershipService)
				}

				if eventService != nil {
					events.RegisterRoutes(protected, eventService)
				}

				if songService != nil {
					songs.RegisterRoutes(protected, songService)
				}

				if repertoireService != nil {
					repertoires.RegisterRoutes(protected, repertoireService)
				}

				if instrumentService != nil {
					instruments.RegisterRoutes(protected, instrumentService)
				}

				if materialService != nil {
					materials.RegisterRoutes(protected, materialService)
				}
			})
		}
	})

	return router
}
