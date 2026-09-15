package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/tamoimages/media-service/internal/config"
	"github.com/tamoimages/media-service/internal/platform/database"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fail(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	db, err := database.Open(ctx, cfg.DatabaseURL, cfg.DatabasePoolMax, cfg.DatabaseConnectTimeout)
	if err != nil {
		fail(err)
	}
	defer db.Close()
	if err = db.Migrate(ctx, "migrations"); err != nil {
		fail(err)
	}
	fmt.Println("database migrations applied")
}

func fail(err error) { fmt.Fprintln(os.Stderr, "migration failed:", err); os.Exit(1) }
