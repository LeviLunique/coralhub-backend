package files

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/LeviLunique/coralhub-backend/internal/platform/requestctx"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, service *Service) {
	router.Route("/voice-kits/{voiceKitID}/files", func(r chi.Router) {
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

			file, err := service.Create(r.Context(), tenant.ID, chi.URLParam(r, "voiceKitID"), actor.ID, input)
			if err != nil {
				switch {
				case errors.Is(err, ErrInvalidVoiceKitID):
					writeError(w, http.StatusBadRequest, "voice kit id is required")
				case errors.Is(err, ErrInvalidOriginalFilename):
					writeError(w, http.StatusBadRequest, "original filename is required")
				case errors.Is(err, ErrInvalidStoredFilename):
					writeError(w, http.StatusBadRequest, "stored filename is required")
				case errors.Is(err, ErrInvalidContentType):
					writeError(w, http.StatusBadRequest, "content type is required")
				case errors.Is(err, ErrInvalidStorageKey):
					writeError(w, http.StatusBadRequest, "storage key is required")
				case errors.Is(err, ErrInvalidSizeBytes):
					writeError(w, http.StatusBadRequest, "size bytes must be greater than zero")
				case errors.Is(err, ErrForbidden):
					writeError(w, http.StatusForbidden, "actor cannot manage this voice kit")
				case errors.Is(err, ErrVoiceKitNotFound):
					writeError(w, http.StatusNotFound, "voice kit not found")
				default:
					writeError(w, http.StatusInternalServerError, "internal server error")
				}
				return
			}

			writeJSON(w, http.StatusCreated, file)
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

			items, err := service.ListByVoiceKit(r.Context(), tenant.ID, chi.URLParam(r, "voiceKitID"), actor.ID)
			if err != nil {
				switch {
				case errors.Is(err, ErrInvalidVoiceKitID):
					writeError(w, http.StatusBadRequest, "voice kit id is required")
				case errors.Is(err, ErrVoiceKitNotFound):
					writeError(w, http.StatusNotFound, "voice kit not found")
				default:
					writeError(w, http.StatusInternalServerError, "internal server error")
				}
				return
			}

			writeJSON(w, http.StatusOK, map[string][]File{"items": items})
		})
	})

	router.Route("/files/{fileID}", func(r chi.Router) {
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

			err := service.Delete(r.Context(), tenant.ID, chi.URLParam(r, "fileID"), actor.ID)
			if err != nil {
				switch {
				case errors.Is(err, ErrInvalidFileID):
					writeError(w, http.StatusBadRequest, "file id is required")
				case errors.Is(err, ErrForbidden):
					writeError(w, http.StatusForbidden, "actor cannot manage this voice kit")
				case errors.Is(err, ErrFileNotFound):
					writeError(w, http.StatusNotFound, "file not found")
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
