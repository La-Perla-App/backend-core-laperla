//go:build smoke

package natsmanager

import (
	"context"
	"log/slog"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/nats-io/nats.go/jetstream"
)

// Un callback que falla tiene que reentregarse **esperando**. Es el arreglo:
// antes se hacía `Nak()` a secas y volvía al instante, así que un consumidor con
// MaxDeliver bajo se quedaba sin intentos antes de que la dependencia caída
// tuviera ocasión de volver.
//
// go test -tags smoke ./pkg/broker/nats/ con SMOKE_NATS_URL.
func TestSmokeAddWorkerBacksOffBeforeRedelivery(t *testing.T) {
	url := os.Getenv("SMOKE_NATS_URL")
	if url == "" {
		t.Skip("SMOKE_NATS_URL vacío")
	}
	ctx := context.Background()
	Connect(Config{URLs: []string{url}, Logger: slog.Default()})
	m, err := Get()
	if err != nil {
		t.Fatal(err)
	}

	const (
		stream  = "CORE_RETRY_SMOKE"
		subject = "core.retry.smoke"
		durable = "core-retry-smoke"
	)
	js := m.GetJS()
	if err := EnsureStreams(ctx, js, jetstream.StreamConfig{
		Name:     stream,
		Subjects: []string{subject},
		Storage:  jetstream.MemoryStorage,
		MaxAge:   5 * time.Minute,
		Replicas: 1,
	}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = js.DeleteStream(ctx, stream) })

	var attempts int32
	done := make(chan time.Time, 2)
	// Backoff corto para no alargar el test; lo que se prueba es que espera.
	cfg := jetstream.ConsumerConfig{
		Durable:    durable,
		AckPolicy:  jetstream.AckExplicitPolicy,
		AckWait:    30 * time.Second,
		MaxDeliver: 5,
		BackOff:    []time.Duration{3 * time.Second, 3 * time.Second, 3 * time.Second, 3 * time.Second},
	}
	if _, err := m.AddWorker(subject, stream, func(msg jetstream.Msg) bool {
		if atomic.AddInt32(&attempts, 1) == 1 {
			return false // primera entrega: simula la dependencia caída
		}
		done <- time.Now()
		return true
	}, cfg); err != nil {
		t.Fatal(err)
	}

	sent := time.Now()
	if _, err := js.Publish(ctx, subject, []byte(`{"hola":"mundo"}`)); err != nil {
		t.Fatal(err)
	}

	select {
	case at := <-done:
		if waited := at.Sub(sent); waited < 2*time.Second {
			t.Fatalf("reentregó en %s: no respetó el backoff", waited)
		}
	case <-time.After(30 * time.Second):
		t.Fatalf("no hubo reentrega (intentos: %d)", atomic.LoadInt32(&attempts))
	}
	if n := atomic.LoadInt32(&attempts); n < 2 {
		t.Fatalf("debía haber reintento, hubo %d", n)
	}
}
