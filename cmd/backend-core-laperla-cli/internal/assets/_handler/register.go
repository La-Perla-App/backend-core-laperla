package __handler_pkg__handler

import (
	"__package__/__proto_path__/generated/services/__pkg__/__version__/__handler_pkg__connect"
	"connectrpc.com/vanguard"
	"github.com/La-Perla-App/backend-core-laperla/pkg/server"
)

var options = server.ServiceHandlerOptions()

func RegisterServiceHandler() *vanguard.Service {
	return vanguard.NewService(__handler_pkg__connect.New__service__ServiceHandler(handler, options...))
}
