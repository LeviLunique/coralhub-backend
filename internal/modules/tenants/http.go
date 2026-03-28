package tenants

import (
	"errors"
	"net/http"

	platformweb "github.com/LeviLunique/coralhub-backend/internal/platform/web"
	"github.com/go-chi/chi/v5"
)

type bootstrapResponse struct {
	Slug        string           `json:"slug"`
	DisplayName string           `json:"display_name"`
	Branding    brandingResponse `json:"branding"`
}

type brandingResponse struct {
	LogoURL        *string `json:"logo_url,omitempty"`
	PrimaryColor   *string `json:"primary_color,omitempty"`
	SecondaryColor *string `json:"secondary_color,omitempty"`
	CustomDomain   *string `json:"custom_domain,omitempty"`
}

func RegisterPublicRoutes(router chi.Router, service *Service) {
	router.Get("/tenants/{tenantSlug}", func(w http.ResponseWriter, r *http.Request) {
		tenant, err := service.GetBootstrap(r.Context(), chi.URLParam(r, "tenantSlug"))
		if err != nil {
			switch {
			case errors.Is(err, ErrInvalidTenantSlug):
				platformweb.WriteError(w, r, http.StatusBadRequest, "invalid_tenant_slug", "tenant slug is required")
			case errors.Is(err, ErrTenantNotFound):
				platformweb.WriteError(w, r, http.StatusNotFound, "tenant_not_found", "tenant not found")
			default:
				platformweb.WriteError(w, r, http.StatusInternalServerError, "internal_error", "internal server error")
			}
			return
		}

		platformweb.WriteJSON(w, http.StatusOK, bootstrapResponse{
			Slug:        tenant.Slug,
			DisplayName: tenant.DisplayName,
			Branding: brandingResponse{
				LogoURL:        tenant.Branding.LogoURL,
				PrimaryColor:   tenant.Branding.PrimaryColor,
				SecondaryColor: tenant.Branding.SecondaryColor,
				CustomDomain:   tenant.Branding.CustomDomain,
			},
		})
	})
}
