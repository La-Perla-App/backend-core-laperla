package handlers

import (
	examplev1handler "__package__/handlers/example/v1"

	"github.com/La-Perla-App/backend-core-laperla/pkg/server"
)

var RegisterServicesFns = []server.RegisterServiceFn{
	examplev1handler.RegisterServiceHandler,
}
