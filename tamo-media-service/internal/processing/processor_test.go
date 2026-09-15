package processing

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tamoimages/media-service/internal/platform/database"
	"github.com/tamoimages/media-service/internal/platform/kafka"
	"github.com/tamoimages/media-service/internal/platform/storage"
)

type fakeRepository struct {
	asset       database.MediaAsset
	claimed     bool
	ready       bool
	failure     string
	variants    []database.MediaVariant
	completeErr error
}

func (f *fakeRepository) ClaimMediaAssetForProcessing(context.Context, string) (database.MediaAsset, bool, error) {
	return f.asset, f.claimed, nil
}
func (f *fakeRepository) CompleteMediaAssetProcessing(_ context.Context, _ string, _ database.ProcessingMetadata, variants []database.MediaVariant) error {
	if f.completeErr != nil {
		return f.completeErr
	}
	f.ready = true
	f.variants = variants
	return nil
}
func (f *fakeRepository) MarkMediaAssetFailed(_ context.Context, _ string, code string) error {
	f.failure = code
	return nil
}

type fakeStore struct {
	err       error
	puts      int
	failPutAt int
	deletes   int
}

func (f fakeStore) Get(context.Context, string) (io.ReadCloser, error) {
	if f.err != nil {
		return nil, f.err
	}
	return io.NopCloser(strings.NewReader("video")), nil
}
func (f *fakeStore) Put(_ context.Context, _ string, _ io.Reader, _ int64, _ string) (storage.ObjectInfo, error) {
	f.puts++
	if f.failPutAt == f.puts {
		return storage.ObjectInfo{}, errors.New("storage unavailable")
	}
	return storage.ObjectInfo{}, nil
}
func (f *fakeStore) Delete(context.Context, string) error { f.deletes++; return nil }

type fakePublisher struct{ topic string }

func (f *fakePublisher) Publish(_ context.Context, topic string, _ kafka.Event) error {
	f.topic = topic
	return nil
}

type fakeProber struct {
	metadata database.ProcessingMetadata
	err      error
}
type fakeScanner struct{ err error }

func (f fakeScanner) Scan(context.Context, string) error { return f.err }

func (f fakeProber) Probe(context.Context, string) (database.ProcessingMetadata, error) {
	return f.metadata, f.err
}

type fakeTranscoder struct{ err error }

func (f fakeTranscoder) Transcode(_ context.Context, _ string, workDirectory, _ string, metadata database.ProcessingMetadata) ([]GeneratedVariant, error) {
	if f.err != nil {
		return nil, f.err
	}
	preview := filepath.Join(workDirectory, "preview.mp4")
	poster := filepath.Join(workDirectory, "poster.jpg")
	if err := os.WriteFile(preview, []byte("preview"), 0o600); err != nil {
		return nil, err
	}
	if err := os.WriteFile(poster, []byte("poster"), 0o600); err != nil {
		return nil, err
	}
	return []GeneratedVariant{{Kind: "preview", Path: preview, ContentType: "video/mp4", Width: 1280, Height: 720}, {Kind: "poster", Path: poster, ContentType: "image/jpeg", Width: metadata.Width, Height: metadata.Height}}, nil
}

func uploadedEvent() kafka.Event {
	return kafka.Event{Type: "media.uploaded", CorrelationID: "correlation-1", Payload: []byte(`{"assetId":"asset-1"}`)}
}

func TestProcessorMarksValidVideoReadyAndPublishes(t *testing.T) {
	repository := &fakeRepository{claimed: true, asset: database.MediaAsset{ID: "asset-1", Kind: "video", StorageKey: "private", SizeBytes: 5}}
	publisher := &fakePublisher{}
	objectStore := &fakeStore{}
	processor := New(repository, objectStore, publisher, fakeScanner{}, fakeProber{metadata: database.ProcessingMetadata{DurationSeconds: 2, Width: 1920, Height: 1080, CodecName: "h264"}}, fakeTranscoder{})
	if err := processor.Handle(context.Background(), uploadedEvent()); err != nil {
		t.Fatal(err)
	}
	if !repository.ready || len(repository.variants) != 2 || objectStore.puts != 2 || publisher.topic != "media.processed.v1" {
		t.Fatalf("unexpected result ready=%v topic=%q", repository.ready, publisher.topic)
	}
}

