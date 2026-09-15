package uploads

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/tamoimages/media-service/internal/platform/database"
	"github.com/tamoimages/media-service/internal/platform/kafka"
	"github.com/tamoimages/media-service/internal/platform/storage"
)

var ErrNotFound = errors.New("media asset not found")
var ErrObjectMismatch = errors.New("uploaded object does not match the declared file")
var ErrNotRetryable = errors.New("media asset is not retryable")

type CreateInput struct {
	ContributorID string
	Kind          string
	Filename      string
	ContentType   string
	SizeBytes     int64
}

type MetadataInput struct {
	Title       string
	Description string
	Keywords    []string
	Location    string
	UsageType   string
}

type Session struct {
	Asset     Asset
	UploadURL string
	ExpiresAt time.Time
}

type Asset struct {
	ID              string     `json:"id"`
	Kind            string     `json:"kind"`
	Filename        string     `json:"filename"`
	ContentType     string     `json:"contentType"`
	SizeBytes       int64      `json:"sizeBytes"`
	Status          string     `json:"status"`
	DurationSeconds float64    `json:"durationSeconds,omitempty"`
	Width           int        `json:"width,omitempty"`
	Height          int        `json:"height,omitempty"`
	CodecName       string     `json:"codecName,omitempty"`
	FailureCode     string     `json:"failureCode,omitempty"`
	Title           string     `json:"title,omitempty"`
	Description     string     `json:"description,omitempty"`
	Keywords        []string   `json:"keywords"`
	Location        string     `json:"location,omitempty"`
	UsageType       string     `json:"usageType,omitempty"`
	ProcessedAt     *time.Time `json:"processedAt,omitempty"`
	Variants        []Variant  `json:"variants"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

type Variant struct {
	Kind        string    `json:"kind"`
	URL         string    `json:"url"`
	ContentType string    `json:"contentType"`
	SizeBytes   int64     `json:"sizeBytes"`
	Width       int       `json:"width,omitempty"`
	Height      int       `json:"height,omitempty"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

type Repository interface {
	CreateMediaAsset(context.Context, database.MediaAsset) (database.MediaAsset, error)
	MediaAssetByID(context.Context, string, string) (database.MediaAsset, error)
	MediaAssetsByContributor(context.Context, string, int) ([]database.MediaAsset, error)
	MediaVariantsByAssetID(context.Context, string) ([]database.MediaVariant, error)
	MarkMediaAssetUploaded(context.Context, string, string) (database.MediaAsset, error)
	RetryFailedMediaAsset(context.Context, string, string) (database.MediaAsset, error)
	UpdateMediaAssetMetadata(context.Context, string, string, database.MediaAsset) (database.MediaAsset, error)
	MediaDeletionTargets(context.Context, string, string) (database.MediaDeletionTargets, error)
	CompleteMediaDeletion(context.Context, database.MediaDeletionTargets, string, string) error
}

type Page struct {
	Items []Asset `json:"items"`
	Total int     `json:"total"`
}

type ObjectStore interface {
	PresignPut(context.Context, string, time.Duration) (*url.URL, error)
	PresignGet(context.Context, string, time.Duration) (*url.URL, error)
	Head(context.Context, string) (storage.ObjectInfo, error)
	Delete(context.Context, string) error
}

func (s *Service) Delete(ctx context.Context, id, contributorID string) error {
	target, err := s.repository.MediaDeletionTargets(ctx, id, contributorID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	keys := append([]string(nil), target.StorageKeys...)
	sort.Strings(keys)
	for _, key := range keys {
		if err = s.storage.Delete(ctx, key); err != nil {
			return fmt.Errorf("delete private media object: %w", err)
		}
	}
	digest := sha256.Sum256([]byte(strings.Join(keys, "\n")))
	return s.repository.CompleteMediaDeletion(ctx, target, "contributor_request", fmt.Sprintf("%x", digest))
}

type EventPublisher interface {
	Publish(context.Context, string, kafka.Event) error
}

type Service struct {
	repository Repository
	storage    ObjectStore
	publisher  EventPublisher
	expiry     time.Duration
	now        func() time.Time
}

func New(repository Repository, objectStore ObjectStore, publisher EventPublisher, expiry time.Duration) *Service {
	return &Service{repository: repository, storage: objectStore, publisher: publisher, expiry: expiry, now: time.Now}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (Session, error) {
	id := uuid.NewString()
	filename := path.Base(strings.ReplaceAll(strings.TrimSpace(input.Filename), `\`, "/"))
	key := fmt.Sprintf("originals/%s/%s/%s", input.ContributorID, id, filename)
	record, err := s.repository.CreateMediaAsset(ctx, database.MediaAsset{
		ID: id, ContributorID: input.ContributorID, Kind: input.Kind, OriginalFilename: filename,
		StorageKey: key, ContentType: input.ContentType, SizeBytes: input.SizeBytes,
	})
	if err != nil {
		return Session{}, err
	}
	uploadURL, err := s.storage.PresignPut(ctx, key, s.expiry)
	if err != nil {
		return Session{}, fmt.Errorf("create presigned upload: %w", err)
	}
	return Session{Asset: publicAsset(record), UploadURL: uploadURL.String(), ExpiresAt: s.now().Add(s.expiry)}, nil
}

func (s *Service) Get(ctx context.Context, id, contributorID string) (Asset, error) {
	record, err := s.repository.MediaAssetByID(ctx, id, contributorID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Asset{}, ErrNotFound
	}
	if err != nil {
		return Asset{}, err
	}
	asset := publicAsset(record)
	if record.Status != "ready" {
		return asset, nil
	}
	if err = s.attachVariants(ctx, &asset); err != nil {
		return Asset{}, err
	}
	return asset, nil
}

func (s *Service) List(ctx context.Context, contributorID string) (Page, error) {
	records, err := s.repository.MediaAssetsByContributor(ctx, contributorID, 50)
	if err != nil {
		return Page{}, err
	}
	page := Page{Items: make([]Asset, 0, len(records)), Total: len(records)}
	for _, record := range records {
		asset := publicAsset(record)
		if record.Status == "ready" {
			if err = s.attachVariants(ctx, &asset); err != nil {
				return Page{}, err
			}
		}
		page.Items = append(page.Items, asset)
	}
	return page, nil
}

func (s *Service) attachVariants(ctx context.Context, asset *Asset) error {
	variants, err := s.repository.MediaVariantsByAssetID(ctx, asset.ID)
	if err != nil {
		return err
	}
	expiresAt := s.now().Add(s.expiry)
	asset.Variants = make([]Variant, 0, len(variants))
	for _, variant := range variants {
		downloadURL, signErr := s.storage.PresignGet(ctx, variant.StorageKey, s.expiry)
		if signErr != nil {
			return fmt.Errorf("create presigned variant download: %w", signErr)
		}
		asset.Variants = append(asset.Variants, Variant{Kind: variant.Kind, URL: downloadURL.String(), ContentType: variant.ContentType,
			SizeBytes: variant.SizeBytes, Width: variant.Width, Height: variant.Height, ExpiresAt: expiresAt})
	}
	return nil
}

func (s *Service) Complete(ctx context.Context, id, contributorID string) (Asset, error) {
	record, err := s.repository.MediaAssetByID(ctx, id, contributorID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Asset{}, ErrNotFound
	}
	if err != nil {
		return Asset{}, err
	}
	if record.Status != "pending" {
		if record.Status == "uploaded" && s.publisher != nil {
			if err = s.publishUploaded(ctx, record); err != nil {
				return Asset{}, err
			}
		}
		return publicAsset(record), nil
	}
	object, err := s.storage.Head(ctx, record.StorageKey)
	if err != nil || object.Size != record.SizeBytes {
		return Asset{}, ErrObjectMismatch
	}
	updated, err := s.repository.MarkMediaAssetUploaded(ctx, id, contributorID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Asset{}, ErrNotFound
	}
	if err != nil {
		return Asset{}, err
	}
	if s.publisher != nil {
		if err = s.publishUploaded(ctx, updated); err != nil {
			return Asset{}, err
		}
	}
	return publicAsset(updated), nil
}

func (s *Service) Retry(ctx context.Context, id, contributorID string) (Asset, error) {
	record, err := s.repository.MediaAssetByID(ctx, id, contributorID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Asset{}, ErrNotFound
	}
	if err != nil {
		return Asset{}, err
	}
	switch record.Status {
	case "failed":
		record, err = s.repository.RetryFailedMediaAsset(ctx, id, contributorID)
		if err != nil {
			return Asset{}, err
		}
	case "uploaded":
		// A prior publish can fail after the durable status change. Republishing is safe.
	case "processing":
		return publicAsset(record), nil
	default:
		return Asset{}, ErrNotRetryable
	}
	if s.publisher != nil {
		if err = s.publishUploaded(ctx, record); err != nil {
			return Asset{}, err
		}
	}
	return publicAsset(record), nil
}

func (s *Service) UpdateMetadata(ctx context.Context, id, contributorID string, input MetadataInput) (Asset, error) {
	record, err := s.repository.UpdateMediaAssetMetadata(ctx, id, contributorID, database.MediaAsset{
		Title: strings.TrimSpace(input.Title), Description: strings.TrimSpace(input.Description), Keywords: input.Keywords,
		Location: strings.TrimSpace(input.Location), UsageType: input.UsageType,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return Asset{}, ErrNotFound
	}
	if err != nil {
		return Asset{}, err
	}
	asset := publicAsset(record)
	if err = s.attachVariants(ctx, &asset); err != nil {
		return Asset{}, err
	}
	return asset, nil
}

func (s *Service) publishUploaded(ctx context.Context, record database.MediaAsset) error {
	payload, err := json.Marshal(map[string]string{"assetId": record.ID, "contributorId": record.ContributorID})
	if err != nil {
		return err
	}
	eventID := uuid.NewString()
	return s.publisher.Publish(ctx, "media.uploaded.v1", kafka.Event{ID: eventID, Type: "media.uploaded", Version: 1,
		OccurredAt: s.now().UTC(), CorrelationID: eventID, Producer: "tamo-media-service", Payload: payload})
}

func publicAsset(record database.MediaAsset) Asset {
	return Asset{ID: record.ID, Kind: record.Kind, Filename: record.OriginalFilename,
		ContentType: record.ContentType, SizeBytes: record.SizeBytes, Status: record.Status,
		DurationSeconds: record.DurationSeconds, Width: record.Width, Height: record.Height,
		CodecName: record.CodecName, FailureCode: record.FailureCode, ProcessedAt: record.ProcessedAt,
		Title: record.Title, Description: record.Description, Keywords: record.Keywords, Location: record.Location,
		UsageType: record.UsageType, Variants: []Variant{}, CreatedAt: record.CreatedAt, UpdatedAt: record.UpdatedAt}
}
