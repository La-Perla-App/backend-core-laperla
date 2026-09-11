package minio

import (
	"context"
	"testing"

	"github.com/La-Perla-App/backend-core-laperla/pkg/config"
)

func TestConfigFromContext_UsesViewOverlay(t *testing.T) {
	if err := config.ReplaceYAML([]byte(`
minio:
  endpoint: minio.minio-system:9000
  accessKey: base-key
  secretKey: base-secret
  useSsl: false
  publicHost: s3.example.com
  publicUseSsl: true
  defaultBucket: laperla-dev
`)); err != nil {
		t.Fatal(err)
	}

	view, err := config.ViewFromYAML([]byte(`
minio:
  endpoint: 100.64.180.61:30900
  accessKey: overlay-key
  secretKey: overlay-secret
  useSsl: false
`))
	if err != nil {
		t.Fatal(err)
	}
	ctx := config.WithView(context.Background(), view)

	cfg := ConfigFromContext(ctx)
	if cfg.Endpoint != "100.64.180.61:30900" {
		t.Fatalf("endpoint: got %q", cfg.Endpoint)
	}
	if cfg.AccessKey != "overlay-key" || cfg.SecretKey != "overlay-secret" {
		t.Fatalf("credentials: %+v", cfg)
	}
	if cfg.PublicHost != "s3.example.com" {
		t.Fatalf("publicHost should fall through from base, got %q", cfg.PublicHost)
	}
	if !cfg.PublicSSL {
		t.Fatal("publicUseSsl should fall through from base")
	}
	if cfg.DefaultBucket != "laperla-dev" {
		t.Fatalf("defaultBucket: got %q", cfg.DefaultBucket)
	}
}

func TestConfigFromContext_StorageAliasAndS3(t *testing.T) {
	if err := config.ReplaceYAML([]byte(`
storage:
  provider: s3
  region: us-east-1
  accessKey: AKIATEST
  secretKey: secret
  useSsl: true
  defaultBucket: laperla-dev
`)); err != nil {
		t.Fatal(err)
	}
	cfg := normalizeConfig(ConfigFromContext(context.Background()))
	if resolveProvider(cfg) != ProviderS3 {
		t.Fatalf("provider: %+v", cfg)
	}
	if cfg.Endpoint == "" {
		t.Fatal("expected default S3 endpoint")
	}
	if cfg.DefaultBucket != "laperla-dev" {
		t.Fatalf("bucket %q", cfg.DefaultBucket)
	}
}

func TestConfigFromContext_FallsBackToProcessConfig(t *testing.T) {
	if err := config.ReplaceYAML([]byte(`
minio:
  endpoint: minio.minio-system:9000
  accessKey: base-key
  secretKey: base-secret
  useSsl: true
`)); err != nil {
		t.Fatal(err)
	}

	cfg := ConfigFromContext(context.Background())
	if cfg.Endpoint != "minio.minio-system:9000" {
		t.Fatalf("endpoint: got %q", cfg.Endpoint)
	}
	if !cfg.UseSSL {
		t.Fatal("useSsl expected true")
	}
}

func TestResolveProvider(t *testing.T) {
	if got := resolveProvider(Config{Endpoint: "s3.us-east-1.amazonaws.com"}); got != ProviderS3 {
		t.Fatalf("got %s", got)
	}
	if got := resolveProvider(Config{Endpoint: "minio.minio-system:9000"}); got != ProviderMinIO {
		t.Fatalf("got %s", got)
	}
	if got := resolveProvider(Config{Provider: "s3", Endpoint: "custom.example"}); got != ProviderS3 {
		t.Fatalf("got %s", got)
	}
}
