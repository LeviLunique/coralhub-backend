package instruments

import (
	"errors"
	"net/http"

	"github.com/LeviLunique/coralhub-backend/internal/modules/tenants"
	"github.com/LeviLunique/coralhub-backend/internal/platform/requestctx"
	platformweb "github.com/LeviLunique/coralhub-backend/internal/platform/web"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, service *Service) {
	router.Route("/choirs/{choirID}/instruments", func(r chi.Router) {
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			tenant, actor, ok := contextActors(w, r)
			if !ok {
				return
			}

			var input CreateInput
			if err := platformweb.DecodeJSONBody(r, &input); err != nil {
				platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_request_body", "request body must be a single valid JSON object")
				return
			}

			instrument, err := service.Create(r.Context(), tenant.ID, chi.URLParam(r, "choirID"), actor.ID, input)
			if err != nil {
				writeInstrumentError(w, r, err)
				return
			}

			platformweb.WriteJSON(w, http.StatusCreated, instrument)
		})

		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			tenant, actor, ok := contextActors(w, r)
			if !ok {
				return
			}

			items, err := service.ListByChoir(r.Context(), tenant.ID, chi.URLParam(r, "choirID"), actor.ID)
			if err != nil {
				writeInstrumentError(w, r, err)
				return
			}

			platformweb.WriteJSON(w, http.StatusOK, map[string][]Instrument{"items": items})
		})
	})

	router.Route("/instruments/{instrumentID}", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			tenant, actor, ok := contextActors(w, r)
			if !ok {
				return
			}

			instrument, err := service.Get(r.Context(), tenant.ID, actor.ID, chi.URLParam(r, "instrumentID"))
			if err != nil {
				writeInstrumentError(w, r, err)
				return
			}

			platformweb.WriteJSON(w, http.StatusOK, instrument)
		})

		r.Put("/", func(w http.ResponseWriter, r *http.Request) {
			tenant, actor, ok := contextActors(w, r)
			if !ok {
				return
			}

			var input UpdateInput
			if err := platformweb.DecodeJSONBody(r, &input); err != nil {
				platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_request_body", "request body must be a single valid JSON object")
				return
			}

			instrument, err := service.Update(r.Context(), tenant.ID, chi.URLParam(r, "instrumentID"), actor.ID, input)
			if err != nil {
				writeInstrumentError(w, r, err)
				return
			}

			platformweb.WriteJSON(w, http.StatusOK, instrument)
		})

		r.Delete("/", func(w http.ResponseWriter, r *http.Request) {
			tenant, actor, ok := contextActors(w, r)
			if !ok {
				return
			}

			if err := service.Delete(r.Context(), tenant.ID, chi.URLParam(r, "instrumentID"), actor.ID); err != nil {
				writeInstrumentError(w, r, err)
				return
			}

			w.WriteHeader(http.StatusNoContent)
		})

		r.Get("/users", func(w http.ResponseWriter, r *http.Request) {
			tenant, actor, ok := contextActors(w, r)
			if !ok {
				return
			}

			items, err := service.ListUsers(r.Context(), tenant.ID, chi.URLParam(r, "instrumentID"), actor.ID)
			if err != nil {
				writeInstrumentError(w, r, err)
				return
			}

			platformweb.WriteJSON(w, http.StatusOK, map[string][]InstrumentUser{"items": items})
		})

		r.Post("/users", func(w http.ResponseWriter, r *http.Request) {
			tenant, actor, ok := contextActors(w, r)
			if !ok {
				return
			}

			var input LinkUserInput
			if err := platformweb.DecodeJSONBody(r, &input); err != nil {
				platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_request_body", "request body must be a single valid JSON object")
				return
			}

			if err := service.AddUser(r.Context(), tenant.ID, chi.URLParam(r, "instrumentID"), actor.ID, input); err != nil {
				writeInstrumentError(w, r, err)
				return
			}

			w.WriteHeader(http.StatusNoContent)
		})

		r.Delete("/users/{userID}", func(w http.ResponseWriter, r *http.Request) {
			tenant, actor, ok := contextActors(w, r)
			if !ok {
				return
			}

			if err := service.RemoveUser(r.Context(), tenant.ID, chi.URLParam(r, "instrumentID"), actor.ID, chi.URLParam(r, "userID")); err != nil {
				writeInstrumentError(w, r, err)
				return
			}

			w.WriteHeader(http.StatusNoContent)
		})
	})
}

func contextActors(w http.ResponseWriter, r *http.Request) (tenants.Context, requestctx.Actor, bool) {
	tenant, ok := requestctx.TenantFromContext(r.Context())
	if !ok {
		platformweb.WriteError(w, r, http.StatusInternalServerError, "tenant_context_missing", "tenant context missing")
		return tenants.Context{}, requestctx.Actor{}, false
	}

	actor, ok := requestctx.ActorFromContext(r.Context())
	if !ok {
		platformweb.WriteError(w, r, http.StatusInternalServerError, "actor_context_missing", "actor context missing")
		return tenants.Context{}, requestctx.Actor{}, false
	}

	return tenant, actor, true
}

func writeInstrumentError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, ErrInvalidChoirID):
		platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_choir_id", "choir id is required")
	case errors.Is(err, ErrInvalidInstrumentID):
		platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_instrument_id", "instrument id is required")
	case errors.Is(err, ErrInvalidName):
		platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_instrument_name", "instrument name is required")
	case errors.Is(err, ErrInvalidUserID):
		platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_user_id", "user id is required")
	case errors.Is(err, ErrInvalidActorID):
		platformweb.WriteError(w, r, http.StatusUnauthorized, "invalid_actor", "actor identity is required")
	case errors.Is(err, ErrForbidden):
		platformweb.WriteError(w, r, http.StatusForbidden, "forbidden", "actor cannot manage this choir")
	case errors.Is(err, ErrInstrumentNameTaken):
		platformweb.WriteError(w, r, http.StatusConflict, "instrument_name_taken", "instrument name already exists")
	case errors.Is(err, ErrUserLinkExists):
		platformweb.WriteError(w, r, http.StatusConflict, "user_link_exists", "user already linked to instrument")
	case errors.Is(err, ErrUserLinkNotFound):
		platformweb.WriteError(w, r, http.StatusNotFound, "user_link_not_found", "user is not linked to instrument")
	case errors.Is(err, ErrInstrumentNotFound):
		platformweb.WriteError(w, r, http.StatusNotFound, "instrument_not_found", "instrument not found")
	default:
		platformweb.WriteError(w, r, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}
