package natsmanager

import (
	"context"
	"errors"
	"strings"

	"github.com/nats-io/nats.go/jetstream"
)

// NATS JetStream: replicas > 1 not supported in non-clustered mode.
const errCodeReplicasUnsupported jetstream.ErrorCode = 10074

// EnsureStreams creates or updates each stream.
// Preferred Replicas from cfg are kept (e.g. 3 on a hub cluster). If the server
// returns 10074 (standalone / leaf), the same stream is retried with Replicas=1.
func EnsureStreams(ctx context.Context, js jetstream.JetStream, streamsConfig ...jetstream.StreamConfig) error {
	for _, streamConfig := range streamsConfig {
		if err := ensureStream(ctx, js, streamConfig); err != nil {
			return err
		}
	}
	return nil
}

func ensureStream(ctx context.Context, js jetstream.JetStream, cfg jetstream.StreamConfig) error {
	_, err := js.CreateOrUpdateStream(ctx, cfg)
	if err == nil {
		return nil
	}
	if cfg.Replicas <= 1 || !isReplicasUnsupported(err) {
		return err
	}
	cfg.Replicas = 1
	_, err = js.CreateOrUpdateStream(ctx, cfg)
	return err
}

func isReplicasUnsupported(err error) bool {
	var apiErr *jetstream.APIError
	if errors.As(err, &apiErr) && apiErr.ErrorCode == errCodeReplicasUnsupported {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "replicas > 1 not supported") ||
		strings.Contains(msg, "non-clustered mode")
}
