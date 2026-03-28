package requestctx

import (
	"context"

	"github.com/LeviLunique/coralhub-backend/internal/modules/tenants"
)

type contextKey string

const (
	tenantContextKey contextKey = "tenant"
	actorContextKey  contextKey = "actor"
)

type Actor struct {
	ID       string
	TenantID string
	Email    string
	FullName string
}

func WithTenant(ctx context.Context, tenant tenants.Context) context.Context {
	return context.WithValue(ctx, tenantContextKey, tenant)
}

func TenantFromContext(ctx context.Context) (tenants.Context, bool) {
	tenant, ok := ctx.Value(tenantContextKey).(tenants.Context)
	return tenant, ok
}

func WithActor(ctx context.Context, actor Actor) context.Context {
	return context.WithValue(ctx, actorContextKey, actor)
}

func ActorFromContext(ctx context.Context) (Actor, bool) {
	actor, ok := ctx.Value(actorContextKey).(Actor)
	return actor, ok
}
