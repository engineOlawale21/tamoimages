package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/twmb/franz-go/pkg/kerr"
)

func TestPublishConsumeRoundTrip(t *testing.T) {
	broker := os.Getenv("KAFKA_INTEGRATION_BROKER")
	if broker == "" {
		t.Skip("KAFKA_INTEGRATION_BROKER is not set")
	}
	topic := "platform.smoke-test.v1"
	group := fmt.Sprintf("media-foundation-%d", time.Now().UnixNano())
	consumer, err := Open([]string{broker}, "media-integration-consumer", group, topic)
	if err != nil {
		t.Fatal(err)
	}
	defer consumer.Close()
	producer, err := Open([]string{broker}, "media-integration-producer", "")
	if err != nil {
		t.Fatal(err)
	}
	defer producer.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	received := make(chan Event, 1)
	consumeDone := make(chan error, 1)
	go func() {
		consumeDone <- consumer.Consume(ctx, func(_ context.Context, event Event) error {
			if event.ID == group {
				received <- event
				return ErrStopConsumer
			}
			return nil
		})
	}()
	event := Event{ID: group, Type: "platform.smoke-test", Version: 1, OccurredAt: time.Now().UTC(), CorrelationID: group, Producer: "tamo-media-service", Payload: json.RawMessage(`{"status":"ready"}`)}
	if err = producer.Publish(ctx, topic, event); err != nil {
		t.Fatal(err)
	}
	select {
	case actual := <-received:
		if actual.CorrelationID != event.CorrelationID {
			t.Fatalf("unexpected correlation ID %q", actual.CorrelationID)
		}
	case <-time.After(20 * time.Second):
		t.Fatal("timed out waiting for kafka event")
	}
	if err = <-consumeDone; err != nil {
		t.Fatal(err)
	}
}

func TestDeadLetterTopicConvention(t *testing.T) {
	if actual := DeadLetterTopic("media.processing-failed.v1"); actual != "media.processing-failed.v1.dlq" {
		t.Fatalf("unexpected dead-letter topic %q", actual)
	}
}

func TestRetryClassification(t *testing.T) {
	if !Retryable(kerr.RequestTimedOut) {
		t.Fatal("expected request timeout to be retryable")
	}
	if Retryable(context.Canceled) {
		t.Fatal("context cancellation must not be retried")
	}
}
