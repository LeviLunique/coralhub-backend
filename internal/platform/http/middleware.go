package platformhttp

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/LeviLunique/coralhub-backend/internal/modules/tenants"
	moduleusers "github.com/LeviLunique/coralhub-backend/internal/modules/users"
	platformobservability "github.com/LeviLunique/coralhub-backend/internal/platform/observability"
	"github.com/LeviLunique/coralhub-backend/internal/platform/requestctx"
	platformweb "github.com/LeviLunique/coralhub-backend/internal/platform/web"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startedAt := time.Now()
			ww := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			statusCode := ww.Status()
			if statusCode == 0 {
				statusCode = http.StatusOK
			}

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
				statusCode,
				"duration",
				time.Since(startedAt).String(),
			)
			platformobservability.DefaultMetrics().ObserveHTTPRequest(r.Method, r.URL.Path, statusCode, time.Since(startedAt))
		})
	}
}

func Timeout(duration time.Duration) func(http.Handler) http.Handler {
	if duration <= 0 {
		duration = 30 * time.Second
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), duration)
			defer cancel()

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireTenantContext(service *tenants.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenantSlug := strings.TrimSpace(r.Header.Get("X-Tenant-Slug"))
			if tenantSlug == "" {
				platformweb.WriteError(w, r, http.StatusBadRequest, "tenant_header_required", "X-Tenant-Slug header is required")
				return
			}

			tenant, err := service.ResolveContext(r.Context(), tenantSlug)
			if err != nil {
				switch {
				case errors.Is(err, tenants.ErrInvalidTenantSlug):
					platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_tenant_slug", "tenant slug is required")
				case errors.Is(err, tenants.ErrTenantNotFound):
					platformweb.WriteError(w, r, http.StatusNotFound, "tenant_not_found", "tenant not found")
				default:
					platformweb.WriteError(w, r, http.StatusInternalServerError, "internal_error", "internal server error")
				}
				return
			}

			next.ServeHTTP(w, r.WithContext(requestctx.WithTenant(r.Context(), tenant)))
		})
	}
}

func RequireActorContext(tenantService *tenants.Service, userService *moduleusers.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenantSlug := strings.TrimSpace(r.Header.Get("X-Tenant-Slug"))
			if tenantSlug == "" {
				platformweb.WriteError(w, r, http.StatusBadRequest, "tenant_header_required", "X-Tenant-Slug header is required")
				return
			}

			userEmail := strings.TrimSpace(r.Header.Get("X-User-Email"))
			if userEmail == "" {
				platformweb.WriteError(w, r, http.StatusBadRequest, "actor_header_required", "X-User-Email header is required")
				return
			}

			tenant, err := tenantService.ResolveContext(r.Context(), tenantSlug)
			if err != nil {
				switch {
				case errors.Is(err, tenants.ErrInvalidTenantSlug):
					platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_tenant_slug", "tenant slug is required")
				case errors.Is(err, tenants.ErrTenantNotFound):
					platformweb.WriteError(w, r, http.StatusNotFound, "tenant_not_found", "tenant not found")
				default:
					platformweb.WriteError(w, r, http.StatusInternalServerError, "internal_error", "internal server error")
				}
				return
			}

			user, err := userService.ResolveActorByEmail(r.Context(), tenant.ID, userEmail)
			if err != nil {
				switch {
				case errors.Is(err, moduleusers.ErrInvalidEmail):
					platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_actor_email", "valid user email is required")
				case errors.Is(err, moduleusers.ErrUserNotFound):
					platformweb.WriteError(w, r, http.StatusUnauthorized, "actor_not_found", "actor user not found")
				default:
					platformweb.WriteError(w, r, http.StatusInternalServerError, "internal_error", "internal server error")
				}
				return
			}

			ctx := requestctx.WithTenant(r.Context(), tenant)
			ctx = requestctx.WithActor(ctx, requestctx.Actor{
				ID:       user.ID,
				TenantID: user.TenantID,
				Email:    user.Email,
				FullName: user.FullName,
			})

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
