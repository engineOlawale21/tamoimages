package collections

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/tamoimages/media-service/internal/platform/database"
)

var ErrNotFound = errors.New("collection not found")
var ErrAssetUnavailable = errors.New("approved media asset not found")

type Collection struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	ItemCount int      `json:"itemCount"`
	MediaIDs  []string `json:"mediaIds"`
}
type Repository interface {
	CreateBuyerCollection(context.Context, string, string, string) (database.BuyerCollection, error)
	BuyerCollections(context.Context, string) ([]database.BuyerCollection, error)
	AddBuyerCollectionItem(context.Context, string, string, string) (bool, error)
	RemoveBuyerCollectionItem(context.Context, string, string, string) (bool, error)
	BuyerFavourites(context.Context, string) ([]string, error)
	SetBuyerFavourite(context.Context, string, string, bool) (bool, error)
}
type Service struct{ repository Repository }

func New(repository Repository) *Service { return &Service{repository: repository} }
func (s *Service) Create(ctx context.Context, buyerID, name string) (Collection, error) {
	record, err := s.repository.CreateBuyerCollection(ctx, uuid.NewString(), buyerID, strings.TrimSpace(name))
	return public(record), err
}
func (s *Service) List(ctx context.Context, buyerID string) ([]Collection, error) {
	records, err := s.repository.BuyerCollections(ctx, buyerID)
	if err != nil {
		return nil, err
	}
	items := make([]Collection, 0, len(records))
	for _, record := range records {
		items = append(items, public(record))
	}
	return items, nil
}
func (s *Service) Add(ctx context.Context, buyerID, collectionID, assetID string) error {
	return mapped(s.repository.AddBuyerCollectionItem(ctx, buyerID, collectionID, assetID))
}
func (s *Service) Remove(ctx context.Context, buyerID, collectionID, assetID string) error {
	return mapped(s.repository.RemoveBuyerCollectionItem(ctx, buyerID, collectionID, assetID))
}
func (s *Service) Favourites(ctx context.Context, buyerID string) ([]string, error) {
	return s.repository.BuyerFavourites(ctx, buyerID)
}
func (s *Service) Favourite(ctx context.Context, buyerID, assetID string, value bool) error {
	ok, err := s.repository.SetBuyerFavourite(ctx, buyerID, assetID, value)
	if err != nil {
		return err
	}
	if !ok {
		return ErrAssetUnavailable
	}
	return nil
}
func public(record database.BuyerCollection) Collection {
	return Collection{ID: record.ID, Name: record.Name, ItemCount: len(record.MediaIDs), MediaIDs: record.MediaIDs}
}
func mapped(ok bool, err error) error {
	if errors.Is(err, pgx.ErrNoRows) || !ok {
		return ErrNotFound
	}
	return err
}
