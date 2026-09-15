package commerce

import (
	"context"
	"github.com/tamoimages/media-service/internal/platform/database"
	"testing"
)

type fakeRepository struct{ cart database.BuyerCart }

func (f *fakeRepository) BuyerCart(context.Context, string) (database.BuyerCart, error) {
	return f.cart, nil
}
func (f *fakeRepository) AddBuyerCartItem(_ context.Context, _ string, asset, license string) (database.BuyerCart, error) {
	f.cart.Items = []database.BuyerCartItem{{MediaAssetID: asset, LicenseCode: license, LicenseName: "Standard image license", UnitAmountMinor: 1500000, Currency: "NGN"}}
	return f.cart, nil
}
func (f *fakeRepository) RemoveBuyerCartItem(context.Context, string, string) (database.BuyerCart, error) {
	f.cart.Items = nil
	return f.cart, nil
}
func TestCartUsesRepositoryPrice(t *testing.T) {
	service := New(&fakeRepository{cart: database.BuyerCart{ID: "cart-1"}})
	cart, err := service.Add(context.Background(), "buyer-1", "asset-1", " STANDARD-IMAGE ")
	if err != nil || cart.SubtotalMinor != 1500000 || cart.Currency != "NGN" {
		t.Fatalf("unexpected cart: %+v %v", cart, err)
	}
	if cart.Items[0].LicenseCode != "standard-image" {
		t.Fatalf("license was not normalized: %+v", cart.Items[0])
	}
}
