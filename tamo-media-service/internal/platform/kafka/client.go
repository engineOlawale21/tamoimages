package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/twmb/franz-go/pkg/kerr"
	"github.com/twmb/franz-go/pkg/kgo"
)

const DeadLetterSuffix = ".dlq"

var ErrStopConsumer = errors.New("stop kafka consumer")

type Event struct {
	ID            string          `json:"id"`
	Type          string          `json:"type"`
	Version       int             `json:"version"`
	OccurredAt    time.Time       `json:"occurredAt"`
	CorrelationID string          `json:"correlationId"`
	Producer      string          `json:"producer"`
	Payload       json.RawMessage `json:"payload"`
}

type Handler func(context.Context, Event) error

type Client struct{ client *kgo.Client }

func Open(brokers []string, clientID, consumerGroup string, topics ...string) (*Client, error) {
	if len(brokers) == 0 {
		return nil, fmt.Errorf("at least one kafka broker is required")
	}
	options := []kgo.Opt{kgo.SeedBrokers(brokers...), kgo.ClientID(clientID)}
	if consumerGroup != "" {
		options = append(options, kgo.ConsumerGroup(consumerGroup), kgo.ConsumeTopics(topics...), kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()), kgo.DisableAutoCommit())
	}
	client, err := kgo.NewClient(options...)
	if err != nil {
		return nil, fmt.Errorf("create kafka client: %w", err)
	}
	return &Client{client: client}, nil
}

func (c *Client) Close() { c.client.Close() }

func (c *Client) Ping(ctx context.Context) error {
	if err := c.client.Ping(ctx); err != nil {
		return fmt.Errorf("ping kafka: %w", err)
	}
	return nil
}

func (c *Client) Publish(ctx context.Context, topic string, event Event) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("encode kafka event: %w", err)
	}
	record := &kgo.Record{Topic: topic, Key: []byte(event.ID), Value: payload, Headers: []kgo.RecordHeader{
		{Key: "correlation-id", Value: []byte(event.CorrelationID)},
		{Key: "event-id", Value: []byte(event.ID)},
	}}
	if err = c.client.ProduceSync(ctx, record).FirstErr(); err != nil {
		return fmt.Errorf("publish kafka event: %w", err)
	}
	return nil
}

func (c *Client) Consume(ctx context.Context, handler Handler) error {
	for {
		fetches := c.client.PollFetches(ctx)
		if ctx.Err() != nil {
			return nil
		}
		if err := fetches.Err(); err != nil {
			return fmt.Errorf("poll kafka: %w", err)
		}
		var handlerErr error
		stopAfterCommit := false
		fetches.EachRecord(func(record *kgo.Record) {
			if handlerErr != nil {
				return
			}
			var event Event
			if err := json.Unmarshal(record.Value, &event); err != nil {
				handlerErr = fmt.Errorf("decode kafka event: %w", err)
				return
			}
			if err := handler(ctx, event); err != nil {
				if errors.Is(err, ErrStopConsumer) {
					stopAfterCommit = true
				} else {
					handlerErr = err
					return
				}
			}
			if err := c.client.CommitRecords(ctx, record); err != nil {
				handlerErr = fmt.Errorf("commit kafka record: %w", err)
			}
		})
		if handlerErr != nil {
			return handlerErr
		}
		if stopAfterCommit {
			return nil
		}
	}
}

func DeadLetterTopic(topic string) string { return topic + DeadLetterSuffix }

func Retryable(err error) bool { return kerr.IsRetriable(err) }
