package redis

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestRedisRoundTrip(t *testing.T) {
	address := os.Getenv("REDIS_INTEGRATION_ADDR")
	if address == "" {
		t.Skip("REDIS_INTEGRATION_ADDR is not set")
	}
	client, err := Open(address, "tamo:test:media", time.Second, 1)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	ctx := context.Background()
	if err = client.Ping(ctx); err != nil {
		t.Fatal(err)
	}
	if err = client.Set(ctx, "foundation", []byte("ready"), time.Minute); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Delete(context.Background(), "foundation") })
	value, err := client.Get(ctx, "foundation")
	if err != nil {
		t.Fatal(err)
	}
	if string(value) != "ready" {
		t.Fatalf("unexpected value %q", value)
	}
}