func TestProcessorRecordsBoundedFailureAndPublishes(t *testing.T) {
	repository := &fakeRepository{claimed: true, asset: database.MediaAsset{ID: "asset-1", Kind: "video", StorageKey: "private", SizeBytes: 5}}
	publisher := &fakePublisher{}
	processor := New(repository, &fakeStore{}, publisher, fakeScanner{}, fakeProber{err: errors.New("private command output")}, fakeTranscoder{})
	if err := processor.Handle(context.Background(), uploadedEvent()); err != nil {
		t.Fatal(err)
	}
	if repository.failure != "invalid_media" || publisher.topic != "media.processing-failed.v1" {
		t.Fatalf("unexpected failure=%q topic=%q", repository.failure, publisher.topic)
	}
}

func TestProcessorIgnoresDuplicateEventWhenAssetCannotBeClaimed(t *testing.T) {
	repository := &fakeRepository{claimed: false}
	publisher := &fakePublisher{}
	if err := New(repository, &fakeStore{}, publisher, fakeScanner{}, fakeProber{}, fakeTranscoder{}).Handle(context.Background(), uploadedEvent()); err != nil {
		t.Fatal(err)
	}
	if publisher.topic != "" {
		t.Fatal("duplicate event must not publish")
	}
}

func TestProcessorRemovesUploadedVariantsWhenLaterUploadFails(t *testing.T) {
	repository := &fakeRepository{claimed: true, asset: database.MediaAsset{ID: "asset-1", Kind: "video", StorageKey: "private", SizeBytes: 5}}
	objectStore := &fakeStore{failPutAt: 2}
	processor := New(repository, objectStore, &fakePublisher{}, fakeScanner{}, fakeProber{metadata: database.ProcessingMetadata{DurationSeconds: 2, Width: 1920, Height: 1080}}, fakeTranscoder{})
	if err := processor.Handle(context.Background(), uploadedEvent()); err != nil {
		t.Fatal(err)
	}
	if objectStore.deletes != 1 || repository.failure != "variant_storage_failed" {
		t.Fatalf("expected one derivative cleanup and bounded failure, deletes=%d failure=%q", objectStore.deletes, repository.failure)
	}
}

func TestProcessorReturnsFinalizationFailureForStaleRetry(t *testing.T) {
	expected := errors.New("database unavailable")
	repository := &fakeRepository{claimed: true, completeErr: expected, asset: database.MediaAsset{ID: "asset-1", Kind: "video", StorageKey: "private", SizeBytes: 5}}
	err := New(repository, &fakeStore{}, &fakePublisher{}, fakeScanner{}, fakeProber{metadata: database.ProcessingMetadata{DurationSeconds: 2, Width: 1920, Height: 1080}}, fakeTranscoder{}).Handle(context.Background(), uploadedEvent())
	if !errors.Is(err, expected) {
		t.Fatalf("expected finalization error, got %v", err)
	}
}

func TestProcessorFailsClosedWhenMalwareScanFails(t *testing.T) {
	repository := &fakeRepository{claimed: true, asset: database.MediaAsset{ID: "asset-1", Kind: "video", StorageKey: "private", SizeBytes: 5}}
	err := New(repository, &fakeStore{}, &fakePublisher{}, fakeScanner{err: errors.New("infected")}, fakeProber{}, fakeTranscoder{}).Handle(context.Background(), uploadedEvent())
	if err != nil {
		t.Fatal(err)
	}
	if repository.failure != "malware_scan_failed" {
		t.Fatalf("unexpected failure %q", repository.failure)
	}
}

func TestPreviewDimensionsNeverUpscaleAndRemainEven(t *testing.T) {
	if width, height := fitBox(640, 360, 1280, 720, true); width != 640 || height != 360 {
		t.Fatalf("small video was resized to %dx%d", width, height)
	}
	if width, height := fitBox(1920, 1081, 1280, 720, true); width != 1278 || height != 720 {
		t.Fatalf("unexpected preview size %dx%d", width, height)
	}
}
