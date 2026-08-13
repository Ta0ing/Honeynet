package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type migrationWriter struct {
	batches [][]string
	failAt  int
}

func (writer *migrationWriter) InsertAttackEvents(_ context.Context, events []AttackEvent) error {
	if writer.failAt > 0 && len(writer.batches)+1 == writer.failAt {
		return errors.New("ClickHouse unavailable")
	}
	ids := make([]string, 0, len(events))
	for _, event := range events {
		ids = append(ids, event.EventID)
	}
	writer.batches = append(writer.batches, ids)
	return nil
}

func migrationDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&AttackEvent{}, &EventMigrationCheckpoint{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func seedMigrationEvents(t *testing.T, db *gorm.DB) []AttackEvent {
	t.Helper()
	base := time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC)
	events := []AttackEvent{
		{EventID: "event-b", NodeID: "node-1", EventType: "web.request", Timestamp: base, CreatedAt: base},
		{EventID: "event-a", NodeID: "node-1", EventType: "web.request", Timestamp: base, CreatedAt: base},
		{EventID: "event-c", NodeID: "node-1", EventType: "port.scan", Timestamp: base.Add(time.Second), CreatedAt: base.Add(time.Second)},
	}
	for index := range events {
		if err := db.Create(&events[index]).Error; err != nil {
			t.Fatal(err)
		}
	}
	return events
}

func TestLoadOrCreateCheckpointUsesNullableInitialCursor(t *testing.T) {
	db := migrationDB(t)
	checkpoint, err := loadOrCreateEventMigrationCheckpoint(db, "test-null-cursor")
	if err != nil {
		t.Fatal(err)
	}
	if checkpoint.LastCreatedAt != nil || checkpoint.LastEventID != "" {
		t.Fatalf("initial cursor must be NULL/empty: %#v", checkpoint)
	}
	var nullCount int64
	if err := db.Raw("SELECT COUNT(*) FROM event_migration_checkpoints WHERE name = ? AND last_created_at IS NULL", checkpoint.Name).Scan(&nullCount).Error; err != nil {
		t.Fatal(err)
	}
	if nullCount != 1 {
		t.Fatalf("initial checkpoint persisted a non-NULL/zero date cursor: count=%d", nullCount)
	}
}

func TestMigrateLegacyEventsOrdersAndCompletes(t *testing.T) {
	db := migrationDB(t)
	seedMigrationEvents(t, db)
	writer := &migrationWriter{}
	checkpoint, err := MigrateLegacyEvents(context.Background(), db, writer, LegacyEventMigrationOptions{Name: "test-complete", BatchSize: 2})
	if err != nil {
		t.Fatal(err)
	}
	if checkpoint.Status != "completed" || checkpoint.MigratedEvents != 3 || checkpoint.LastEventID != "event-c" {
		t.Fatalf("unexpected checkpoint: %#v", checkpoint)
	}
	if len(writer.batches) != 2 || writer.batches[0][0] != "event-a" || writer.batches[0][1] != "event-b" || writer.batches[1][0] != "event-c" {
		t.Fatalf("unexpected batches: %#v", writer.batches)
	}
	var legacyCount int64
	if err := db.Model(&AttackEvent{}).Count(&legacyCount).Error; err != nil || legacyCount != 3 {
		t.Fatalf("legacy events were modified: count=%d err=%v", legacyCount, err)
	}
}

func TestMigrateLegacyEventsResumesAfterFailure(t *testing.T) {
	db := migrationDB(t)
	seedMigrationEvents(t, db)
	failing := &migrationWriter{failAt: 2}
	checkpoint, err := MigrateLegacyEvents(context.Background(), db, failing, LegacyEventMigrationOptions{Name: "test-resume", BatchSize: 2})
	if err == nil {
		t.Fatal("expected ClickHouse failure")
	}
	if checkpoint.Status != "failed" || checkpoint.MigratedEvents != 2 || checkpoint.LastEventID != "event-b" {
		t.Fatalf("unexpected failed checkpoint: %#v", checkpoint)
	}

	recovered := &migrationWriter{}
	checkpoint, err = MigrateLegacyEvents(context.Background(), db, recovered, LegacyEventMigrationOptions{Name: "test-resume", BatchSize: 2})
	if err != nil {
		t.Fatal(err)
	}
	if checkpoint.Status != "completed" || checkpoint.MigratedEvents != 3 || len(recovered.batches) != 1 || recovered.batches[0][0] != "event-c" {
		t.Fatalf("migration did not resume: checkpoint=%#v batches=%#v", checkpoint, recovered.batches)
	}
}

func TestMigrateLegacyEventsCompletedIsNoOp(t *testing.T) {
	db := migrationDB(t)
	seedMigrationEvents(t, db)
	writer := &migrationWriter{}
	if _, err := MigrateLegacyEvents(context.Background(), db, writer, LegacyEventMigrationOptions{Name: "test-noop", BatchSize: 10}); err != nil {
		t.Fatal(err)
	}
	second := &migrationWriter{}
	checkpoint, err := MigrateLegacyEvents(context.Background(), db, second, LegacyEventMigrationOptions{Name: "test-noop", BatchSize: 10})
	if err != nil {
		t.Fatal(err)
	}
	if checkpoint.Status != "completed" || len(second.batches) != 0 {
		t.Fatalf("completed migration reran: %#v", second.batches)
	}
}

func TestMigrateLegacyEventsPicksUpRollbackTail(t *testing.T) {
	db := migrationDB(t)
	events := seedMigrationEvents(t, db)
	writer := &migrationWriter{}
	checkpoint, err := MigrateLegacyEvents(context.Background(), db, writer, LegacyEventMigrationOptions{Name: "test-tail", BatchSize: 10})
	if err != nil {
		t.Fatal(err)
	}
	tailTime := events[len(events)-1].CreatedAt.Add(time.Second)
	tail := AttackEvent{EventID: "event-after-rollback", NodeID: "node-1", EventType: "web.request", Timestamp: tailTime, CreatedAt: tailTime}
	if err := db.Create(&tail).Error; err != nil {
		t.Fatal(err)
	}

	recovered := &migrationWriter{}
	checkpoint, err = MigrateLegacyEvents(context.Background(), db, recovered, LegacyEventMigrationOptions{Name: "test-tail", BatchSize: 10})
	if err != nil {
		t.Fatal(err)
	}
	if checkpoint.Status != "completed" || checkpoint.MigratedEvents != 4 || len(recovered.batches) != 1 || recovered.batches[0][0] != tail.EventID {
		t.Fatalf("rollback tail was not migrated: checkpoint=%#v batches=%#v", checkpoint, recovered.batches)
	}
}

func TestTruncateMigrationErrorRedactsClickHousePassword(t *testing.T) {
	message := truncateMigrationError("write clickhouse://writer:super-secret@127.0.0.1:9000 failed")
	if strings.Contains(message, "super-secret") || !strings.Contains(message, "[redacted]") {
		t.Fatalf("secret was not redacted: %q", message)
	}
}
