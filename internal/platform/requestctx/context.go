package requestctx

import (
	"context"

	"github.com/LeviLunique/coralhub-backend/internal/modules/tenants"
)

type contextKey string

const tenantContextKey contextKey = "tenant"

func WithTenant(ctx context.Context, tenant tenants.Context) context.Context {
	return context.WithValue(ctx, tenantContextKey, tenant)
}

func TenantFromContext(ctx context.Context) (tenants.Context, bool) {
	tenant, ok := ctx.Value(tenantContextKey).(tenants.Context)
	return tenant, ok
}
