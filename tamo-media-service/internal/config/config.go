package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Environment            string
	Port                   int
	WebOrigin              string
	DatabaseURL            string
	RedisAddress           string
	RedisKeyPrefix         string
	RedisOperationTimeout  time.Duration
	RedisMaxRetries        int
	WorkerConcurrency      int
	KafkaBrokers           []string
	S3Endpoint             string
	S3Bucket               string
	S3AccessKey            string
	S3SecretKey            string
	MaxBodyBytes           int64
	DatabasePoolMax        int32
	DatabaseConnectTimeout time.Duration
	JWTSecret              string
	JWTIssuer              string
	JWTAudience            string
	UploadURLTTL           time.Duration
	FFprobePath            string
	FFmpegPath             string
	ClamScanPath           string
	PaystackSecretKey      string
	PaystackAPIURL         string
	PaymentCallbackURL     string
}

func Load() (Config, error) {
	cfg := Config{
		Environment:       value("APP_ENV", "development"),
		WebOrigin:         value("WEB_ORIGIN", "http://localhost:3000"),
		DatabaseURL:       os.Getenv("DATABASE_URL"),
		RedisAddress:      os.Getenv("REDIS_ADDR"),
		RedisKeyPrefix:    value("REDIS_KEY_PREFIX", "tamo:development:media"),
		KafkaBrokers:      split(value("KAFKA_BROKERS", "localhost:9092")),
		S3Endpoint:        os.Getenv("S3_ENDPOINT"),
		S3Bucket:          os.Getenv("S3_BUCKET"),
		S3AccessKey:       os.Getenv("S3_ACCESS_KEY"),
		S3SecretKey:       os.Getenv("S3_SECRET_KEY"),
		MaxBodyBytes:      32 << 20,
		JWTSecret:         os.Getenv("JWT_SECRET"),
		JWTIssuer:         value("JWT_ISSUER", "tamo-identity"),
		JWTAudience:       value("JWT_AUDIENCE", "tamo-platform"),
		FFprobePath:       value("FFPROBE_PATH", "ffprobe"),
		FFmpegPath:        value("FFMPEG_PATH", "ffmpeg"),
		ClamScanPath:      value("CLAMSCAN_PATH", "clamscan"),
		PaystackSecretKey: os.Getenv("PAYSTACK_SECRET_KEY"), PaystackAPIURL: value("PAYSTACK_API_URL", "https://api.paystack.co"), PaymentCallbackURL: value("PAYMENT_CALLBACK_URL", "http://localhost:3000/cart"),
	}
	var err error
	if cfg.Port, err = integer("PORT", 5000, 1, 65535); err != nil {
		return Config{}, err
	}
	if cfg.MaxBodyBytes, err = integer64("MAX_UPLOAD_BYTES", cfg.MaxBodyBytes, 1, 1<<30); err != nil {
		return Config{}, err
	}
	poolMax, poolErr := integer("DATABASE_POOL_MAX", 10, 1, 50)
	if poolErr != nil {
		return Config{}, poolErr
	}
	cfg.DatabasePoolMax = int32(poolMax)
	connectMilliseconds, connectErr := integer("DATABASE_CONNECT_TIMEOUT_MS", 3000, 100, 30000)
	if connectErr != nil {
		return Config{}, connectErr
	}
	cfg.DatabaseConnectTimeout = time.Duration(connectMilliseconds) * time.Millisecond
	redisMilliseconds, redisErr := integer("REDIS_OPERATION_TIMEOUT_MS", 1000, 100, 10000)
	if redisErr != nil {
		return Config{}, redisErr
	}
	cfg.RedisOperationTimeout = time.Duration(redisMilliseconds) * time.Millisecond
	if cfg.RedisMaxRetries, err = integer("REDIS_MAX_RETRIES", 2, 0, 10); err != nil {
		return Config{}, err
	}
	if cfg.WorkerConcurrency, err = integer("WORKER_CONCURRENCY", 4, 1, 64); err != nil {
		return Config{}, err
	}
	uploadTTLSeconds, uploadTTLErr := integer("UPLOAD_URL_TTL_SECONDS", 900, 60, 3600)
	if uploadTTLErr != nil {
		return Config{}, uploadTTLErr
	}
	cfg.UploadURLTTL = time.Duration(uploadTTLSeconds) * time.Second
	if len(cfg.JWTSecret) < 32 {
		return Config{}, fmt.Errorf("JWT_SECRET must contain at least 32 characters")
	}
	if cfg.Environment != "development" && cfg.Environment != "test" && cfg.Environment != "staging" && cfg.Environment != "production" {
		return Config{}, fmt.Errorf("APP_ENV must be development, test, staging, or production")
	}
	for key, raw := range map[string]string{"WEB_ORIGIN": cfg.WebOrigin, "DATABASE_URL": cfg.DatabaseURL, "S3_ENDPOINT": cfg.S3Endpoint} {
		parsed, parseErr := url.ParseRequestURI(raw)
		if parseErr != nil || parsed.Scheme == "" || parsed.Host == "" {
			return Config{}, fmt.Errorf("%s must be a valid absolute URL", key)
		}
	}
	if _, _, err = net.SplitHostPort(cfg.RedisAddress); err != nil {
		return Config{}, fmt.Errorf("REDIS_ADDR must be host:port: %w", err)
	}
	for _, broker := range cfg.KafkaBrokers {
		if _, _, err = net.SplitHostPort(broker); err != nil {
			return Config{}, fmt.Errorf("KAFKA_BROKERS contains invalid address %q", broker)
		}
	}
	if strings.TrimSpace(cfg.S3Bucket) == "" {
		return Config{}, fmt.Errorf("S3_BUCKET is required")
	}
	if strings.TrimSpace(cfg.RedisKeyPrefix) == "" {
		return Config{}, fmt.Errorf("REDIS_KEY_PREFIX is required")
	}
	if cfg.Environment == "production" {
		for _, key := range []string{"WEB_ORIGIN", "DATABASE_URL", "REDIS_ADDR", "REDIS_KEY_PREFIX", "KAFKA_BROKERS", "S3_ENDPOINT", "S3_BUCKET", "S3_ACCESS_KEY", "S3_SECRET_KEY", "JWT_SECRET", "PAYSTACK_SECRET_KEY", "PAYMENT_CALLBACK_URL"} {
			if strings.TrimSpace(os.Getenv(key)) == "" {
				return Config{}, fmt.Errorf("%s is required in production", key)
			}
		}
	}
	return cfg, nil
}

func (c Config) Address() string { return fmt.Sprintf(":%d", c.Port) }

func value(key, fallback string) string {
	if current := strings.TrimSpace(os.Getenv(key)); current != "" {
		return current
	}
	return fallback
}

func split(raw string) []string {
	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			values = append(values, trimmed)
		}
	}
	return values
}

func integer(key string, fallback, minimum, maximum int) (int, error) {
	parsed, err := integer64(key, int64(fallback), int64(minimum), int64(maximum))
	return int(parsed), err
}

func integer64(key string, fallback, minimum, maximum int64) (int64, error) {
	raw := value(key, strconv.FormatInt(fallback, 10))
	parsed, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || parsed < minimum || parsed > maximum {
		return 0, fmt.Errorf("%s must be between %d and %d", key, minimum, maximum)
	}
	return parsed, nil
}

const DependencyTimeout = 750 * time.Millisecond
