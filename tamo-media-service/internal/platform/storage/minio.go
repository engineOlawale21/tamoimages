package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIO struct {
	client *minio.Client
	bucket string
}

func OpenMinIO(endpoint, bucket, accessKey, secretKey string) (*MinIO, error) {
	parsed, err := url.Parse(endpoint)
	if err != nil || parsed.Host == "" {
		return nil, fmt.Errorf("invalid object-storage endpoint")
	}
	if strings.TrimSpace(bucket) == "" {
		return nil, fmt.Errorf("object-storage bucket is required")
	}
	client, err := minio.New(parsed.Host, &minio.Options{
		Creds: credentials.NewStaticV4(accessKey, secretKey, ""), Secure: parsed.Scheme == "https",
	})
	if err != nil {
		return nil, fmt.Errorf("create object-storage client: %w", err)
	}
	return &MinIO{client: client, bucket: bucket}, nil
}

func (m *MinIO) EnsureBucket(ctx context.Context) error {
	exists, err := m.client.BucketExists(ctx, m.bucket)
	if err != nil {
		return fmt.Errorf("check object-storage bucket: %w", err)
	}
	if !exists {
		return fmt.Errorf("required object-storage bucket %q does not exist", m.bucket)
	}
	return nil
}

func (m *MinIO) Put(ctx context.Context, key string, reader io.Reader, size int64, contentType string) (ObjectInfo, error) {
	result, err := m.client.PutObject(ctx, m.bucket, key, reader, size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return ObjectInfo{}, fmt.Errorf("put object: %w", err)
	}
	return ObjectInfo{Key: result.Key, Size: result.Size, ContentType: contentType}, nil
}

func (m *MinIO) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	object, err := m.client.GetObject(ctx, m.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("get object: %w", err)
	}
	if _, err = object.Stat(); err != nil {
		_ = object.Close()
		return nil, fmt.Errorf("stat downloaded object: %w", err)
	}
	return object, nil
}

func (m *MinIO) Head(ctx context.Context, key string) (ObjectInfo, error) {
	result, err := m.client.StatObject(ctx, m.bucket, key, minio.StatObjectOptions{})
	if err != nil {
		return ObjectInfo{}, fmt.Errorf("head object: %w", err)
	}
	return ObjectInfo{Key: result.Key, Size: result.Size, ContentType: result.ContentType, LastModified: result.LastModified}, nil
}

func (m *MinIO) Delete(ctx context.Context, key string) error {
	if err := m.client.RemoveObject(ctx, m.bucket, key, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("delete object: %w", err)
	}
	return nil
}

func (m *MinIO) PresignGet(ctx context.Context, key string, expiry time.Duration) (*url.URL, error) {
	result, err := m.client.PresignedGetObject(ctx, m.bucket, key, expiry, nil)
	if err != nil {
		return nil, fmt.Errorf("presign object download: %w", err)
	}
	return result, nil
}

func (m *MinIO) PresignPut(ctx context.Context, key string, expiry time.Duration) (*url.URL, error) {
	result, err := m.client.PresignedPutObject(ctx, m.bucket, key, expiry)
	if err != nil {
		return nil, fmt.Errorf("presign object upload: %w", err)
	}
	return result, nil
}
