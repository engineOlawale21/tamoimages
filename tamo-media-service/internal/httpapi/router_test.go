package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/tamoimages/media-service/internal/auth"
	"github.com/tamoimages/media-service/internal/batches"
	"github.com/tamoimages/media-service/internal/commerce"
	"github.com/tamoimages/media-service/internal/uploads"
)

type fakeVerifier struct{ principal auth.Principal }

func (f fakeVerifier) Verify(string) (auth.Principal, error) { return f.principal, nil }

type fakeUploads struct{ created uploads.CreateInput }
type fakeBatches struct{ createdName string }
type fakeCommerce struct{ licenseCode string }

func (f *fakeCommerce) Get(context.Context, string) (commerce.Cart, error) {
	return commerce.Cart{ID: "cart-1", Items: []commerce.CartItem{}, Currency: "NGN"}, nil
}
func (f *fakeCommerce) Add(_ context.Context, _ string, _ string, license string) (commerce.Cart, error) {
	f.licenseCode = license
	return commerce.Cart{ID: "cart-1", Currency: "NGN", SubtotalMinor: 1500000}, nil
}
func (f *fakeCommerce) Remove(context.Context, string, string) (commerce.Cart, error) {
	return commerce.Cart{ID: "cart-1", Items: []commerce.CartItem{}, Currency: "NGN"}, nil
}
func (f *fakeCommerce) Checkout(context.Context, string, string, string) (commerce.Order, error) {
	return commerce.Order{ID: "order-1", Status: "pending_payment", Currency: "NGN", TotalAmountMinor: 1500000}, nil
}
func (f *fakeCommerce) Webhook(context.Context, []byte, string) error { return nil }

func (f *fakeBatches) Create(_ context.Context, contributorID, name string) (batches.Batch, error) {
	f.createdName = name
	return batches.Batch{ID: "batch-1", Name: name, Status: "draft"}, nil
}
func (f *fakeBatches) List(context.Context, string) ([]batches.Batch, error) {
	return []batches.Batch{{ID: "batch-1", Name: "Lagos", Status: "draft"}}, nil
}
func (f *fakeBatches) AddItem(context.Context, string, string, string) error { return nil }
func (f *fakeBatches) Submit(context.Context, string, string) (batches.Batch, error) {
	return batches.Batch{ID: "batch-1", Status: "submitted", ItemCount: 1}, nil
}
func (f *fakeBatches) Review(_ context.Context, _ string, status, feedback string) (batches.Batch, error) {
	return batches.Batch{ID: "batch-1", Status: status, ReviewFeedback: feedback}, nil
}

func (f *fakeUploads) Create(_ context.Context, input uploads.CreateInput) (uploads.Session, error) {
	f.created = input
	now := time.Date(2026, 9, 7, 12, 0, 0, 0, time.UTC)
	return uploads.Session{Asset: uploads.Asset{ID: "asset-1", Kind: input.Kind, Filename: input.Filename, ContentType: input.ContentType, SizeBytes: input.SizeBytes, Status: "pending", CreatedAt: now, UpdatedAt: now}, UploadURL: "http://storage.test/upload", ExpiresAt: now.Add(15 * time.Minute)}, nil
}
func (f *fakeUploads) Get(_ context.Context, id, contributorID string) (uploads.Asset, error) {
	if id != "asset-1" || contributorID != "contributor-1" {
		return uploads.Asset{}, uploads.ErrNotFound
	}
	return uploads.Asset{ID: id, Status: "pending"}, nil
}
func (f *fakeUploads) List(_ context.Context, contributorID string) (uploads.Page, error) {
	if contributorID != "contributor-1" {
		return uploads.Page{Items: []uploads.Asset{}, Total: 0}, nil
	}
	return uploads.Page{Items: []uploads.Asset{{ID: "asset-1", Status: "ready", Variants: []uploads.Variant{}}}, Total: 1}, nil
}
func (f *fakeUploads) Complete(_ context.Context, id, contributorID string) (uploads.Asset, error) {
	if id != "asset-1" || contributorID != "contributor-1" {
		return uploads.Asset{}, uploads.ErrNotFound
	}
	return uploads.Asset{ID: id, Status: "uploaded"}, nil
}
func (f *fakeUploads) Retry(_ context.Context, id, contributorID string) (uploads.Asset, error) {
	if id != "asset-1" || contributorID != "contributor-1" {
		return uploads.Asset{}, uploads.ErrNotFound
	}
	return uploads.Asset{ID: id, Status: "uploaded", Variants: []uploads.Variant{}}, nil
}
func (f *fakeUploads) UpdateMetadata(_ context.Context, id, contributorID string, input uploads.MetadataInput) (uploads.Asset, error) {
	if id != "asset-1" || contributorID != "contributor-1" {
		return uploads.Asset{}, uploads.ErrNotFound
	}
	return uploads.Asset{ID: id, Status: "ready", Title: input.Title, Keywords: input.Keywords, UsageType: input.UsageType, Variants: []uploads.Variant{}}, nil
}
func (f *fakeUploads) Delete(_ context.Context, id, contributorID string) error {
	if id != "asset-1" || contributorID != "contributor-1" {
		return uploads.ErrNotFound
	}
	return nil
}

