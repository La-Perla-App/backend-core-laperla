package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache define la interfaz para un almacén de caché.
type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value any, expiration time.Duration) error
	SetNX(ctx context.Context, key string, value any, expiration time.Duration) (bool, error)
	HashSet(ctx context.Context, key, field string, value any, expiration time.Duration) error
	HashGet(ctx context.Context, key, field string) (string, error)
	Del(ctx context.Context, keys ...string) error
	Incrby(ctx context.Context, key string, increment int64) (int64, error)
	Decrby(ctx context.Context, key string, decrement int64) (int64, error)
	Keys(ctx context.Context, pattern string) ([]string, error)
	Ping(ctx context.Context) error
	Close() error
	SAdd(ctx context.Context, key string, members ...any) error
	SRem(ctx context.Context, key string, members ...any) error
	SMembers(ctx context.Context, key string) ([]string, error)
	Eval(ctx context.Context, script string, keys []string, args ...any) (any, error) // New method

	// New methods for batch operations and pipelining
	MGet(ctx context.Context, keys ...string) ([]any, error)
	MSet(ctx context.Context, pairs ...any) error
	NewPipeline() Pipeline
}

// Pipeline define la interfaz para operaciones pipelined en Redis
type Pipeline interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Exec(ctx context.Context) ([]redis.Cmder, error)
}
