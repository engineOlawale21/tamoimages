package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
)

func TestMinIORoundTrip(t *testing.T) {
	endpoint := os.Getenv("S3_INTEGRATION_ENDPOINT")
	if endpoint == "" {
		t.Skip("S3_INTEGRATION_ENDPOINT is not set")
	}
	bucket := fmt.Sprintf("tamo-media-test-%d", time.Now().UnixNano())
	store, err := OpenMinIO(endpoint, bucket, os.Getenv("S3_INTEGRATION_ACCESS_KEY"), os.Getenv("S3_INTEGRATION_SECRET_KEY"))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err = store.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.client.RemoveBucket(context.Background(), bucket) })
	if err = store.EnsureBucket(ctx); err != nil {
		t.Fatal(err)
	}
	key := "integration/foundation.txt"
	t.Cleanup(func() { _ = store.Delete(context.Background(), key) })
	if _, err = store.Put(ctx, key, bytes.NewReader([]byte("ready")), 5, "text/plain"); err != nil {
		t.Fatal(err)
	}
	info, err := store.Head(ctx, key)
	if err != nil || info.Size != 5 {
		t.Fatalf("unexpected object info %+v: %v", info, err)
	}
	reader, err := store.Get(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	value, err := io.ReadAll(reader)
	if err != nil || string(value) != "ready" {
		t.Fatalf("unexpected object %q: %v", value, err)
	}
	if _, err = store.PresignGet(ctx, key, time.Minute); err != nil {
		t.Fatal(err)
	}
	if _, err = store.PresignPut(ctx, key, time.Minute); err != nil {
		t.Fatal(err)
	}
}
