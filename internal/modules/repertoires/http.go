package repertoires

import (
	"errors"
	"net/http"

	"github.com/LeviLunique/coralhub-backend/internal/modules/tenants"
	"github.com/LeviLunique/coralhub-backend/internal/platform/requestctx"
	platformweb "github.com/LeviLunique/coralhub-backend/internal/platform/web"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, service *Service) {
	router.Route("/choirs/{choirID}/repertoires", func(r chi.Router) {
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

			repertoire, err := service.Create(r.Context(), tenant.ID, chi.URLParam(r, "choirID"), actor.ID, input)
			if err != nil {
				writeRepertoireError(w, r, err)
				return
			}

			platformweb.WriteJSON(w, http.StatusCreated, repertoire)
		})

		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			tenant, actor, ok := contextActors(w, r)
			if !ok {
				return
			}

			items, err := service.ListByChoir(r.Context(), tenant.ID, chi.URLParam(r, "choirID"), actor.ID)
			if err != nil {
				writeRepertoireError(w, r, err)
				return
			}

			platformweb.WriteJSON(w, http.StatusOK, map[string][]Repertoire{"items": items})
		})
	})

	router.Route("/repertoires/{repertoireID}", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			tenant, actor, ok := contextActors(w, r)
			if !ok {
				return
			}

			repertoire, err := service.Get(r.Context(), tenant.ID, actor.ID, chi.URLParam(r, "repertoireID"))
			if err != nil {
				writeRepertoireError(w, r, err)
				return
			}

			platformweb.WriteJSON(w, http.StatusOK, repertoire)
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

			repertoire, err := service.Update(r.Context(), tenant.ID, chi.URLParam(r, "repertoireID"), actor.ID, input)
			if err != nil {
				writeRepertoireError(w, r, err)
				return
			}

			platformweb.WriteJSON(w, http.StatusOK, repertoire)
		})

		r.Delete("/", func(w http.ResponseWriter, r *http.Request) {
			tenant, actor, ok := contextActors(w, r)
			if !ok {
				return
			}

			if err := service.Delete(r.Context(), tenant.ID, chi.URLParam(r, "repertoireID"), actor.ID); err != nil {
				writeRepertoireError(w, r, err)
				return
			}

			w.WriteHeader(http.StatusNoContent)
		})

		r.Get("/songs", func(w http.ResponseWriter, r *http.Request) {
			tenant, actor, ok := contextActors(w, r)
			if !ok {
				return
			}

			items, err := service.ListSongs(r.Context(), tenant.ID, chi.URLParam(r, "repertoireID"), actor.ID)
			if err != nil {
				writeRepertoireError(w, r, err)
				return
			}

			platformweb.WriteJSON(w, http.StatusOK, map[string][]RepertoireSong{"items": items})
		})

		r.Post("/songs", func(w http.ResponseWriter, r *http.Request) {
			tenant, actor, ok := contextActors(w, r)
			if !ok {
				return
			}

			var input AddSongInput
			if err := platformweb.DecodeJSONBody(r, &input); err != nil {
				platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_request_body", "request body must be a single valid JSON object")
				return
			}

			if err := service.AddSong(r.Context(), tenant.ID, chi.URLParam(r, "repertoireID"), actor.ID, input); err != nil {
				writeRepertoireError(w, r, err)
				return
			}

			w.WriteHeader(http.StatusNoContent)
		})

		r.Delete("/songs/{songID}", func(w http.ResponseWriter, r *http.Request) {
			tenant, actor, ok := contextActors(w, r)
			if !ok {
				return
			}

			if err := service.RemoveSong(r.Context(), tenant.ID, chi.URLParam(r, "repertoireID"), actor.ID, chi.URLParam(r, "songID")); err != nil {
				writeRepertoireError(w, r, err)
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

func writeRepertoireError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, ErrInvalidChoirID):
		platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_choir_id", "choir id is required")
	case errors.Is(err, ErrInvalidRepertoireID):
		platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_repertoire_id", "repertoire id is required")
	case errors.Is(err, ErrInvalidName):
		platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_repertoire_name", "repertoire name is required")
	case errors.Is(err, ErrInvalidSongID):
		platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_song_id", "song id is required")
	case errors.Is(err, ErrInvalidActorID):
		platformweb.WriteError(w, r, http.StatusUnauthorized, "invalid_actor", "actor identity is required")
	case errors.Is(err, ErrForbidden):
		platformweb.WriteError(w, r, http.StatusForbidden, "forbidden", "actor cannot manage this choir")
	case errors.Is(err, ErrRepertoireNameTaken):
		platformweb.WriteError(w, r, http.StatusConflict, "repertoire_name_taken", "repertoire name already exists")
	case errors.Is(err, ErrSongLinkExists):
		platformweb.WriteError(w, r, http.StatusConflict, "song_link_exists", "song already linked to repertoire")
	case errors.Is(err, ErrSongLinkNotFound):
		platformweb.WriteError(w, r, http.StatusNotFound, "song_link_not_found", "song is not linked to repertoire")
	case errors.Is(err, ErrRepertoireNotFound):
		platformweb.WriteError(w, r, http.StatusNotFound, "repertoire_not_found", "repertoire not found")
	default:
		platformweb.WriteError(w, r, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}
