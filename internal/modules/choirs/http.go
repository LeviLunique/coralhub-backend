package choirs

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/LeviLunique/coralhub-backend/internal/platform/requestctx"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, service *Service) {
	router.Route("/choirs", func(r chi.Router) {
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			tenant, ok := requestctx.TenantFromContext(r.Context())
			if !ok {
				writeError(w, http.StatusInternalServerError, "tenant context missing")
				return
			}

			var input CreateInput
			if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
				writeError(w, http.StatusBadRequest, "invalid request body")
				return
			}

			choir, err := service.Create(r.Context(), tenant.ID, input)
			if err != nil {
				switch {
				case errors.Is(err, ErrInvalidChoirName):
					writeError(w, http.StatusBadRequest, "choir name is required")
				case errors.Is(err, ErrChoirNameTaken):
					writeError(w, http.StatusConflict, "choir name already exists")
				default:
					writeError(w, http.StatusInternalServerError, "internal server error")
				}
				return
			}

			writeJSON(w, http.StatusCreated, choir)
		})

		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			tenant, ok := requestctx.TenantFromContext(r.Context())
			if !ok {
				writeError(w, http.StatusInternalServerError, "tenant context missing")
				return
			}

			items, err := service.List(r.Context(), tenant.ID)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "internal server error")
				return
			}

			writeJSON(w, http.StatusOK, map[string][]Choir{"items": items})
		})

		r.Get("/{choirID}", func(w http.ResponseWriter, r *http.Request) {
			tenant, ok := requestctx.TenantFromContext(r.Context())
			if !ok {
				writeError(w, http.StatusInternalServerError, "tenant context missing")
				return
			}

			choir, err := service.Get(r.Context(), tenant.ID, chi.URLParam(r, "choirID"))
			if err != nil {
				switch {
				case errors.Is(err, ErrInvalidChoirID):
					writeError(w, http.StatusBadRequest, "choir id is required")
				case errors.Is(err, ErrChoirNotFound):
					writeError(w, http.StatusNotFound, "choir not found")
				default:
					writeError(w, http.StatusInternalServerError, "internal server error")
				}
				return
			}

			writeJSON(w, http.StatusOK, choir)
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
