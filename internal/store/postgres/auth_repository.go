package postgres

import (
	"context"

	"github.com/LeviLunique/coralhub-backend/internal/modules/auth"
	"github.com/LeviLunique/coralhub-backend/internal/store/postgres/sqlc"
)

type AuthRepository struct {
	queries *sqlc.Queries
}

func NewAuthRepository(queries *sqlc.Queries) *AuthRepository {
	return &AuthRepository{queries: queries}
}

func (r *AuthRepository) ListLoginIdentitiesByEmail(ctx context.Context, email string) ([]auth.LoginIdentity, error) {
	rows, err := r.queries.ListLoginIdentitiesByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	items := make([]auth.LoginIdentity, 0, len(rows))
	for _, row := range rows {
		items = append(items, auth.LoginIdentity{
			TenantID:          uuidString(row.TenantID),
			TenantSlug:        row.TenantSlug,
			TenantDisplayName: row.TenantDisplayName,
			UserID:            uuidString(row.ID),
			Email:             row.Email,
			FullName:          row.FullName,
			PasswordHash:      textPointer(row.PasswordHash),
			Manager:           row.Manager,
		})
	}

	return items, nil
}
