package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/tamoimages/media-service/internal/auth"
	"github.com/tamoimages/media-service/internal/batches"
	"github.com/tamoimages/media-service/internal/catalog"
	"github.com/tamoimages/media-service/internal/collections"
	"github.com/tamoimages/media-service/internal/releases"
	"github.com/tamoimages/media-service/internal/uploads"
)

type Config struct {
	WebOrigin     string
	DatabaseURL   string
	RedisAddress  string
	KafkaBrokers  []string
	S3Endpoint    string
	MaxBodyBytes  int64
	Uploads       uploadService
	Batches       batchService
	Releases      releaseService
	Catalog       catalogService
	Collections   collectionService
	TokenVerifier auth.Verifier
}
type collectionService interface {
	Create(context.Context, string, string) (collections.Collection, error)
	List(context.Context, string) ([]collections.Collection, error)
	Add(context.Context, string, string, string) error
	Remove(context.Context, string, string, string) error
	Favourites(context.Context, string) ([]string, error)
	Favourite(context.Context, string, string, bool) error
}
type catalogService interface {
	Search(context.Context, catalog.Query) (catalog.Page, error)
	Get(context.Context, string) (catalog.Asset, error)
}
type releaseService interface {
	Create(context.Context, releases.CreateInput) (releases.Session, error)
	Complete(context.Context, string, string, string) (releases.Release, error)
	List(context.Context, string, string) ([]releases.Release, error)
}

type batchService interface {
	Create(context.Context, string, string) (batches.Batch, error)
	List(context.Context, string) ([]batches.Batch, error)
	AddItem(context.Context, string, string, string) error
	Submit(context.Context, string, string) (batches.Batch, error)
	Review(context.Context, string, string, string) (batches.Batch, error)
}

type uploadService interface {
	Create(context.Context, uploads.CreateInput) (uploads.Session, error)
	List(context.Context, string) (uploads.Page, error)
	Get(context.Context, string, string) (uploads.Asset, error)
	Complete(context.Context, string, string) (uploads.Asset, error)
	Retry(context.Context, string, string) (uploads.Asset, error)
	UpdateMetadata(context.Context, string, string, uploads.MetadataInput) (uploads.Asset, error)
	Delete(context.Context, string, string) error
}

type createUploadRequest struct {
	Kind        string `json:"kind"`
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
	SizeBytes   int64  `json:"sizeBytes"`
}
type createBatchRequest struct {
	Name string `json:"name"`
}
type addBatchItemRequest struct {
	AssetID string `json:"assetId"`
}
type updateMetadataRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Keywords    []string `json:"keywords"`
	Location    string   `json:"location"`
	UsageType   string   `json:"usageType"`
}
type createReleaseRequest struct {
	ReleaseType string `json:"releaseType"`
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
	SizeBytes   int64  `json:"sizeBytes"`
}
type reviewBatchRequest struct {
	Status   string `json:"status"`
	Feedback string `json:"feedback"`
}
type createCollectionRequest struct {
	Name string `json:"name"`
}
type problem struct {
	Type          string `json:"type"`
	Title         string `json:"title"`
	Status        int    `json:"status"`
	Detail        string `json:"detail"`
	Instance      string `json:"instance"`
	CorrelationID string `json:"correlationId"`
}
type contextKey string

const correlationIDKey contextKey = "correlation-id"

var validCorrelationID = regexp.MustCompile(`^[A-Za-z0-9._:-]{1,128}$`)

