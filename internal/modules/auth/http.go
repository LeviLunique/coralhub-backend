package auth

import (
	"errors"
	"net/http"

	platformweb "github.com/LeviLunique/coralhub-backend/internal/platform/web"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, service *Service) {
	router.Route("/auth", func(r chi.Router) {
		r.Post("/login", func(w http.ResponseWriter, r *http.Request) {
			var input LoginInput
			if err := platformweb.DecodeJSONBody(r, &input); err != nil {
				platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_request_body", "request body must be a single valid JSON object")
				return
			}

			session, err := service.Login(r.Context(), input)
			if err != nil {
				switch {
				case errors.Is(err, ErrInvalidEmail):
					platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_email", "valid email is required")
				case errors.Is(err, ErrInvalidPassword):
					platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_password", "password is required")
				case errors.Is(err, ErrAmbiguousUserTenant):
					platformweb.WriteError(w, r, http.StatusConflict, "ambiguous_user_tenant", "user belongs to more than one tenant")
				case errors.Is(err, ErrInvalidCredentials):
					platformweb.WriteError(w, r, http.StatusUnauthorized, "invalid_credentials", "email or password is invalid")
				default:
					platformweb.WriteError(w, r, http.StatusInternalServerError, "internal_error", "internal server error")
				}
				return
			}

			platformweb.WriteJSON(w, http.StatusOK, session)
		})
	})
}
