package tenants

import "context"

type Repository interface {
	GetBootstrapBySlug(ctx context.Context, slug string) (Bootstrap, error)
}
