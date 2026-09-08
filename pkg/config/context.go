package config

import (
	"context"
)

type ctxKey struct{}

// WithView attaches a request-scoped config View to ctx.
func WithView(ctx context.Context, view *View) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if view == nil {
		return ctx
	}
	return context.WithValue(ctx, ctxKey{}, view)
}

// ViewFromContext returns the request-scoped View, or nil if absent.
func ViewFromContext(ctx context.Context) *View {
	if ctx == nil {
		return nil
	}
	v, _ := ctx.Value(ctxKey{}).(*View)
	return v
}

// FromContext returns the request-scoped View if present, otherwise the process configurator.
func FromContext(ctx context.Context) Configurator {
	if v := ViewFromContext(ctx); v != nil {
		return v
	}
	return configurator
}

// GetStringCtx reads from the context View when present, else process config.
func GetStringCtx(ctx context.Context, key string) string {
	return FromContext(ctx).GetString(key)
}

// GetIntCtx reads from the context View when present, else process config.
func GetIntCtx(ctx context.Context, key string) int {
	return FromContext(ctx).GetInt(key)
}

// GetInt64Ctx reads from the context View when present, else process config.
func GetInt64Ctx(ctx context.Context, key string) int64 {
	return FromContext(ctx).GetInt64(key)
}

// GetBoolCtx reads from the context View when present, else process config.
func GetBoolCtx(ctx context.Context, key string) bool {
	return FromContext(ctx).GetBool(key)
}

// GetCtx reads from the context View when present, else process config.
func GetCtx(ctx context.Context, key string) any {
	return FromContext(ctx).Get(key)
}

// GetArrayMapsCtx reads from the context View when present, else process config.
func GetArrayMapsCtx(ctx context.Context, key string) []map[string]any {
	return FromContext(ctx).GetArrayMaps(key)
}

// GetArrayStringsCtx reads from the context View when present, else process config.
func GetArrayStringsCtx(ctx context.Context, key string) []string {
	return FromContext(ctx).GetArrayStrings(key)
}
