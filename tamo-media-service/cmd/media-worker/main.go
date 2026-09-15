package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/tamoimages/media-service/internal/config"
	"github.com/tamoimages/media-service/internal/platform/database"
	"github.com/tamoimages/media-service/internal/platform/kafka"
	"github.com/tamoimages/media-service/internal/platform/storage"
	"github.com/tamoimages/media-service/internal/processing"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fatal("configuration.invalid", err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	db, err := database.Open(ctx, cfg.DatabaseURL, cfg.DatabasePoolMax, cfg.DatabaseConnectTimeout)
	if err != nil {
		fatal("database.open_failed", err)
	}
	defer db.Close()
	objectStore, err := storage.OpenMinIO(cfg.S3Endpoint, cfg.S3Bucket, cfg.S3AccessKey, cfg.S3SecretKey)
	if err != nil {
		fatal("storage.open_failed", err)
	}
	publisher, err := kafka.Open(cfg.KafkaBrokers, "tamo-media-worker-producer", "")
	if err != nil {
		fatal("kafka.producer_open_failed", err)
	}
	defer publisher.Close()
	processor := processing.New(db, objectStore, publisher, processing.ClamAV{Path: cfg.ClamScanPath}, processing.FFprobe{Path: cfg.FFprobePath}, processing.FFmpeg{Path: cfg.FFmpegPath})

	workerContext, cancel := context.WithCancel(ctx)
	defer cancel()
	errorsChannel := make(chan error, cfg.WorkerConcurrency)
	var workers sync.WaitGroup
	for index := range cfg.WorkerConcurrency {
		consumer, openErr := kafka.Open(cfg.KafkaBrokers, fmt.Sprintf("tamo-media-worker-%d", index+1), "tamo-media-processing-v1", "media.uploaded.v1")
		if openErr != nil {
			fatal("kafka.consumer_open_failed", openErr)
		}
		workers.Add(1)
		go func() {
			defer workers.Done()
			defer consumer.Close()
			if consumeErr := consumer.Consume(workerContext, processor.Handle); consumeErr != nil {
				select {
				case errorsChannel <- consumeErr:
				default:
				}
				cancel()
			}
		}()
	}
	log("info", "worker.started", map[string]any{"concurrency": cfg.WorkerConcurrency})
	select {
	case <-ctx.Done():
		cancel()
	case err = <-errorsChannel:
		log("error", "worker.failed", map[string]any{"error": err.Error()})
		cancel()
	}
	workers.Wait()
	if err != nil {
		os.Exit(1)
	}
}

func fatal(event string, err error) {
	log("fatal", event, map[string]any{"error": err.Error()})
	os.Exit(1)
}
func log(severity, event string, fields map[string]any) {
	entry := map[string]any{"timestamp": time.Now().UTC().Format(time.RFC3339Nano), "severity": severity, "service": "tamo-media-worker", "event": event}
	for key, value := range fields {
		entry[key] = value
	}
	_ = json.NewEncoder(os.Stdout).Encode(entry)
}
