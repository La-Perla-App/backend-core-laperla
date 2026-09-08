package examplev1handler

import (
	"__package__/proto/generated/services/example/v1/examplev1connect"
	"connectrpc.com/vanguard"
	"github.com/La-Perla-App/backend-core-laperla/pkg/server"
)

var options = server.ServiceHandlerOptions()

func RegisterServiceHandler() *vanguard.Service {
	return vanguard.NewService(examplev1connect.NewExampleServiceHandler(handler, options...))
}