func testConfig() Config {
	return Config{WebOrigin: "http://localhost:3000", DatabaseURL: "postgresql://localhost:1/tamo_media", RedisAddress: "localhost:1", KafkaBrokers: []string{"localhost:1"}, S3Endpoint: "http://localhost:1", MaxBodyBytes: 1024}
}

func TestLivenessDoesNotRequireDependencies(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health/live", nil)
	res := httptest.NewRecorder()
	New(testConfig()).ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("wanted 200, got %d", res.Code)
	}
	if res.Header().Get("X-Correlation-ID") == "" {
		t.Fatal("expected a correlation ID")
	}
	if res.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatal("expected security headers")
	}
}

func TestReadinessFailsWhenDependenciesAreUnavailable(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health/ready", nil)
	res := httptest.NewRecorder()
	New(testConfig()).ServeHTTP(res, req)
	if res.Code != http.StatusServiceUnavailable {
		t.Fatalf("wanted 503, got %d", res.Code)
	}
}

func TestUploadSessionRequiresContributor(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/upload-sessions", strings.NewReader(`{"kind":"video","filename":"clip.mp4","contentType":"video/mp4","sizeBytes":500}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token")
	res := httptest.NewRecorder()
	cfg := testConfig()
	cfg.TokenVerifier = fakeVerifier{principal: auth.Principal{Subject: "buyer-1", Role: "buyer"}}
	New(cfg).ServeHTTP(res, req)
	if res.Code != http.StatusForbidden {
		t.Fatalf("wanted 403, got %d", res.Code)
	}
}

func TestBuyerAddsServerPricedCartItem(t *testing.T) {
	service := &fakeCommerce{}
	cfg := testConfig()
	cfg.Commerce = service
	cfg.TokenVerifier = fakeVerifier{principal: auth.Principal{Subject: "buyer-1", Role: "buyer"}}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cart/items", strings.NewReader(`{"mediaAssetId":"asset-1","licenseCode":"standard-image"}`))
	req.Header.Set("Authorization", "Bearer token")
	res := httptest.NewRecorder()
	New(cfg).ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("wanted 200, got %d: %s", res.Code, res.Body.String())
	}
	if service.licenseCode != "standard-image" {
		t.Fatalf("unexpected license %q", service.licenseCode)
	}
	var cart commerce.Cart
	if err := json.Unmarshal(res.Body.Bytes(), &cart); err != nil || cart.SubtotalMinor != 1500000 {
		t.Fatalf("unexpected cart: %+v %v", cart, err)
	}
}

func TestContributorCreatesPresignedUploadSession(t *testing.T) {
	service := &fakeUploads{}
	cfg := testConfig()
	cfg.Uploads = service
	cfg.TokenVerifier = fakeVerifier{principal: auth.Principal{Subject: "contributor-1", Role: "contributor"}}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/upload-sessions", strings.NewReader(`{"kind":"video","filename":"clip.mp4","contentType":"video/mp4","sizeBytes":500}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token")
	res := httptest.NewRecorder()
	New(cfg).ServeHTTP(res, req)
	if res.Code != http.StatusCreated {
		t.Fatalf("wanted 201, got %d: %s", res.Code, res.Body.String())
	}
	if service.created.ContributorID != "contributor-1" {
		t.Fatalf("unexpected owner %q", service.created.ContributorID)
	}
	var result struct {
		Asset  uploads.Asset                `json:"asset"`
		Upload struct{ Method, URL string } `json:"upload"`
	}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if result.Asset.Status != "pending" || result.Upload.Method != "PUT" || result.Upload.URL == "" {
		t.Fatalf("unexpected response %+v", result)
	}
}

func TestInvalidUploadUsesProblemDetails(t *testing.T) {
	cfg := testConfig()
	cfg.Uploads = &fakeUploads{}
	cfg.TokenVerifier = fakeVerifier{principal: auth.Principal{Subject: "contributor-1", Role: "contributor"}}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/upload-sessions", strings.NewReader(`{"kind":"document","filename":"private.pdf","contentType":"application/pdf","sizeBytes":10}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token")
	res := httptest.NewRecorder()
	New(cfg).ServeHTTP(res, req)
	if res.Code != http.StatusUnprocessableEntity {
		t.Fatalf("wanted 422, got %d", res.Code)
	}
	if res.Header().Get("Content-Type") != "application/problem+json" {
		t.Fatalf("unexpected content type %q", res.Header().Get("Content-Type"))
	}
	var result problem
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if result.CorrelationID == "" {
		t.Fatal("expected problem correlation ID")
	}
}

