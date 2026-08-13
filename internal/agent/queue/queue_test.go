package queue

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/honeynet/honeynet/internal/agent/protocol"
)

func TestQueuePersistsUntilAck(t *testing.T) {
	path := filepath.Join(t.TempDir(), "events.json")
	q, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = q.Add(protocol.Event{EventID: "event-1", EventType: "web.request"}); err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if reopened.Len() != 1 || reopened.Batch(1)[0].EventID != "event-1" {
		t.Fatalf("reopened queue = %#v", reopened.Batch(10))
	}
	if err = reopened.Ack([]string{"event-1"}); err != nil {
		t.Fatal(err)
	}
	again, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if again.Len() != 0 {
		t.Fatalf("queue length after ack = %d", again.Len())
	}
}

func TestQueueFullPreservesOldestUnacknowledgedEvents(t *testing.T) {
	path := filepath.Join(t.TempDir(), "events.json")
	q, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	q.events = make([]protocol.Event, maxPendingEvents)
	for index := range q.events {
		q.events[index] = protocol.Event{EventID: "pending"}
	}
	if err := q.Add(protocol.Event{EventID: "new-event", EventType: "web.request"}); err != ErrFull {
		t.Fatalf("Add error = %v, want ErrFull", err)
	}
	if q.Len() != maxPendingEvents || q.Batch(1)[0].EventID != "pending" {
		t.Fatalf("full queue mutated existing evidence")
	}
}

func TestMoveToDeadLetterPersistsCompleteEventBeforeRemoval(t *testing.T) {
	path := filepath.Join(t.TempDir(), "events.json")
	q, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	event := protocol.Event{
		EventID: "event-fatal", EventType: "web.request", RawPacket: "GET /secret HTTP/1.1",
		Payload: map[string]any{"path": "/secret", "credential": "evidence"}, Tags: []string{"web"},
	}
	if err := q.Add(event); err != nil {
		t.Fatal(err)
	}
	moved, err := q.MoveToDeadLetter([]DeadLetterRequest{{EventID: event.EventID, Reason: "permanently invalid"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(moved) != 1 || q.Len() != 0 {
		t.Fatalf("moved=%#v pending=%d", moved, q.Len())
	}
	info, err := os.Stat(q.DeadLetterPath())
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("dead-letter mode=%#o, want 0600", info.Mode().Perm())
	}
	file, err := os.Open(q.DeadLetterPath())
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	var record DeadLetterRecord
	if err := json.NewDecoder(file).Decode(&record); err != nil {
		t.Fatal(err)
	}
	if record.Event.EventID != event.EventID || record.Event.RawPacket != event.RawPacket || record.Event.Payload["credential"] != "evidence" || record.Reason != "permanently invalid" || record.RejectedAt.IsZero() {
		t.Fatalf("dead-letter record=%#v", record)
	}
	reopened, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if reopened.Len() != 0 {
		t.Fatalf("reopened pending=%d", reopened.Len())
	}
}

func TestMoveToDeadLetterFailureKeepsPendingEvent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "events.json")
	q, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := q.Add(protocol.Event{EventID: "event-kept", EventType: "web.request", Payload: map[string]any{}}); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(q.DeadLetterPath(), 0700); err != nil {
		t.Fatal(err)
	}
	if _, err := q.MoveToDeadLetter([]DeadLetterRequest{{EventID: "event-kept", Reason: "fatal"}}); err == nil {
		t.Fatal("expected dead-letter write failure")
	}
	if q.Len() != 1 || q.Batch(1)[0].EventID != "event-kept" {
		t.Fatalf("pending evidence was removed: %#v", q.Batch(10))
	}
	reopened, err := Open(path)
	if err == nil {
		// A directory at the dead-letter path is itself an invalid state; remove
		// it to prove the still-persisted main queue can be reopened.
		t.Fatalf("Open unexpectedly accepted invalid dead-letter path: %#v", reopened)
	}
	if err := os.Remove(q.DeadLetterPath()); err != nil {
		t.Fatal(err)
	}
	reopened, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if reopened.Len() != 1 {
		t.Fatalf("reopened pending=%d", reopened.Len())
	}
}

func TestMoveToDeadLetterDoesNotDuplicateRecordAfterInterruptedRemoval(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "events.json")
	q, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	event := protocol.Event{EventID: "event-recovery", EventType: "web.request", Payload: map[string]any{"path": "/"}}
	if err := q.Add(event); err != nil {
		t.Fatal(err)
	}
	// Simulate a process loss after JSONL fsync but before rewriting the main
	// queue: both files contain the same event when the Agent restarts.
	record := DeadLetterRecord{Event: event, Reason: "fatal", RejectedAt: time.Now().UTC()}
	data, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(filepath.Join(dir, "dead-letter-events.jsonl"), data, 0600); err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := reopened.MoveToDeadLetter([]DeadLetterRequest{{EventID: event.EventID, Reason: "fatal"}}); err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(reopened.DeadLetterPath())
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	lines := 0
	for scanner.Scan() {
		if len(scanner.Bytes()) > 0 {
			lines++
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if lines != 1 || reopened.Len() != 0 {
		t.Fatalf("dead-letter lines=%d pending=%d", lines, reopened.Len())
	}
}
