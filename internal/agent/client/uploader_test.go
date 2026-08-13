package client

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	agentconfig "github.com/honeynet/honeynet/internal/agent/config"
	"github.com/honeynet/honeynet/internal/agent/protocol"
	"github.com/honeynet/honeynet/internal/agent/queue"
)

func TestFlushMovesOnlyFatalRejectsToDeadLetter(t *testing.T) {
	fatalID, retryID := "fatal-event", "retry-event"
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		decompressed, err := gzip.NewReader(request.Body)
		if err != nil {
			t.Errorf("gzip request: %v", err)
			http.Error(writer, "bad gzip", http.StatusBadRequest)
			return
		}
		defer decompressed.Close()
		var events []protocol.Event
		if err := json.NewDecoder(decompressed).Decode(&events); err != nil || len(events) != 2 {
			t.Errorf("events=%#v err=%v", events, err)
		}
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(map[string]any{"data": map[string]any{
			"ack_ids": []string{},
			"reject": []map[string]any{
				{"event_id": fatalID, "reason": "invalid UUID", "fatal": true},
				{"event_id": retryID, "reason": "temporary database failure"},
			},
		}})
	}))
	defer server.Close()

	client, eventQueue := uploaderTestClient(t, server.URL)
	if err := eventQueue.Add(protocol.Event{EventID: fatalID, EventType: "web.request", RawPacket: "evidence", Payload: map[string]any{}}); err != nil {
		t.Fatal(err)
	}
	if err := eventQueue.Add(protocol.Event{EventID: retryID, EventType: "web.request", Payload: map[string]any{}}); err != nil {
		t.Fatal(err)
	}
	if err := client.flush(context.Background()); err != nil {
		t.Fatal(err)
	}
	remaining := eventQueue.Batch(10)
	if len(remaining) != 1 || remaining[0].EventID != retryID {
		t.Fatalf("remaining=%#v", remaining)
	}
	data, err := os.ReadFile(eventQueue.DeadLetterPath())
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid(data[:len(data)-1]) { // JSONL contains one JSON record plus newline.
		t.Fatalf("dead-letter=%q", data)
	}
}

func TestFlushRemainsCompatibleWithLegacyAckResponse(t *testing.T) {
	eventID := "legacy-event"
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"data":{"ack_ids":["legacy-event"]}}`))
	}))
	defer server.Close()
	client, eventQueue := uploaderTestClient(t, server.URL)
	if err := eventQueue.Add(protocol.Event{EventID: eventID, EventType: "web.request", Payload: map[string]any{}}); err != nil {
		t.Fatal(err)
	}
	if err := client.flush(context.Background()); err != nil {
		t.Fatal(err)
	}
	if eventQueue.Len() != 0 {
		t.Fatalf("legacy ACK left %d events", eventQueue.Len())
	}
	if _, err := os.Stat(eventQueue.DeadLetterPath()); !os.IsNotExist(err) {
		t.Fatalf("unexpected dead-letter file: %v", err)
	}
}

func uploaderTestClient(t *testing.T, serverURL string) (*Client, *queue.Queue) {
	t.Helper()
	eventQueue, err := queue.Open(filepath.Join(t.TempDir(), "pending-events.json"))
	if err != nil {
		t.Fatal(err)
	}
	return &Client{cfg: &agentconfig.Config{AgentURL: serverURL}, http: http.DefaultClient, queue: eventQueue}, eventQueue
}
