package config

import "testing"

func TestMergeYAMLOverridesBase(t *testing.T) {
	if err := MergeYAML([]byte(`
server:
  address: "0.0.0.0:8080"
redis:
  pool:
    size: 10
    host: base-host
`)); err != nil {
		t.Fatal(err)
	}

	if err := MergeYAML([]byte(`
redis:
  pool:
    host: overlay-host
  timeouts:
    dialMs: 500
grpc:
  clientAddresses:
    directory: http://directory:8080
`)); err != nil {
		t.Fatal(err)
	}

	if got := GetString("server.address"); got != "0.0.0.0:8080" {
		t.Fatalf("server.address: %q", got)
	}
	if got := GetString("redis.pool.host"); got != "overlay-host" {
		t.Fatalf("host: %q", got)
	}
	if got := GetInt("redis.pool.size"); got != 10 {
		t.Fatalf("size: %d", got)
	}
	if got := GetInt("redis.timeouts.dialMs"); got != 500 {
		t.Fatalf("dialMs: %d", got)
	}
	if got := GetString("grpc.clientAddresses.directory"); got != "http://directory:8080" {
		t.Fatalf("directory: %q", got)
	}
}
