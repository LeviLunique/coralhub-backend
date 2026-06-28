package auth

import "context"

type Repository interface {
	ListLoginIdentitiesByEmail(ctx context.Context, email string) ([]LoginIdentity, error)
}
