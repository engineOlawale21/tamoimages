package releases

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/tamoimages/media-service/internal/platform/database"
	"github.com/tamoimages/media-service/internal/platform/storage"
)

var ErrAssetNotFound = errors.New("media asset not found")
var ErrLimitReached = errors.New("media release limit reached")
var ErrReleaseNotFound = errors.New("media release not found")
var ErrObjectMismatch = errors.New("release object does not match")

type CreateInput struct {
	ContributorID, AssetID, ReleaseType, Filename, ContentType string
	SizeBytes                                                  int64
}
type Release struct {
	ID          string    `json:"id"`
	AssetID     string    `json:"assetId"`
	ReleaseType string    `json:"releaseType"`
	Filename    string    `json:"filename"`
	ContentType string    `json:"contentType"`
	Status      string    `json:"status"`
	SizeBytes   int64     `json:"sizeBytes"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
type Session struct {
	Release   Release   `json:"release"`
	UploadURL string    `json:"-"`
	ExpiresAt time.Time `json:"-"`
}
type Repository interface {
	CreateMediaRelease(context.Context, database.MediaRelease) (database.MediaRelease, bool, error)
	MediaReleaseByID(context.Context, string, string, string) (database.MediaRelease, error)
	MarkMediaReleaseUploaded(context.Context, string, string, string) (database.MediaRelease, error)
	MediaReleasesByAsset(context.Context, string, string) ([]database.MediaRelease, error)
}
type ObjectStore interface {
	PresignPut(context.Context, string, time.Duration) (*url.URL, error)
	Head(context.Context, string) (storage.ObjectInfo, error)
}
type Service struct {
	repository Repository
	storage    ObjectStore
	expiry     time.Duration
	now        func() time.Time
}

func New(repository Repository, objectStore ObjectStore, expiry time.Duration) *Service {
	return &Service{repository: repository, storage: objectStore, expiry: expiry, now: time.Now}
}
func (s *Service) Create(ctx context.Context, input CreateInput) (Session, error) {
	id := uuid.NewString()
	expiresAt := s.now().Add(s.expiry)
	filename := path.Base(strings.ReplaceAll(strings.TrimSpace(input.Filename), `\`, "/"))
	key := fmt.Sprintf("releases/%s/%s/%s/%s", input.ContributorID, input.AssetID, id, filename)
	record, limit, err := s.repository.CreateMediaRelease(ctx, database.MediaRelease{ID: id, MediaAssetID: input.AssetID, ContributorID: input.ContributorID, ReleaseType: input.ReleaseType, Filename: filename, StorageKey: key, ContentType: input.ContentType, SizeBytes: input.SizeBytes, UploadExpiresAt: expiresAt})
	if errors.Is(err, pgx.ErrNoRows) {
		return Session{}, ErrAssetNotFound
	}
	if limit {
		return Session{}, ErrLimitReached
	}
	if err != nil {
		return Session{}, err
	}
	uploadURL, err := s.storage.PresignPut(ctx, key, s.expiry)
	if err != nil {
		return Session{}, fmt.Errorf("presign release upload: %w", err)
	}
	return Session{Release: publicRelease(record), UploadURL: uploadURL.String(), ExpiresAt: expiresAt}, nil
}
func (s *Service) Complete(ctx context.Context, assetID, releaseID, contributorID string) (Release, error) {
	record, err := s.repository.MediaReleaseByID(ctx, assetID, releaseID, contributorID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Release{}, ErrReleaseNotFound
	}
	if err != nil {
		return Release{}, err
	}
	if record.Status == "uploaded" {
		return publicRelease(record), nil
	}
	object, err := s.storage.Head(ctx, record.StorageKey)
	if err != nil || object.Size != record.SizeBytes {
		return Release{}, ErrObjectMismatch
	}
	record, err = s.repository.MarkMediaReleaseUploaded(ctx, assetID, releaseID, contributorID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Release{}, ErrReleaseNotFound
	}
	return publicRelease(record), err
}
func (s *Service) List(ctx context.Context, assetID, contributorID string) ([]Release, error) {
	records, err := s.repository.MediaReleasesByAsset(ctx, assetID, contributorID)
	if err != nil {
		return nil, err
	}
	result := make([]Release, 0, len(records))
	for _, record := range records {
		result = append(result, publicRelease(record))
	}
	return result, nil
}
func publicRelease(record database.MediaRelease) Release {
	return Release{ID: record.ID, AssetID: record.MediaAssetID, ReleaseType: record.ReleaseType, Filename: record.Filename, ContentType: record.ContentType, SizeBytes: record.SizeBytes, Status: record.Status, CreatedAt: record.CreatedAt, UpdatedAt: record.UpdatedAt}
}
