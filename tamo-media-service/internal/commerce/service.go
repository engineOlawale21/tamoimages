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
	"net/url"
	"time"
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
type Purchase struct {
	ID               string     `json:"id"`
	Status           string     `json:"status"`
	PaymentReference string     `json:"paymentReference"`
	TotalAmountMinor int64      `json:"totalAmountMinor"`
	Currency         string     `json:"currency"`
	CreatedAt        time.Time  `json:"createdAt"`
	PaidAt           *time.Time `json:"paidAt,omitempty"`
	EntitlementIDs   []string   `json:"entitlementIds"`
}
type Download struct {
	URL         string    `json:"url"`
	Filename    string    `json:"filename"`
	ContentType string    `json:"contentType"`
	LicenseCode string    `json:"licenseCode"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

type Repository interface {
	BuyerCart(context.Context, string) (database.BuyerCart, error)
	AddBuyerCartItem(context.Context, string, string, string) (database.BuyerCart, error)
	RemoveBuyerCartItem(context.Context, string, string) (database.BuyerCart, error)
	CreateOrderFromCart(context.Context, string, string, string, string) (database.CommerceOrder, error)
	SetOrderCheckoutURL(context.Context, string, string, string) (database.CommerceOrder, error)
	MarkOrderPaid(context.Context, string, string, int64, string) (bool, error)
	BuyerPurchases(context.Context, string) ([]database.Purchase, error)
	LicensedDownload(context.Context, string, string, string) (database.LicensedDownload, error)
}
type Signer interface {
	PresignGet(context.Context, string, time.Duration) (*url.URL, error)
}
type Gateway interface {
	Initialize(context.Context, string, int64, string, string, string) (payments.Checkout, error)
	ValidSignature([]byte, string) bool
}

type Service struct {
	repository  Repository
	gateway     Gateway
	callbackURL string
	signer      Signer
	downloadTTL time.Duration
}

func New(repository Repository) *Service { return &Service{repository: repository} }
func NewCheckout(repository Repository, gateway Gateway, callbackURL string, signer Signer, downloadTTL time.Duration) *Service {
	return &Service{repository: repository, gateway: gateway, callbackURL: callbackURL, signer: signer, downloadTTL: downloadTTL}
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
func (s *Service) Purchases(ctx context.Context, buyerID string) ([]Purchase, error) {
	records, err := s.repository.BuyerPurchases(ctx, buyerID)
	if err != nil {
		return nil, err
	}
	items := make([]Purchase, 0, len(records))
	for _, r := range records {
		items = append(items, Purchase{ID: r.ID, Status: r.Status, PaymentReference: r.PaymentReference, TotalAmountMinor: r.TotalAmountMinor, Currency: r.Currency, CreatedAt: r.CreatedAt, PaidAt: r.PaidAt, EntitlementIDs: r.EntitlementIDs})
	}
	return items, nil
}
func (s *Service) Download(ctx context.Context, buyerID, entitlementID, correlationID string) (Download, error) {
	record, err := s.repository.LicensedDownload(ctx, buyerID, entitlementID, correlationID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Download{}, ErrUnavailable
	}
	if err != nil {
		return Download{}, err
	}
	signed, err := s.signer.PresignGet(ctx, record.StorageKey, s.downloadTTL)
	if err != nil {
		return Download{}, err
	}
	return Download{URL: signed.String(), Filename: record.Filename, ContentType: record.ContentType, LicenseCode: record.LicenseCode, ExpiresAt: time.Now().UTC().Add(s.downloadTTL)}, nil
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