func New(cfg Config) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health/live", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]string{"status": "alive", "service": "tamo-media-service"}, http.StatusOK)
	})
	mux.HandleFunc("GET /api/v1/health/ready", func(w http.ResponseWriter, r *http.Request) {
		checks := readinessChecks(r.Context(), cfg)
		status, state := http.StatusOK, "ready"
		for _, check := range checks {
			if check.Status != "up" {
				status, state = http.StatusServiceUnavailable, "not-ready"
			}
		}
		writeJSON(w, map[string]any{"status": state, "checks": checks}, status)
	})
	mux.HandleFunc("GET /api/v1/catalog", func(w http.ResponseWriter, r *http.Request) {
		if cfg.Catalog == nil {
			writeProblem(w, r, http.StatusServiceUnavailable, "Catalog unavailable", "The public catalog is not configured.")
			return
		}
		query, valid := catalogQuery(r.URL.Query())
		if !valid {
			writeProblem(w, r, http.StatusUnprocessableEntity, "Invalid catalog filters", "Use supported media, usage, orientation, sorting, and pagination values.")
			return
		}
		page, err := cfg.Catalog.Search(r.Context(), query)
		if err != nil {
			writeProblem(w, r, http.StatusServiceUnavailable, "Catalog unavailable", "The catalog could not be searched.")
			return
		}
		writeJSON(w, page, http.StatusOK)
	})
	mux.HandleFunc("GET /api/v1/catalog/{id}", func(w http.ResponseWriter, r *http.Request) {
		if cfg.Catalog == nil {
			writeProblem(w, r, http.StatusServiceUnavailable, "Catalog unavailable", "The public catalog is not configured.")
			return
		}
		asset, err := cfg.Catalog.Get(r.Context(), r.PathValue("id"))
		if errors.Is(err, catalog.ErrNotFound) {
			writeProblem(w, r, http.StatusNotFound, "Media not found", "The approved media asset does not exist.")
			return
		}
		if err != nil {
			writeProblem(w, r, http.StatusServiceUnavailable, "Catalog unavailable", "The media asset could not be retrieved.")
			return
		}
		writeJSON(w, asset, http.StatusOK)
	})
	mux.Handle("GET /api/v1/media", contributorOnly(cfg.TokenVerifier, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page, err := cfg.Uploads.List(r.Context(), principalFrom(r.Context()).Subject)
		if err != nil {
			writeProblem(w, r, http.StatusServiceUnavailable, "Media unavailable", "The media library could not be retrieved.")
			return
		}
		writeJSON(w, page, http.StatusOK)
	})))
	mux.Handle("POST /api/v1/upload-sessions", contributorOnly(cfg.TokenVerifier, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, cfg.MaxBodyBytes)
		if !strings.HasPrefix(strings.ToLower(r.Header.Get("Content-Type")), "application/json") {
			writeProblem(w, r, http.StatusUnsupportedMediaType, "Unsupported media type", "Content-Type must be application/json.")
			return
		}
		var input createUploadRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil {
			writeProblem(w, r, http.StatusBadRequest, "Invalid upload session", "The request body must be valid JSON with supported fields.")
			return
		}
		input.Kind = strings.ToLower(strings.TrimSpace(input.Kind))
		input.Filename = strings.TrimSpace(input.Filename)
		input.ContentType = strings.ToLower(strings.TrimSpace(input.ContentType))
		if input.Kind != "image" && input.Kind != "video" && input.Kind != "illustration" {
			writeProblem(w, r, http.StatusUnprocessableEntity, "Invalid media kind", "Kind must be image, video, or illustration.")
			return
		}
		if input.Filename == "" || len(input.Filename) > 255 || input.ContentType == "" || input.SizeBytes < 1 || input.SizeBytes > cfg.MaxBodyBytes {
			writeProblem(w, r, http.StatusUnprocessableEntity, "Invalid upload metadata", "Filename, content type, and an allowed positive file size are required.")
			return
		}
		if !validContentType(input.Kind, input.ContentType) {
			writeProblem(w, r, http.StatusUnprocessableEntity, "Invalid content type", "The declared content type is not supported for this media kind.")
			return
		}
		principal := principalFrom(r.Context())
		session, err := cfg.Uploads.Create(r.Context(), uploads.CreateInput{ContributorID: principal.Subject, Kind: input.Kind, Filename: input.Filename, ContentType: input.ContentType, SizeBytes: input.SizeBytes})
		if err != nil {
			writeProblem(w, r, http.StatusServiceUnavailable, "Upload unavailable", "The upload session could not be created.")
			return
		}
		writeJSON(w, map[string]any{"asset": session.Asset, "upload": map[string]any{"method": "PUT", "url": session.UploadURL, "expiresAt": session.ExpiresAt}}, http.StatusCreated)
	})))
	mux.Handle("GET /api/v1/media/{id}", contributorOnly(cfg.TokenVerifier, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		asset, err := cfg.Uploads.Get(r.Context(), r.PathValue("id"), principalFrom(r.Context()).Subject)
		if errors.Is(err, uploads.ErrNotFound) {
			writeProblem(w, r, http.StatusNotFound, "Media not found", "The media asset does not exist.")
			return
		}
		if err != nil {
			writeProblem(w, r, http.StatusServiceUnavailable, "Media unavailable", "The media asset could not be retrieved.")
			return
		}
		writeJSON(w, asset, http.StatusOK)
	})))
	mux.Handle("DELETE /api/v1/media/{id}", contributorOnly(cfg.TokenVerifier, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := cfg.Uploads.Delete(r.Context(), r.PathValue("id"), principalFrom(r.Context()).Subject)
		if errors.Is(err, uploads.ErrNotFound) {
			writeProblem(w, r, http.StatusNotFound, "Media not found", "The media asset does not exist.")
			return
		}
		if err != nil {
			writeProblem(w, r, http.StatusServiceUnavailable, "Deletion unavailable", "Media deletion could not be completed.")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})))
	mux.Handle("POST /api/v1/upload-sessions/{id}/complete", contributorOnly(cfg.TokenVerifier, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		asset, err := cfg.Uploads.Complete(r.Context(), r.PathValue("id"), principalFrom(r.Context()).Subject)
		if errors.Is(err, uploads.ErrNotFound) {
			writeProblem(w, r, http.StatusNotFound, "Upload not found", "The upload session does not exist.")
			return
		}
		if errors.Is(err, uploads.ErrObjectMismatch) {
			writeProblem(w, r, http.StatusConflict, "Upload incomplete", "The uploaded object is missing or does not match the declared size.")
			return
		}
		if err != nil {
			writeProblem(w, r, http.StatusServiceUnavailable, "Upload unavailable", "The upload could not be completed.")
			return
		}
		writeJSON(w, asset, http.StatusOK)
	})))
	mux.Handle("POST /api/v1/media/{id}/retry", contributorOnly(cfg.TokenVerifier, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		asset, err := cfg.Uploads.Retry(r.Context(), r.PathValue("id"), principalFrom(r.Context()).Subject)
		if errors.Is(err, uploads.ErrNotFound) {
			writeProblem(w, r, http.StatusNotFound, "Media not found", "The media asset does not exist.")
			return
		}
		if errors.Is(err, uploads.ErrNotRetryable) {
			writeProblem(w, r, http.StatusConflict, "Media is not retryable", "Only failed media can be retried.")
			return
		}
		if err != nil {
			writeProblem(w, r, http.StatusServiceUnavailable, "Retry unavailable", "Media processing could not be retried.")
			return
		}
		writeJSON(w, asset, http.StatusAccepted)
	})))
	mux.Handle("PATCH /api/v1/media/{id}/metadata", contributorOnly(cfg.TokenVerifier, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, cfg.MaxBodyBytes)
		var input updateMetadataRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil {
			writeProblem(w, r, http.StatusBadRequest, "Invalid metadata", "The request body must contain supported metadata fields.")
			return
		}
		input.Title = strings.TrimSpace(input.Title)
		input.Description = strings.TrimSpace(input.Description)
		input.Location = strings.TrimSpace(input.Location)
		input.UsageType = strings.ToLower(strings.TrimSpace(input.UsageType))
		keywords := make([]string, 0, len(input.Keywords))
		seen := map[string]bool{}
		valid := len(input.Title) > 0 && len(input.Title) <= 160 && len(input.Description) <= 2000 && len(input.Location) <= 200 && (input.UsageType == "creative" || input.UsageType == "editorial") && len(input.Keywords) <= 50
		for _, keyword := range input.Keywords {
			keyword = strings.ToLower(strings.TrimSpace(keyword))
			if keyword == "" || len(keyword) > 64 {
				valid = false
				continue
			}
			if !seen[keyword] {
				seen[keyword] = true
				keywords = append(keywords, keyword)
			}
		}
		if !valid || len(keywords) == 0 {
			writeProblem(w, r, http.StatusUnprocessableEntity, "Invalid metadata", "Title, classification, and at least one valid keyword are required.")
			return
		}
		asset, err := cfg.Uploads.UpdateMetadata(r.Context(), r.PathValue("id"), principalFrom(r.Context()).Subject, uploads.MetadataInput{Title: input.Title, Description: input.Description, Keywords: keywords, Location: input.Location, UsageType: input.UsageType})
		if errors.Is(err, uploads.ErrNotFound) {
			writeProblem(w, r, http.StatusNotFound, "Media not found", "Only owned, ready media can be edited.")
			return
		}
		if err != nil {
			writeProblem(w, r, http.StatusServiceUnavailable, "Metadata unavailable", "Media metadata could not be saved.")
			return
		}
		writeJSON(w, asset, http.StatusOK)
	})))
	mux.Handle("GET /api/v1/batches", contributorOnly(cfg.TokenVerifier, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		items, err := cfg.Batches.List(r.Context(), principalFrom(r.Context()).Subject)
		if err != nil {
			writeProblem(w, r, http.StatusServiceUnavailable, "Batches unavailable", "Batches could not be retrieved.")
			return
		}
		writeJSON(w, map[string]any{"items": items, "total": len(items)}, http.StatusOK)
	})))
	mux.Handle("POST /api/v1/batches", contributorOnly(cfg.TokenVerifier, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, cfg.MaxBodyBytes)
		var input createBatchRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil || len(strings.TrimSpace(input.Name)) < 1 || len(strings.TrimSpace(input.Name)) > 120 {
			writeProblem(w, r, http.StatusUnprocessableEntity, "Invalid batch", "A batch name between 1 and 120 characters is required.")
			return
		}
		batch, err := cfg.Batches.Create(r.Context(), principalFrom(r.Context()).Subject, input.Name)
		if err != nil {
			writeProblem(w, r, http.StatusServiceUnavailable, "Batch unavailable", "The batch could not be created.")
			return
		}
		writeJSON(w, batch, http.StatusCreated)
	})))
	mux.Handle("POST /api/v1/batches/{id}/items", contributorOnly(cfg.TokenVerifier, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input addBatchItemRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil || strings.TrimSpace(input.AssetID) == "" {
			writeProblem(w, r, http.StatusUnprocessableEntity, "Invalid batch item", "A media asset ID is required.")
			return
		}
		err := cfg.Batches.AddItem(r.Context(), r.PathValue("id"), principalFrom(r.Context()).Subject, input.AssetID)
		if errors.Is(err, batches.ErrAssetUnavailable) {
			writeProblem(w, r, http.StatusConflict, "Media unavailable", "Only owned, ready, unassigned media can be added to a draft batch.")
			return
		}
		if err != nil {
			writeProblem(w, r, http.StatusServiceUnavailable, "Batch unavailable", "The media could not be added to the batch.")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})))
	mux.Handle("POST /api/v1/batches/{id}/submit", contributorOnly(cfg.TokenVerifier, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		batch, err := cfg.Batches.Submit(r.Context(), r.PathValue("id"), principalFrom(r.Context()).Subject)
		if errors.Is(err, batches.ErrNotSubmittable) {
			writeProblem(w, r, http.StatusConflict, "Batch is not submittable", "The batch must be a non-empty draft.")
			return
		}
		if err != nil {
			writeProblem(w, r, http.StatusServiceUnavailable, "Batch unavailable", "The batch could not be submitted.")
			return
		}
		writeJSON(w, batch, http.StatusOK)
	})))
	mux.Handle("POST /api/v1/media/{id}/releases", contributorOnly(cfg.TokenVerifier, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, cfg.MaxBodyBytes)
		var input createReleaseRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil {
			writeProblem(w, r, http.StatusBadRequest, "Invalid release", "The request must contain supported release fields.")
			return
		}
		input.ReleaseType = strings.ToLower(strings.TrimSpace(input.ReleaseType))
		input.Filename = strings.TrimSpace(input.Filename)
		input.ContentType = strings.ToLower(strings.TrimSpace(input.ContentType))
		validType := input.ReleaseType == "model" || input.ReleaseType == "property"
		validContent := input.ContentType == "application/pdf" || input.ContentType == "image/jpeg" || input.ContentType == "image/png"
		if !validType || !validContent || input.Filename == "" || len(input.Filename) > 255 || input.SizeBytes < 1 || input.SizeBytes > cfg.MaxBodyBytes {
			writeProblem(w, r, http.StatusUnprocessableEntity, "Invalid release", "A valid model/property PDF, JPEG, or PNG release is required.")
			return
		}
		principal := principalFrom(r.Context())
		session, err := cfg.Releases.Create(r.Context(), releases.CreateInput{ContributorID: principal.Subject, AssetID: r.PathValue("id"), ReleaseType: input.ReleaseType, Filename: input.Filename, ContentType: input.ContentType, SizeBytes: input.SizeBytes})
		if errors.Is(err, releases.ErrAssetNotFound) {
			writeProblem(w, r, http.StatusNotFound, "Media not found", "Only owned, ready media can receive releases.")
			return
		}
		if errors.Is(err, releases.ErrLimitReached) {
			writeProblem(w, r, http.StatusConflict, "Release limit reached", "A media asset can have at most three release forms.")
			return
		}
		if err != nil {
			writeProblem(w, r, http.StatusServiceUnavailable, "Release unavailable", "The release upload could not be created.")
			return
		}
		writeJSON(w, map[string]any{"release": session.Release, "upload": map[string]any{"method": "PUT", "url": session.UploadURL, "expiresAt": session.ExpiresAt}}, http.StatusCreated)
	})))
	mux.Handle("GET /api/v1/media/{id}/releases", contributorOnly(cfg.TokenVerifier, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		items, err := cfg.Releases.List(r.Context(), r.PathValue("id"), principalFrom(r.Context()).Subject)
		if err != nil {
			writeProblem(w, r, http.StatusServiceUnavailable, "Releases unavailable", "Release forms could not be retrieved.")
			return
		}
		writeJSON(w, map[string]any{"items": items, "total": len(items), "limit": 3}, http.StatusOK)
	})))
	mux.Handle("POST /api/v1/media/{id}/releases/{releaseId}/complete", contributorOnly(cfg.TokenVerifier, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		release, err := cfg.Releases.Complete(r.Context(), r.PathValue("id"), r.PathValue("releaseId"), principalFrom(r.Context()).Subject)
		if errors.Is(err, releases.ErrReleaseNotFound) {
			writeProblem(w, r, http.StatusNotFound, "Release not found", "The release upload does not exist.")
			return
		}
		if errors.Is(err, releases.ErrObjectMismatch) {
			writeProblem(w, r, http.StatusConflict, "Release upload incomplete", "The release object is missing or has the wrong size.")
			return
		}
		if err != nil {
			writeProblem(w, r, http.StatusServiceUnavailable, "Release unavailable", "The release upload could not be completed.")
			return
		}
		writeJSON(w, release, http.StatusOK)
	})))
	mux.Handle("PATCH /api/v1/batches/{id}/review", roleOnly(cfg.TokenVerifier, "reviewer", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input reviewBatchRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil {
			writeProblem(w, r, http.StatusBadRequest, "Invalid review", "The request body must contain supported review fields.")
			return
		}
		input.Status = strings.ToLower(strings.TrimSpace(input.Status))
		input.Feedback = strings.TrimSpace(input.Feedback)
		valid := input.Status == "in_review" || input.Status == "approved" || input.Status == "rejected"
		if !valid || len(input.Feedback) > 2000 || (input.Status == "rejected" && input.Feedback == "") {
			writeProblem(w, r, http.StatusUnprocessableEntity, "Invalid review", "Status must be in_review, approved, or rejected; rejection requires bounded feedback.")
			return
		}
		batch, err := cfg.Batches.Review(r.Context(), r.PathValue("id"), input.Status, input.Feedback)
		if errors.Is(err, batches.ErrNotReviewable) {
			writeProblem(w, r, http.StatusConflict, "Batch is not reviewable", "Only submitted or in-review batches can be reviewed.")
			return
		}
		if err != nil {
			writeProblem(w, r, http.StatusServiceUnavailable, "Review unavailable", "The review could not be saved.")
			return
		}
		writeJSON(w, batch, http.StatusOK)
	})))
	mux.Handle("GET /api/v1/collections", roleOnly(cfg.TokenVerifier, "buyer", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		items, err := cfg.Collections.List(r.Context(), principalFrom(r.Context()).Subject)
		if err != nil {
			writeProblem(w, r, 503, "Collections unavailable", "Collections could not be retrieved.")
			return
		}
		writeJSON(w, map[string]any{"items": items, "total": len(items)}, 200)
	})))
	mux.Handle("POST /api/v1/collections", roleOnly(cfg.TokenVerifier, "buyer", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, cfg.MaxBodyBytes)
		var input createCollectionRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if decoder.Decode(&input) != nil || len(strings.TrimSpace(input.Name)) < 1 || len(strings.TrimSpace(input.Name)) > 120 {
			writeProblem(w, r, 422, "Invalid collection", "A collection name between 1 and 120 characters is required.")
			return
		}
		item, err := cfg.Collections.Create(r.Context(), principalFrom(r.Context()).Subject, input.Name)
		if err != nil {
			writeProblem(w, r, 503, "Collections unavailable", "The collection could not be created.")
			return
		}
		writeJSON(w, item, 201)
	})))
	mux.Handle("PUT /api/v1/collections/{id}/items/{assetId}", roleOnly(cfg.TokenVerifier, "buyer", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := cfg.Collections.Add(r.Context(), principalFrom(r.Context()).Subject, r.PathValue("id"), r.PathValue("assetId"))
		if errors.Is(err, collections.ErrNotFound) {
			writeProblem(w, r, 404, "Collection or media unavailable", "The owned collection or approved media asset was not found.")
			return
		}
		if err != nil {
			writeProblem(w, r, 503, "Collections unavailable", "The item could not be saved.")
			return
		}
		w.WriteHeader(204)
	})))
	mux.Handle("DELETE /api/v1/collections/{id}/items/{assetId}", roleOnly(cfg.TokenVerifier, "buyer", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := cfg.Collections.Remove(r.Context(), principalFrom(r.Context()).Subject, r.PathValue("id"), r.PathValue("assetId"))
		if errors.Is(err, collections.ErrNotFound) {
			writeProblem(w, r, 404, "Collection item not found", "The collection item was not found.")
			return
		}
		if err != nil {
			writeProblem(w, r, 503, "Collections unavailable", "The item could not be removed.")
			return
		}
		w.WriteHeader(204)
	})))
	mux.Handle("GET /api/v1/favourites", roleOnly(cfg.TokenVerifier, "buyer", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ids, err := cfg.Collections.Favourites(r.Context(), principalFrom(r.Context()).Subject)
		if err != nil {
			writeProblem(w, r, 503, "Favourites unavailable", "Favourites could not be retrieved.")
			return
		}
		writeJSON(w, map[string]any{"mediaIds": ids}, 200)
	})))
	mux.Handle("PUT /api/v1/favourites/{assetId}", roleOnly(cfg.TokenVerifier, "buyer", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := cfg.Collections.Favourite(r.Context(), principalFrom(r.Context()).Subject, r.PathValue("assetId"), true); err != nil {
			status := 503
			if errors.Is(err, collections.ErrAssetUnavailable) {
				status = 404
			}
			writeProblem(w, r, status, "Favourite unavailable", "The approved media asset could not be saved.")
			return
		}
		w.WriteHeader(204)
	})))
	mux.Handle("DELETE /api/v1/favourites/{assetId}", roleOnly(cfg.TokenVerifier, "buyer", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := cfg.Collections.Favourite(r.Context(), principalFrom(r.Context()).Subject, r.PathValue("assetId"), false); err != nil {
			writeProblem(w, r, 503, "Favourite unavailable", "The favourite could not be removed.")
			return
		}
		w.WriteHeader(204)
	})))
	return requestContext(recoverPanics(securityHeaders(cors(cfg.WebOrigin, mux))))
}

