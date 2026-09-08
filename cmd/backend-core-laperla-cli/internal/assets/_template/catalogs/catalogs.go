package catalogs

import (
	"context"
	"os"

	"github.com/La-Perla-App/backend-core-laperla/pkg/config"
)

var (
	IsProd = os.Getenv("MODE") == "PROD"

	SpecialRoutes = struct {
		DebugRoute     string
		SwaggerRoute   string
		ProtosDownload string
	}{
		DebugRoute:     "__api_prefix__/debug",
		SwaggerRoute:   "__api_prefix__/swagger",
		ProtosDownload: "__api_prefix__/protos_download",
	}
)

// ClientAddress returns the service base URL from process (or request-scoped) config.
func ClientAddress(ctx context.Context) string {
	key := "grpc.clientAddresses.__package_short__"
	addr := config.GetStringCtx(ctx, key)
	if addr == "" {
		addr = config.GetString(key)
	}
	return addr
}
