package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tamoimages/media-service/internal/auth"
	"github.com/tamoimages/media-service/internal/batches"
	"github.com/tamoimages/media-service/internal/catalog"
	"github.com/tamoimages/media-service/internal/collections"
	"github.com/tamoimages/media-service/internal/config"
	"github.com/tamoimages/media-service/internal/httpapi"
	"github.com/tamoimages/media-service/internal/platform/database"
	"github.com/tamoimages/media-service/internal/platform/kafka"
	"github.com/tamoimages/media-service/internal/platform/storage"
	"github.com/tamoimages/media-service/internal/releases"
	"github.com/tamoimages/media-service/internal/uploads"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log("fatal", "configuration.invalid", map[string]any{"error": err.Error()})
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	db, err := database.Open(ctx, cfg.DatabaseURL, cfg.DatabasePoolMax, cfg.DatabaseConnectTimeout)
	if err != nil {
		cancel()
		log("fatal", "database.open_failed", map[string]any{"error": err.Error()})
		os.Exit(1)
	}
	objectStore, err := storage.OpenMinIO(cfg.S3Endpoint, cfg.S3Bucket, cfg.S3AccessKey, cfg.S3SecretKey)
	cancel()
	if err != nil {
		db.Close()
		log("fatal", "storage.open_failed", map[string]any{"error": err.Error()})
		os.Exit(1)
	}
	defer db.Close()
	producer, err := kafka.Open(cfg.KafkaBrokers, "tamo-media-api", "")
	if err != nil {
		log("fatal", "kafka.open_failed", map[string]any{"error": err.Error()})
		os.Exit(1)
	}
	defer producer.Close()
	uploadService := uploads.New(db, objectStore, producer, cfg.UploadURLTTL)
	batchService := batches.New(db, producer)
	releaseService := releases.New(db, objectStore, cfg.UploadURLTTL)
	catalogService := catalog.New(db, objectStore, cfg.UploadURLTTL)
	collectionService := collections.New(db)
	server := &http.Server{
		Addr: cfg.Address(), Handler: httpapi.New(httpapi.Config{
			WebOrigin: cfg.WebOrigin, DatabaseURL: cfg.DatabaseURL, RedisAddress: cfg.RedisAddress,
			KafkaBrokers: cfg.KafkaBrokers, S3Endpoint: cfg.S3Endpoint, MaxBodyBytes: cfg.MaxBodyBytes,
			Uploads: uploadService, Batches: batchService, Releases: releaseService, Catalog: catalogService, Collections: collectionService, TokenVerifier: auth.NewHMACVerifier(cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTAudience),
		}),
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 30 * time.Second,
		WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second,
	}
	stopped := make(chan os.Signal, 1)
	signal.Notify(stopped, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-stopped
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if shutdownErr := server.Shutdown(ctx); shutdownErr != nil {
			log("error", "server.shutdown_failed", map[string]any{"error": shutdownErr.Error()})
		}
	}()
	log("info", "server.started", map[string]any{"address": cfg.Address(), "environment": cfg.Environment})
	if err = server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log("fatal", "server.failed", map[string]any{"error": err.Error()})
		os.Exit(1)
	}
}

func log(severity, event string, fields map[string]any) {
	entry := map[string]any{"timestamp": time.Now().UTC().Format(time.RFC3339Nano), "severity": severity, "service": "tamo-media-service", "event": event}
	for key, value := range fields {
		entry[key] = value
	}
	_ = json.NewEncoder(os.Stdout).Encode(entry)
}
