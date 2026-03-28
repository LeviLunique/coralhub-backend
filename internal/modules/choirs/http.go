package choirs

import (
	"errors"
	"net/http"

	"github.com/LeviLunique/coralhub-backend/internal/platform/requestctx"
	platformweb "github.com/LeviLunique/coralhub-backend/internal/platform/web"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, service *Service) {
	router.Route("/choirs", func(r chi.Router) {
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			tenant, ok := requestctx.TenantFromContext(r.Context())
			if !ok {
				platformweb.WriteError(w, r, http.StatusInternalServerError, "tenant_context_missing", "tenant context missing")
				return
			}

			actor, ok := requestctx.ActorFromContext(r.Context())
			if !ok {
				platformweb.WriteError(w, r, http.StatusInternalServerError, "actor_context_missing", "actor context missing")
				return
			}

			var input CreateInput
			if err := platformweb.DecodeJSONBody(r, &input); err != nil {
				platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_request_body", "request body must be a single valid JSON object")
				return
			}

			choir, err := service.Create(r.Context(), tenant.ID, actor.ID, input)
			if err != nil {
				switch {
				case errors.Is(err, ErrInvalidChoirName):
					platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_choir_name", "choir name is required")
				case errors.Is(err, ErrChoirNameTaken):
					platformweb.WriteError(w, r, http.StatusConflict, "choir_name_taken", "choir name already exists")
				case errors.Is(err, ErrInvalidActorID):
					platformweb.WriteError(w, r, http.StatusUnauthorized, "invalid_actor", "actor identity is required")
				default:
					platformweb.WriteError(w, r, http.StatusInternalServerError, "internal_error", "internal server error")
				}
				return
			}

			platformweb.WriteJSON(w, http.StatusCreated, choir)
		})

		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			tenant, ok := requestctx.TenantFromContext(r.Context())
			if !ok {
				platformweb.WriteError(w, r, http.StatusInternalServerError, "tenant_context_missing", "tenant context missing")
				return
			}

			actor, ok := requestctx.ActorFromContext(r.Context())
			if !ok {
				platformweb.WriteError(w, r, http.StatusInternalServerError, "actor_context_missing", "actor context missing")
				return
			}

			items, err := service.List(r.Context(), tenant.ID, actor.ID)
			if err != nil {
				platformweb.WriteError(w, r, http.StatusInternalServerError, "internal_error", "internal server error")
				return
			}

			platformweb.WriteJSON(w, http.StatusOK, map[string][]Choir{"items": items})
		})

		r.Get("/{choirID}", func(w http.ResponseWriter, r *http.Request) {
			tenant, ok := requestctx.TenantFromContext(r.Context())
			if !ok {
				platformweb.WriteError(w, r, http.StatusInternalServerError, "tenant_context_missing", "tenant context missing")
				return
			}

			actor, ok := requestctx.ActorFromContext(r.Context())
			if !ok {
				platformweb.WriteError(w, r, http.StatusInternalServerError, "actor_context_missing", "actor context missing")
				return
			}

			choir, err := service.Get(r.Context(), tenant.ID, actor.ID, chi.URLParam(r, "choirID"))
			if err != nil {
				switch {
				case errors.Is(err, ErrInvalidChoirID):
					platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_choir_id", "choir id is required")
				case errors.Is(err, ErrInvalidActorID):
					platformweb.WriteError(w, r, http.StatusUnauthorized, "invalid_actor", "actor identity is required")
				case errors.Is(err, ErrChoirNotFound):
					platformweb.WriteError(w, r, http.StatusNotFound, "choir_not_found", "choir not found")
				default:
					platformweb.WriteError(w, r, http.StatusInternalServerError, "internal_error", "internal server error")
				}
				return
			}

			platformweb.WriteJSON(w, http.StatusOK, choir)
		})
	})
}
