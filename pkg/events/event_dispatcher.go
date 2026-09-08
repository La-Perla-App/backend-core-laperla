package events

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

const (
	maxRetries             = 3
	retryDelay             = 500 * time.Millisecond
	eventChannelBufferSize = 1000
	dispatchTimeout        = 100 * time.Millisecond
)

// NatsEvent represents an event that should be published to NATS.
type NatsEvent interface {
	Event
	Subjects() []string
	Payload() ([]byte, error)
	JetStream() bool
}

type FanoutEvent struct {
	Data      any
	OnFanount func(context.Context, FanoutEvent)
}

func (e FanoutEvent) EventType() string {
	return "FanoutEvent"
}

type Event interface {
	EventType() string
}

type eventContext struct {
	ctx   context.Context
	event Event
}

// EventDispatcher handles asynchronous processing of events.
type EventDispatcher struct {
	logger       *slog.Logger
	nc           *nats.Conn
	js           jetstream.JetStream
	eventChannel chan eventContext
	done         chan bool
}

// NewEventDispatcher creates and starts a new dispatcher.
func NewEventDispatcher(nc *nats.Conn, logger *slog.Logger, numWorkers int) (*EventDispatcher, error) {
	js, err := jetstream.New(nc)
	if err != nil {
		return nil, fmt.Errorf("could not get JetStream context: %w", err)
	}

	dispatcher := &EventDispatcher{
		logger:       logger,
		nc:           nc,
		js:           js,
		eventChannel: make(chan eventContext, eventChannelBufferSize),
		done:         make(chan bool),
	}

	dispatcher.startWorkers(numWorkers)
	logger.Info("Event dispatcher started", "numWorkers", numWorkers)
	return dispatcher, nil
}

// Dispatch sends an event to the dispatcher's channel for asynchronous processing.
func (d *EventDispatcher) Dispatch(ctx context.Context, event Event) error {
	currentLen := len(d.eventChannel)
	if currentLen > (eventChannelBufferSize / 2) {
		d.logger.Info("Event channel is over 50% full", "currentLength", currentLen, "capacity", eventChannelBufferSize)
	}

	select {
	case d.eventChannel <- eventContext{event: event, ctx: ctx}:
		// Event successfully queued
	case <-time.After(dispatchTimeout):
		err := errors.New("Event dispatcher channel full, event dropped after timeout")
		d.logger.Error(
			err.Error(),
			"eventType", event.EventType(),
			"timeout", dispatchTimeout,
		)
	}

	return nil
}

func (d *EventDispatcher) startWorkers(numWorkers int) {
	for i := range numWorkers {
		go d.worker(i + 1)
	}
}

func (d *EventDispatcher) worker(id int) {
	d.logger.Info("Worker started", "id", id)
	for {
		select {
		case v := <-d.eventChannel:
			switch e := v.event.(type) {
			case NatsEvent:
				d.publishWithRetry(v.ctx, e)
			case FanoutEvent:
				d.handleFanoutEvent(v.ctx, e)
			default:
				d.logger.Warn("Unknown event type in dispatcher", "eventType", v.event.EventType())
			}
		case <-d.done:
			d.logger.Info("Worker shutting down", "id", id)
			return
		}
	}
}

func (d *EventDispatcher) handleFanoutEvent(ctx context.Context, event FanoutEvent) {
	d.logger.Info("Handling fanout event...")

	if event.OnFanount != nil {
		event.OnFanount(ctx, event)
	}
}

func (d *EventDispatcher) publishWithRetry(ctx context.Context, event NatsEvent) {
	payload, err := event.Payload()
	if err != nil {
		d.logger.Error("Failed to serialize event payload", "error", err, "eventType", event.EventType())
		return
	}

	subjects := slices.Compact(event.Subjects())

outer:
	for _, subject := range subjects {
		d.logger.Info(
			"Trying to publish event",
			"subject", subject, "eventType", event.EventType(),
		)
		for i := range maxRetries {
			msg := &nats.Msg{
				Subject: subject,
				Data:    payload,
			}

			if event.JetStream() {
				_, err = d.js.PublishMsg(ctx, msg)
			} else {
				err = d.nc.PublishMsg(msg)
			}

			if err == nil {
				d.logger.Info("Event published successfully", "subject", subject, "eventType", event.EventType())
				continue outer
			}

			d.logger.Warn(
				"Failed to publish event, retrying...",
				"error", err,
				"subject", subject,
				"retry", i+1,
				"maxRetries", maxRetries,
			)

			if i < maxRetries-1 {
				time.Sleep(retryDelay)
			}

			d.logger.Error(
				"Failed to publish event after multiple retries, event dropped",
				"subject", subject,
				"eventType", event.EventType(),
			)
		}
	}
}

func (d *EventDispatcher) Shutdown() {
	close(d.done)
	close(d.eventChannel)
	d.logger.Info("Event dispatcher shut down.")
}
