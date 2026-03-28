package memberships

import (
	"errors"
	"net/http"

	"github.com/LeviLunique/coralhub-backend/internal/platform/requestctx"
	platformweb "github.com/LeviLunique/coralhub-backend/internal/platform/web"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, service *Service) {
	router.Route("/choirs/{choirID}/memberships", func(r chi.Router) {
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

			membership, err := service.AddMember(r.Context(), tenant.ID, chi.URLParam(r, "choirID"), actor.ID, input)
			if err != nil {
				switch {
				case errors.Is(err, ErrInvalidChoirID):
					platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_choir_id", "choir id is required")
				case errors.Is(err, ErrInvalidUserID):
					platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_user_id", "user id is required")
				case errors.Is(err, ErrInvalidRole):
					platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_membership_role", "role must be manager or member")
				case errors.Is(err, ErrForbidden):
					platformweb.WriteError(w, r, http.StatusForbidden, "forbidden", "actor cannot manage this choir")
				case errors.Is(err, ErrMembershipAlreadyExist):
					platformweb.WriteError(w, r, http.StatusConflict, "membership_exists", "membership already exists")
				case errors.Is(err, ErrMembershipNotFound):
					platformweb.WriteError(w, r, http.StatusForbidden, "membership_not_found", "actor is not a member of this choir")
				default:
					platformweb.WriteError(w, r, http.StatusInternalServerError, "internal_error", "internal server error")
				}
				return
			}

			platformweb.WriteJSON(w, http.StatusCreated, membership)
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
				case errors.Is(err, ErrMembershipNotFound):
					platformweb.WriteError(w, r, http.StatusForbidden, "membership_not_found", "actor is not a member of this choir")
				default:
					platformweb.WriteError(w, r, http.StatusInternalServerError, "internal_error", "internal server error")
				}
				return
			}

			platformweb.WriteJSON(w, http.StatusOK, map[string][]Membership{"items": items})
		})
	})
}
