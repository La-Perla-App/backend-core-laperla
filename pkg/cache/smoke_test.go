//go:build smoke

package cache

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestSmokeDragonflyClusterSafeOps(t *testing.T) {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "dragonfly.laperla-develop-dragonfly.svc.cluster.local:6379"
	}
	c, err := NewRedisCache(Config{Addrs: []string{addr}})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	ctx := context.Background()

	tag := "smoke-core"
	a := ClusterKey(tag, "sess", "a")
	b := ClusterKey(tag, "sess", "index")
	if !SameSlot(a, b) {
		t.Fatal("hash tags must colocate")
	}
	if err := c.Set(ctx, a, "1", time.Minute); err != nil {
		t.Fatal(err)
	}
	if err := c.SAdd(ctx, b, "a"); err != nil {
		t.Fatal(err)
	}
	if err := c.Expire(ctx, b, time.Minute); err != nil {
		t.Fatal(err)
	}
	n, err := c.IncrExpire(ctx, ClusterKey(tag, "rl"), time.Minute)
	if err != nil || n < 1 {
		t.Fatalf("incr %d %v", n, err)
	}
	if err := c.Del(ctx, a, b, ClusterKey(tag, "rl")); err != nil {
		t.Fatal(err)
	}
}
