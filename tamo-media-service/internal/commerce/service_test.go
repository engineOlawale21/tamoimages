package commerce

import (
	"context"
	"github.com/tamoimages/media-service/internal/payments"
	"github.com/tamoimages/media-service/internal/platform/database"
	"testing"
)

type fakeRepository struct{ cart database.BuyerCart }
type fakeGateway struct {
	amount    int64
	reference string
}

func (f *fakeGateway) Initialize(_ context.Context, _ string, amount int64, _, reference, _ string) (payments.Checkout, error) {
	f.amount = amount
	f.reference = reference
	return payments.Checkout{AuthorizationURL: "https://checkout.test/start", Reference: reference}, nil
}
func (f *fakeGateway) ValidSignature([]byte, string) bool { return true }

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
func (f *fakeRepository) CreateOrderFromCart(_ context.Context, id, buyer, _, reference string) (database.CommerceOrder, error) {
	return database.CommerceOrder{ID: id, BuyerID: buyer, Status: "pending_payment", PaymentReference: reference, Currency: "NGN", TotalAmountMinor: 1500000}, nil
}
func (f *fakeRepository) SetOrderCheckoutURL(_ context.Context, id, buyer, url string) (database.CommerceOrder, error) {
	return database.CommerceOrder{ID: id, BuyerID: buyer, Status: "pending_payment", PaymentReference: "order-" + id, Currency: "NGN", TotalAmountMinor: 1500000, CheckoutURL: url}, nil
}
func TestCheckoutSendsImmutableServerTotal(t *testing.T) {
	repo := &fakeRepository{}
	gateway := &fakeGateway{}
	service := NewCheckout(repo, gateway, "https://app.test/cart")
	order, err := service.Checkout(context.Background(), "buyer-1", "buyer@example.test", "request-123")
	if err != nil || gateway.amount != 1500000 || gateway.reference != order.PaymentReference || order.CheckoutURL == "" {
		t.Fatalf("unexpected checkout %+v %v", order, err)
	}
}
func (f *fakeRepository) MarkOrderPaid(context.Context, string, string, int64, string) (bool, error) {
	return true, nil
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
