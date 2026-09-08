package events

import (
	"context"
	"log/slog"
	"sync"

	"connectrpc.com/connect"
	"github.com/La-Perla-App/backend-core-laperla/pkg/api"
)

const messageBufferSize = 1000

// clientStream representa la conexión de un cliente y su búfer de envío.
type clientStream[T any] struct {
	stream  *connect.ServerStream[T]
	msgChan chan *T
	done    chan struct{}
	logger  *slog.Logger
	ctx     context.Context // Contexto para manejar la cancelación del stream
}

// senderLoop es una goroutine que se encarga de enviar mensajes desde el canal al stream del cliente.
func (cs *clientStream[T]) senderLoop(clientID string) {
	defer cs.logger.Info("Sender loop stopped for client", "clientID", clientID)

	for {
		select {
		case msg := <-cs.msgChan:
			if err := cs.stream.Send(msg); err != nil {
				cs.logger.Error("Failed to send event to stream, stopping loop", "clientID", clientID, "error", err)
				// La conexión está rota. La goroutine termina.
				return
			}
		case <-cs.done:
			return // Termina la goroutine cuando se llama a Unregister.
		case <-cs.ctx.Done():
			cs.logger.Info("Stream context cancelled, stopping sender loop", "clientID", clientID, "error", cs.ctx.Err())
			return // El contexto del stream fue cancelado (cliente desconectado), termina la goroutine.
		}
	}
}

// StreamManager gestiona los streams activos de los usuarios en una única instancia de servicio.
// Es seguro para uso concurrente.
type StreamManager[T any] struct {
	mu      sync.RWMutex
	streams map[string]*clientStream[T]
	logger  *slog.Logger
}

func NewStreamManager[T any](logger *slog.Logger) *StreamManager[T] {
	return &StreamManager[T]{
		streams: make(map[string]*clientStream[T]),
		logger:  logger,
	}
}

// Register crea un nuevo clientStream, lo registra y lanza su goroutine de envío.
func (sm *StreamManager[T]) Register(ctx context.Context, generalParams api.GeneralParams, stream *connect.ServerStream[T]) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	cs := &clientStream[T]{
		stream:  stream,
		msgChan: make(chan *T, messageBufferSize),
		done:    make(chan struct{}),
		logger:  sm.logger,
		ctx:     ctx,
	}

	sm.streams[generalParams.ClientId] = cs
	go cs.senderLoop(generalParams.ClientId) // Lanza la goroutine de envío para este cliente.
	sm.logger.Info("Stream registered and sender loop started", "clientID", generalParams.ClientId)
}

// Unregister cierra el canal 'done' para detener la goroutine de envío y elimina el stream del mapa.
func (sm *StreamManager[T]) Unregister(generalParams api.GeneralParams) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if cs, ok := sm.streams[generalParams.ClientId]; ok {
		delete(sm.streams, generalParams.ClientId)
		close(cs.done) // Señala a la senderLoop que debe terminar.
		sm.logger.Info("Stream unregistered", "clientID", generalParams.ClientId)
	}
}

// Send envía un evento al canal de un usuario de forma bloqueante hasta que el mensaje pueda ser encolado o el cliente se desconecte.
func (sm *StreamManager[T]) Send(generalParams api.GeneralParams, event *T) {
	sm.mu.RLock()
	cs, ok := sm.streams[generalParams.ClientId]
	sm.mu.RUnlock()

	if !ok {
		// El usuario no está conectado a esta instancia.
		return
	}

	select {
	case cs.msgChan <- event:
		// Evento encolado exitosamente.
	case <-cs.done:
		// El cliente se desconectó mientras esperábamos enviar el mensaje.
		sm.logger.Info("Client disconnected while waiting to send event", "clientID", generalParams.ClientId)
	}
}
