package batches_test

import (
	"context"
	"errors"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tamoimages/media-service/internal/batches"
	"github.com/tamoimages/media-service/internal/platform/database"
	"github.com/tamoimages/media-service/internal/platform/kafka"
	"github.com/tamoimages/media-service/internal/platform/storage"
	"github.com/tamoimages/media-service/internal/releases"
)

type releaseStore struct{ size int64 }

func (s releaseStore) PresignPut(context.Context, string, time.Duration) (*url.URL, error) {
	return url.Parse("https://storage.example/release")
}
func (s releaseStore) Head(context.Context, string) (storage.ObjectInfo, error) {
	return storage.ObjectInfo{Size: s.size}, nil
}

type eventPublisher struct{ topics []string }

func (p *eventPublisher) Publish(_ context.Context, topic string, _ kafka.Event) error {
	p.topics = append(p.topics, topic)
	return nil
}

func TestMilestone4ContributorJourney(t *testing.T) {
	connectionString := os.Getenv("MILESTONE4_DATABASE_URL")
	if connectionString == "" {
		t.Skip("MILESTONE4_DATABASE_URL is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	db, err := database.Open(ctx, connectionString, 3, 3*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	cleanupPool, err := pgxpool.New(ctx, connectionString)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanupPool.Close()
	contributorID := uuid.NewString()
	assetID := uuid.NewString()
	batchID := ""
	defer func() {
		if batchID != "" {
			_, _ = cleanupPool.Exec(context.Background(), "DELETE FROM media_batches WHERE id=$1", batchID)
		}
		_, _ = cleanupPool.Exec(context.Background(), "DELETE FROM media_assets WHERE id=$1", assetID)
	}()
	record, err := db.CreateMediaAsset(ctx, database.MediaAsset{ID: assetID, ContributorID: contributorID, Kind: "image", OriginalFilename: "lagos.jpg", StorageKey: "integration/original/" + assetID, ContentType: "image/jpeg", SizeBytes: 100})
	if err != nil {
		t.Fatal(err)
	}
	if record, err = db.MarkMediaAssetUploaded(ctx, record.ID, contributorID); err != nil {
		t.Fatal(err)
	}
	if _, claimed, claimErr := db.ClaimMediaAssetForProcessing(ctx, record.ID); claimErr != nil || !claimed {
		t.Fatalf("claim failed claimed=%v error=%v", claimed, claimErr)
	}
	if err = db.CompleteMediaAssetProcessing(ctx, record.ID, database.ProcessingMetadata{Width: 1600, Height: 1200, CodecName: "mjpeg"}, []database.MediaVariant{{Kind: "thumbnail", StorageKey: "integration/variant/" + assetID, ContentType: "image/webp", SizeBytes: 50, Width: 320, Height: 240}}); err != nil {
		t.Fatal(err)
	}
	if _, err = db.UpdateMediaAssetMetadata(ctx, record.ID, contributorID, database.MediaAsset{Title: "Lagos street life", Description: "Editorial street scene", Keywords: []string{"lagos", "street"}, Location: "Lagos", UsageType: "editorial"}); err != nil {
		t.Fatal(err)
	}
	releaseService := releases.New(db, releaseStore{size: 100}, time.Minute)
	for index := 0; index < 3; index++ {
		session, createErr := releaseService.Create(ctx, releases.CreateInput{ContributorID: contributorID, AssetID: record.ID, ReleaseType: "model", Filename: "release.pdf", ContentType: "application/pdf", SizeBytes: 100})
		if createErr != nil {
			t.Fatal(createErr)
		}
		if _, completeErr := releaseService.Complete(ctx, record.ID, session.Release.ID, contributorID); completeErr != nil {
			t.Fatal(completeErr)
		}
	}
	if _, err = releaseService.Create(ctx, releases.CreateInput{ContributorID: contributorID, AssetID: record.ID, ReleaseType: "property", Filename: "fourth.pdf", ContentType: "application/pdf", SizeBytes: 100}); !errors.Is(err, releases.ErrLimitReached) {
		t.Fatalf("expected fourth release rejection, got %v", err)
	}
	publisher := &eventPublisher{}
	batchService := batches.New(db, publisher)
	batch, err := batchService.Create(ctx, contributorID, "Lagos collection")
	if err != nil {
		t.Fatal(err)
	}
	batchID = batch.ID
	if err = batchService.AddItem(ctx, batch.ID, contributorID, record.ID); err != nil {
		t.Fatal(err)
	}
	batch, err = batchService.Submit(ctx, batch.ID, contributorID)
	if err != nil {
		t.Fatal(err)
	}
	if batch.Status != "submitted" {
		t.Fatalf("unexpected submitted batch %+v", batch)
	}
	batch, err = batchService.Review(ctx, batch.ID, "rejected", "Add a more specific location.")
	if err != nil {
		t.Fatal(err)
	}
	if batch.Status != "rejected" || batch.ReviewFeedback != "Add a more specific location." {
		t.Fatalf("unexpected reviewed batch %+v", batch)
	}
	if len(publisher.topics) != 2 || publisher.topics[0] != "media.submitted-for-review.v1" || publisher.topics[1] != "media.review-completed.v1" {
		t.Fatalf("unexpected topics %v", publisher.topics)
	}
}
