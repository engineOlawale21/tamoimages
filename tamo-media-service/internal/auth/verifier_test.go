package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"
)

func token(t *testing.T, secret string, claims map[string]any) string {
	t.Helper()
	header, _ := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	payload, _ := json.Marshal(claims)
	unsigned := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(unsigned))
	return unsigned + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func TestHMACVerifierValidatesSignatureClaimsAndRole(t *testing.T) {
	verifier := NewHMACVerifier("test-secret-with-at-least-32-characters", "tamo-identity", "tamo-platform")
	raw := token(t, "test-secret-with-at-least-32-characters", map[string]any{"sub": "user-1", "role": "contributor", "iss": "tamo-identity", "aud": "tamo-platform", "exp": time.Now().Add(time.Minute).Unix()})
	principal, err := verifier.Verify(raw)
	if err != nil || principal.Subject != "user-1" || principal.Role != "contributor" {
		t.Fatalf("unexpected verification result %+v: %v", principal, err)
	}
}

func TestHMACVerifierRejectsTamperingAndExpiredTokens(t *testing.T) {
	verifier := NewHMACVerifier("test-secret-with-at-least-32-characters", "tamo-identity", "tamo-platform")
	expired := token(t, "test-secret-with-at-least-32-characters", map[string]any{"sub": "user-1", "role": "contributor", "iss": "tamo-identity", "aud": "tamo-platform", "exp": time.Now().Add(-time.Minute).Unix()})
	if _, err := verifier.Verify(expired); err == nil {
		t.Fatal("expected expired token rejection")
	}
	if _, err := verifier.Verify(expired + "x"); err == nil {
		t.Fatal("expected tampered token rejection")
	}
}
