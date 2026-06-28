package events

import (
	"errors"
	"net/http"

	"github.com/LeviLunique/coralhub-backend/internal/platform/requestctx"
	platformweb "github.com/LeviLunique/coralhub-backend/internal/platform/web"
	"github.com/go-chi/chi/v5"
)

func registerRepertoireRoutes(router chi.Router, service *Service) {
	router.Route("/events/{eventID}/repertoires", func(r chi.Router) {
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

			items, err := service.ListRepertoires(r.Context(), tenant.ID, chi.URLParam(r, "eventID"), actor.ID)
			if err != nil {
				writeRepertoireLinkError(w, r, err)
				return
			}

			platformweb.WriteJSON(w, http.StatusOK, map[string][]Repertoire{"items": items})
		})

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

			var input LinkRepertoireInput
			if err := platformweb.DecodeJSONBody(r, &input); err != nil {
				platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_request_body", "request body must be a single valid JSON object")
				return
			}

			if err := service.AddRepertoire(r.Context(), tenant.ID, chi.URLParam(r, "eventID"), actor.ID, input); err != nil {
				writeRepertoireLinkError(w, r, err)
				return
			}

			w.WriteHeader(http.StatusNoContent)
		})

		r.Delete("/{repertoireID}", func(w http.ResponseWriter, r *http.Request) {
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

			if err := service.RemoveRepertoire(r.Context(), tenant.ID, chi.URLParam(r, "eventID"), actor.ID, chi.URLParam(r, "repertoireID")); err != nil {
				writeRepertoireLinkError(w, r, err)
				return
			}

			w.WriteHeader(http.StatusNoContent)
		})
	})
}

func writeRepertoireLinkError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, ErrInvalidEventID):
		platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_event_id", "event id is required")
	case errors.Is(err, ErrInvalidRepertoireID):
		platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_repertoire_id", "repertoire id is required")
	case errors.Is(err, ErrInvalidActorID):
		platformweb.WriteError(w, r, http.StatusUnauthorized, "invalid_actor", "actor identity is required")
	case errors.Is(err, ErrForbidden):
		platformweb.WriteError(w, r, http.StatusForbidden, "forbidden", "actor cannot manage this choir")
	case errors.Is(err, ErrRepertoireLinkExists):
		platformweb.WriteError(w, r, http.StatusConflict, "repertoire_link_exists", "repertoire already linked to event")
	case errors.Is(err, ErrRepertoireLinkNotFound):
		platformweb.WriteError(w, r, http.StatusNotFound, "repertoire_link_not_found", "repertoire is not linked to event")
	case errors.Is(err, ErrEventNotFound):
		platformweb.WriteError(w, r, http.StatusNotFound, "event_not_found", "event not found")
	default:
		platformweb.WriteError(w, r, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}
