package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

var ErrInvalidToken = errors.New("invalid access token")

type Principal struct {
	Subject string
	Role    string
}

type Verifier interface {
	Verify(string) (Principal, error)
}

type HMACVerifier struct {
	secret   []byte
	issuer   string
	audience string
	now      func() time.Time
}

func NewHMACVerifier(secret, issuer, audience string) *HMACVerifier {
	return &HMACVerifier{secret: []byte(secret), issuer: issuer, audience: audience, now: time.Now}
}

func (v *HMACVerifier) Verify(raw string) (Principal, error) {
	parts := strings.Split(raw, ".")
	if len(parts) != 3 {
		return Principal{}, ErrInvalidToken
	}
	var header struct {
		Algorithm string `json:"alg"`
	}
	if !decode(parts[0], &header) || header.Algorithm != "HS256" {
		return Principal{}, ErrInvalidToken
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return Principal{}, ErrInvalidToken
	}
	mac := hmac.New(sha256.New, v.secret)
	_, _ = mac.Write([]byte(parts[0] + "." + parts[1]))
	if !hmac.Equal(signature, mac.Sum(nil)) {
		return Principal{}, ErrInvalidToken
	}
	var claims struct {
		Subject  string          `json:"sub"`
		Role     string          `json:"role"`
		Issuer   string          `json:"iss"`
		Audience json.RawMessage `json:"aud"`
		Expires  int64           `json:"exp"`
	}
	if !decode(parts[1], &claims) || claims.Subject == "" || claims.Expires <= v.now().Unix() || claims.Issuer != v.issuer || !hasAudience(claims.Audience, v.audience) {
		return Principal{}, ErrInvalidToken
	}
	return Principal{Subject: claims.Subject, Role: claims.Role}, nil
}

func decode(value string, target any) bool {
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	return err == nil && json.Unmarshal(decoded, target) == nil
}

func hasAudience(raw json.RawMessage, expected string) bool {
	var single string
	if json.Unmarshal(raw, &single) == nil {
		return single == expected
	}
	var multiple []string
	if json.Unmarshal(raw, &multiple) != nil {
		return false
	}
	for _, audience := range multiple {
		if audience == expected {
			return true
		}
	}
	return false
}
