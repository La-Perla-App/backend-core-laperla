package examplev1handler

import (
	"context"

	examplev1 "__package__/proto/generated/services/example/v1"
	"__package__/proto/generated/services/example/v1/examplev1connect"
	"buf.build/go/protovalidate"
	"connectrpc.com/connect"
)

var handler examplev1connect.ExampleServiceHandler = &handlerImpl{}

type handlerImpl struct{}

func (h *handlerImpl) Ping(ctx context.Context, req *connect.Request[examplev1.PingRequest]) (*connect.Response[examplev1.PingResponse], error) {
	// Validate request
	if err := protovalidate.Validate(req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	response := &examplev1.PingResponse{
		Pong: req.Msg.GetText(),
	}

	return connect.NewResponse(response), nil
}
