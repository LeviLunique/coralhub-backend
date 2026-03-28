package events

import (
	"errors"
	"net/http"

	"github.com/LeviLunique/coralhub-backend/internal/platform/requestctx"
	platformweb "github.com/LeviLunique/coralhub-backend/internal/platform/web"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, service *Service) {
	router.Route("/choirs/{choirID}/events", func(r chi.Router) {
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

			event, err := service.Create(r.Context(), tenant.ID, chi.URLParam(r, "choirID"), actor.ID, input)
			if err != nil {
				switch {
				case errors.Is(err, ErrInvalidChoirID):
					platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_choir_id", "choir id is required")
				case errors.Is(err, ErrInvalidTitle):
					platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_event_title", "title is required")
				case errors.Is(err, ErrInvalidEventType):
					platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_event_type", "event type must be rehearsal, presentation, or other")
				case errors.Is(err, ErrInvalidStartAt):
					platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_event_start_at", "start_at is required")
				case errors.Is(err, ErrForbidden):
					platformweb.WriteError(w, r, http.StatusForbidden, "forbidden", "actor cannot manage this choir")
				default:
					platformweb.WriteError(w, r, http.StatusInternalServerError, "internal_error", "internal server error")
				}
				return
			}

			platformweb.WriteJSON(w, http.StatusCreated, event)
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

			items, err := service.ListByChoir(r.Context(), tenant.ID, chi.URLParam(r, "choirID"), actor.ID)
			if err != nil {
				switch {
				case errors.Is(err, ErrInvalidChoirID):
					platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_choir_id", "choir id is required")
				case errors.Is(err, ErrForbidden):
					platformweb.WriteError(w, r, http.StatusForbidden, "forbidden", "actor is not a member of this choir")
				default:
					platformweb.WriteError(w, r, http.StatusInternalServerError, "internal_error", "internal server error")
				}
				return
			}

			platformweb.WriteJSON(w, http.StatusOK, map[string][]Event{"items": items})
		})
	})

	router.Route("/events/{eventID}", func(r chi.Router) {
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

			event, err := service.Get(r.Context(), tenant.ID, actor.ID, chi.URLParam(r, "eventID"))
			if err != nil {
				switch {
				case errors.Is(err, ErrInvalidEventID):
					platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_event_id", "event id is required")
				case errors.Is(err, ErrEventNotFound):
					platformweb.WriteError(w, r, http.StatusNotFound, "event_not_found", "event not found")
				default:
					platformweb.WriteError(w, r, http.StatusInternalServerError, "internal_error", "internal server error")
				}
				return
			}

			platformweb.WriteJSON(w, http.StatusOK, event)
		})

		r.Put("/", func(w http.ResponseWriter, r *http.Request) {
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

			var input UpdateInput
			if err := platformweb.DecodeJSONBody(r, &input); err != nil {
				platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_request_body", "request body must be a single valid JSON object")
				return
			}

			event, err := service.Update(r.Context(), tenant.ID, chi.URLParam(r, "eventID"), actor.ID, input)
			if err != nil {
				switch {
				case errors.Is(err, ErrInvalidEventID):
					platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_event_id", "event id is required")
				case errors.Is(err, ErrInvalidTitle):
					platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_event_title", "title is required")
				case errors.Is(err, ErrInvalidEventType):
					platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_event_type", "event type must be rehearsal, presentation, or other")
				case errors.Is(err, ErrInvalidStartAt):
					platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_event_start_at", "start_at is required")
				case errors.Is(err, ErrForbidden):
					platformweb.WriteError(w, r, http.StatusForbidden, "forbidden", "actor cannot manage this choir")
				case errors.Is(err, ErrEventNotFound):
					platformweb.WriteError(w, r, http.StatusNotFound, "event_not_found", "event not found")
				default:
					platformweb.WriteError(w, r, http.StatusInternalServerError, "internal_error", "internal server error")
				}
				return
			}

			platformweb.WriteJSON(w, http.StatusOK, event)
		})

		r.Delete("/", func(w http.ResponseWriter, r *http.Request) {
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

			err := service.Cancel(r.Context(), tenant.ID, chi.URLParam(r, "eventID"), actor.ID)
			if err != nil {
				switch {
				case errors.Is(err, ErrInvalidEventID):
					platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_event_id", "event id is required")
				case errors.Is(err, ErrForbidden):
					platformweb.WriteError(w, r, http.StatusForbidden, "forbidden", "actor cannot manage this choir")
				case errors.Is(err, ErrEventNotFound):
					platformweb.WriteError(w, r, http.StatusNotFound, "event_not_found", "event not found")
				default:
					platformweb.WriteError(w, r, http.StatusInternalServerError, "internal_error", "internal server error")
				}
				return
			}

			w.WriteHeader(http.StatusNoContent)
		})
	})
}
