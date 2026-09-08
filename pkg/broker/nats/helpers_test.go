package natsmanager

import (
	"errors"
	"testing"

	"github.com/nats-io/nats.go/jetstream"
)

func TestIsReplicasUnsupported(t *testing.T) {
	apiErr := &jetstream.APIError{
		Code:        500,
		ErrorCode:   10074,
		Description: "replicas > 1 not supported in non-clustered mode",
	}
	if !isReplicasUnsupported(apiErr) {
		t.Fatal("expected APIError 10074 to match")
	}
	if !isReplicasUnsupported(errors.New("nats: API error: code=500 err_code=10074 description=replicas > 1 not supported in non-clustered mode")) {
		t.Fatal("expected string form to match")
	}
	if isReplicasUnsupported(errors.New("stream not found")) {
		t.Fatal("did not expect unrelated error")
	}
}
