package alerting

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/honeynet/honeynet/internal/store"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const maxAttempts = 4

type Dispatcher struct {
	db         *gorm.DB
	sender     *Sender
	consoleURL string
	wake       chan struct{}
}

func NewDispatcher(db *gorm.DB, consoleURL string) *Dispatcher {
	return &Dispatcher{db: db, sender: NewSender(), consoleURL: strings.TrimRight(consoleURL, "/"), wake: make(chan struct{}, 1)}
}

func (d *Dispatcher) Start(ctx context.Context) {
	d.db.Model(&store.AlertDelivery{}).Where("status = ?", "sending").Updates(map[string]any{"status": "retrying", "next_attempt": time.Now()})
	go d.loop(ctx)
}

func (d *Dispatcher) Enqueue(alert store.Alert, channelIDs []string) error {
	return d.EnqueueWithDB(d.db, alert, channelIDs)
}

// EnqueueWithDB allows event ingestion to create AlertDelivery rows in the
// same MySQL transaction as the alert and receipt. Delivery dispatch remains
// asynchronous; the unique alert/channel index makes retries idempotent.
func (d *Dispatcher) EnqueueWithDB(db *gorm.DB, alert store.Alert, channelIDs []string) error {
	query := db.Where("enabled = ?", true)
	if len(channelIDs) > 0 {
		query = query.Where("id IN ?", channelIDs)
	}
	var channels []store.AlertChannel
	if err := query.Find(&channels).Error; err != nil {
		return err
	}
	now := time.Now()
	for _, channel := range channels {
		delivery := store.AlertDelivery{
			Base: store.NewBase(), AlertID: alert.ID, ChannelID: channel.ID, ChannelName: channel.Name,
			ChannelType: channel.Type, Status: "pending", NextAttempt: now,
		}
		if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&delivery).Error; err != nil {
			return err
		}
	}
	if db == d.db {
		d.notify()
	}
	return nil
}

func (d *Dispatcher) Notify() { d.notify() }

func (d *Dispatcher) Test(ctx context.Context, channel store.AlertChannel) error {
	ctx, cancel := context.WithTimeout(ctx, 12*time.Second)
	defer cancel()
	if err := d.sender.Send(ctx, channel, TestMessage(d.consoleURL)); err != nil {
		return fmt.Errorf("%s", redactError(channel, err))
	}
	return nil
}

func (d *Dispatcher) Retry(id string) error {
	result := d.db.Model(&store.AlertDelivery{}).Where("id = ?", id).Updates(map[string]any{
		"status": "pending", "attempt": 0, "last_error": "", "next_attempt": time.Now(), "delivered_at": nil,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	d.notify()
	return nil
}

func (d *Dispatcher) ProcessDue(ctx context.Context) error {
	var deliveries []store.AlertDelivery
	if err := d.db.Where("status IN ? AND next_attempt <= ?", []string{"pending", "retrying"}, time.Now()).Order("next_attempt ASC").Limit(20).Find(&deliveries).Error; err != nil {
		return err
	}
	for i := range deliveries {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		d.process(ctx, &deliveries[i])
	}
	return nil
}

func (d *Dispatcher) loop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		if err := d.ProcessDue(ctx); err != nil && ctx.Err() == nil {
			log.Printf("alert dispatcher: %v", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-d.wake:
		case <-ticker.C:
		}
	}
}

func (d *Dispatcher) process(parent context.Context, delivery *store.AlertDelivery) {
	now := time.Now()
	claimed := d.db.Model(&store.AlertDelivery{}).
		Where("id = ? AND status IN ?", delivery.ID, []string{"pending", "retrying"}).
		Updates(map[string]any{"status": "sending", "attempt": gorm.Expr("attempt + 1"), "last_attempt": now})
	if claimed.Error != nil || claimed.RowsAffected == 0 {
		return
	}
	delivery.Attempt++
	var alert store.Alert
	var channel store.AlertChannel
	if d.db.First(&alert, "id = ?", delivery.AlertID).Error != nil || d.db.First(&channel, "id = ?", delivery.ChannelID).Error != nil {
		d.failPermanently(delivery.ID, "alert or channel no longer exists")
		return
	}
	if !channel.Enabled {
		d.db.Model(delivery).Updates(map[string]any{"status": "skipped", "last_error": "channel is disabled"})
		return
	}
	ctx, cancel := context.WithTimeout(parent, 12*time.Second)
	err := d.sender.Send(ctx, channel, Message{Alert: alert, ConsoleURL: d.consoleURL})
	cancel()
	if err == nil {
		d.db.Model(delivery).Updates(map[string]any{"status": "sent", "last_error": "", "delivered_at": time.Now()})
		return
	}
	message := redactError(channel, err)
	if delivery.Attempt >= maxAttempts {
		d.failPermanently(delivery.ID, message)
		return
	}
	d.db.Model(delivery).Updates(map[string]any{
		"status": "retrying", "last_error": message, "next_attempt": time.Now().Add(retryDelay(delivery.Attempt)),
	})
}

func (d *Dispatcher) failPermanently(id, message string) {
	d.db.Model(&store.AlertDelivery{}).Where("id = ?", id).Updates(map[string]any{"status": "failed", "last_error": truncate(message, 1000)})
}

func (d *Dispatcher) notify() {
	select {
	case d.wake <- struct{}{}:
	default:
	}
}

func retryDelay(attempt int) time.Duration {
	switch attempt {
	case 1:
		return 5 * time.Second
	case 2:
		return 30 * time.Second
	default:
		return 2 * time.Minute
	}
}

func redactError(channel store.AlertChannel, err error) string {
	message := err.Error()
	values := map[string]any{}
	_ = json.Unmarshal(channel.Config, &values)
	for _, key := range append(secretKeys(channel.Type), "password") {
		if value := strings.TrimSpace(fmt.Sprint(values[key])); value != "" && value != "<nil>" {
			message = strings.ReplaceAll(message, value, "[redacted]")
		}
	}
	if headers, ok := values["headers"].(map[string]any); ok {
		for _, value := range headers {
			message = strings.ReplaceAll(message, fmt.Sprint(value), "[redacted]")
		}
	}
	return truncate(message, 1000)
}
