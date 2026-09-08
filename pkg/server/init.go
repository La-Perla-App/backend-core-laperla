package server

import (
	"context"
	"log"
	"log/slog"
	"os"
	"strconv"

	natsmanager "github.com/La-Perla-App/backend-core-laperla/pkg/broker/nats"
	"github.com/La-Perla-App/backend-core-laperla/pkg/cache"
	"github.com/La-Perla-App/backend-core-laperla/pkg/config"
	"github.com/La-Perla-App/backend-core-laperla/pkg/storage/minio"
)

type noopHandler struct{}

func (h *noopHandler) Enabled(_ context.Context, _ slog.Level) bool  { return false }
func (h *noopHandler) Handle(_ context.Context, _ slog.Record) error { return nil }
func (h *noopHandler) WithAttrs(_ []slog.Attr) slog.Handler          { return h }
func (h *noopHandler) WithGroup(_ string) slog.Handler               { return h }

func init() {
	handler := slog.NewJSONHandler(os.Stdout, nil)
	if nologs, _ := strconv.ParseBool(os.Getenv("NO_LOGS")); nologs {
		slog.SetDefault(slog.New(&noopHandler{}))
	} else {
		logger := slog.New(handler)
		slog.SetDefault(logger)
	}
}

func InitEnvironment() {
}

func InitRedis() {
	var addrs []string
	addrsAny := config.Get("cache.redis.addrs")
	if addrsAny, ok := addrsAny.([]any); ok {
		for _, addr := range addrsAny {
			if addr, ok := addr.(string); ok {
				addrs = append(addrs, addr)
			}
		}
	}
	if err := cache.Init(cache.Config{
		Addrs:     addrs,
		Password:  config.GetString("cache.redis.password"),
		IsCluster: config.GetBool("cache.redis.isCluster"),
	}); err != nil {
		log.Fatal("ERROR connecting to cache: ", err)
	}
}

func InitNats() {
	var addrs []string
	addrsAny := config.Get("nats.addrs")
	if addrsAny, ok := addrsAny.([]any); ok {
		for _, addr := range addrsAny {
			if addr, ok := addr.(string); ok {
				addrs = append(addrs, addr)
			}
		}
	}
	natsmanager.Connect(natsmanager.Config{
		URLs:     addrs,
		Logger:   slog.Default(),
		User:     config.GetString("nats.user"),
		Token:    config.GetString("nats.token"),
		Password: config.GetString("nats.password"),
		JSDomain: config.GetString("nats.jsDomain"),
	})
}

func InitMinio() {
	cfg := minio.ConfigFromContext(context.Background())
	if err := minio.Init(cfg); err != nil {
		log.Fatal("ERROR initializing MinIO: ", err)
	}
}

func CloseEnvironment() {
	cache.Close()
	if nats, err := natsmanager.Get(); err == nil {
		nats.Close()
	}
}
