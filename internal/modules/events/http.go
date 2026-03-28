package events

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/LeviLunique/coralhub-backend/internal/platform/requestctx"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, service *Service) {
	router.Route("/choirs/{choirID}/events", func(r chi.Router) {
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			tenant, ok := requestctx.TenantFromContext(r.Context())
			if !ok {
				writeError(w, http.StatusInternalServerError, "tenant context missing")
				return
			}
			actor, ok := requestctx.ActorFromContext(r.Context())
			if !ok {
				writeError(w, http.StatusInternalServerError, "actor context missing")
				return
			}

			var input CreateInput
			if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
				writeError(w, http.StatusBadRequest, "invalid request body")
				return
			}

			event, err := service.Create(r.Context(), tenant.ID, chi.URLParam(r, "choirID"), actor.ID, input)
			if err != nil {
				switch {
				case errors.Is(err, ErrInvalidChoirID):
					writeError(w, http.StatusBadRequest, "choir id is required")
				case errors.Is(err, ErrInvalidTitle):
					writeError(w, http.StatusBadRequest, "title is required")
				case errors.Is(err, ErrInvalidEventType):
					writeError(w, http.StatusBadRequest, "event type must be rehearsal, presentation, or other")
				case errors.Is(err, ErrInvalidStartAt):
					writeError(w, http.StatusBadRequest, "start_at is required")
				case errors.Is(err, ErrForbidden):
					writeError(w, http.StatusForbidden, "actor cannot manage this choir")
				default:
					writeError(w, http.StatusInternalServerError, "internal server error")
				}
				return
			}

			writeJSON(w, http.StatusCreated, event)
		})

		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			tenant, ok := requestctx.TenantFromContext(r.Context())
			if !ok {
				writeError(w, http.StatusInternalServerError, "tenant context missing")
				return
			}
			actor, ok := requestctx.ActorFromContext(r.Context())
			if !ok {
				writeError(w, http.StatusInternalServerError, "actor context missing")
				return
			}

			items, err := service.ListByChoir(r.Context(), tenant.ID, chi.URLParam(r, "choirID"), actor.ID)
			if err != nil {
				switch {
				case errors.Is(err, ErrInvalidChoirID):
					writeError(w, http.StatusBadRequest, "choir id is required")
				case errors.Is(err, ErrForbidden):
					writeError(w, http.StatusForbidden, "actor is not a member of this choir")
				default:
					writeError(w, http.StatusInternalServerError, "internal server error")
				}
				return
			}

			writeJSON(w, http.StatusOK, map[string][]Event{"items": items})
		})
	})

	router.Route("/events/{eventID}", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			tenant, ok := requestctx.TenantFromContext(r.Context())
			if !ok {
				writeError(w, http.StatusInternalServerError, "tenant context missing")
				return
			}
			actor, ok := requestctx.ActorFromContext(r.Context())
			if !ok {
				writeError(w, http.StatusInternalServerError, "actor context missing")
				return
			}

			event, err := service.Get(r.Context(), tenant.ID, actor.ID, chi.URLParam(r, "eventID"))
			if err != nil {
				switch {
				case errors.Is(err, ErrInvalidEventID):
					writeError(w, http.StatusBadRequest, "event id is required")
				case errors.Is(err, ErrEventNotFound):
					writeError(w, http.StatusNotFound, "event not found")
				default:
					writeError(w, http.StatusInternalServerError, "internal server error")
				}
				return
			}

			writeJSON(w, http.StatusOK, event)
		})

		r.Put("/", func(w http.ResponseWriter, r *http.Request) {
			tenant, ok := requestctx.TenantFromContext(r.Context())
			if !ok {
				writeError(w, http.StatusInternalServerError, "tenant context missing")
				return
			}
			actor, ok := requestctx.ActorFromContext(r.Context())
			if !ok {
				writeError(w, http.StatusInternalServerError, "actor context missing")
				return
			}

			var input UpdateInput
			if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
				writeError(w, http.StatusBadRequest, "invalid request body")
				return
			}

			event, err := service.Update(r.Context(), tenant.ID, chi.URLParam(r, "eventID"), actor.ID, input)
			if err != nil {
				switch {
				case errors.Is(err, ErrInvalidEventID):
					writeError(w, http.StatusBadRequest, "event id is required")
				case errors.Is(err, ErrInvalidTitle):
					writeError(w, http.StatusBadRequest, "title is required")
				case errors.Is(err, ErrInvalidEventType):
					writeError(w, http.StatusBadRequest, "event type must be rehearsal, presentation, or other")
				case errors.Is(err, ErrInvalidStartAt):
					writeError(w, http.StatusBadRequest, "start_at is required")
				case errors.Is(err, ErrForbidden):
					writeError(w, http.StatusForbidden, "actor cannot manage this choir")
				case errors.Is(err, ErrEventNotFound):
					writeError(w, http.StatusNotFound, "event not found")
				default:
					writeError(w, http.StatusInternalServerError, "internal server error")
				}
				return
			}

			writeJSON(w, http.StatusOK, event)
		})

		r.Delete("/", func(w http.ResponseWriter, r *http.Request) {
			tenant, ok := requestctx.TenantFromContext(r.Context())
			if !ok {
				writeError(w, http.StatusInternalServerError, "tenant context missing")
				return
			}
			actor, ok := requestctx.ActorFromContext(r.Context())
			if !ok {
				writeError(w, http.StatusInternalServerError, "actor context missing")
				return
			}

			err := service.Cancel(r.Context(), tenant.ID, chi.URLParam(r, "eventID"), actor.ID)
			if err != nil {
				switch {
				case errors.Is(err, ErrInvalidEventID):
					writeError(w, http.StatusBadRequest, "event id is required")
				case errors.Is(err, ErrForbidden):
					writeError(w, http.StatusForbidden, "actor cannot manage this choir")
				case errors.Is(err, ErrEventNotFound):
					writeError(w, http.StatusNotFound, "event not found")
				default:
					writeError(w, http.StatusInternalServerError, "internal server error")
				}
				return
			}

			w.WriteHeader(http.StatusNoContent)
		})
	})
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, map[string]string{
		"error": message,
	})
}
