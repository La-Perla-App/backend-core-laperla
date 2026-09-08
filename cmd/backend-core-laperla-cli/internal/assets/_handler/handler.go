package __handler_pkg__handler

import (
	"context"

	__handler_pkg__ "__package__/__proto_path__/generated/services/__pkg__/__version__"
	"__package__/__proto_path__/generated/services/__pkg__/__version__/__handler_pkg__connect"
	"buf.build/go/protovalidate"
	"connectrpc.com/connect"
)

var handler __handler_pkg__connect.__service__ServiceHandler = &handlerImpl{}

type handlerImpl struct{}

func (h *handlerImpl) Ping(ctx context.Context, req *connect.Request[__handler_pkg__.PingRequest]) (*connect.Response[__handler_pkg__.PingResponse], error) {
	// Validate request
	if err := protovalidate.Validate(req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	response := &__handler_pkg__.PingResponse{
		Pong: req.Msg.GetText(),
	}

	return connect.NewResponse(response), nil
}
