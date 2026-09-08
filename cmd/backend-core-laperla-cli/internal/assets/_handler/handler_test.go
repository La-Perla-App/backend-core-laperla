package __handler_pkg__handler

import (
	"testing"

	__handler_pkg__ "__package__/__proto_path__/generated/services/__pkg__/__version__"
	"buf.build/go/protovalidate"
	"google.golang.org/protobuf/proto"
)

func TestRequestValidation(t *testing.T) {
	requests := []proto.Message{
		&__handler_pkg__.PingRequest{
			Text: "foo",
		},
	}
	for _, request := range requests {
		if err := protovalidate.Validate(request); err != nil {
			t.Fatalf(`[validation error] %q, got: %v`, request, err)
		}
	}
}
