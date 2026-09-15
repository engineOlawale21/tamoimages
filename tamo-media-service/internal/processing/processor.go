package processing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/tamoimages/media-service/internal/platform/database"
	"github.com/tamoimages/media-service/internal/platform/kafka"
	"github.com/tamoimages/media-service/internal/platform/storage"
)

type Repository interface {
	ClaimMediaAssetForProcessing(context.Context, string) (database.MediaAsset, bool, error)
	CompleteMediaAssetProcessing(context.Context, string, database.ProcessingMetadata, []database.MediaVariant) error
	MarkMediaAssetFailed(context.Context, string, string) error
}

type ObjectStore interface {
	Get(context.Context, string) (io.ReadCloser, error)
	Put(context.Context, string, io.Reader, int64, string) (storage.ObjectInfo, error)
	Delete(context.Context, string) error
}
type Publisher interface {
	Publish(context.Context, string, kafka.Event) error
}
type Prober interface {
	Probe(context.Context, string) (database.ProcessingMetadata, error)
}
type MalwareScanner interface {
	Scan(context.Context, string) error
}
type Transcoder interface {
	Transcode(context.Context, string, string, string, database.ProcessingMetadata) ([]GeneratedVariant, error)
}

type GeneratedVariant struct {
	Kind        string
	Path        string
	ContentType string
	Width       int
	Height      int
}

type Processor struct {
	repository Repository
	storage    ObjectStore
	publisher  Publisher
	prober     Prober
	transcoder Transcoder
	scanner    MalwareScanner
	now        func() time.Time
}

func New(repository Repository, objectStore ObjectStore, publisher Publisher, scanner MalwareScanner, prober Prober, transcoder Transcoder) *Processor {
	return &Processor{repository: repository, storage: objectStore, publisher: publisher, scanner: scanner, prober: prober, transcoder: transcoder, now: time.Now}
}

func (p *Processor) Handle(ctx context.Context, event kafka.Event) error {
	var payload struct {
		AssetID string `json:"assetId"`
	}
	if event.Type != "media.uploaded" || json.Unmarshal(event.Payload, &payload) != nil || payload.AssetID == "" {
		return nil
	}
	asset, claimed, err := p.repository.ClaimMediaAssetForProcessing(ctx, payload.AssetID)
	if err != nil || !claimed {
		return err
	}
	processingContext, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	metadata, variants, failureCode, processErr := p.process(processingContext, asset)
	if processErr != nil {
		if err = p.repository.MarkMediaAssetFailed(ctx, asset.ID, failureCode); err != nil {
			return err
		}
		return p.publish(ctx, "media.processing-failed.v1", "media.processing-failed", asset.ID, event.CorrelationID, failureCode)
	}
	if err = p.repository.CompleteMediaAssetProcessing(ctx, asset.ID, metadata, variants); err != nil {
		return err
	}
	return p.publish(ctx, "media.processed.v1", "media.processed", asset.ID, event.CorrelationID, "")
}

