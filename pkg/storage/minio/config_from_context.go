package minio

import (
	"context"
	"strings"

	"github.com/La-Perla-App/backend-core-laperla/pkg/config"
)

// ConfigFromContext reads MinIO settings from the request-scoped config View
// when present, otherwise from process config.
//
// Use FromContext to build a client, or this helper when exporting connection
// params to another process.
func ConfigFromContext(ctx context.Context) Config {
	return Config{
		Endpoint:   strings.TrimSpace(config.GetStringCtx(ctx, "minio.endpoint")),
		AccessKey:  strings.TrimSpace(config.GetStringCtx(ctx, "minio.accessKey")),
		SecretKey:  strings.TrimSpace(config.GetStringCtx(ctx, "minio.secretKey")),
		UseSSL:     config.GetBoolCtx(ctx, "minio.useSsl"),
		PublicHost: strings.TrimSpace(config.GetStringCtx(ctx, "minio.publicHost")),
		PublicSSL:  config.GetBoolCtx(ctx, "minio.publicUseSsl"),
	}
}
