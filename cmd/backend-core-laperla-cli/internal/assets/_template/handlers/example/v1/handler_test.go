package examplev1handler

import (
	"testing"

	examplev1 "__package__/proto/generated/services/example/v1"
	"buf.build/go/protovalidate"
	"google.golang.org/protobuf/proto"
)

func TestRequestValidation(t *testing.T) {
	requests := []proto.Message{
		&examplev1.PingRequest{
			Text: "foo",
		},
	}
	for _, request := range requests {
		if err := protovalidate.Validate(request); err != nil {
			t.Fatalf(`[validation error] %q, got: %v`, request, err)
		}
	}
}
