package postgres

import (
	"context"
	"errors"

	moduleusers "github.com/LeviLunique/coralhub-backend/internal/modules/users"
	"github.com/LeviLunique/coralhub-backend/internal/store/postgres/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type UserRepository struct {
	queries *sqlc.Queries
}

func NewUserRepository(queries *sqlc.Queries) *UserRepository {
	return &UserRepository{queries: queries}
}

func (r *UserRepository) Create(ctx context.Context, params moduleusers.CreateParams) (moduleusers.User, error) {
	tenantID, err := parseUUID(params.TenantID)
	if err != nil {
		return moduleusers.User{}, moduleusers.ErrInvalidTenantID
	}

	row, err := r.queries.CreateUser(ctx, sqlc.CreateUserParams{
		TenantID: tenantID,
		Email:    params.Email,
		FullName: params.FullName,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return moduleusers.User{}, moduleusers.ErrUserEmailTaken
		}

		return moduleusers.User{}, err
	}

	return moduleusers.User{
		ID:       uuidString(row.ID),
		TenantID: uuidString(row.TenantID),
		Email:    row.Email,
		FullName: row.FullName,
		Active:   row.Active,
	}, nil
}

func (r *UserRepository) GetByID(ctx context.Context, tenantID string, userID string) (moduleusers.User, error) {
	tenantUUID, err := parseUUID(tenantID)
	if err != nil {
		return moduleusers.User{}, moduleusers.ErrInvalidTenantID
	}

	userUUID, err := parseUUID(userID)
	if err != nil {
		return moduleusers.User{}, moduleusers.ErrInvalidUserID
	}

	row, err := r.queries.GetUserByID(ctx, sqlc.GetUserByIDParams{
		TenantID: tenantUUID,
		ID:       userUUID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return moduleusers.User{}, moduleusers.ErrUserNotFound
		}

		return moduleusers.User{}, err
	}

	return moduleusers.User{
		ID:       uuidString(row.ID),
		TenantID: uuidString(row.TenantID),
		Email:    row.Email,
		FullName: row.FullName,
		Active:   row.Active,
	}, nil
}

func (r *UserRepository) ListByTenantID(ctx context.Context, tenantID string) ([]moduleusers.User, error) {
	tenantUUID, err := parseUUID(tenantID)
	if err != nil {
		return nil, moduleusers.ErrInvalidTenantID
	}

	rows, err := r.queries.ListUsersByTenantID(ctx, tenantUUID)
	if err != nil {
		return nil, err
	}

	items := make([]moduleusers.User, 0, len(rows))
	for _, row := range rows {
		items = append(items, moduleusers.User{
			ID:       uuidString(row.ID),
			TenantID: uuidString(row.TenantID),
			Email:    row.Email,
			FullName: row.FullName,
			Active:   row.Active,
		})
	}

	return items, nil
}
