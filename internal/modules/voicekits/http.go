package voicekits

import (
	"errors"
	"net/http"

	"github.com/LeviLunique/coralhub-backend/internal/platform/requestctx"
	platformweb "github.com/LeviLunique/coralhub-backend/internal/platform/web"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, service *Service) {
	router.Route("/choirs/{choirID}/voice-kits", func(r chi.Router) {
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

			voiceKit, err := service.Create(r.Context(), tenant.ID, chi.URLParam(r, "choirID"), actor.ID, input)
			if err != nil {
				switch {
				case errors.Is(err, ErrInvalidChoirID):
					platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_choir_id", "choir id is required")
				case errors.Is(err, ErrInvalidVoiceKitName):
					platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_voice_kit_name", "voice kit name is required")
				case errors.Is(err, ErrInvalidActorID):
					platformweb.WriteError(w, r, http.StatusUnauthorized, "invalid_actor", "actor identity is required")
				case errors.Is(err, ErrForbidden):
					platformweb.WriteError(w, r, http.StatusForbidden, "forbidden", "actor cannot manage this choir")
				case errors.Is(err, ErrVoiceKitNameTaken):
					platformweb.WriteError(w, r, http.StatusConflict, "voice_kit_name_taken", "voice kit name already exists")
				default:
					platformweb.WriteError(w, r, http.StatusInternalServerError, "internal_error", "internal server error")
				}
				return
			}

			platformweb.WriteJSON(w, http.StatusCreated, voiceKit)
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

			platformweb.WriteJSON(w, http.StatusOK, map[string][]VoiceKit{"items": items})
		})
	})

	router.Route("/voice-kits/{voiceKitID}", func(r chi.Router) {
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

			voiceKit, err := service.Get(r.Context(), tenant.ID, actor.ID, chi.URLParam(r, "voiceKitID"))
			if err != nil {
				switch {
				case errors.Is(err, ErrInvalidVoiceKitID):
					platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_voice_kit_id", "voice kit id is required")
				case errors.Is(err, ErrInvalidActorID):
					platformweb.WriteError(w, r, http.StatusUnauthorized, "invalid_actor", "actor identity is required")
				case errors.Is(err, ErrVoiceKitNotFound):
					platformweb.WriteError(w, r, http.StatusNotFound, "voice_kit_not_found", "voice kit not found")
				default:
					platformweb.WriteError(w, r, http.StatusInternalServerError, "internal_error", "internal server error")
				}
				return
			}

			platformweb.WriteJSON(w, http.StatusOK, voiceKit)
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

			err := service.Delete(r.Context(), tenant.ID, chi.URLParam(r, "voiceKitID"), actor.ID)
			if err != nil {
				switch {
				case errors.Is(err, ErrInvalidVoiceKitID):
					platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_voice_kit_id", "voice kit id is required")
				case errors.Is(err, ErrInvalidActorID):
					platformweb.WriteError(w, r, http.StatusUnauthorized, "invalid_actor", "actor identity is required")
				case errors.Is(err, ErrForbidden):
					platformweb.WriteError(w, r, http.StatusForbidden, "forbidden", "actor cannot manage this choir")
				case errors.Is(err, ErrVoiceKitNotFound):
					platformweb.WriteError(w, r, http.StatusNotFound, "voice_kit_not_found", "voice kit not found")
				default:
					platformweb.WriteError(w, r, http.StatusInternalServerError, "internal_error", "internal server error")
				}
				return
			}

			w.WriteHeader(http.StatusNoContent)
		})
	})
}
