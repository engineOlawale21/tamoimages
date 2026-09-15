package payments

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

var ErrRejected = errors.New("payment provider rejected request")

type Checkout struct {
	AuthorizationURL string `json:"authorizationUrl"`
	Reference        string `json:"reference"`
}
type Paystack struct {
	secret, endpoint string
	client           *http.Client
}

func NewPaystack(secret, endpoint string, timeout time.Duration) *Paystack {
	return &Paystack{secret: secret, endpoint: endpoint, client: &http.Client{Timeout: timeout}}
}
func (p *Paystack) Initialize(ctx context.Context, email string, amount int64, currency, reference, callbackURL string) (Checkout, error) {
	body, _ := json.Marshal(map[string]string{"email": email, "amount": strconv.FormatInt(amount, 10), "currency": currency, "reference": reference, "callback_url": callbackURL})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint+"/transaction/initialize", bytes.NewReader(body))
	if err != nil {
		return Checkout{}, err
	}
	req.Header.Set("Authorization", "Bearer "+p.secret)
	req.Header.Set("Content-Type", "application/json")
	res, err := p.client.Do(req)
	if err != nil {
		return Checkout{}, fmt.Errorf("initialize payment: %w", err)
	}
	defer res.Body.Close()
	var result struct {
		Status bool `json:"status"`
		Data   struct {
			AuthorizationURL string `json:"authorization_url"`
			Reference        string `json:"reference"`
		} `json:"data"`
	}
	if json.NewDecoder(res.Body).Decode(&result) != nil || res.StatusCode != http.StatusOK || !result.Status {
		return Checkout{}, ErrRejected
	}
	parsed, err := url.ParseRequestURI(result.Data.AuthorizationURL)
	if err != nil || parsed.Scheme != "https" {
		return Checkout{}, ErrRejected
	}
	return Checkout{AuthorizationURL: result.Data.AuthorizationURL, Reference: result.Data.Reference}, nil
}
func (p *Paystack) ValidSignature(body []byte, signature string) bool {
	provided, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}
	mac := hmac.New(sha512.New, []byte(p.secret))
	_, _ = mac.Write(body)
	return hmac.Equal(provided, mac.Sum(nil))
}
