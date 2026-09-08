package config

import "testing"

func TestViewFromYAML_DoesNotMutateSingleton(t *testing.T) {
	if err := MergeYAML([]byte(`
redis:
  pool:
    host: base-host
    size: 10
grpc:
  clientAddresses:
    directory: http://base:8080
`)); err != nil {
		t.Fatal(err)
	}

	view, err := ViewFromYAML([]byte(`
redis:
  pool:
    host: overlay-host
grpc:
  clientAddresses:
    directory: http://overlay:8080
`))
	if err != nil {
		t.Fatal(err)
	}

	if got := view.GetString("redis.pool.host"); got != "overlay-host" {
		t.Fatalf("view host=%q", got)
	}
	if got := view.GetInt("redis.pool.size"); got != 10 {
		t.Fatalf("view size=%d", got)
	}
	if got := GetString("redis.pool.host"); got != "base-host" {
		t.Fatalf("singleton mutated: host=%q", got)
	}
	if got := GetString("grpc.clientAddresses.directory"); got != "http://base:8080" {
		t.Fatalf("singleton mutated: directory=%q", got)
	}
}

func TestGetStringCtx_UsesView(t *testing.T) {
	_ = MergeYAML([]byte("redis:\n  pool:\n    host: base\n"))
	view, err := ViewFromYAML([]byte("redis:\n  pool:\n    host: from-ctx\n"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := WithView(t.Context(), view)
	if got := GetStringCtx(ctx, "redis.pool.host"); got != "from-ctx" {
		t.Fatalf("got %q", got)
	}
	if got := GetStringCtx(t.Context(), "redis.pool.host"); got != "base" {
		t.Fatalf("fallback got %q", got)
	}
}
