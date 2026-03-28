package users

import (
	"errors"
	"net/http"

	"github.com/LeviLunique/coralhub-backend/internal/platform/requestctx"
	platformweb "github.com/LeviLunique/coralhub-backend/internal/platform/web"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, service *Service) {
	router.Route("/users", func(r chi.Router) {
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			tenant, ok := requestctx.TenantFromContext(r.Context())
			if !ok {
				platformweb.WriteError(w, r, http.StatusInternalServerError, "tenant_context_missing", "tenant context missing")
				return
			}

			var input CreateInput
			if err := platformweb.DecodeJSONBody(r, &input); err != nil {
				platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_request_body", "request body must be a single valid JSON object")
				return
			}

			user, err := service.Create(r.Context(), tenant.ID, input)
			if err != nil {
				switch {
				case errors.Is(err, ErrInvalidEmail):
					platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_email", "valid email is required")
				case errors.Is(err, ErrInvalidFullName):
					platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_full_name", "full name is required")
				case errors.Is(err, ErrUserEmailTaken):
					platformweb.WriteError(w, r, http.StatusConflict, "user_email_taken", "user email already exists")
				default:
					platformweb.WriteError(w, r, http.StatusInternalServerError, "internal_error", "internal server error")
				}
				return
			}

			platformweb.WriteJSON(w, http.StatusCreated, user)
		})

		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			tenant, ok := requestctx.TenantFromContext(r.Context())
			if !ok {
				platformweb.WriteError(w, r, http.StatusInternalServerError, "tenant_context_missing", "tenant context missing")
				return
			}

			items, err := service.List(r.Context(), tenant.ID)
			if err != nil {
				platformweb.WriteError(w, r, http.StatusInternalServerError, "internal_error", "internal server error")
				return
			}

			platformweb.WriteJSON(w, http.StatusOK, map[string][]User{"items": items})
		})

		r.Get("/{userID}", func(w http.ResponseWriter, r *http.Request) {
			tenant, ok := requestctx.TenantFromContext(r.Context())
			if !ok {
				platformweb.WriteError(w, r, http.StatusInternalServerError, "tenant_context_missing", "tenant context missing")
				return
			}

			user, err := service.Get(r.Context(), tenant.ID, chi.URLParam(r, "userID"))
			if err != nil {
				switch {
				case errors.Is(err, ErrInvalidUserID):
					platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_user_id", "user id is required")
				case errors.Is(err, ErrUserNotFound):
					platformweb.WriteError(w, r, http.StatusNotFound, "user_not_found", "user not found")
				default:
					platformweb.WriteError(w, r, http.StatusInternalServerError, "internal_error", "internal server error")
				}
				return
			}

			platformweb.WriteJSON(w, http.StatusOK, user)
		})
	})
}
