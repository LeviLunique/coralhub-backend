package materials

import (
	"errors"
	"net/http"

	"github.com/LeviLunique/coralhub-backend/internal/modules/instruments"
	"github.com/LeviLunique/coralhub-backend/internal/modules/tenants"
	"github.com/LeviLunique/coralhub-backend/internal/platform/requestctx"
	platformweb "github.com/LeviLunique/coralhub-backend/internal/platform/web"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, service *Service) {
	router.Route("/choirs/{choirID}/materials", func(r chi.Router) {
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

			material, err := service.Create(r.Context(), tenant.ID, chi.URLParam(r, "choirID"), actor.ID, input)
			if err != nil {
				writeMaterialError(w, r, err)
				return
			}

			platformweb.WriteJSON(w, http.StatusCreated, material)
		})

		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			tenant, actor, ok := contextActors(w, r)
			if !ok {
				return
			}

			items, err := service.ListByChoir(r.Context(), tenant.ID, chi.URLParam(r, "choirID"), actor.ID)
			if err != nil {
				writeMaterialError(w, r, err)
				return
			}

			platformweb.WriteJSON(w, http.StatusOK, map[string][]Material{"items": items})
		})
	})

	router.Route("/songs/{songID}/materials", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			tenant, actor, ok := contextActors(w, r)
			if !ok {
				return
			}

			items, err := service.ListBySong(r.Context(), tenant.ID, chi.URLParam(r, "songID"), actor.ID)
			if err != nil {
				writeMaterialError(w, r, err)
				return
			}

			platformweb.WriteJSON(w, http.StatusOK, map[string][]Material{"items": items})
		})
	})

	router.Route("/materials/{materialID}", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			tenant, actor, ok := contextActors(w, r)
			if !ok {
				return
			}

			material, err := service.Get(r.Context(), tenant.ID, actor.ID, chi.URLParam(r, "materialID"))
			if err != nil {
				writeMaterialError(w, r, err)
				return
			}

			platformweb.WriteJSON(w, http.StatusOK, material)
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

			material, err := service.Update(r.Context(), tenant.ID, chi.URLParam(r, "materialID"), actor.ID, input)
			if err != nil {
				writeMaterialError(w, r, err)
				return
			}

			platformweb.WriteJSON(w, http.StatusOK, material)
		})

		r.Delete("/", func(w http.ResponseWriter, r *http.Request) {
			tenant, actor, ok := contextActors(w, r)
			if !ok {
				return
			}

			if err := service.Delete(r.Context(), tenant.ID, chi.URLParam(r, "materialID"), actor.ID); err != nil {
				writeMaterialError(w, r, err)
				return
			}

			w.WriteHeader(http.StatusNoContent)
		})

		r.Get("/instruments", func(w http.ResponseWriter, r *http.Request) {
			tenant, actor, ok := contextActors(w, r)
			if !ok {
				return
			}

			items, err := service.ListInstruments(r.Context(), tenant.ID, chi.URLParam(r, "materialID"), actor.ID)
			if err != nil {
				writeMaterialError(w, r, err)
				return
			}

			platformweb.WriteJSON(w, http.StatusOK, map[string][]instruments.Instrument{"items": items})
		})

		r.Post("/instruments", func(w http.ResponseWriter, r *http.Request) {
			tenant, actor, ok := contextActors(w, r)
			if !ok {
				return
			}

			var input LinkInstrumentInput
			if err := platformweb.DecodeJSONBody(r, &input); err != nil {
				platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_request_body", "request body must be a single valid JSON object")
				return
			}

			if err := service.AddInstrument(r.Context(), tenant.ID, chi.URLParam(r, "materialID"), actor.ID, input); err != nil {
				writeMaterialError(w, r, err)
				return
			}

			w.WriteHeader(http.StatusNoContent)
		})

		r.Delete("/instruments/{instrumentID}", func(w http.ResponseWriter, r *http.Request) {
			tenant, actor, ok := contextActors(w, r)
			if !ok {
				return
			}

			if err := service.RemoveInstrument(r.Context(), tenant.ID, chi.URLParam(r, "materialID"), actor.ID, chi.URLParam(r, "instrumentID")); err != nil {
				writeMaterialError(w, r, err)
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

func writeMaterialError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, ErrInvalidChoirID):
		platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_choir_id", "choir id is required")
	case errors.Is(err, ErrInvalidMaterialID):
		platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_material_id", "material id is required")
	case errors.Is(err, ErrInvalidSongID):
		platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_song_id", "song id is required")
	case errors.Is(err, ErrInvalidName):
		platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_material_name", "material name is required")
	case errors.Is(err, ErrInvalidMaterialType):
		platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_material_type", "material type must be audio_guide, sheet_music, playback, lyrics, or other")
	case errors.Is(err, ErrInvalidTargetType):
		platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_target_type", "target type must be voice or instrument")
	case errors.Is(err, ErrInvalidInstrumentID):
		platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_instrument_id", "instrument id is required")
	case errors.Is(err, ErrInvalidActorID):
		platformweb.WriteError(w, r, http.StatusUnauthorized, "invalid_actor", "actor identity is required")
	case errors.Is(err, ErrForbidden):
		platformweb.WriteError(w, r, http.StatusForbidden, "forbidden", "actor cannot manage this choir")
	case errors.Is(err, ErrInstrumentLinkExists):
		platformweb.WriteError(w, r, http.StatusConflict, "instrument_link_exists", "instrument already linked to material")
	case errors.Is(err, ErrInstrumentLinkNotFound):
		platformweb.WriteError(w, r, http.StatusNotFound, "instrument_link_not_found", "instrument is not linked to material")
	case errors.Is(err, ErrMaterialNotFound):
		platformweb.WriteError(w, r, http.StatusNotFound, "material_not_found", "material not found")
	default:
		platformweb.WriteError(w, r, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}
