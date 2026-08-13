package store

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"
)

const LegacyEventsMigrationName = "legacy-mysql-events-to-clickhouse-v1"

var migrationSecretPattern = regexp.MustCompile(`(?i)(://[^:/@\s]+:)[^@\s]+(@)`)

type LegacyEventWriter interface {
	InsertAttackEvents(context.Context, []AttackEvent) error
}

type LegacyEventMigrationOptions struct {
	Name      string
	BatchSize int
}

// MigrateLegacyEvents copies the immutable legacy MySQL event history without
// deleting or modifying it. Checkpoint advancement happens only after the
// corresponding ClickHouse batch is committed. InsertAttackEvents is stable
// and idempotent by event_id, so a crash between those operations is safe.
func MigrateLegacyEvents(ctx context.Context, db *gorm.DB, writer LegacyEventWriter, options LegacyEventMigrationOptions) (EventMigrationCheckpoint, error) {
	if db == nil {
		return EventMigrationCheckpoint{}, errors.New("MySQL database is required")
	}
	if writer == nil {
		return EventMigrationCheckpoint{}, errors.New("ClickHouse event writer is required")
	}
	name := strings.TrimSpace(options.Name)
	if name == "" {
		name = LegacyEventsMigrationName
	}
	batchSize := options.BatchSize
	if batchSize <= 0 {
		batchSize = 500
	}
	if batchSize > 5000 {
		batchSize = 5000
	}
	checkpoint, err := loadOrCreateEventMigrationCheckpoint(db.WithContext(ctx), name)
	if err != nil {
		return EventMigrationCheckpoint{}, err
	}
	if checkpoint.Status == "completed" {
		hasMore, moreErr := hasLegacyEventsAfterCheckpoint(db.WithContext(ctx), checkpoint)
		if moreErr != nil {
			return checkpoint, moreErr
		}
		if !hasMore {
			return checkpoint, nil
		}
		// A rollback to a legacy Server can append MySQL events after a
		// completed migration. Re-open the same stable cursor and migrate only
		// that tail when the ClickHouse-enabled version starts again.
		checkpoint.CompletedAt = nil
	}
	if checkpoint.StartedAt == nil {
		now := time.Now().UTC()
		checkpoint.StartedAt = &now
	}
	// Defensive repair for a legacy/partially-created checkpoint: the cursor
	// tuple must advance atomically. Keeping only a timestamp could skip rows
	// sharing that timestamp on resume.
	if checkpoint.LastEventID == "" && checkpoint.LastCreatedAt != nil {
		checkpoint.LastCreatedAt = nil
		checkpoint.MigratedEvents = 0
	}
	if checkpoint.LastEventID != "" && checkpoint.LastCreatedAt == nil {
		checkpoint.LastEventID = ""
		checkpoint.MigratedEvents = 0
	}
	checkpoint.Status, checkpoint.LastError = "running", ""
	checkpoint.UpdatedAt = time.Now().UTC()
	if err := db.WithContext(ctx).Save(&checkpoint).Error; err != nil {
		return checkpoint, fmt.Errorf("start legacy event migration: %w", err)
	}

	for {
		if err := ctx.Err(); err != nil {
			return failEventMigration(db, checkpoint, err)
		}
		var events []AttackEvent
		query := db.WithContext(ctx).Order("created_at ASC").Order("event_id ASC").Limit(batchSize)
		if checkpoint.LastEventID != "" && checkpoint.LastCreatedAt != nil {
			query = query.Where("created_at > ? OR (created_at = ? AND event_id > ?)", *checkpoint.LastCreatedAt, *checkpoint.LastCreatedAt, checkpoint.LastEventID)
		}
		if err := query.Find(&events).Error; err != nil {
			return failEventMigration(db, checkpoint, fmt.Errorf("read legacy event batch: %w", err))
		}
		if len(events) == 0 {
			now := time.Now().UTC()
			checkpoint.Status, checkpoint.LastError, checkpoint.CompletedAt, checkpoint.UpdatedAt = "completed", "", &now, now
			if err := db.WithContext(ctx).Save(&checkpoint).Error; err != nil {
				return checkpoint, fmt.Errorf("complete legacy event migration: %w", err)
			}
			return checkpoint, nil
		}
		if err := writer.InsertAttackEvents(ctx, events); err != nil {
			return failEventMigration(db, checkpoint, fmt.Errorf("write legacy events to ClickHouse: %w", err))
		}
		last := events[len(events)-1]
		lastCreatedAt := last.CreatedAt
		checkpoint.LastCreatedAt, checkpoint.LastEventID = &lastCreatedAt, last.EventID
		checkpoint.MigratedEvents += int64(len(events))
		checkpoint.Status, checkpoint.LastError, checkpoint.UpdatedAt = "running", "", time.Now().UTC()
		if err := db.WithContext(ctx).Save(&checkpoint).Error; err != nil {
			return checkpoint, fmt.Errorf("persist legacy event migration checkpoint: %w", err)
		}
	}
}

func hasLegacyEventsAfterCheckpoint(db *gorm.DB, checkpoint EventMigrationCheckpoint) (bool, error) {
	query := db.Model(&AttackEvent{})
	if checkpoint.LastEventID != "" && checkpoint.LastCreatedAt != nil {
		query = query.Where("created_at > ? OR (created_at = ? AND event_id > ?)", *checkpoint.LastCreatedAt, *checkpoint.LastCreatedAt, checkpoint.LastEventID)
	}
	var count int64
	if err := query.Limit(1).Count(&count).Error; err != nil {
		return false, fmt.Errorf("check legacy event migration tail: %w", err)
	}
	return count > 0, nil
}

func loadOrCreateEventMigrationCheckpoint(db *gorm.DB, name string) (EventMigrationCheckpoint, error) {
	checkpoint := EventMigrationCheckpoint{Name: name}
	err := db.First(&checkpoint, "name = ?", name).Error
	if err == nil {
		return checkpoint, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return checkpoint, fmt.Errorf("load legacy event migration checkpoint: %w", err)
	}
	checkpoint.Status, checkpoint.UpdatedAt = "pending", time.Now().UTC()
	if err := db.Create(&checkpoint).Error; err != nil {
		if retryErr := db.First(&checkpoint, "name = ?", name).Error; retryErr == nil {
			return checkpoint, nil
		}
		return checkpoint, fmt.Errorf("create legacy event migration checkpoint: %w", err)
	}
	return checkpoint, nil
}

func failEventMigration(db *gorm.DB, checkpoint EventMigrationCheckpoint, migrationErr error) (EventMigrationCheckpoint, error) {
	checkpoint.Status = "failed"
	checkpoint.LastError = truncateMigrationError(migrationErr.Error())
	checkpoint.UpdatedAt = time.Now().UTC()
	if saveErr := db.Save(&checkpoint).Error; saveErr != nil {
		return checkpoint, fmt.Errorf("%v; persist migration failure: %w", migrationErr, saveErr)
	}
	return checkpoint, migrationErr
}

func truncateMigrationError(message string) string {
	message = migrationSecretPattern.ReplaceAllString(message, `${1}[redacted]${2}`)
	if len(message) <= 1024 {
		return message
	}
	return message[:1024]
}
