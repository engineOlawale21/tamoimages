package uploads

import (
	"context"
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/tamoimages/media-service/internal/platform/database"
	"github.com/tamoimages/media-service/internal/platform/kafka"
	"github.com/tamoimages/media-service/internal/platform/storage"
)

type repositoryFake struct {
	record   database.MediaAsset
	variants []database.MediaVariant
}

func (f *repositoryFake) MediaDeletionTargets(_ context.Context, id, owner string) (database.MediaDeletionTargets, error) {
	return database.MediaDeletionTargets{AssetID: id, ContributorID: owner, StorageKeys: []string{"original", "variant"}}, nil
}
func (f *repositoryFake) CompleteMediaDeletion(_ context.Context, _ database.MediaDeletionTargets, _ string, _ string) error {
	return nil
}

func (f *repositoryFake) CreateMediaAsset(_ context.Context, record database.MediaAsset) (database.MediaAsset, error) {
	record.Status = "pending"
	f.record = record
	return record, nil
}
func (f *repositoryFake) MediaAssetByID(_ context.Context, id, owner string) (database.MediaAsset, error) {
	return f.record, nil
}
func (f *repositoryFake) MediaAssetsByContributor(context.Context, string, int) ([]database.MediaAsset, error) {
	return []database.MediaAsset{f.record}, nil
}
func (f *repositoryFake) MediaVariantsByAssetID(context.Context, string) ([]database.MediaVariant, error) {
	return f.variants, nil
}
func (f *repositoryFake) MarkMediaAssetUploaded(_ context.Context, id, owner string) (database.MediaAsset, error) {
	f.record.Status = "uploaded"
	return f.record, nil
}
func (f *repositoryFake) RetryFailedMediaAsset(_ context.Context, id, owner string) (database.MediaAsset, error) {
	f.record.Status = "uploaded"
	f.record.FailureCode = ""
	return f.record, nil
}
func (f *repositoryFake) UpdateMediaAssetMetadata(_ context.Context, id, owner string, metadata database.MediaAsset) (database.MediaAsset, error) {
	metadata.ID = id
	metadata.ContributorID = owner
	metadata.Status = "ready"
	f.record = metadata
	return metadata, nil
}

type storageFake struct {
	size    int64
	key     string
	deleted []string
}

func (f *storageFake) Delete(_ context.Context, key string) error {
	f.deleted = append(f.deleted, key)
	return nil
}

func (f *storageFake) PresignPut(_ context.Context, key string, _ time.Duration) (*url.URL, error) {
	f.key = key
	return url.Parse("https://storage.example/upload")
}
func (f *storageFake) PresignGet(_ context.Context, key string, _ time.Duration) (*url.URL, error) {
	return url.Parse("https://storage.example/" + key)
}
func (f *storageFake) Head(context.Context, string) (storage.ObjectInfo, error) {
	return storage.ObjectInfo{Size: f.size}, nil
}

type publisherFake struct {
	topic string
	event kafka.Event
}

func (f *publisherFake) Publish(_ context.Context, topic string, event kafka.Event) error {
	f.topic = topic
	f.event = event
	return nil
}

func TestCreateAndCompletePublishesUploadedEvent(t *testing.T) {
	repository := &repositoryFake{}
	objectStore := &storageFake{size: 500}
	publisher := &publisherFake{}
	service := New(repository, objectStore, publisher, 15*time.Minute)
	session, err := service.Create(context.Background(), CreateInput{ContributorID: "contributor-1", Kind: "video", Filename: `..\clip.mp4`, ContentType: "video/mp4", SizeBytes: 500})
	if err != nil {
		t.Fatal(err)
	}
	if session.Asset.Filename != "clip.mp4" || objectStore.key == "" {
		t.Fatalf("unsafe or missing object key: %+v %q", session.Asset, objectStore.key)
	}
	asset, err := service.Complete(context.Background(), session.Asset.ID, "contributor-1")
	if err != nil {
		t.Fatal(err)
	}
	if asset.Status != "uploaded" || publisher.topic != "media.uploaded.v1" || publisher.event.Type != "media.uploaded" {
		t.Fatalf("unexpected completion %+v topic=%q event=%+v", asset, publisher.topic, publisher.event)
	}
}

