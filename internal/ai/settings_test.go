package ai

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestAPIKeyEncryptionRoundTrip(t *testing.T) {
	key := sha256.Sum256([]byte("settings-key"))
	first, err := sealAPIKey(key, "sk-sensitive-value")
	if err != nil {
		t.Fatal(err)
	}
	second, err := sealAPIKey(key, "sk-sensitive-value")
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(first, second) {
		t.Fatal("ciphertexts must use independent nonces")
	}
	if bytes.Contains(first, []byte("sk-sensitive-value")) {
		t.Fatal("ciphertext contains plaintext API Key")
	}
	plaintext, err := openAPIKey(key, first)
	if err != nil || plaintext != "sk-sensitive-value" {
		t.Fatalf("unexpected decryption result %q: %v", plaintext, err)
	}
	wrongKey := sha256.Sum256([]byte("wrong-key"))
	if _, err := openAPIKey(wrongKey, first); err == nil {
		t.Fatal("decrypting with a different master key must fail")
	}
}

func TestSettingsUpdatePreservesBlankAPIKey(t *testing.T) {
	current := Config{Enabled: true, Provider: "deepseek", BaseURL: "https://api.example/v1", APIKey: "existing-secret", Model: "deepseek-chat", Timeout: 45 * time.Second}
	blank, mask := "", SecretMask
	for _, apiKey := range []*string{nil, &blank, &mask} {
		updated, err := applySettingsUpdate(current, SettingsUpdate{APIKey: apiKey})
		if err != nil {
			t.Fatal(err)
		}
		if updated.APIKey != current.APIKey {
			t.Fatalf("blank or masked key did not preserve the stored value: %q", updated.APIKey)
		}
	}
	newKey := "replacement-secret"
	updated, err := applySettingsUpdate(current, SettingsUpdate{APIKey: &newKey})
	if err != nil || updated.APIKey != newKey {
		t.Fatalf("new API Key was not applied: %#v, %v", updated, err)
	}
	disabled := false
	cleared, err := applySettingsUpdate(current, SettingsUpdate{Enabled: &disabled, ClearAPIKey: true})
	if err != nil || cleared.APIKey != "" {
		t.Fatalf("explicit key clearing failed: %#v, %v", cleared, err)
	}
	if _, err := applySettingsUpdate(current, SettingsUpdate{ClearAPIKey: true}); !errors.Is(err, ErrInvalidSettings) {
		t.Fatalf("enabled config without a key should fail, got %v", err)
	}
}

func TestSettingsViewNeverSerializesAPIKey(t *testing.T) {
	settings := StoredSettings{Config: Config{Enabled: true, Provider: "glm", BaseURL: "https://api.example/v1", APIKey: "never-return-this", Model: "glm-4", Timeout: time.Minute}, Revision: 7}
	raw, err := json.Marshal(settings.View())
	if err != nil {
		t.Fatal(err)
	}
	serialized := string(raw)
	var fields map[string]any
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(serialized, "never-return-this") {
		t.Fatalf("API Key leaked in settings response: %s", serialized)
	}
	if _, exists := fields["api_key"]; exists {
		t.Fatalf("settings response must not contain an api_key field: %s", serialized)
	}
	if !strings.Contains(serialized, `"has_api_key":true`) {
		t.Fatalf("settings response omitted key presence indicator: %s", serialized)
	}
}

func TestRuntimeHotReplaceAndConcurrentReads(t *testing.T) {
	provider := func(expectedKey, summary string) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
			if request.Header.Get("Authorization") != "Bearer "+expectedKey {
				t.Errorf("unexpected provider credential: %q", request.Header.Get("Authorization"))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"message": map[string]any{"content": fmt.Sprintf(`{"summary":%q}`, summary)}}}})
		}))
	}
	first := provider("first-key", "first")
	defer first.Close()
	second := provider("second-key", "second")
	defer second.Close()

	runtime, err := NewRuntime(Config{Enabled: true, Provider: "first", BaseURL: first.URL, APIKey: "first-key", Model: "model-a", Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	response, err := runtime.Analyze(context.Background(), Request{Task: "test", Evidence: map[string]any{}})
	if err != nil || response.Provider != "first" {
		t.Fatalf("first runtime request failed: %#v, %v", response, err)
	}
	if err := runtime.Replace(Config{Enabled: true, Provider: "second", BaseURL: second.URL, APIKey: "second-key", Model: "model-b", Timeout: time.Second, SendRawPacket: true}); err != nil {
		t.Fatal(err)
	}
	response, err = runtime.Analyze(context.Background(), Request{Task: "test", Evidence: map[string]any{}})
	if err != nil || response.Provider != "second" || !runtime.Status().SendRawPacket {
		t.Fatalf("replacement runtime request failed: %#v, %v", response, err)
	}

	var wait sync.WaitGroup
	for i := 0; i < 20; i++ {
		wait.Add(2)
		go func() {
			defer wait.Done()
			_, _ = runtime.Analyze(context.Background(), Request{Task: "test", Evidence: map[string]any{}})
		}()
		go func() {
			defer wait.Done()
			_ = runtime.Replace(Config{Enabled: true, Provider: "second", BaseURL: second.URL, APIKey: "second-key", Model: "model-b", Timeout: time.Second})
		}()
	}
	wait.Wait()
}

func TestProviderErrorRedactsAPIKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = fmt.Fprint(w, `{"error":"credential sk-secret-value was rejected"}`)
	}))
	defer server.Close()
	client := New(Config{Enabled: true, BaseURL: server.URL, APIKey: "sk-secret-value", Model: "test", Timeout: time.Second})
	_, err := client.Analyze(context.Background(), Request{Task: "test", Evidence: map[string]any{}})
	if err == nil || strings.Contains(err.Error(), "sk-secret-value") || !strings.Contains(err.Error(), "[redacted]") {
		t.Fatalf("provider error was not redacted: %v", err)
	}
}
