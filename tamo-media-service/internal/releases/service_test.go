package releases

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/tamoimages/media-service/internal/platform/database"
	"github.com/tamoimages/media-service/internal/platform/storage"
)

type repositoryFake struct {
	record  database.MediaRelease
	limit   bool
	missing bool
}

func (f *repositoryFake) CreateMediaRelease(_ context.Context, record database.MediaRelease) (database.MediaRelease, bool, error) {
	if f.missing {
		return database.MediaRelease{}, false, pgx.ErrNoRows
	}
	record.Status = "pending"
	f.record = record
	return record, f.limit, nil
}
func (f *repositoryFake) MediaReleaseByID(context.Context, string, string, string) (database.MediaRelease, error) {
	if f.missing {
		return database.MediaRelease{}, pgx.ErrNoRows
	}
	return f.record, nil
}
func (f *repositoryFake) MarkMediaReleaseUploaded(context.Context, string, string, string) (database.MediaRelease, error) {
	f.record.Status = "uploaded"
	return f.record, nil
}
func (f *repositoryFake) MediaReleasesByAsset(context.Context, string, string) ([]database.MediaRelease, error) {
	return []database.MediaRelease{f.record}, nil
}

type storeFake struct {
	size int64
	key  string
}

func (f *storeFake) PresignPut(_ context.Context, key string, _ time.Duration) (*url.URL, error) {
	f.key = key
	return url.Parse("https://storage.example/release")
}
func (f *storeFake) Head(context.Context, string) (storage.ObjectInfo, error) {
	return storage.ObjectInfo{Size: f.size}, nil
}

func TestCreateAndCompleteRelease(t *testing.T) {
	repository := &repositoryFake{}
	objectStore := &storeFake{size: 100}
	service := New(repository, objectStore, 15*time.Minute)
	session, err := service.Create(context.Background(), CreateInput{ContributorID: "contributor-1", AssetID: "asset-1", ReleaseType: "model", Filename: `..\release.pdf`, ContentType: "application/pdf", SizeBytes: 100})
	if err != nil {
		t.Fatal(err)
	}
	if session.Release.Filename != "release.pdf" || objectStore.key == "" {
		t.Fatalf("unsafe release session %+v", session)
	}
	completed, err := service.Complete(context.Background(), "asset-1", session.Release.ID, "contributor-1")
	if err != nil || completed.Status != "uploaded" {
		t.Fatalf("unexpected completion %+v error=%v", completed, err)
	}
}
func TestCreateEnforcesThreeReleaseLimit(t *testing.T) {
	_, err := New(&repositoryFake{limit: true}, &storeFake{}, time.Minute).Create(context.Background(), CreateInput{ContributorID: "contributor-1", AssetID: "asset-1", ReleaseType: "model", Filename: "release.pdf", ContentType: "application/pdf", SizeBytes: 100})
	if err != ErrLimitReached {
		t.Fatalf("expected limit error, got %v", err)
	}
}
func TestCompleteRejectsWrongObjectSize(t *testing.T) {
	repository := &repositoryFake{record: database.MediaRelease{ID: "release-1", MediaAssetID: "asset-1", Status: "pending", SizeBytes: 100, StorageKey: "private"}}
	_, err := New(repository, &storeFake{size: 99}, time.Minute).Complete(context.Background(), "asset-1", "release-1", "contributor-1")
	if err != ErrObjectMismatch {
		t.Fatalf("expected mismatch, got %v", err)
	}
}
