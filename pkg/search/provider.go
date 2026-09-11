package search

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

var (
	client Searcher
	once   sync.Once

	ErrProviderNotInitialized = errors.New("opensearch provider no inicializado")
)

// Init initializes the process-wide OpenSearch client (same pattern as cache / nats).
func Init(cfg Config) (err error) {
	once.Do(func() {
		var c Searcher
		c, err = New(cfg)
		if err != nil {
			err = fmt.Errorf("fallo al inicializar OpenSearch: %w", err)
			return
		}
		client = c
	})
	return err
}

// Get returns the singleton client, or ErrProviderNotInitialized.
func Get() (Searcher, error) {
	if client == nil {
		return nil, ErrProviderNotInitialized
	}
	return client, nil
}

// Ready reports whether Init succeeded.
func Ready() bool {
	return client != nil
}

func Ping(ctx context.Context) error {
	c, err := Get()
	if err != nil {
		return err
	}
	return c.Ping(ctx)
}

func EnsureIndex(ctx context.Context, index string, mapping []byte) error {
	c, err := Get()
	if err != nil {
		return err
	}
	return c.EnsureIndex(ctx, index, mapping)
}

func Index(ctx context.Context, index, id string, doc any) error {
	c, err := Get()
	if err != nil {
		return err
	}
	return c.Index(ctx, index, id, doc)
}

func Delete(ctx context.Context, index, id string) error {
	c, err := Get()
	if err != nil {
		return err
	}
	return c.Delete(ctx, index, id)
}

func Search(ctx context.Context, index string, req Query) (*Result, error) {
	c, err := Get()
	if err != nil {
		return nil, err
	}
	return c.Search(ctx, index, req)
}

func Close() error {
	if client == nil {
		return nil
	}
	return client.Close()
}
