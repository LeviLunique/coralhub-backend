package postgres

import (
	"context"
	"errors"

	"github.com/LeviLunique/coralhub-backend/internal/modules/tenants"
	"github.com/LeviLunique/coralhub-backend/internal/store/postgres/sqlc"
	"github.com/jackc/pgx/v5"
)

type TenantRepository struct {
	queries *sqlc.Queries
}

func NewTenantRepository(queries *sqlc.Queries) *TenantRepository {
	return &TenantRepository{queries: queries}
}

func (r *TenantRepository) GetBootstrapBySlug(ctx context.Context, slug string) (tenants.Bootstrap, error) {
	row, err := r.queries.GetTenantBootstrapBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return tenants.Bootstrap{}, tenants.ErrTenantNotFound
		}

		return tenants.Bootstrap{}, err
	}

	return tenants.Bootstrap{
		Slug:        row.Slug,
		DisplayName: row.DisplayName,
		Branding: tenants.Branding{
			LogoURL:        textPointer(row.LogoUrl),
			PrimaryColor:   textPointer(row.PrimaryColor),
			SecondaryColor: textPointer(row.SecondaryColor),
			CustomDomain:   textPointer(row.CustomDomain),
		},
	}, nil
}
