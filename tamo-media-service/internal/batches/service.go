package batches

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/tamoimages/media-service/internal/platform/database"
	"github.com/tamoimages/media-service/internal/platform/kafka"
)

var ErrNotFound = errors.New("batch not found")
var ErrAssetUnavailable = errors.New("ready media asset is unavailable")
var ErrNotSubmittable = errors.New("batch is not submittable")
var ErrNotReviewable = errors.New("batch is not reviewable")

type Batch struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	Status         string     `json:"status"`
	ItemCount      int        `json:"itemCount"`
	SubmittedAt    *time.Time `json:"submittedAt,omitempty"`
	ReviewFeedback string     `json:"reviewFeedback,omitempty"`
	ReviewedAt     *time.Time `json:"reviewedAt,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

type Repository interface {
	CreateMediaBatch(context.Context, database.MediaBatch) (database.MediaBatch, error)
	MediaBatchesByContributor(context.Context, string, int) ([]database.MediaBatch, error)
	AddMediaAssetToBatch(context.Context, string, string, string) (bool, error)
	SubmitMediaBatch(context.Context, string, string) (database.MediaBatch, error)
	MediaBatchByID(context.Context, string) (database.MediaBatch, error)
	ReviewMediaBatch(context.Context, string, string, string) (database.MediaBatch, error)
}

type Publisher interface {
	Publish(context.Context, string, kafka.Event) error
}
type Service struct {
	repository Repository
	publisher  Publisher
	now        func() time.Time
}

func New(repository Repository, publisher Publisher) *Service {
	return &Service{repository: repository, publisher: publisher, now: time.Now}
}

func (s *Service) Create(ctx context.Context, contributorID, name string) (Batch, error) {
	record, err := s.repository.CreateMediaBatch(ctx, database.MediaBatch{ID: uuid.NewString(), ContributorID: contributorID, Name: strings.TrimSpace(name)})
	return publicBatch(record), err
}

func (s *Service) List(ctx context.Context, contributorID string) ([]Batch, error) {
	records, err := s.repository.MediaBatchesByContributor(ctx, contributorID, 50)
	if err != nil {
		return nil, err
	}
	result := make([]Batch, 0, len(records))
	for _, record := range records {
		result = append(result, publicBatch(record))
	}
	return result, nil
}

func (s *Service) AddItem(ctx context.Context, batchID, contributorID, assetID string) error {
	added, err := s.repository.AddMediaAssetToBatch(ctx, batchID, contributorID, assetID)
	if err != nil {
		return err
	}
	if !added {
		return ErrAssetUnavailable
	}
	return nil
}

func (s *Service) Submit(ctx context.Context, batchID, contributorID string) (Batch, error) {
	record, err := s.repository.SubmitMediaBatch(ctx, batchID, contributorID)
	if errors.Is(err, pgx.ErrNoRows) {
		record, err = s.repository.MediaBatchByID(ctx, batchID)
		if err != nil || record.ContributorID != contributorID || record.Status != "submitted" {
			return Batch{}, ErrNotSubmittable
		}
	}
	if err != nil {
		return Batch{}, err
	}
	if s.publisher != nil {
		payload, _ := json.Marshal(map[string]string{"batchId": record.ID, "contributorId": record.ContributorID})
		eventID := uuid.NewString()
		if err = s.publisher.Publish(ctx, "media.submitted-for-review.v1", kafka.Event{ID: eventID, Type: "media.submitted-for-review", Version: 1, OccurredAt: s.now().UTC(), CorrelationID: eventID, Producer: "tamo-media-service", Payload: payload}); err != nil {
			return Batch{}, err
		}
	}
	return publicBatch(record), nil
}

func (s *Service) Review(ctx context.Context, batchID, status, feedback string) (Batch, error) {
	record, err := s.repository.ReviewMediaBatch(ctx, batchID, status, strings.TrimSpace(feedback))
	if errors.Is(err, pgx.ErrNoRows) {
		record, err = s.repository.MediaBatchByID(ctx, batchID)
		if err != nil || record.Status != status {
			return Batch{}, ErrNotReviewable
		}
	}
	if err != nil {
		return Batch{}, err
	}
	if s.publisher != nil {
		payload, _ := json.Marshal(map[string]string{"batchId": record.ID, "contributorId": record.ContributorID, "status": record.Status})
		eventID := uuid.NewString()
		if err = s.publisher.Publish(ctx, "media.review-completed.v1", kafka.Event{ID: eventID, Type: "media.review-completed", Version: 1, OccurredAt: s.now().UTC(), CorrelationID: eventID, Producer: "tamo-media-service", Payload: payload}); err != nil {
			return Batch{}, err
		}
	}
	return publicBatch(record), nil
}

func publicBatch(record database.MediaBatch) Batch {
	return Batch{ID: record.ID, Name: record.Name, Status: record.Status, ItemCount: record.ItemCount, SubmittedAt: record.SubmittedAt, ReviewFeedback: record.ReviewFeedback, ReviewedAt: record.ReviewedAt, CreatedAt: record.CreatedAt, UpdatedAt: record.UpdatedAt}
}
