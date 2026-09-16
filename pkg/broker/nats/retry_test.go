package natsmanager

import (
	"testing"
	"time"
)

func TestRetryDelayUsesDefaultWhenUnset(t *testing.T) {
	if got := RetryDelay(nil, 1); got != DefaultRetryBackoff[0] {
		t.Fatalf("%v", got)
	}
	// Nunca inmediato por defecto: eso es lo que quemaba los reintentos.
	if RetryDelay(nil, 1) <= 0 {
		t.Fatal("el backoff por defecto no puede ser 0")
	}
}

func TestRetryDelayGrowsAndPlateaus(t *testing.T) {
	backoff := []time.Duration{time.Second, 2 * time.Second, 4 * time.Second}
	if got := RetryDelay(backoff, 0); got != time.Second {
		t.Fatalf("primera entrega: %v", got)
	}
	if got := RetryDelay(backoff, 1); got != time.Second {
		t.Fatalf("intento 1: %v", got)
	}
	if got := RetryDelay(backoff, 3); got != 4*time.Second {
		t.Fatalf("intento 3: %v", got)
	}
	// Pasado el último escalón se queda en él en vez de salirse del slice.
	if got := RetryDelay(backoff, 99); got != 4*time.Second {
		t.Fatalf("intento 99: %v", got)
	}
}

func TestRetryDelayHonorsExplicitZero(t *testing.T) {
	// Escotilla para quien quiera la reentrega inmediata de antes.
	if got := RetryDelay([]time.Duration{0}, 5); got != 0 {
		t.Fatalf("%v", got)
	}
}

func TestDefaultRetryBackoffCoversAnOutage(t *testing.T) {
	var total time.Duration
	for _, d := range DefaultRetryBackoff {
		total += d
	}
	if total < 30*time.Minute {
		t.Fatalf("ventana de %s: muy corta para un reinicio o un despliegue", total)
	}
}