func catalogQuery(values url.Values) (catalog.Query, bool) {
	page, pageSize := 1, 24
	var err error
	if values.Get("page") != "" {
		page, err = strconv.Atoi(values.Get("page"))
		if err != nil {
			return catalog.Query{}, false
		}
	}
	if values.Get("pageSize") != "" {
		pageSize, err = strconv.Atoi(values.Get("pageSize"))
		if err != nil {
			return catalog.Query{}, false
		}
	}
	query := catalog.Query{Text: strings.TrimSpace(values.Get("q")), Kind: strings.ToLower(values.Get("kind")),
		UsageType: strings.ToLower(values.Get("usageType")), Orientation: strings.ToLower(values.Get("orientation")),
		Location: strings.TrimSpace(values.Get("location")), Sort: strings.ToLower(values.Get("sort")), Page: page, PageSize: pageSize}
	if query.Sort == "" {
		query.Sort = "newest"
	}
	valid := len(query.Text) <= 200 && len(query.Location) <= 200 && page >= 1 && pageSize >= 1 && pageSize <= 100 &&
		(query.Kind == "" || query.Kind == "image" || query.Kind == "video" || query.Kind == "illustration") &&
		(query.UsageType == "" || query.UsageType == "creative" || query.UsageType == "editorial") &&
		(query.Orientation == "" || query.Orientation == "portrait" || query.Orientation == "landscape" || query.Orientation == "square") &&
		(query.Sort == "newest" || query.Sort == "relevance")
	return query, valid
}

