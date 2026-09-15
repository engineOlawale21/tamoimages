package batches

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/tamoimages/media-service/internal/platform/database"
	"github.com/tamoimages/media-service/internal/platform/kafka"
)

type repositoryFake struct {
	batch     database.MediaBatch
	added     bool
	submitErr error
}

func (f *repositoryFake) CreateMediaBatch(_ context.Context, batch database.MediaBatch) (database.MediaBatch, error) {
	batch.Status = "draft"
	f.batch = batch
	return batch, nil
}
func (f *repositoryFake) MediaBatchesByContributor(context.Context, string, int) ([]database.MediaBatch, error) {
	return []database.MediaBatch{f.batch}, nil
}
func (f *repositoryFake) AddMediaAssetToBatch(context.Context, string, string, string) (bool, error) {
	return f.added, nil
}
func (f *repositoryFake) SubmitMediaBatch(context.Context, string, string) (database.MediaBatch, error) {
	if f.submitErr != nil {
		return database.MediaBatch{}, f.submitErr
	}
	f.batch.Status = "submitted"
	return f.batch, nil
}
func (f *repositoryFake) MediaBatchByID(context.Context, string) (database.MediaBatch, error) {
	return f.batch, nil
}
func (f *repositoryFake) ReviewMediaBatch(_ context.Context, _ string, status, feedback string) (database.MediaBatch, error) {
	f.batch.Status = status
	f.batch.ReviewFeedback = feedback
	return f.batch, nil
}

type publisherFake struct{ topic string }

func (f *publisherFake) Publish(_ context.Context, topic string, _ kafka.Event) error {
	f.topic = topic
	return nil
}

func TestCreateTrimsNameAndListsBatch(t *testing.T) {
	repository := &repositoryFake{}
	service := New(repository, nil)
	created, err := service.Create(context.Background(), "contributor-1", "  Lagos stories  ")
	if err != nil {
		t.Fatal(err)
	}
	if created.Name != "Lagos stories" || created.Status != "draft" {
		t.Fatalf("unexpected batch %+v", created)
	}
	listed, err := service.List(context.Background(), "contributor-1")
	if err != nil || len(listed) != 1 {
		t.Fatalf("unexpected list %+v error=%v", listed, err)
	}
}
func TestAddItemRejectsUnavailableAsset(t *testing.T) {
	err := New(&repositoryFake{}, nil).AddItem(context.Background(), "batch-1", "contributor-1", "asset-1")
	if err != ErrAssetUnavailable {
		t.Fatalf("expected unavailable asset, got %v", err)
	}
}
func TestSubmitRejectsEmptyOrNonDraftBatch(t *testing.T) {
	_, err := New(&repositoryFake{submitErr: pgx.ErrNoRows}, nil).Submit(context.Background(), "batch-1", "contributor-1")
	if err != ErrNotSubmittable {
		t.Fatalf("expected not submittable, got %v", err)
	}
}
func TestSubmitPublishesReviewEvent(t *testing.T) {
	repository := &repositoryFake{batch: database.MediaBatch{ID: "batch-1", ContributorID: "contributor-1"}}
	publisher := &publisherFake{}
	batch, err := New(repository, publisher).Submit(context.Background(), "batch-1", "contributor-1")
	if err != nil {
		t.Fatal(err)
	}
	if batch.Status != "submitted" || publisher.topic != "media.submitted-for-review.v1" {
		t.Fatalf("unexpected batch %+v topic=%q", batch, publisher.topic)
	}
}
func TestReviewStoresContributorFeedback(t *testing.T) {
	repository := &repositoryFake{batch: database.MediaBatch{ID: "batch-1"}}
	publisher := &publisherFake{}
	batch, err := New(repository, publisher).Review(context.Background(), "batch-1", "rejected", " Add clearer keywords ")
	if err != nil {
		t.Fatal(err)
	}
	if batch.Status != "rejected" || batch.ReviewFeedback != "Add clearer keywords" || publisher.topic != "media.review-completed.v1" {
		t.Fatalf("unexpected review %+v", batch)
	}
}
