package natsmanager

import (
	"time"

	"github.com/nats-io/nats.go/jetstream"
)

// DefaultRetryBackoff es la espera entre reentregas cuando el consumidor no
// declara `BackOff` propio.
//
// Existe porque el fallo de un worker casi nunca se arregla reintentando al
// instante: lo que falla suele ser una dependencia caída (OpenSearch, SMTP, un
// servicio) y necesita tiempo, no prisa. Con reentrega inmediata, un consumidor
// con `MaxDeliver: 8` se come los ocho intentos en menos de un segundo y pierde
// el mensaje justo cuando la dependencia aún no ha vuelto.
//
// Escalado: ~50 min por mensaje, que da margen a un reinicio o a un despliegue.
var DefaultRetryBackoff = []time.Duration{
	5 * time.Second,
	15 * time.Second,
	time.Minute,
	5 * time.Minute,
	15 * time.Minute,
	30 * time.Minute,
}

// RetryDelay dice cuánto esperar antes de la siguiente entrega, dado el backoff
// del consumidor y las veces que el mensaje ya se entregó.
//
// `backoff` vacío usa DefaultRetryBackoff. Pasado el último escalón se repite
// ese, en vez de salirse del slice. Un `backoff` de `{0}` reproduce el
// comportamiento antiguo: reentrega inmediata.
func RetryDelay(backoff []time.Duration, numDelivered uint64) time.Duration {
	if len(backoff) == 0 {
		backoff = DefaultRetryBackoff
	}
	if numDelivered == 0 {
		return backoff[0]
	}
	i := int(numDelivered) - 1
	if i >= len(backoff) {
		i = len(backoff) - 1
	}
	return backoff[i]
}

// retryDelayForMsg saca el número de entrega del propio mensaje; si no viene
// (no debería), trata al mensaje como si fuera la primera.
func retryDelayForMsg(backoff []time.Duration, msg jetstream.Msg) time.Duration {
	md, err := msg.Metadata()
	if err != nil {
		return RetryDelay(backoff, 1)
	}
	return RetryDelay(backoff, md.NumDelivered)
}
