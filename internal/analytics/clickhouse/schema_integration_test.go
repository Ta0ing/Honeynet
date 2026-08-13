package clickhouse

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestValidateSchemaIntegration(t *testing.T) {
	dsn := os.Getenv("HONEYPOT_TEST_CLICKHOUSE_DSN")
	if dsn == "" {
		t.Skip("HONEYPOT_TEST_CLICKHOUSE_DSN is not set")
	}
	repository, err := Open(Config{
		DSN: dsn, Database: "honeynet_analytics", Table: defaultTable,
		DialTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer repository.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := repository.Ping(ctx); err != nil {
		t.Fatal(err)
	}
	if err := repository.ValidateSchema(ctx); err != nil {
		t.Fatal(err)
	}
	status := repository.Status(ctx)
	if !status.Healthy || status.SchemaVersion < minimumSchemaVersion {
		t.Fatalf("unexpected schema-aware status: %+v", status)
	}
}
