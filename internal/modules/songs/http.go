package songs

import (
	"errors"
	"net/http"

	"github.com/LeviLunique/coralhub-backend/internal/platform/requestctx"
	platformweb "github.com/LeviLunique/coralhub-backend/internal/platform/web"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, service *Service) {
	router.Route("/choirs/{choirID}/songs", func(r chi.Router) {
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

			song, err := service.Create(r.Context(), tenant.ID, chi.URLParam(r, "choirID"), actor.ID, input)
			if err != nil {
				writeSongError(w, r, err)
				return
			}

			platformweb.WriteJSON(w, http.StatusCreated, song)
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
				writeSongError(w, r, err)
				return
			}

			platformweb.WriteJSON(w, http.StatusOK, map[string][]Song{"items": items})
		})
	})

	router.Route("/songs/{songID}", func(r chi.Router) {
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

			song, err := service.Get(r.Context(), tenant.ID, actor.ID, chi.URLParam(r, "songID"))
			if err != nil {
				writeSongError(w, r, err)
				return
			}

			platformweb.WriteJSON(w, http.StatusOK, song)
		})

		r.Put("/", func(w http.ResponseWriter, r *http.Request) {
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

			var input UpdateInput
			if err := platformweb.DecodeJSONBody(r, &input); err != nil {
				platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_request_body", "request body must be a single valid JSON object")
				return
			}

			song, err := service.Update(r.Context(), tenant.ID, chi.URLParam(r, "songID"), actor.ID, input)
			if err != nil {
				writeSongError(w, r, err)
				return
			}

			platformweb.WriteJSON(w, http.StatusOK, song)
		})

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

			err := service.Delete(r.Context(), tenant.ID, chi.URLParam(r, "songID"), actor.ID)
			if err != nil {
				writeSongError(w, r, err)
				return
			}

			w.WriteHeader(http.StatusNoContent)
		})
	})
}

func writeSongError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, ErrInvalidChoirID):
		platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_choir_id", "choir id is required")
	case errors.Is(err, ErrInvalidSongID):
		platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_song_id", "song id is required")
	case errors.Is(err, ErrInvalidTitle):
		platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_song_title", "song title is required")
	case errors.Is(err, ErrInvalidActorID):
		platformweb.WriteError(w, r, http.StatusUnauthorized, "invalid_actor", "actor identity is required")
	case errors.Is(err, ErrForbidden):
		platformweb.WriteError(w, r, http.StatusForbidden, "forbidden", "actor cannot manage this choir")
	case errors.Is(err, ErrSongTitleTaken):
		platformweb.WriteError(w, r, http.StatusConflict, "song_title_taken", "song title already exists")
	case errors.Is(err, ErrSongNotFound):
		platformweb.WriteError(w, r, http.StatusNotFound, "song_not_found", "song not found")
	default:
		platformweb.WriteError(w, r, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}
