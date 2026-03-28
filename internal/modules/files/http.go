package files

import (
	"errors"
	"net"
	"net/http"

	"github.com/LeviLunique/coralhub-backend/internal/platform/requestctx"
	platformweb "github.com/LeviLunique/coralhub-backend/internal/platform/web"
	"github.com/go-chi/chi/v5"
)

const maxUploadRequestBytes = maxUploadSizeBytes + (1 << 20)

func RegisterRoutes(router chi.Router, service *Service) {
	router.Route("/voice-kits/{voiceKitID}/files", func(r chi.Router) {
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

			r.Body = http.MaxBytesReader(w, r.Body, maxUploadRequestBytes)
			if err := r.ParseMultipartForm(1 << 20); err != nil {
				var maxBytesErr *http.MaxBytesError
				if errors.As(err, &maxBytesErr) {
					platformweb.WriteError(w, r, http.StatusBadRequest, "request_body_too_large", "request body exceeds the maximum allowed size")
					return
				}
				var netErr net.Error
				if errors.As(err, &netErr) {
					platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_multipart_form", "invalid multipart form")
					return
				}
				platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_multipart_form", "invalid multipart form")
				return
			}

			uploadedFile, header, err := r.FormFile("file")
			if err != nil {
				platformweb.WriteError(w, r, http.StatusBadRequest, "file_field_required", "file form field is required")
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
					platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_voice_kit_id", "voice kit id is required")
				case errors.Is(err, ErrInvalidOriginalFilename):
					platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_uploaded_filename", "uploaded filename is required")
				case errors.Is(err, ErrInvalidContentType):
					platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_content_type", "valid content type is required")
				case errors.Is(err, ErrUnsupportedContentType):
					platformweb.WriteError(w, r, http.StatusBadRequest, "unsupported_content_type", "content type must be audio/* or application/pdf")
				case errors.Is(err, ErrInvalidSizeBytes):
					platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_file_size", "file size must be greater than zero")
				case errors.Is(err, ErrFileTooLarge):
					platformweb.WriteError(w, r, http.StatusBadRequest, "file_too_large", "file exceeds the maximum allowed size")
				case errors.Is(err, ErrForbidden):
					platformweb.WriteError(w, r, http.StatusForbidden, "forbidden", "actor cannot manage this voice kit")
				case errors.Is(err, ErrVoiceKitNotFound):
					platformweb.WriteError(w, r, http.StatusNotFound, "voice_kit_not_found", "voice kit not found")
				case errors.Is(err, ErrStorageUnavailable):
					platformweb.WriteError(w, r, http.StatusServiceUnavailable, "storage_unavailable", "storage unavailable")
				default:
					platformweb.WriteError(w, r, http.StatusInternalServerError, "internal_error", "internal server error")
				}
				return
			}

			platformweb.WriteJSON(w, http.StatusCreated, file)
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

			items, err := service.ListByVoiceKit(r.Context(), tenant.ID, chi.URLParam(r, "voiceKitID"), actor.ID)
			if err != nil {
				switch {
				case errors.Is(err, ErrInvalidVoiceKitID):
					platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_voice_kit_id", "voice kit id is required")
				case errors.Is(err, ErrVoiceKitNotFound):
					platformweb.WriteError(w, r, http.StatusNotFound, "voice_kit_not_found", "voice kit not found")
				default:
					platformweb.WriteError(w, r, http.StatusInternalServerError, "internal_error", "internal server error")
				}
				return
			}

			platformweb.WriteJSON(w, http.StatusOK, map[string][]File{"items": items})
		})
	})

	router.Route("/files/{fileID}", func(r chi.Router) {
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

			err := service.Delete(r.Context(), tenant.ID, chi.URLParam(r, "fileID"), actor.ID)
			if err != nil {
				switch {
				case errors.Is(err, ErrInvalidFileID):
					platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_file_id", "file id is required")
				case errors.Is(err, ErrForbidden):
					platformweb.WriteError(w, r, http.StatusForbidden, "forbidden", "actor cannot manage this voice kit")
				case errors.Is(err, ErrFileNotFound):
					platformweb.WriteError(w, r, http.StatusNotFound, "file_not_found", "file not found")
				case errors.Is(err, ErrStorageUnavailable):
					platformweb.WriteError(w, r, http.StatusServiceUnavailable, "storage_unavailable", "storage unavailable")
				default:
					platformweb.WriteError(w, r, http.StatusInternalServerError, "internal_error", "internal server error")
				}
				return
			}

			w.WriteHeader(http.StatusNoContent)
		})

		r.Get("/download-url", func(w http.ResponseWriter, r *http.Request) {
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

			result, err := service.GetDownloadURL(r.Context(), tenant.ID, chi.URLParam(r, "fileID"), actor.ID)
			if err != nil {
				switch {
				case errors.Is(err, ErrInvalidFileID):
					platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_file_id", "file id is required")
				case errors.Is(err, ErrFileNotFound):
					platformweb.WriteError(w, r, http.StatusNotFound, "file_not_found", "file not found")
				case errors.Is(err, ErrStorageUnavailable):
					platformweb.WriteError(w, r, http.StatusServiceUnavailable, "storage_unavailable", "storage unavailable")
				default:
					platformweb.WriteError(w, r, http.StatusInternalServerError, "internal_error", "internal server error")
				}
				return
			}

			platformweb.WriteJSON(w, http.StatusOK, result)
		})
	})
}
