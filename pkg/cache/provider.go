package cache

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	// cacheClient es la instancia única (singleton) de nuestra caché.
	// No se exporta para evitar que otros paquetes la modifiquen directamente.
	cacheClient Cache
	// once se usa para asegurar que la inicialización ocurra solo una vez.
	once sync.Once

	// ErrProviderNotInitialized se devuelve si se intenta usar la caché sin inicializarla.
	ErrProviderNotInitialized = errors.New("cache provider no inicializado")
)

// Init inicializa el provider de caché con la configuración dada.
// Utiliza sync.Once para garantizar que la inicialización sea a prueba de concurrencia
// y solo se ejecute una vez.
// Devuelve un error si la inicialización falla.
func Init(cfg Config) (err error) {
	once.Do(func() {
		var c Cache
		c, err = NewRedisCache(cfg)
		if err != nil {
			// Añadimos contexto al error para facilitar la depuración.
			err = fmt.Errorf("fallo al inicializar el cliente de caché de Redis: %w", err)
			return
		}
		cacheClient = c
	})
	return err
}

// Get es una función envoltorio que utiliza el provider para obtener un valor.
func Get(ctx context.Context, key string) (string, error) {
	if cacheClient == nil {
		return "", ErrProviderNotInitialized
	}
	return cacheClient.Get(ctx, key)
}

// Set es una función envoltorio que utiliza el provider para establecer un valor.
func Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	if cacheClient == nil {
		return ErrProviderNotInitialized
	}
	return cacheClient.Set(ctx, key, value, expiration)
}

// SetNX es una función envoltorio que utiliza el provider para establecer un valor si no existe.
func SetNX(ctx context.Context, key string, value any, expiration time.Duration) (bool, error) {
	if cacheClient == nil {
		return false, ErrProviderNotInitialized
	}
	return cacheClient.SetNX(ctx, key, value, expiration)
}

// Del es una función envoltorio que utiliza el provider para eliminar un valor.
func Del(ctx context.Context, keys ...string) error {
	if cacheClient == nil {
		return ErrProviderNotInitialized
	}
	return cacheClient.Del(ctx, keys...)
}

// Incrby
func Incrby(ctx context.Context, key string, increment int64) (int64, error) {
	if cacheClient == nil {
		return 0, ErrProviderNotInitialized
	}
	return cacheClient.Incrby(ctx, key, increment)
}

// Incrby
func Decrby(ctx context.Context, key string, decrement int64) (int64, error) {
	if cacheClient == nil {
		return 0, ErrProviderNotInitialized
	}
	return cacheClient.Decrby(ctx, key, decrement)
}

// Keys
func Keys(ctx context.Context, pattern string) ([]string, error) {
	if cacheClient == nil {
		return nil, ErrProviderNotInitialized
	}
	return cacheClient.Keys(ctx, pattern)
}

// Ping comprueba la conexión del provider.
func Ping(ctx context.Context) error {
	if cacheClient == nil {
		return ErrProviderNotInitialized
	}
	return cacheClient.Ping(ctx)
}

// Close cierra la conexión de la caché del provider.
// Es importante llamar a esta función para una finalización limpia de la aplicación.
func Close() error {
	if cacheClient == nil {
		return nil // No hay nada que cerrar
	}
	return cacheClient.Close()
}

// SAdd es una función envoltorio que utiliza el provider para agregar miembros a un set.
func SAdd(ctx context.Context, key string, members ...any) error {
	if cacheClient == nil {
		return ErrProviderNotInitialized
	}
	return cacheClient.SAdd(ctx, key, members...)
}

// SRem es una función envoltorio que utiliza el provider para eliminar miembros de un set.
func SRem(ctx context.Context, key string, members ...any) error {
	if cacheClient == nil {
		return ErrProviderNotInitialized
	}
	return cacheClient.SRem(ctx, key, members...)
}

// SMembers es una función envoltorio que utiliza el provider para obtener miembros de un set.
func SMembers(ctx context.Context, key string) ([]string, error) {
	if cacheClient == nil {
		return nil, ErrProviderNotInitialized
	}
	return cacheClient.SMembers(ctx, key)
}

func HashSet(ctx context.Context, key, field string, value any, expiration time.Duration) error {
	if cacheClient == nil {
		return ErrProviderNotInitialized
	}
	return cacheClient.HashSet(ctx, key, field, value, expiration)
}

func HashGet(ctx context.Context, key, field string) (string, error) {
	if cacheClient == nil {
		return "", ErrProviderNotInitialized
	}
	return cacheClient.HashGet(ctx, key, field)
}

// Eval ejecuta un script Lua en el servidor Redis.
func Eval(ctx context.Context, script string, keys []string, args ...any) (any, error) {
	if cacheClient == nil {
		return nil, ErrProviderNotInitialized
	}
	return cacheClient.Eval(ctx, script, keys, args...)
}

// MGet obtiene múltiples valores de la caché en una sola llamada.
func MGet(ctx context.Context, keys ...string) ([]any, error) {
	if cacheClient == nil {
		return nil, ErrProviderNotInitialized
	}
	return cacheClient.MGet(ctx, keys...)
}

// MSet establece múltiples valores en la caché en una sola llamada.
func MSet(ctx context.Context, pairs ...any) error {
	if cacheClient == nil {
		return ErrProviderNotInitialized
	}
	return cacheClient.MSet(ctx, pairs...)
}

// NewPipeline crea un nuevo pipeline para ejecutar múltiples comandos en un solo viaje de red.
func NewPipeline() Pipeline {
	if cacheClient == nil {
		return nil
	}
	return cacheClient.NewPipeline()
}
