package client

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/honeynet/honeynet/internal/agent/queue"
)

func (c *Client) runUploader(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		case <-c.queue.Notify():
		}
		if err := c.flush(ctx); err != nil {
			log.Printf("upload events: %v", err)
		}
	}
}
func (c *Client) flush(ctx context.Context) error {
	events := c.queue.Batch(100)
	if len(events) == 0 {
		return nil
	}
	var body bytes.Buffer
	writer := gzip.NewWriter(&body)
	if err := json.NewEncoder(writer).Encode(events); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.AgentURL+"/agent/v1/events:batch", &body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	resp, err := c.httpClient().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var result apiEnvelope[struct {
		AckIDs []string `json:"ack_ids"`
		Reject []struct {
			EventID string `json:"event_id"`
			Reason  string `json:"reason"`
			Fatal   bool   `json:"fatal"`
		} `json:"reject"`
	}]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d: %s", resp.StatusCode, result.Message)
	}
	fatal := make([]queue.DeadLetterRequest, 0, len(result.Data.Reject))
	for _, rejected := range result.Data.Reject {
		if rejected.Fatal {
			fatal = append(fatal, queue.DeadLetterRequest{EventID: rejected.EventID, Reason: rejected.Reason})
		}
	}
	if len(fatal) > 0 {
		moved, err := c.queue.MoveToDeadLetter(fatal)
		if err != nil {
			return fmt.Errorf("persist permanently rejected events: %w", err)
		}
		for _, record := range moved {
			log.Printf("event %s permanently rejected (%s); preserved in %s", record.Event.EventID, record.Reason, c.queue.DeadLetterPath())
		}
	}
	return c.queue.Ack(result.Data.AckIDs)
}
