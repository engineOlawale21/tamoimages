package config

import "testing"

func setValidEnvironment(t *testing.T) {
	t.Helper()
	t.Setenv("APP_ENV", "test")
	t.Setenv("PORT", "5000")
	t.Setenv("WEB_ORIGIN", "http://localhost:3000")
	t.Setenv("DATABASE_URL", "postgresql://localhost:5432/tamo_media")
	t.Setenv("REDIS_ADDR", "localhost:6379")
	t.Setenv("KAFKA_BROKERS", "localhost:9092")
	t.Setenv("S3_ENDPOINT", "http://localhost:9000")
	t.Setenv("S3_BUCKET", "tamo-media")
	t.Setenv("S3_ACCESS_KEY", "test-access")
	t.Setenv("S3_SECRET_KEY", "test-secret")
	t.Setenv("JWT_SECRET", "test-secret-with-at-least-32-characters")
}

func TestLoadValidatesAndNormalizesEnvironment(t *testing.T) {
	setValidEnvironment(t)
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != 5000 || cfg.MaxBodyBytes != 32<<20 {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}

func TestLoadRejectsInvalidPort(t *testing.T) {
	setValidEnvironment(t)
	t.Setenv("PORT", "70000")
	if _, err := Load(); err == nil {
		t.Fatal("expected invalid port error")
	}
}

func TestLoadRequiresExplicitProductionDependencies(t *testing.T) {
	setValidEnvironment(t)
	t.Setenv("APP_ENV", "production")
	t.Setenv("KAFKA_BROKERS", "")
	if _, err := Load(); err == nil {
		t.Fatal("expected production configuration error")
	}
}
