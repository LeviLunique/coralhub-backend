package users

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/LeviLunique/coralhub-backend/internal/platform/requestctx"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, service *Service) {
	router.Route("/users", func(r chi.Router) {
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

			user, err := service.Create(r.Context(), tenant.ID, input)
			if err != nil {
				switch {
				case errors.Is(err, ErrInvalidEmail):
					writeError(w, http.StatusBadRequest, "valid email is required")
				case errors.Is(err, ErrInvalidFullName):
					writeError(w, http.StatusBadRequest, "full name is required")
				case errors.Is(err, ErrUserEmailTaken):
					writeError(w, http.StatusConflict, "user email already exists")
				default:
					writeError(w, http.StatusInternalServerError, "internal server error")
				}
				return
			}

			writeJSON(w, http.StatusCreated, user)
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

			writeJSON(w, http.StatusOK, map[string][]User{"items": items})
		})

		r.Get("/{userID}", func(w http.ResponseWriter, r *http.Request) {
			tenant, ok := requestctx.TenantFromContext(r.Context())
			if !ok {
				writeError(w, http.StatusInternalServerError, "tenant context missing")
				return
			}

			user, err := service.Get(r.Context(), tenant.ID, chi.URLParam(r, "userID"))
			if err != nil {
				switch {
				case errors.Is(err, ErrInvalidUserID):
					writeError(w, http.StatusBadRequest, "user id is required")
				case errors.Is(err, ErrUserNotFound):
					writeError(w, http.StatusNotFound, "user not found")
				default:
					writeError(w, http.StatusInternalServerError, "internal server error")
				}
				return
			}

			writeJSON(w, http.StatusOK, user)
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
