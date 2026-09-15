package commerce

import (
	"context"
	"errors"
	"strings"

	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/tamoimages/media-service/internal/payments"
	"github.com/tamoimages/media-service/internal/platform/database"
)

var ErrUnavailable = errors.New("asset or license unavailable")

type CartItem struct {
	MediaAssetID    string `json:"mediaAssetId"`
	LicenseCode     string `json:"licenseCode"`
	LicenseName     string `json:"licenseName"`
	UnitAmountMinor int64  `json:"unitAmountMinor"`
	Currency        string `json:"currency"`
}

type Cart struct {
	ID            string     `json:"id"`
	Items         []CartItem `json:"items"`
	SubtotalMinor int64      `json:"subtotalMinor"`
	Currency      string     `json:"currency"`
}
type Order struct {
	ID               string `json:"id"`
	Status           string `json:"status"`
	PaymentReference string `json:"paymentReference"`
	TotalAmountMinor int64  `json:"totalAmountMinor"`
	Currency         string `json:"currency"`
	CheckoutURL      string `json:"checkoutUrl,omitempty"`
}

type Repository interface {
	BuyerCart(context.Context, string) (database.BuyerCart, error)
	AddBuyerCartItem(context.Context, string, string, string) (database.BuyerCart, error)
	RemoveBuyerCartItem(context.Context, string, string) (database.BuyerCart, error)
	CreateOrderFromCart(context.Context, string, string, string, string) (database.CommerceOrder, error)
	SetOrderCheckoutURL(context.Context, string, string, string) (database.CommerceOrder, error)
	MarkOrderPaid(context.Context, string, string, int64, string) (bool, error)
}
type Gateway interface {
	Initialize(context.Context, string, int64, string, string, string) (payments.Checkout, error)
	ValidSignature([]byte, string) bool
}

type Service struct {
	repository  Repository
	gateway     Gateway
	callbackURL string
}

func New(repository Repository) *Service { return &Service{repository: repository} }
func NewCheckout(repository Repository, gateway Gateway, callbackURL string) *Service {
	return &Service{repository: repository, gateway: gateway, callbackURL: callbackURL}
}
func (s *Service) Get(ctx context.Context, buyerID string) (Cart, error) {
	record, err := s.repository.BuyerCart(ctx, buyerID)
	return public(record), err
}
func (s *Service) Checkout(ctx context.Context, buyerID, email, idempotencyKey string) (Order, error) {
	id := uuid.NewString()
	reference := "order-" + id
	record, err := s.repository.CreateOrderFromCart(ctx, id, buyerID, idempotencyKey, reference)
	if err != nil {
		return Order{}, err
	}
	if record.CheckoutURL != "" {
		return orderPublic(record), nil
	}
	checkout, err := s.gateway.Initialize(ctx, email, record.TotalAmountMinor, record.Currency, record.PaymentReference, s.callbackURL)
	if err != nil {
		return Order{}, err
	}
	if checkout.Reference != record.PaymentReference {
		return Order{}, payments.ErrRejected
	}
	record, err = s.repository.SetOrderCheckoutURL(ctx, record.ID, buyerID, checkout.AuthorizationURL)
	return orderPublic(record), err
}
func (s *Service) Webhook(ctx context.Context, body []byte, signature string) error {
	if s.gateway == nil || !s.gateway.ValidSignature(body, signature) {
		return payments.ErrRejected
	}
	var event struct {
		Event string `json:"event"`
		Data  struct {
			ID        json.Number `json:"id"`
			Reference string      `json:"reference"`
			Status    string      `json:"status"`
			Amount    int64       `json:"amount"`
			Currency  string      `json:"currency"`
		} `json:"data"`
	}
	if json.Unmarshal(body, &event) != nil {
		return payments.ErrRejected
	}
	if event.Event != "charge.success" || event.Data.Status != "success" {
		return nil
	}
	changed, err := s.repository.MarkOrderPaid(ctx, event.Data.ID.String(), event.Data.Reference, event.Data.Amount, event.Data.Currency)
	if err != nil {
		return err
	}
	if !changed {
		return fmt.Errorf("payment does not match order")
	}
	return nil
}
func orderPublic(record database.CommerceOrder) Order {
	return Order{ID: record.ID, Status: record.Status, PaymentReference: record.PaymentReference, TotalAmountMinor: record.TotalAmountMinor, Currency: record.Currency, CheckoutURL: record.CheckoutURL}
}
func (s *Service) Add(ctx context.Context, buyerID, assetID, licenseCode string) (Cart, error) {
	record, err := s.repository.AddBuyerCartItem(ctx, buyerID, assetID, strings.ToLower(strings.TrimSpace(licenseCode)))
	if errors.Is(err, pgx.ErrNoRows) {
		return Cart{}, ErrUnavailable
	}
	return public(record), err
}
func (s *Service) Remove(ctx context.Context, buyerID, assetID string) (Cart, error) {
	record, err := s.repository.RemoveBuyerCartItem(ctx, buyerID, assetID)
	return public(record), err
}
func public(record database.BuyerCart) Cart {
	items := make([]CartItem, 0, len(record.Items))
	var subtotal int64
	for _, item := range record.Items {
		items = append(items, CartItem{MediaAssetID: item.MediaAssetID, LicenseCode: item.LicenseCode, LicenseName: item.LicenseName, UnitAmountMinor: item.UnitAmountMinor, Currency: item.Currency})
		subtotal += item.UnitAmountMinor
	}
	currency := "NGN"
	if len(items) > 0 {
		currency = items[0].Currency
	}
	return Cart{ID: record.ID, Items: items, SubtotalMinor: subtotal, Currency: currency}
}
