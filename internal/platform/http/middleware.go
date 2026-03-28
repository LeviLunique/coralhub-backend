package platformhttp

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/LeviLunique/coralhub-backend/internal/modules/tenants"
	moduleusers "github.com/LeviLunique/coralhub-backend/internal/modules/users"
	"github.com/LeviLunique/coralhub-backend/internal/platform/requestctx"
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

func RequireTenantContext(service *tenants.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenantSlug := strings.TrimSpace(r.Header.Get("X-Tenant-Slug"))
			if tenantSlug == "" {
				WriteError(w, http.StatusBadRequest, "X-Tenant-Slug header is required")
				return
			}

			tenant, err := service.ResolveContext(r.Context(), tenantSlug)
			if err != nil {
				switch {
				case errors.Is(err, tenants.ErrInvalidTenantSlug):
					WriteError(w, http.StatusBadRequest, "tenant slug is required")
				case errors.Is(err, tenants.ErrTenantNotFound):
					WriteError(w, http.StatusNotFound, "tenant not found")
				default:
					WriteError(w, http.StatusInternalServerError, "internal server error")
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
				WriteError(w, http.StatusBadRequest, "X-Tenant-Slug header is required")
				return
			}

			userEmail := strings.TrimSpace(r.Header.Get("X-User-Email"))
			if userEmail == "" {
				WriteError(w, http.StatusBadRequest, "X-User-Email header is required")
				return
			}

			tenant, err := tenantService.ResolveContext(r.Context(), tenantSlug)
			if err != nil {
				switch {
				case errors.Is(err, tenants.ErrInvalidTenantSlug):
					WriteError(w, http.StatusBadRequest, "tenant slug is required")
				case errors.Is(err, tenants.ErrTenantNotFound):
					WriteError(w, http.StatusNotFound, "tenant not found")
				default:
					WriteError(w, http.StatusInternalServerError, "internal server error")
				}
				return
			}

			user, err := userService.ResolveActorByEmail(r.Context(), tenant.ID, userEmail)
			if err != nil {
				switch {
				case errors.Is(err, moduleusers.ErrInvalidEmail):
					WriteError(w, http.StatusBadRequest, "valid user email is required")
				case errors.Is(err, moduleusers.ErrUserNotFound):
					WriteError(w, http.StatusUnauthorized, "actor user not found")
				default:
					WriteError(w, http.StatusInternalServerError, "internal server error")
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