func validContentType(kind, contentType string) bool {
	allowed := map[string]map[string]bool{
		"image":        {"image/jpeg": true, "image/png": true, "image/webp": true},
		"illustration": {"image/jpeg": true, "image/png": true, "image/webp": true},
		"video":        {"video/mp4": true, "video/quicktime": true, "video/webm": true},
	}
	return allowed[kind][contentType]
}

type principalKeyType string

const principalKey principalKeyType = "principal"

func contributorOnly(verifier auth.Verifier, next http.Handler) http.Handler {
	return roleOnly(verifier, "contributor", next)
}

func roleOnly(verifier auth.Verifier, role string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if verifier == nil {
			writeProblem(w, r, http.StatusServiceUnavailable, "Authentication unavailable", "Authentication is not configured.")
			return
		}
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			writeProblem(w, r, http.StatusUnauthorized, "Authentication required", "A valid bearer token is required.")
			return
		}
		principal, err := verifier.Verify(strings.TrimSpace(strings.TrimPrefix(header, "Bearer ")))
		if err != nil {
			writeProblem(w, r, http.StatusUnauthorized, "Authentication required", "A valid bearer token is required.")
			return
		}
		if principal.Role != role {
			writeProblem(w, r, http.StatusForbidden, "Role access required", "The authenticated account cannot access this operation.")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), principalKey, principal)))
	})
}