func TestMediaStatusIsScopedToContributor(t *testing.T) {
	cfg := testConfig()
	cfg.Uploads = &fakeUploads{}
	cfg.TokenVerifier = fakeVerifier{principal: auth.Principal{Subject: "other-contributor", Role: "contributor"}}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/media/asset-1", nil)
	req.Header.Set("Authorization", "Bearer token")
	res := httptest.NewRecorder()
	New(cfg).ServeHTTP(res, req)
	if res.Code != http.StatusNotFound {
		t.Fatalf("wanted ownership-safe 404, got %d", res.Code)
	}
}

func TestContributorListsOwnedMedia(t *testing.T) {
	cfg := testConfig()
	cfg.Uploads = &fakeUploads{}
	cfg.TokenVerifier = fakeVerifier{principal: auth.Principal{Subject: "contributor-1", Role: "contributor"}}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/media", nil)
	req.Header.Set("Authorization", "Bearer token")
	res := httptest.NewRecorder()
	New(cfg).ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("wanted 200, got %d", res.Code)
	}
	var page uploads.Page
	if err := json.NewDecoder(res.Body).Decode(&page); err != nil {
		t.Fatal(err)
	}
	if page.Total != 1 || page.Items[0].ID != "asset-1" {
		t.Fatalf("unexpected page %+v", page)
	}
}

func TestContributorCompletesOwnedUpload(t *testing.T) {
	cfg := testConfig()
	cfg.Uploads = &fakeUploads{}
	cfg.TokenVerifier = fakeVerifier{principal: auth.Principal{Subject: "contributor-1", Role: "contributor"}}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/upload-sessions/asset-1/complete", nil)
	req.Header.Set("Authorization", "Bearer token")
	res := httptest.NewRecorder()
	New(cfg).ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("wanted 200, got %d", res.Code)
	}
	var asset uploads.Asset
	if err := json.NewDecoder(res.Body).Decode(&asset); err != nil {
		t.Fatal(err)
	}
	if asset.Status != "uploaded" {
		t.Fatalf("unexpected status %q", asset.Status)
	}
}

func TestContributorRetriesOwnedFailedMedia(t *testing.T) {
	cfg := testConfig()
	cfg.Uploads = &fakeUploads{}
	cfg.TokenVerifier = fakeVerifier{principal: auth.Principal{Subject: "contributor-1", Role: "contributor"}}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/media/asset-1/retry", nil)
	req.Header.Set("Authorization", "Bearer token")
	res := httptest.NewRecorder()
	New(cfg).ServeHTTP(res, req)
	if res.Code != http.StatusAccepted {
		t.Fatalf("wanted 202, got %d", res.Code)
	}
}

func TestContributorUpdatesReadyMediaMetadata(t *testing.T) {
	cfg := testConfig()
	cfg.Uploads = &fakeUploads{}
	cfg.TokenVerifier = fakeVerifier{principal: auth.Principal{Subject: "contributor-1", Role: "contributor"}}
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/media/asset-1/metadata", strings.NewReader(`{"title":" Lagos Life ","description":"Street scene","keywords":["Lagos","Street"],"location":"Lagos","usageType":"creative"}`))
	req.Header.Set("Authorization", "Bearer token")
	res := httptest.NewRecorder()
	New(cfg).ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("wanted 200, got %d: %s", res.Code, res.Body.String())
	}
	var asset uploads.Asset
	if err := json.NewDecoder(res.Body).Decode(&asset); err != nil {
		t.Fatal(err)
	}
	if asset.Title != "Lagos Life" || asset.UsageType != "creative" || len(asset.Keywords) != 2 {
		t.Fatalf("unexpected metadata %+v", asset)
	}
}

func TestContributorCreatesAndListsBatch(t *testing.T) {
	service := &fakeBatches{}
	cfg := testConfig()
	cfg.Batches = service
	cfg.TokenVerifier = fakeVerifier{principal: auth.Principal{Subject: "contributor-1", Role: "contributor"}}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/batches", strings.NewReader(`{"name":"Lagos stories"}`))
	req.Header.Set("Authorization", "Bearer token")
	res := httptest.NewRecorder()
	New(cfg).ServeHTTP(res, req)
	if res.Code != http.StatusCreated || service.createdName != "Lagos stories" {
		t.Fatalf("unexpected create status=%d name=%q", res.Code, service.createdName)
	}
	req = httptest.NewRequest(http.MethodGet, "/api/v1/batches", nil)
	req.Header.Set("Authorization", "Bearer token")
	res = httptest.NewRecorder()
	New(cfg).ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("unexpected list status %d", res.Code)
	}
}

func TestPanicRecoveryHidesInternalDetails(t *testing.T) {
	handler := requestContext(recoverPanics(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("private implementation detail")
	})))
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusInternalServerError {
		t.Fatalf("wanted 500, got %d", res.Code)
	}
	if bytes.Contains(res.Body.Bytes(), []byte("private implementation detail")) {
		t.Fatal("panic detail leaked in response")
	}
}
