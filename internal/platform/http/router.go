package platformhttp

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/LeviLunique/coralhub-backend/internal/modules/tenants"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func NewRouter(logger *slog.Logger, tenantService *tenants.Service) http.Handler {
	router := chi.NewRouter()

	router.Use(chimiddleware.RequestID)
	router.Use(chimiddleware.RealIP)
	router.Use(chimiddleware.Recoverer)
	router.Use(chimiddleware.Timeout(30 * time.Second))
	router.Use(RequestLogger(logger))

	router.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, http.StatusOK, map[string]string{
			"service": "coralhub-api",
			"status":  "ok",
		})
	})

	router.Route("/api/v1", func(r chi.Router) {
		r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
			WriteJSON(w, http.StatusOK, map[string]string{
				"service": "coralhub-api",
				"status":  "ok",
			})
		})

		r.Route("/public", func(public chi.Router) {
			if tenantService != nil {
				tenants.RegisterPublicRoutes(public, tenantService)
			}
		})
	})

	return router
}