func (p *Processor) process(ctx context.Context, asset database.MediaAsset) (database.ProcessingMetadata, []database.MediaVariant, string, error) {
	object, err := p.storage.Get(ctx, asset.StorageKey)
	if err != nil {
		return database.ProcessingMetadata{}, nil, "object_unavailable", err
	}
	defer object.Close()
	workDirectory, err := os.MkdirTemp("", "tamo-media-*")
	if err != nil {
		return database.ProcessingMetadata{}, nil, "temporary_storage_unavailable", err
	}
	defer os.RemoveAll(workDirectory)
	inputPath := filepath.Join(workDirectory, "original")
	file, err := os.Create(inputPath)
	if err != nil {
		return database.ProcessingMetadata{}, nil, "temporary_storage_unavailable", err
	}
	written, copyErr := io.Copy(file, io.LimitReader(object, asset.SizeBytes+1))
	if copyErr != nil {
		_ = file.Close()
		return database.ProcessingMetadata{}, nil, "object_read_failed", copyErr
	}
	if written != asset.SizeBytes {
		_ = file.Close()
		return database.ProcessingMetadata{}, nil, "object_size_mismatch", errors.New("object size changed after upload completion")
	}
	if err = file.Close(); err != nil {
		return database.ProcessingMetadata{}, nil, "object_read_failed", err
	}
	if err = p.scanner.Scan(ctx, inputPath); err != nil {
		return database.ProcessingMetadata{}, nil, "malware_scan_failed", err
	}
	metadata, err := p.prober.Probe(ctx, inputPath)
	if err != nil {
		return database.ProcessingMetadata{}, nil, "invalid_media", err
	}
	if asset.Kind == "video" && metadata.DurationSeconds <= 0 {
		return database.ProcessingMetadata{}, nil, "invalid_video", errors.New("video duration is missing")
	}
	generated, err := p.transcoder.Transcode(ctx, inputPath, workDirectory, asset.Kind, metadata)
	if err != nil {
		return database.ProcessingMetadata{}, nil, "transcode_failed", err
	}
	variants := make([]database.MediaVariant, 0, len(generated))
	uploadedKeys := make([]string, 0, len(generated))
	cleanup := func() {
		cleanupContext, cancelCleanup := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelCleanup()
		for _, key := range uploadedKeys {
			_ = p.storage.Delete(cleanupContext, key)
		}
	}
	for _, variant := range generated {
		info, statErr := os.Stat(variant.Path)
		if statErr != nil || info.Size() < 1 {
			cleanup()
			return database.ProcessingMetadata{}, nil, "transcode_output_missing", errors.New("transcode output missing")
		}
		reader, openErr := os.Open(variant.Path)
		if openErr != nil {
			cleanup()
			return database.ProcessingMetadata{}, nil, "transcode_output_missing", openErr
		}
		key := fmt.Sprintf("variants/%s/%s/%s", asset.ContributorID, asset.ID, filepath.Base(variant.Path))
		_, putErr := p.storage.Put(ctx, key, reader, info.Size(), variant.ContentType)
		_ = reader.Close()
		if putErr != nil {
			cleanup()
			return database.ProcessingMetadata{}, nil, "variant_storage_failed", putErr
		}
		uploadedKeys = append(uploadedKeys, key)
		variants = append(variants, database.MediaVariant{Kind: variant.Kind, StorageKey: key, ContentType: variant.ContentType, SizeBytes: info.Size(), Width: variant.Width, Height: variant.Height})
	}
	return metadata, variants, "", nil
}

// ClamAV scans the private original before metadata extraction or derivative creation.
// Exit status 1 (infected) and scanner failures are both fail-closed.
type ClamAV struct{ Path string }

func (c ClamAV) Scan(ctx context.Context, path string) error {
	output, err := exec.CommandContext(ctx, c.Path, "--no-summary", path).CombinedOutput()
	if err != nil {
		return fmt.Errorf("malware scan rejected upload: %w (%s)", err, boundedOutput(output))
	}
	return nil
}

func (p *Processor) publish(ctx context.Context, topic, eventType, assetID, correlationID, failureCode string) error {
	payload := map[string]string{"assetId": assetID}
	if failureCode != "" {
		payload["failureCode"] = failureCode
	}
	encoded, _ := json.Marshal(payload)
	return p.publisher.Publish(ctx, topic, kafka.Event{ID: uuid.NewString(), Type: eventType, Version: 1,
		OccurredAt: p.now().UTC(), CorrelationID: correlationID, Producer: "tamo-media-worker", Payload: encoded})
}

type FFprobe struct{ Path string }

func (f FFprobe) Probe(ctx context.Context, path string) (database.ProcessingMetadata, error) {
	output, err := exec.CommandContext(ctx, f.Path, "-v", "error", "-select_streams", "v:0", "-show_entries", "stream=codec_name,width,height:format=duration", "-of", "json", path).Output()
	if err != nil {
		return database.ProcessingMetadata{}, fmt.Errorf("ffprobe rejected video: %w", err)
	}
	var result struct {
		Streams []struct {
			Codec  string `json:"codec_name"`
			Width  int    `json:"width"`
			Height int    `json:"height"`
		} `json:"streams"`
		Format struct {
			Duration string `json:"duration"`
		} `json:"format"`
	}
	if json.Unmarshal(output, &result) != nil || len(result.Streams) == 0 || result.Streams[0].Width < 1 || result.Streams[0].Height < 1 {
		return database.ProcessingMetadata{}, errors.New("ffprobe returned no valid video stream")
	}
	duration, _ := strconv.ParseFloat(result.Format.Duration, 64)
	stream := result.Streams[0]
	return database.ProcessingMetadata{DurationSeconds: duration, Width: stream.Width, Height: stream.Height, CodecName: stream.Codec}, nil
}

type FFmpeg struct{ Path string }

func (f FFmpeg) Transcode(ctx context.Context, inputPath, workDirectory, kind string, metadata database.ProcessingMetadata) ([]GeneratedVariant, error) {
	if kind == "video" {
		return f.videoVariants(ctx, inputPath, workDirectory, metadata)
	}
	return f.imageVariants(ctx, inputPath, workDirectory, metadata)
}