func principalFrom(ctx context.Context) auth.Principal {
	principal, _ := ctx.Value(principalKey).(auth.Principal)
	return principal
}

type dependencyCheck struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

func readinessChecks(ctx context.Context, cfg Config) []dependencyCheck {
	database, _ := url.Parse(cfg.DatabaseURL)
	databaseAddress := net.JoinHostPort(database.Hostname(), defaultPort(database.Port(), "5432"))
	addresses := []struct{ name, address string }{{"postgres", databaseAddress}, {"redis", cfg.RedisAddress}}
	if len(cfg.KafkaBrokers) > 0 {
		addresses = append(addresses, struct{ name, address string }{"kafka", cfg.KafkaBrokers[0]})
	}
	storage, _ := url.Parse(cfg.S3Endpoint)
	addresses = append(addresses, struct{ name, address string }{"object-storage", net.JoinHostPort(storage.Hostname(), defaultPort(storage.Port(), "9000"))})
	checks := make([]dependencyCheck, 0, len(addresses))
	for _, dependency := range addresses {
		check := dependencyCheck{Name: dependency.name, Status: "down"}
		connection, err := (&net.Dialer{Timeout: 750 * time.Millisecond}).DialContext(ctx, "tcp", dependency.address)
		if err == nil {
			check.Status = "up"
			_ = connection.Close()
		}
		checks = append(checks, check)
	}
	return checks
}

