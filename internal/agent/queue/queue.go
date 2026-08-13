package queue

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/honeynet/honeynet/internal/agent/protocol"
)

const maxPendingEvents = 10000

var ErrFull = errors.New("durable event queue is full")

type Queue struct {
	mu             sync.Mutex
	path           string
	deadLetterPath string
	deadLetterIDs  map[string]struct{}
	events         []protocol.Event
	notify         chan struct{}
}

// DeadLetterRequest identifies a server rejection that is known to be
// permanent. Retryable rejects must never be passed to MoveToDeadLetter.
type DeadLetterRequest struct {
	EventID string
	Reason  string
}

// DeadLetterRecord preserves the complete forensic event together with the
// server's rejection reason and the time the Agent stopped retrying it.
type DeadLetterRecord struct {
	Event      protocol.Event `json:"event"`
	Reason     string         `json:"reason"`
	RejectedAt time.Time      `json:"rejected_at"`
}

func Open(path string) (*Queue, error) {
	q := &Queue{
		path:           path,
		deadLetterPath: filepath.Join(filepath.Dir(path), "dead-letter-events.jsonl"),
		deadLetterIDs:  make(map[string]struct{}),
		notify:         make(chan struct{}, 1),
	}
	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	if len(data) > 0 {
		if err := json.Unmarshal(data, &q.events); err != nil {
			return nil, err
		}
	}
	if err := q.loadDeadLetterIDs(); err != nil {
		return nil, fmt.Errorf("read dead-letter events: %w", err)
	}
	return q, nil
}

func (q *Queue) Add(event protocol.Event) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if event.EventID == "" {
		event.EventID = uuid.NewString()
	}
	if len(q.events) >= maxPendingEvents {
		// Never discard an unacknowledged forensic event silently. Backpressure
		// is preferable to corrupting the evidence trail; the caller logs this
		// condition and the node exposes queue depth in heartbeats.
		return ErrFull
	}
	next := append(append([]protocol.Event(nil), q.events...), event)
	if err := q.persistEventsLocked(next); err != nil {
		return err
	}
	q.events = next
	select {
	case q.notify <- struct{}{}:
	default:
	}
	return nil
}

func (q *Queue) Batch(limit int) []protocol.Event {
	q.mu.Lock()
	defer q.mu.Unlock()
	if limit <= 0 || limit > len(q.events) {
		limit = len(q.events)
	}
	return append([]protocol.Event(nil), q.events[:limit]...)
}

func (q *Queue) Ack(ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	set := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		set[id] = struct{}{}
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	kept := make([]protocol.Event, 0, len(q.events))
	for _, event := range q.events {
		if _, ok := set[event.EventID]; !ok {
			kept = append(kept, event)
		}
	}
	if err := q.persistEventsLocked(kept); err != nil {
		return err
	}
	q.events = kept
	return nil
}

// MoveToDeadLetter first appends and fsyncs complete events to a mode-0600
// JSONL file. Only after that durable write succeeds are those events removed
// from the pending queue. Valid records already present in the JSONL file are
// not appended again after a crash between these two operations.
func (q *Queue) MoveToDeadLetter(requests []DeadLetterRequest) ([]DeadLetterRecord, error) {
	if len(requests) == 0 {
		return nil, nil
	}
	reasons := make(map[string]string, len(requests))
	for _, request := range requests {
		if request.EventID != "" {
			reasons[request.EventID] = request.Reason
		}
	}
	if len(reasons) == 0 {
		return nil, nil
	}

	q.mu.Lock()
	defer q.mu.Unlock()
	moved := make([]DeadLetterRecord, 0, len(reasons))
	toAppend := make([]DeadLetterRecord, 0, len(reasons))
	remove := make(map[string]struct{}, len(reasons))
	for _, event := range q.events {
		reason, rejected := reasons[event.EventID]
		if !rejected {
			continue
		}
		if _, duplicate := remove[event.EventID]; duplicate {
			continue
		}
		record := DeadLetterRecord{Event: event, Reason: reason, RejectedAt: time.Now().UTC()}
		moved = append(moved, record)
		remove[event.EventID] = struct{}{}
		if _, exists := q.deadLetterIDs[event.EventID]; !exists {
			toAppend = append(toAppend, record)
		}
	}
	if len(moved) == 0 {
		return nil, nil
	}
	if len(toAppend) > 0 {
		if err := q.appendDeadLettersLocked(toAppend); err != nil {
			return nil, err
		}
		for _, record := range toAppend {
			q.deadLetterIDs[record.Event.EventID] = struct{}{}
		}
	}
	kept := make([]protocol.Event, 0, len(q.events)-len(moved))
	for _, event := range q.events {
		if _, rejected := remove[event.EventID]; !rejected {
			kept = append(kept, event)
		}
	}
	if err := q.persistEventsLocked(kept); err != nil {
		return nil, err
	}
	q.events = kept
	return moved, nil
}

func (q *Queue) Len() int                { q.mu.Lock(); defer q.mu.Unlock(); return len(q.events) }
func (q *Queue) Notify() <-chan struct{} { return q.notify }
func (q *Queue) DeadLetterPath() string  { return q.deadLetterPath }

func (q *Queue) persistEventsLocked(events []protocol.Event) error {
	if err := os.MkdirAll(filepath.Dir(q.path), 0700); err != nil {
		return err
	}
	data, err := json.Marshal(events)
	if err != nil {
		return err
	}
	tmp := q.path + ".tmp"
	file, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	if err := file.Chmod(0600); err != nil {
		_ = file.Close()
		return err
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, q.path)
}

func (q *Queue) appendDeadLettersLocked(records []DeadLetterRecord) error {
	if err := os.MkdirAll(filepath.Dir(q.deadLetterPath), 0700); err != nil {
		return err
	}
	needsSeparator := false
	if existing, err := os.Open(q.deadLetterPath); err == nil {
		if info, statErr := existing.Stat(); statErr != nil {
			_ = existing.Close()
			return statErr
		} else if info.Size() > 0 {
			last := []byte{0}
			if _, readErr := existing.ReadAt(last, info.Size()-1); readErr != nil {
				_ = existing.Close()
				return readErr
			}
			needsSeparator = last[0] != '\n'
		}
		if err := existing.Close(); err != nil {
			return err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	var payload bytes.Buffer
	if needsSeparator {
		payload.WriteByte('\n')
	}
	encoder := json.NewEncoder(&payload)
	for _, record := range records {
		if err := encoder.Encode(record); err != nil {
			return err
		}
	}
	file, err := os.OpenFile(q.deadLetterPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	if err := file.Chmod(0600); err != nil {
		_ = file.Close()
		return err
	}
	if _, err := io.Copy(file, &payload); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

func (q *Queue) loadDeadLetterIDs() error {
	file, err := os.Open(q.deadLetterPath)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64<<10), 4<<20)
	for scanner.Scan() {
		var record DeadLetterRecord
		// A power loss can leave one partial JSONL line. Keep the pending event
		// retryable and ignore only that malformed dead-letter fragment.
		if json.Unmarshal(scanner.Bytes(), &record) == nil && record.Event.EventID != "" {
			q.deadLetterIDs[record.Event.EventID] = struct{}{}
		}
	}
	return scanner.Err()
}
