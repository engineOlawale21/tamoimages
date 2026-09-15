package redis

import (
	"context"
	"fmt"
	"strings"
	"time"

	redisclient "github.com/redis/go-redis/v9"
)

type Client struct {
	client  *redisclient.Client
	prefix  string
	timeout time.Duration
}

func Open(address, prefix string, timeout time.Duration, maxRetries int) (*Client, error) {
	if strings.TrimSpace(address) == "" {
		return nil, fmt.Errorf("redis address is required")
	}
	if strings.TrimSpace(prefix) == "" {
		return nil, fmt.Errorf("redis key prefix is required")
	}
	if timeout <= 0 {
		return nil, fmt.Errorf("redis operation timeout must be positive")
	}
	return &Client{
		client: redisclient.NewClient(&redisclient.Options{
			Addr: address, DialTimeout: timeout, ReadTimeout: timeout,
			WriteTimeout: timeout, MaxRetries: maxRetries,
		}),
		prefix: strings.TrimSuffix(prefix, ":"), timeout: timeout,
	}, nil
}

func (c *Client) Close() error { return c.client.Close() }

func (c *Client) Ping(ctx context.Context) error {
	operationContext, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	if err := c.client.Ping(operationContext).Err(); err != nil {
		return fmt.Errorf("ping redis: %w", err)
	}
	return nil
}

func (c *Client) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	operationContext, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	if err := c.client.Set(operationContext, c.key(key), value, expiration).Err(); err != nil {
		return fmt.Errorf("set redis key: %w", err)
	}
	return nil
}

func (c *Client) Get(ctx context.Context, key string) ([]byte, error) {
	operationContext, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	value, err := c.client.Get(operationContext, c.key(key)).Bytes()
	if err != nil {
		return nil, fmt.Errorf("get redis key: %w", err)
	}
	return value, nil
}

func (c *Client) Delete(ctx context.Context, key string) error {
	operationContext, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	if err := c.client.Del(operationContext, c.key(key)).Err(); err != nil {
		return fmt.Errorf("delete redis key: %w", err)
	}
	return nil
}

func (c *Client) key(key string) string { return c.prefix + ":" + strings.TrimPrefix(key, ":") }
