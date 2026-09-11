package minio

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/La-Perla-App/backend-core-laperla/pkg/config"
)

// ConfigFromContext reads object-store settings from the request-scoped config
// View when present, otherwise from process config.
//
// Uses `storage.*` when any of provider/endpoint/accessKey is set there;
// otherwise legacy `minio.*`. Fields are not mixed across prefixes.
func ConfigFromContext(ctx context.Context) Config {
	prefix := "minio"
	if firstString(ctx, "storage.provider", "storage.endpoint", "storage.accessKey") != "" {
		prefix = "storage"
	}
	cfg := Config{
		Provider:      firstString(ctx, prefix+".provider"),
		Endpoint:      firstString(ctx, prefix+".endpoint"),
		Region:        firstString(ctx, prefix+".region"),
		AccessKey:     firstString(ctx, prefix+".accessKey"),
		SecretKey:     firstString(ctx, prefix+".secretKey"),
		UseSSL:        firstBool(ctx, prefix+".useSsl"),
		PublicHost:    firstString(ctx, prefix+".publicHost"),
		PublicSSL:     firstBool(ctx, prefix+".publicUseSsl"),
		DefaultBucket: firstString(ctx, prefix+".defaultBucket"),
	}
	if ps, ok := optionalBool(ctx, prefix+".pathStyle"); ok {
		cfg.PathStyle = &ps
	}
	if d := firstDuration(ctx, prefix+".defaultPutExpiry"); d > 0 {
		cfg.DefaultPutExpiry = d
	}
	if d := firstDuration(ctx, prefix+".defaultGetExpiry"); d > 0 {
		cfg.DefaultGetExpiry = d
	}
	return cfg
}

func firstString(ctx context.Context, keys ...string) string {
	for _, k := range keys {
		if v := strings.TrimSpace(config.GetStringCtx(ctx, k)); v != "" {
			return v
		}
	}
	return ""
}

func firstBool(ctx context.Context, keys ...string) bool {
	for _, k := range keys {
		// GetBoolCtx returns false both for missing and explicit false; prefer
		// the first key that has a non-empty string form or is set via YAML bool.
		if raw := strings.TrimSpace(config.GetStringCtx(ctx, k)); raw != "" {
			v, err := strconv.ParseBool(raw)
			if err == nil {
				return v
			}
		}
		if config.GetBoolCtx(ctx, k) {
			return true
		}
	}
	return false
}

func optionalBool(ctx context.Context, keys ...string) (bool, bool) {
	for _, k := range keys {
		raw := config.GetCtx(ctx, k)
		if raw == nil {
			continue
		}
		switch v := raw.(type) {
		case bool:
			return v, true
		case string:
			if b, err := strconv.ParseBool(strings.TrimSpace(v)); err == nil {
				return b, true
			}
		case int, int32, int64, float32, float64:
			return config.GetBoolCtx(ctx, k), true
		}
	}
	return false, false
}

func firstDuration(ctx context.Context, keys ...string) time.Duration {
	for _, k := range keys {
		raw := strings.TrimSpace(config.GetStringCtx(ctx, k))
		if raw == "" {
			continue
		}
		if d, err := time.ParseDuration(raw); err == nil {
			return d
		}
		if secs, err := strconv.Atoi(raw); err == nil {
			return time.Duration(secs) * time.Second
		}
	}
	return 0
}

// Configured reports whether enough settings exist to build a client.
func (cfg Config) Configured() bool {
	n := normalizeConfig(cfg)
	return n.Endpoint != "" && n.AccessKey != "" && n.SecretKey != ""
}
