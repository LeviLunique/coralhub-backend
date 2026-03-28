package files

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/LeviLunique/coralhub-backend/internal/platform/requestctx"
	"github.com/go-chi/chi/v5"
)

const maxUploadRequestBytes = maxUploadSizeBytes + (1 << 20)

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

			r.Body = http.MaxBytesReader(w, r.Body, maxUploadRequestBytes)
			if err := r.ParseMultipartForm(1 << 20); err != nil {
				writeError(w, http.StatusBadRequest, "invalid multipart form")
				return
			}

			uploadedFile, header, err := r.FormFile("file")
			if err != nil {
				writeError(w, http.StatusBadRequest, "file form field is required")
				return
			}
			defer uploadedFile.Close()

			file, err := service.Upload(r.Context(), tenant.ID, tenant.Slug, chi.URLParam(r, "voiceKitID"), actor.ID, UploadInput{
				OriginalFilename: header.Filename,
				ContentType:      header.Header.Get("Content-Type"),
				SizeBytes:        header.Size,
				Content:          uploadedFile,
			})
			if err != nil {
				switch {
				case errors.Is(err, ErrInvalidVoiceKitID):
					writeError(w, http.StatusBadRequest, "voice kit id is required")
				case errors.Is(err, ErrInvalidOriginalFilename):
					writeError(w, http.StatusBadRequest, "uploaded filename is required")
				case errors.Is(err, ErrInvalidContentType):
					writeError(w, http.StatusBadRequest, "valid content type is required")
				case errors.Is(err, ErrUnsupportedContentType):
					writeError(w, http.StatusBadRequest, "content type must be audio/* or application/pdf")
				case errors.Is(err, ErrInvalidSizeBytes):
					writeError(w, http.StatusBadRequest, "file size must be greater than zero")
				case errors.Is(err, ErrFileTooLarge):
					writeError(w, http.StatusBadRequest, "file exceeds the maximum allowed size")
				case errors.Is(err, ErrForbidden):
					writeError(w, http.StatusForbidden, "actor cannot manage this voice kit")
				case errors.Is(err, ErrVoiceKitNotFound):
					writeError(w, http.StatusNotFound, "voice kit not found")
				case errors.Is(err, ErrStorageUnavailable):
					writeError(w, http.StatusServiceUnavailable, "storage unavailable")
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
				case errors.Is(err, ErrStorageUnavailable):
					writeError(w, http.StatusServiceUnavailable, "storage unavailable")
				default:
					writeError(w, http.StatusInternalServerError, "internal server error")
				}
				return
			}

			w.WriteHeader(http.StatusNoContent)
		})

		r.Get("/download-url", func(w http.ResponseWriter, r *http.Request) {
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

			result, err := service.GetDownloadURL(r.Context(), tenant.ID, chi.URLParam(r, "fileID"), actor.ID)
			if err != nil {
				switch {
				case errors.Is(err, ErrInvalidFileID):
					writeError(w, http.StatusBadRequest, "file id is required")
				case errors.Is(err, ErrFileNotFound):
					writeError(w, http.StatusNotFound, "file not found")
				case errors.Is(err, ErrStorageUnavailable):
					writeError(w, http.StatusServiceUnavailable, "storage unavailable")
				default:
					writeError(w, http.StatusInternalServerError, "internal server error")
				}
				return
			}

			writeJSON(w, http.StatusOK, result)
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
