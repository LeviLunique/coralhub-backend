package tenants

import (
	"encoding/json"
	"errors"
	"net/http"

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
				writeError(w, http.StatusBadRequest, "tenant slug is required")
			case errors.Is(err, ErrTenantNotFound):
				writeError(w, http.StatusNotFound, "tenant not found")
			default:
				writeError(w, http.StatusInternalServerError, "internal server error")
			}
			return
		}

		writeJSON(w, http.StatusOK, bootstrapResponse{
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

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, map[string]string{
		"error": message,
	})
}