func TestCompleteRejectsChangedObjectSize(t *testing.T) {
	repository := &repositoryFake{record: database.MediaAsset{ID: "asset-1", ContributorID: "contributor-1", Status: "pending", StorageKey: "private", SizeBytes: 500}}
	service := New(repository, &storageFake{size: 499}, &publisherFake{}, time.Minute)
	if _, err := service.Complete(context.Background(), "asset-1", "contributor-1"); err != ErrObjectMismatch {
		t.Fatalf("expected object mismatch, got %v", err)
	}
}

func TestDeleteRemovesAllPrivateObjects(t *testing.T) {
	repository := &repositoryFake{}
	objectStore := &storageFake{}
	service := New(repository, objectStore, &publisherFake{}, time.Minute)
	if err := service.Delete(context.Background(), "asset-1", "contributor-1"); err != nil {
		t.Fatal(err)
	}
	if len(objectStore.deleted) != 2 {
		t.Fatalf("expected two deleted objects, got %d", len(objectStore.deleted))
	}
}

func TestGetReadyAssetReturnsSignedVariantsWithoutStorageKeys(t *testing.T) {
	repository := &repositoryFake{
		record:   database.MediaAsset{ID: "asset-1", ContributorID: "contributor-1", Status: "ready", Width: 1600, Height: 1200},
		variants: []database.MediaVariant{{Kind: "thumbnail", StorageKey: "variants/private/thumbnail.webp", ContentType: "image/webp", SizeBytes: 1200, Width: 320, Height: 240}},
	}
	service := New(repository, &storageFake{}, &publisherFake{}, 15*time.Minute)
	asset, err := service.Get(context.Background(), "asset-1", "contributor-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(asset.Variants) != 1 || asset.Variants[0].Kind != "thumbnail" || asset.Variants[0].URL == "" {
		t.Fatalf("unexpected variants: %+v", asset.Variants)
	}
	if asset.Variants[0].ExpiresAt.IsZero() || asset.Width != 1600 || asset.Height != 1200 {
		t.Fatalf("missing metadata: %+v", asset)
	}
}

func TestListReturnsContributorAssetsWithSignedReadyVariants(t *testing.T) {
	repository := &repositoryFake{record: database.MediaAsset{ID: "asset-1", Status: "ready"}, variants: []database.MediaVariant{{Kind: "small", StorageKey: "variants/private/small.webp", ContentType: "image/webp", SizeBytes: 100}}}
	page, err := New(repository, &storageFake{}, &publisherFake{}, time.Minute).List(context.Background(), "contributor-1")
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 1 || len(page.Items) != 1 || len(page.Items[0].Variants) != 1 || page.Items[0].Variants[0].URL == "" {
		t.Fatalf("unexpected page: %+v", page)
	}
}

func TestRetryFailedAssetPublishesUploadedEvent(t *testing.T) {
	repository := &repositoryFake{record: database.MediaAsset{ID: "asset-1", ContributorID: "contributor-1", Status: "failed", FailureCode: "transcode_failed"}}
	publisher := &publisherFake{}
	asset, err := New(repository, &storageFake{}, publisher, time.Minute).Retry(context.Background(), "asset-1", "contributor-1")
	if err != nil {
		t.Fatal(err)
	}
	if asset.Status != "uploaded" || asset.FailureCode != "" || publisher.topic != "media.uploaded.v1" {
		t.Fatalf("unexpected retry result asset=%+v topic=%q", asset, publisher.topic)
	}
}

func TestRetryReadyAssetIsRejected(t *testing.T) {
	repository := &repositoryFake{record: database.MediaAsset{ID: "asset-1", ContributorID: "contributor-1", Status: "ready"}}
	_, err := New(repository, &storageFake{}, &publisherFake{}, time.Minute).Retry(context.Background(), "asset-1", "contributor-1")
	if !errors.Is(err, ErrNotRetryable) {
		t.Fatalf("expected not retryable, got %v", err)
	}
}
