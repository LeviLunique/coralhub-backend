package platformhttp

import (
	"log/slog"
	"net/http"
	"time"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startedAt := time.Now()
			ww := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			logger.InfoContext(
				r.Context(),
				"http request",
				"request_id",
				chimiddleware.GetReqID(r.Context()),
				"method",
				r.Method,
				"path",
				r.URL.Path,
				"status",
				ww.Status(),
				"duration",
				time.Since(startedAt).String(),
			)
		})
	}
}
