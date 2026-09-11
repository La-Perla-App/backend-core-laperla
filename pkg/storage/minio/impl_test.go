package minio

import (
	"strings"
	"testing"
)

func TestObjectURL_S3VirtualHost(t *testing.T) {
	c, err := NewClient(Config{
		Provider:  ProviderS3,
		Endpoint:  "s3.us-east-1.amazonaws.com",
		Region:    "us-east-1",
		AccessKey: "AKIA",
		SecretKey: "secret",
		UseSSL:    true,
		DefaultBucket: "laperla-dev",
	})
	if err != nil {
		t.Fatal(err)
	}
	u := c.ObjectURL("laperla-dev", "photos/a.jpg")
	if !strings.Contains(u, "laperla-dev.s3") && !strings.HasPrefix(u, "https://laperla-dev.") {
		t.Fatalf("expected virtual-hosted URL, got %s", u)
	}
	if !strings.HasSuffix(u, "/photos/a.jpg") {
		t.Fatalf("key path: %s", u)
	}
}

func TestObjectURL_MinIOPathStyle(t *testing.T) {
	c, err := NewClient(Config{
		Provider:  ProviderMinIO,
		Endpoint:  "minio.minio-system:9000",
		AccessKey: "minio",
		SecretKey: "minio123",
		UseSSL:    false,
		DefaultBucket: "laperla-dev",
	})
	if err != nil {
		t.Fatal(err)
	}
	u := c.ObjectURL("", "x.png")
	if u != "http://minio.minio-system:9000/laperla-dev/x.png" {
		t.Fatalf("got %s", u)
	}
}

func TestNewClient_PublicHostSeparate(t *testing.T) {
	c, err := NewClient(Config{
		Provider:   ProviderMinIO,
		Endpoint:   "minio.minio-system:9000",
		PublicHost: "files-dev.laperlaapp.biz",
		PublicSSL:  true,
		AccessKey:  "minio",
		SecretKey:  "minio123",
		UseSSL:     false,
	})
	if err != nil {
		t.Fatal(err)
	}
	if c.ops == c.sign {
		t.Fatal("expected distinct ops and sign clients")
	}
	u := c.ObjectURL("b", "k")
	if !strings.Contains(u, "files-dev.laperlaapp.biz") {
		t.Fatalf("object URL should use public host: %s", u)
	}
}