func defaultPort(port, fallback string) string {
	if port == "" {
		return fallback
	}
	return port
}

func writeJSON(w http.ResponseWriter, value any, status int) {
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "application/json")
	}
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeProblem(w http.ResponseWriter, r *http.Request, status int, title, detail string) {
	w.Header().Set("Content-Type", "application/problem+json")
	writeJSON(w, problem{Type: "https://httpstatuses.io/" + strconv.Itoa(status), Title: title, Status: status, Detail: detail, Instance: r.URL.Path, CorrelationID: correlationID(r.Context())}, status)
}

func requestContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Correlation-ID")
		if !validCorrelationID.MatchString(id) {
			id = newID()
		}
		started := time.Now()
		w.Header().Set("X-Correlation-ID", id)
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, r.WithContext(context.WithValue(r.Context(), correlationIDKey, id)))
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{
			"timestamp": time.Now().UTC().Format(time.RFC3339Nano), "severity": "info", "service": "tamo-media-service",
			"event": "http.request.completed", "correlationId": id, "method": r.Method, "route": r.URL.Path,
			"status": recorder.status, "durationMs": time.Since(started).Milliseconds(),
		})
	})
}

func recoverPanics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				_ = json.NewEncoder(os.Stderr).Encode(map[string]any{
					"timestamp": time.Now().UTC().Format(time.RFC3339Nano), "severity": "error",
					"service": "tamo-media-service", "event": "http.request.panic",
					"correlationId": correlationID(r.Context()),
				})
				writeProblem(w, r, http.StatusInternalServerError, "Internal server error", "The request could not be completed.")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		next.ServeHTTP(w, r)
	})
}

func cors(origin string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Correlation-ID")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		w.Header().Add("Vary", "Origin")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func correlationID(ctx context.Context) string {
	value, _ := ctx.Value(correlationIDKey).(string)
	return value
}
func newID() string { return time.Now().UTC().Format("20060102T150405.000000000") }
