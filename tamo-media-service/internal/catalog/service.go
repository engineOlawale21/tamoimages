package catalog

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/tamoimages/media-service/internal/platform/database"
)

var ErrNotFound = errors.New("catalog asset not found")

type Query struct {
	Text, Kind, UsageType, Orientation, Location, Sort string
	Page, PageSize                                     int
}

type Variant struct {
	Kind, URL, ContentType string
	Width, Height          int
	ExpiresAt              time.Time
}

type Asset struct {
	ID, Kind, Title, Description, Location, UsageType, Orientation string
	Keywords                                                       []string
	Width, Height                                                  int
	DurationSeconds                                                float64
	Variants                                                       []Variant
	CreatedAt                                                      time.Time
}

type Page struct {
	Items      []Asset `json:"items"`
	Page       int     `json:"page"`
	PageSize   int     `json:"pageSize"`
	Total      int     `json:"total"`
	TotalPages int     `json:"totalPages"`
}

type Repository interface {
	CatalogAssets(context.Context, database.CatalogQuery) ([]database.MediaAsset, int, error)
	CatalogAssetByID(context.Context, string) (database.MediaAsset, error)
	MediaVariantsByAssetID(context.Context, string) ([]database.MediaVariant, error)
}

type ObjectStore interface {
	PresignGet(context.Context, string, time.Duration) (*url.URL, error)
}

type Service struct {
	repository Repository
	storage    ObjectStore
	expiry     time.Duration
	now        func() time.Time
}

func New(repository Repository, storage ObjectStore, expiry time.Duration) *Service {
	return &Service{repository: repository, storage: storage, expiry: expiry, now: time.Now}
}

func (s *Service) Search(ctx context.Context, query Query) (Page, error) {
	records, total, err := s.repository.CatalogAssets(ctx, database.CatalogQuery{
		Text: query.Text, Kind: query.Kind, UsageType: query.UsageType, Orientation: query.Orientation,
		Location: query.Location, Sort: query.Sort, Page: query.Page, PageSize: query.PageSize,
	})
	if err != nil {
		return Page{}, err
	}
	items := make([]Asset, 0, len(records))
	for _, record := range records {
		asset, attachErr := s.publicAsset(ctx, record)
		if attachErr != nil {
			return Page{}, attachErr
		}
		items = append(items, asset)
	}
	pages := 0
	if total > 0 {
		pages = (total + query.PageSize - 1) / query.PageSize
	}
	return Page{Items: items, Page: query.Page, PageSize: query.PageSize, Total: total, TotalPages: pages}, nil
}

func (s *Service) Get(ctx context.Context, id string) (Asset, error) {
	record, err := s.repository.CatalogAssetByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Asset{}, ErrNotFound
	}
	if err != nil {
		return Asset{}, err
	}
	return s.publicAsset(ctx, record)
}

func (s *Service) publicAsset(ctx context.Context, record database.MediaAsset) (Asset, error) {
	asset := Asset{ID: record.ID, Kind: record.Kind, Title: record.Title, Description: record.Description,
		Keywords: record.Keywords, Location: record.Location, UsageType: record.UsageType, Width: record.Width,
		Height: record.Height, DurationSeconds: record.DurationSeconds, CreatedAt: record.CreatedAt}
	if record.Width > record.Height {
		asset.Orientation = "landscape"
	} else if record.Height > record.Width {
		asset.Orientation = "portrait"
	} else {
		asset.Orientation = "square"
	}
	variants, err := s.repository.MediaVariantsByAssetID(ctx, record.ID)
	if err != nil {
		return Asset{}, err
	}
	expiresAt := s.now().Add(s.expiry)
	for _, variant := range variants {
		signed, signErr := s.storage.PresignGet(ctx, variant.StorageKey, s.expiry)
		if signErr != nil {
			return Asset{}, fmt.Errorf("sign catalog variant: %w", signErr)
		}
		asset.Variants = append(asset.Variants, Variant{Kind: variant.Kind, URL: signed.String(), ContentType: variant.ContentType, Width: variant.Width, Height: variant.Height, ExpiresAt: expiresAt})
	}
	return asset, nil
}
