package commerce

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
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

type Repository interface {
	BuyerCart(context.Context, string) (database.BuyerCart, error)
	AddBuyerCartItem(context.Context, string, string, string) (database.BuyerCart, error)
	RemoveBuyerCartItem(context.Context, string, string) (database.BuyerCart, error)
}

type Service struct{ repository Repository }

func New(repository Repository) *Service { return &Service{repository: repository} }
func (s *Service) Get(ctx context.Context, buyerID string) (Cart, error) {
	record, err := s.repository.BuyerCart(ctx, buyerID)
	return public(record), err
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
