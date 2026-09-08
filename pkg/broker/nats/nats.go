package natsmanager

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

var (
	// defaultManager es la instancia única (Singleton) de nuestro gestor.
	defaultManager *Manager
	// once se asegura de que la inicialización ocurra solo una vez.
	once sync.Once
	// errInit guarda un posible error de inicialización para devolverlo en Get().
	errInit error
)

// Config contiene todos los parámetros para establecer la conexión a NATS.
type Config struct {
	URLs              []string      // Lista de URLs de servidores NATS.
	User              string        // Usuario para la autenticación.
	Password          string        // Contraseña para la autenticación.
	Token             string        // Token para la autenticación.
	JSDomain          string        // Optional JetStream domain (leaf edge vs hub).
	TLSConfig         *tls.Config   // Configuración de TLS para conexiones seguras.
	ReconnectAttempts int           // Cuántas veces intentar la reconexión (-1 para siempre).
	ReconnectWait     time.Duration // Tiempo de espera entre intentos de reconexión.
	Logger            *slog.Logger  // Logger para registrar eventos de la conexión.
}

// Manager encapsula la conexión a NATS y su configuración.
type Manager struct {
	conn      *nats.Conn
	config    Config
	js        jetstream.JetStream
	consumers map[string]jetstream.ConsumeContext
}

// Connect inicializa el gestor de NATS como un Singleton.
// Esta función debe ser llamada una sola vez al inicio de la aplicación.
// Es seguro llamar a Connect desde múltiples gorutinas.
func Connect(config Config) {
	once.Do(func() {
		// Asignar valores por defecto si no se proporcionan
		if len(config.URLs) == 0 {
			config.URLs = []string{nats.DefaultURL}
		}
		if config.ReconnectAttempts == 0 {
			config.ReconnectAttempts = -1 // Reintentar para siempre
		}
		if config.ReconnectWait == 0 {
			config.ReconnectWait = 5 * time.Second // Espera razonable
		}
		if config.Logger == nil {
			config.Logger = slog.Default()
		}

		// Crear las opciones de conexión
		opts := []nats.Option{
			nats.Name("laperla"), // Nombre de cliente útil para debugging
			nats.MaxReconnects(config.ReconnectAttempts),
			nats.ReconnectWait(config.ReconnectWait),
			nats.Timeout(5 * time.Second), // Timeout para la conexión inicial

			// Handlers para monitorear el estado de la conexión
			nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
				config.Logger.Error("NATS desconectado", "error", err)
			}),
			nats.ReconnectHandler(func(nc *nats.Conn) {
				config.Logger.Info("NATS reconectado", "url", nc.ConnectedUrl())
			}),
			nats.ClosedHandler(func(nc *nats.Conn) {
				config.Logger.Warn("Conexión a NATS cerrada permanentemente")
			}),
		}

		// Añadir credenciales si se proporcionan
		if config.User != "" && config.Password != "" {
			opts = append(opts, nats.UserInfo(config.User, config.Password))
		}
		if config.Token != "" {
			opts = append(opts, nats.Token(config.Token))
		}
		if config.TLSConfig != nil {
			opts = append(opts, nats.Secure(config.TLSConfig))
		}

		// Intentar conectar
		nc, err := nats.Connect(strings.Join(config.URLs, ", "), opts...)
		if err != nil {
			errInit = err // Guardar el error para devolverlo en Get()
			return
		}

		defaultManager = &Manager{
			conn:      nc,
			config:    config,
			consumers: make(map[string]jetstream.ConsumeContext),
		}
		var js jetstream.JetStream
		var jsErr error
		if domain := strings.TrimSpace(config.JSDomain); domain != "" {
			js, jsErr = jetstream.NewWithDomain(nc, domain)
			config.Logger.Info("JetStream domain selected", "domain", domain)
		} else {
			js, jsErr = jetstream.New(nc)
		}
		if jsErr != nil {
			config.Logger.Error("Failed to create JetStream context", "error", jsErr)
		} else {
			defaultManager.js = js
		}

		config.Logger.Info("Conectado a NATS exitosamente", "url", nc.ConnectedUrl())
	})
}

// Get devuelve la instancia Singleton del gestor.
// Devuelve un error si la conexión no se ha establecido.
func Get() (*Manager, error) {
	if defaultManager == nil {
		if errInit != nil {
			return nil, errInit // La inicialización falló
		}
		return nil, errors.New("el gestor de NATS no ha sido inicializado; llama a Connect() primero")
	}
	return defaultManager, nil
}

// GetConn devuelve el objeto de conexión nats.Conn subyacente.
// Útil para cuando necesitas usar una función del cliente de NATS que no está encapsulada.
func (m *Manager) GetConn() *nats.Conn {
	return m.conn
}

func (m *Manager) GetJS() jetstream.JetStream {
	return m.js
}

// Close realiza un cierre "graceful" de la conexión.
// Intenta enviar todos los mensajes en búfer antes de cerrar.
func (m *Manager) Close() {
	if m.conn != nil && !m.conn.IsClosed() {
		m.config.Logger.Info("Drenando y cerrando la conexión a NATS...")
		for _, consumer := range m.consumers {
			consumer.Drain()
		}
		if err := m.conn.Drain(); err != nil {
			m.config.Logger.Error("Error al drenar la conexión a NATS", "error", err)
		} else {
			m.config.Logger.Info("Conexión a NATS cerrada correctamente")
		}
	}
}

