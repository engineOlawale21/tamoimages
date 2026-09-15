package database

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

func TestDatabasePingAndTransaction(t *testing.T) {
	connectionString := os.Getenv("DATABASE_INTEGRATION_URL")
	if connectionString == "" {
		t.Skip("DATABASE_INTEGRATION_URL is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	database, err := Open(ctx, connectionString, 2, 3*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if err = database.Ping(ctx); err != nil {
		t.Fatal(err)
	}
	var value string
	if err = database.InTransaction(ctx, func(transaction pgx.Tx) error {
		if _, executeErr := transaction.Exec(ctx, "CREATE TEMPORARY TABLE foundation_check (value text NOT NULL)"); executeErr != nil {
			return executeErr
		}
		if _, executeErr := transaction.Exec(ctx, "INSERT INTO foundation_check(value) VALUES ('ready')"); executeErr != nil {
			return executeErr
		}
		return transaction.QueryRow(ctx, "SELECT value FROM foundation_check").Scan(&value)
	}); err != nil {
		t.Fatal(err)
	}
	if value != "ready" {
		t.Fatalf("unexpected transaction value %q", value)
	}
}
