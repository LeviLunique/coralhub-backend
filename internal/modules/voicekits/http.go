package voicekits

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/LeviLunique/coralhub-backend/internal/platform/requestctx"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, service *Service) {
	router.Route("/choirs/{choirID}/voice-kits", func(r chi.Router) {
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

			voiceKit, err := service.Create(r.Context(), tenant.ID, chi.URLParam(r, "choirID"), actor.ID, input)
			if err != nil {
				switch {
				case errors.Is(err, ErrInvalidChoirID):
					writeError(w, http.StatusBadRequest, "choir id is required")
				case errors.Is(err, ErrInvalidVoiceKitName):
					writeError(w, http.StatusBadRequest, "voice kit name is required")
				case errors.Is(err, ErrInvalidActorID):
					writeError(w, http.StatusUnauthorized, "actor identity is required")
				case errors.Is(err, ErrForbidden):
					writeError(w, http.StatusForbidden, "actor cannot manage this choir")
				case errors.Is(err, ErrVoiceKitNameTaken):
					writeError(w, http.StatusConflict, "voice kit name already exists")
				default:
					writeError(w, http.StatusInternalServerError, "internal server error")
				}
				return
			}

			writeJSON(w, http.StatusCreated, voiceKit)
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

			writeJSON(w, http.StatusOK, map[string][]VoiceKit{"items": items})
		})
	})

	router.Route("/voice-kits/{voiceKitID}", func(r chi.Router) {
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

			voiceKit, err := service.Get(r.Context(), tenant.ID, actor.ID, chi.URLParam(r, "voiceKitID"))
			if err != nil {
				switch {
				case errors.Is(err, ErrInvalidVoiceKitID):
					writeError(w, http.StatusBadRequest, "voice kit id is required")
				case errors.Is(err, ErrInvalidActorID):
					writeError(w, http.StatusUnauthorized, "actor identity is required")
				case errors.Is(err, ErrVoiceKitNotFound):
					writeError(w, http.StatusNotFound, "voice kit not found")
				default:
					writeError(w, http.StatusInternalServerError, "internal server error")
				}
				return
			}

			writeJSON(w, http.StatusOK, voiceKit)
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

			err := service.Delete(r.Context(), tenant.ID, chi.URLParam(r, "voiceKitID"), actor.ID)
			if err != nil {
				switch {
				case errors.Is(err, ErrInvalidVoiceKitID):
					writeError(w, http.StatusBadRequest, "voice kit id is required")
				case errors.Is(err, ErrInvalidActorID):
					writeError(w, http.StatusUnauthorized, "actor identity is required")
				case errors.Is(err, ErrForbidden):
					writeError(w, http.StatusForbidden, "actor cannot manage this choir")
				case errors.Is(err, ErrVoiceKitNotFound):
					writeError(w, http.StatusNotFound, "voice kit not found")
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
