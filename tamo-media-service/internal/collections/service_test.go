package collections

import (
	"context"
	"github.com/tamoimages/media-service/internal/platform/database"
	"testing"
)

type fakeRepository struct {
	collection database.BuyerCollection
	changed    bool
}

func (f *fakeRepository) CreateBuyerCollection(_ context.Context, id, buyer, name string) (database.BuyerCollection, error) {
	f.collection = database.BuyerCollection{ID: id, BuyerID: buyer, Name: name}
	return f.collection, nil
}
func (f *fakeRepository) BuyerCollections(context.Context, string) ([]database.BuyerCollection, error) {
	return []database.BuyerCollection{f.collection}, nil
}
func (f *fakeRepository) AddBuyerCollectionItem(context.Context, string, string, string) (bool, error) {
	f.changed = true
	return true, nil
}
func (f *fakeRepository) RemoveBuyerCollectionItem(context.Context, string, string, string) (bool, error) {
	return true, nil
}
func (f *fakeRepository) BuyerFavourites(context.Context, string) ([]string, error) {
	return []string{"asset-1"}, nil
}
func (f *fakeRepository) SetBuyerFavourite(context.Context, string, string, bool) (bool, error) {
	f.changed = true
	return true, nil
}
func TestBuyerCanCreateAndPopulateCollection(t *testing.T) {
	repo := &fakeRepository{}
	service := New(repo)
	collection, err := service.Create(context.Background(), "buyer-1", " Campaign ")
	if err != nil || collection.Name != "Campaign" {
		t.Fatalf("unexpected collection %+v %v", collection, err)
	}
	if err = service.Add(context.Background(), "buyer-1", collection.ID, "asset-1"); err != nil || !repo.changed {
		t.Fatal("item was not added")
	}
}