func (m *Manager) AddWorker(subject, streamName string, callback func(msg jetstream.Msg) bool, consumerConfig jetstream.ConsumerConfig) (jetstream.ConsumeContext, error) {
	consumerConfig.FilterSubject = subject

	// Create or update the consumer
	cons, err := m.js.CreateOrUpdateConsumer(context.Background(), streamName, consumerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create or update consumer for subject %s: %w", subject, err)
	}

	consumeCtx, err := cons.Consume(func(msg jetstream.Msg) {
		if callback != nil {
			if !callback(msg) {
				// If callback returns false, it means the message was not successfully delivered to the client stream.
				// Nak the message to request redelivery.
				msg.Nak()
				return
			}
		}
		// If no callback, just acknowledge
		msg.Ack()
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start consuming messages for subject %s: %w", subject, err)
	}

	m.consumers[subject] = consumeCtx

	return consumeCtx, nil
}

// AddFanOutTaskWorker crea un consumidor duradero con un Queue Group, asegurando que:
// 1. Solo UNA réplica del microservicio procese el mensaje (Fan-Out con exclusión mutua).
// 2. El mensaje se descarte después de la primera ejecución exitosa o después del primer reintento fallido.
// Se usa MaxDeliver: 2 para permitir un reintento en caso de fallo de una réplica (resiliencia).
func (m *Manager) AddFanOutTaskWorker(
	subject, streamName, durableName, queueGroupName string,
	callback func(msg jetstream.Msg) error,
	numWorkers int, // New parameter for concurrency
) (jetstream.ConsumeContext, error) {
	// 1. Configuración del Consumidor para Tarea Única (Task)
	consumerConfig := jetstream.ConsumerConfig{
		Durable:       durableName,                 // OBLIGATORIO: Define el estado compartido para la tarea.
		AckPolicy:     jetstream.AckExplicitPolicy, // OBLIGATORIO: Requiere Ack explícito para avanzar el puntero.
		DeliverPolicy: jetstream.DeliverNewPolicy,  // Empezar con mensajes nuevos.
		FilterSubject: subject,
		MaxDeliver:    2, // Permite 1 reintento si la primera réplica muere antes del Ack.

		// Usar DeliverGroup para definir el Queue Group.
		// Esto activa la distribución de carga entre las réplicas que usen este durable.
		DeliverGroup: queueGroupName,
	}

	// 2. Crear o actualizar el consumidor duradero (obtenemos la interfaz Consumer)
	// El DeliverGroup se configura en este paso.
	cons, err := m.js.CreateOrUpdateConsumer(context.Background(), streamName, consumerConfig)
	if err != nil {
		return nil, fmt.Errorf("fallo al crear o actualizar el consumidor duradero %s: %w", durableName, err)
	}

	// Create a channel to buffer incoming NATS messages
	// Default buffer size to 100, can be made configurable if needed
	msgChan := make(chan jetstream.Msg, 100)

	// Start worker goroutines
	if numWorkers <= 0 {
		numWorkers = 5 // Default to 5 workers if not specified or invalid
	}
	for i := 0; i < numWorkers; i++ {
		go func() {
			for msg := range msgChan {
				ack := true
				if err := callback(msg); err != nil {
					ack = false
					m.config.Logger.Error("Error processing NATS message in worker", "error", err, "subject", msg.Subject())
				}

				if ack {
					msg.Ack()
				} else {
					// If processing failed, NATS will redeliver based on MaxDeliver
					// No explicit Nak() here, as not Acknowledging will trigger redelivery
				}
			}
		}()
	}

	// 3. Suscribirse como Push Consumer. No se necesita una opción de Queue Group adicional.
	consumeCtx, err := cons.Consume(func(msg jetstream.Msg) {
		select {
		case msgChan <- msg:
			// Message successfully sent to worker pool
		case <-time.After(5 * time.Second): // Timeout if channel is full
			m.config.Logger.Warn("Failed to push NATS message to worker channel, channel full. Message will be redelivered.", "subject", msg.Subject())
			// Do not Ack or Nak here, let NATS redeliver after timeout
		}
	})
	if err != nil {
		return nil, fmt.Errorf("fallo al iniciar la suscripción Push con queue group %s: %w", queueGroupName, err)
	}

	// 4. Almacenar el contexto de consumo
	m.consumers[durableName] = consumeCtx

	return consumeCtx, nil
}

// PublishFanOutTask publica un mensaje en el sujeto dado utilizando JetStream.
// Esto garantiza que el mensaje se persista en el Stream (durabilidad) antes de
// ser entregado a una única réplica del grupo de consumidores.
func (m *Manager) PublishFanOutTask(subject string, data []byte) error {
	if m.js == nil {
		return errors.New("el contexto de JetStream no está inicializado")
	}

	// Publicar en el sujeto asociado con el Worker de Tarea Fan-Out
	_, err := m.js.Publish(context.Background(), subject, data)
	if err != nil {
		return fmt.Errorf("fallo al publicar el mensaje en el sujeto %s: %w", subject, err)
	}

	m.config.Logger.Info("Mensaje de tarea Fan-Out publicado con éxito", "subject", subject)
	return nil
}
