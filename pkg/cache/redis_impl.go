package cache

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// redisCache es una implementación de la interfaz Cache que utiliza Redis.
type redisCache struct {
	client redis.Cmdable
}

// redisPipeline es una implementación de la interfaz Pipeline que utiliza el pipeline de Redis.
type redisPipeline struct {
	pipe redis.Pipeliner
}

// Get añade un comando GET al pipeline.
func (p *redisPipeline) Get(ctx context.Context, key string) *redis.StringCmd {
	return p.pipe.Get(ctx, key)
}

// Set añade un comando SET al pipeline.
func (p *redisPipeline) Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd {
	return p.pipe.Set(ctx, key, value, expiration)
}

// Del añade un comando DEL al pipeline.
func (p *redisPipeline) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	return p.pipe.Del(ctx, keys...)
}

// Exec ejecuta todos los comandos en el pipeline.
func (p *redisPipeline) Exec(ctx context.Context) ([]redis.Cmder, error) {
	return p.pipe.Exec(ctx)
}

// Config contiene la configuración para el cliente de Redis.
type Config struct {
	// IsCluster indica si la conexión es a un clúster de Redis.
	IsCluster bool
	// Addrs son las direcciones de los nodos de Redis.
	Addrs []string
	// Password es la contraseña para la autenticación en Redis.
	Password string
	// DB es el número de la base de datos a utilizar (solo para modo no clúster).
	DB int
}

// NewRedisCache crea un nuevo cliente de caché de Redis.
// Se conecta en modo cliente normal o en modo clúster según la configuración.
// La biblioteca go-redis maneja la reconexión automáticamente.
func NewRedisCache(cfg Config) (Cache, error) {
	var client redis.Cmdable

	if cfg.IsCluster {
		if len(cfg.Addrs) == 0 {
			return nil, errors.New("se requieren direcciones para el modo clúster")
		}
		clusterClient := redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:    cfg.Addrs,
			Password: cfg.Password,
			// go-redis reintentará los comandos fallidos hasta 3 veces por defecto.
			// También maneja la redirección de slots del clúster.
		})
		client = clusterClient
	} else {
		if len(cfg.Addrs) != 1 {
			return nil, errors.New("se requiere exactamente una dirección para el modo de cliente normal")
		}
		singleClient := redis.NewClient(&redis.Options{
			Addr:     cfg.Addrs[0],
			Password: cfg.Password,
			DB:       cfg.DB,
		})
		client = singleClient
	}

	// Comprueba la conexión inicial
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return &redisCache{client: client}, nil
}

// Get obtiene un valor de la caché.
func (r *redisCache) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil // La clave no existe
	}
	return val, err
}

// Set establece un valor en la caché con una expiración.
func (r *redisCache) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

// Del elimina un valor de la caché.
func (r *redisCache) Del(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

// Ping comprueba la conexión con Redis.
func (r *redisCache) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *redisCache) Incrby(ctx context.Context, key string, increment int64) (int64, error) {
	val, err := r.client.IncrBy(ctx, key, increment).Result()
	if err == redis.Nil {
		return 0, nil // La clave no existe
	}
	return val, err
}

func (r *redisCache) Decrby(ctx context.Context, key string, decrement int64) (int64, error) {
	val, err := r.client.DecrBy(ctx, key, decrement).Result()
	if err == redis.Nil {
		return 0, nil // La clave no existe
	}
	return val, err
}

func (r *redisCache) Keys(ctx context.Context, pattern string) ([]string, error) {
	return r.client.Keys(ctx, pattern).Result()
}

func (r *redisCache) SAdd(ctx context.Context, key string, members ...any) error {
	return r.client.SAdd(ctx, key, members...).Err()
}

func (r *redisCache) SRem(ctx context.Context, key string, members ...any) error {
	return r.client.SRem(ctx, key, members...).Err()
}

func (r *redisCache) SMembers(ctx context.Context, key string) ([]string, error) {
	return r.client.SMembers(ctx, key).Result()
}

// SetNX establece un valor en la caché si la clave no existe.
func (r *redisCache) SetNX(ctx context.Context, key string, value any, expiration time.Duration) (bool, error) {
	return r.client.SetNX(ctx, key, value, expiration).Result()
}

func (r *redisCache) HashSet(ctx context.Context, key, field string, value any, expiration time.Duration) error {
	if err := r.client.HSet(ctx, key, field, value).Err(); err != nil {
		return err
	}
	if expiration > 0 {
		return r.client.Expire(ctx, key, expiration).Err()
	}
	return nil
}

func (r *redisCache) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, key, expiration).Err()
}

func (r *redisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.client.TTL(ctx, key).Result()
}

func (r *redisCache) Exists(ctx context.Context, keys ...string) (int64, error) {
	return r.client.Exists(ctx, keys...).Result()
}

// IncrExpire INCR + PEXPIRE on first increment. Single key → cluster-safe.
func (r *redisCache) IncrExpire(ctx context.Context, key string, expiration time.Duration) (int64, error) {
	if expiration <= 0 {
		return r.Incrby(ctx, key, 1)
	}
	ms := expiration.Milliseconds()
	if ms < 1 {
		ms = 1
	}
	n, err := r.client.Eval(ctx, incrExpireScript, []string{key}, ms).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return n, err
}

const incrExpireScript = `
local n = redis.call('INCR', KEYS[1])
if n == 1 then
  redis.call('PEXPIRE', KEYS[1], ARGV[1])
end
return n
`

func (r *redisCache) HashGet(ctx context.Context, key, field string) (string, error) {
	return r.client.HGet(ctx, key, field).Result()
}

func (r *redisCache) Eval(ctx context.Context, script string, keys []string, args ...any) (any, error) {
	return r.client.Eval(ctx, script, keys, args...).Result()
}

// MGet obtiene múltiples valores de la caché en una sola llamada.
func (r *redisCache) MGet(ctx context.Context, keys ...string) ([]any, error) {
	return r.client.MGet(ctx, keys...).Result()
}

// MSet establece múltiples valores en la caché en una sola llamada.
func (r *redisCache) MSet(ctx context.Context, pairs ...any) error {
	return r.client.MSet(ctx, pairs...).Err()
}

// NewPipeline crea un nuevo pipeline para ejecutar múltiples comandos en un solo viaje de red.
func (r *redisCache) NewPipeline() Pipeline {
	return &redisPipeline{pipe: r.client.Pipeline()}
}

// Close cierra la conexión con Redis.
func (r *redisCache) Close() error {
	// La interfaz redis.Cmdable no tiene un método Close().
	// Necesitamos hacer una aserción de tipo para llamar al método Close() del cliente subyacente.
	switch c := r.client.(type) {
	case *redis.Client:
		return c.Close()
	case *redis.ClusterClient:
		return c.Close()
	}
	return nil
}