type variantSize struct {
	kind   string
	width  int
	height int
}

func (f FFmpeg) imageVariants(ctx context.Context, inputPath, workDirectory string, metadata database.ProcessingMetadata) ([]GeneratedVariant, error) {
	sizes := []variantSize{{"thumbnail", 320, 320}, {"small", 640, 640}, {"medium", 1280, 1280}, {"large", 2048, 2048}}
	return f.renderSizes(ctx, inputPath, workDirectory, metadata, sizes, "webp")
}

func (f FFmpeg) videoVariants(ctx context.Context, inputPath, workDirectory string, metadata database.ProcessingMetadata) ([]GeneratedVariant, error) {
	posterSizes := []variantSize{{"thumbnail", 320, 320}, {"poster", 1280, 1280}}
	variants, err := f.renderSizes(ctx, inputPath, workDirectory, metadata, posterSizes, "jpg")
	if err != nil {
		return nil, err
	}
	videoSizes := []variantSize{{"small", 854, 480}, {"medium", 1280, 720}, {"large", 1920, 1080}}
	seen := map[string]bool{}
	for _, size := range videoSizes {
		width, height := fitBox(metadata.Width, metadata.Height, size.width, size.height, true)
		dimension := fmt.Sprintf("%dx%d", width, height)
		if seen[dimension] {
			continue
		}
		seen[dimension] = true
		outputPath := filepath.Join(workDirectory, size.kind+".mp4")
		filter := fmt.Sprintf("scale=%d:%d", width, height)
		command := exec.CommandContext(ctx, f.Path, "-v", "error", "-y", "-i", inputPath, "-map", "0:v:0", "-map", "0:a?", "-vf", filter, "-c:v", "libx264", "-preset", "medium", "-crf", "23", "-c:a", "aac", "-b:a", "128k", "-movflags", "+faststart", outputPath)
		if output, runErr := command.CombinedOutput(); runErr != nil {
			return nil, fmt.Errorf("ffmpeg %s video failed: %w (%s)", size.kind, runErr, boundedOutput(output))
		}
		variants = append(variants, GeneratedVariant{Kind: size.kind, Path: outputPath, ContentType: "video/mp4", Width: width, Height: height})
	}
	return variants, nil
}

func (f FFmpeg) renderSizes(ctx context.Context, inputPath, workDirectory string, metadata database.ProcessingMetadata, sizes []variantSize, extension string) ([]GeneratedVariant, error) {
	variants := make([]GeneratedVariant, 0, len(sizes))
	seen := map[string]bool{}
	for _, size := range sizes {
		width, height := fitBox(metadata.Width, metadata.Height, size.width, size.height, false)
		dimension := fmt.Sprintf("%dx%d", width, height)
		if seen[dimension] {
			continue
		}
		seen[dimension] = true
		outputPath := filepath.Join(workDirectory, size.kind+"."+extension)
		filter := fmt.Sprintf("scale=%d:%d", width, height)
		args := []string{"-v", "error", "-y", "-i", inputPath, "-frames:v", "1", "-vf", filter}
		contentType := "image/webp"
		if extension == "webp" {
			args = append(args, "-c:v", "libwebp", "-quality", "82")
		} else {
			args = append(args, "-q:v", "3")
			contentType = "image/jpeg"
		}
		args = append(args, outputPath)
		if output, runErr := exec.CommandContext(ctx, f.Path, args...).CombinedOutput(); runErr != nil {
			return nil, fmt.Errorf("ffmpeg %s image failed: %w (%s)", size.kind, runErr, boundedOutput(output))
		}
		variants = append(variants, GeneratedVariant{Kind: size.kind, Path: outputPath, ContentType: contentType, Width: width, Height: height})
	}
	return variants, nil
}

func fitBox(width, height, maxWidth, maxHeight int, even bool) (int, int) {
	scale := math.Min(1, math.Min(float64(maxWidth)/float64(width), float64(maxHeight)/float64(height)))
	resultWidth := int(math.Round(float64(width) * scale))
	resultHeight := int(math.Round(float64(height) * scale))
	if even {
		if resultWidth%2 != 0 {
			resultWidth--
		}
		if resultHeight%2 != 0 {
			resultHeight--
		}
	}
	return max(1, resultWidth), max(1, resultHeight)
}
func boundedOutput(output []byte) string {
	if len(output) > 512 {
		output = output[:512]
	}
	return string(output)
}
