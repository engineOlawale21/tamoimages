package storage

import (
	"context"
	"io"
	"net/url"
	"time"
)

type ObjectInfo struct {
	Key          string
	Size         int64
	ContentType  string
	LastModified time.Time
}

type Store interface {
	EnsureBucket(context.Context) error
	Put(context.Context, string, io.Reader, int64, string) (ObjectInfo, error)
	Get(context.Context, string) (io.ReadCloser, error)
	Head(context.Context, string) (ObjectInfo, error)
	Delete(context.Context, string) error
	PresignGet(context.Context, string, time.Duration) (*url.URL, error)
	PresignPut(context.Context, string, time.Duration) (*url.URL, error)
}
