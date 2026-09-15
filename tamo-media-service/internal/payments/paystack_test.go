package payments

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPaystackInitializesServerAmountAndVerifiesSignature(t *testing.T) {
	var authorization string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization = r.Header.Get("Authorization")
		w.Header().Set("content-type", "application/json")
		_, _ = w.Write([]byte(`{"status":true,"data":{"authorization_url":"https://checkout.paystack.test/start","reference":"order-1"}}`))
	}))
	defer server.Close()
	client := NewPaystack("secret", server.URL, time.Second)
	checkout, err := client.Initialize(context.Background(), "buyer@example.test", 1500000, "NGN", "order-1", "https://app.test/checkout/return")
	if err != nil || checkout.Reference != "order-1" || authorization != "Bearer secret" {
		t.Fatalf("unexpected checkout %+v %v", checkout, err)
	}
	body := []byte(`{"event":"charge.success"}`)
	mac := hmac.New(sha512.New, []byte("secret"))
	_, _ = mac.Write(body)
	if !client.ValidSignature(body, hex.EncodeToString(mac.Sum(nil))) || client.ValidSignature(body, "bad") {
		t.Fatal("signature validation failed")
	}
}
