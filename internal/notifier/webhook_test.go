package notifier_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/user/portwatch/internal/notifier"
	"github.com/user/portwatch/internal/state"
)

// helpers

func makeChange(kind, proto string, port uint16) state.Change {
	return state.Change{
		Kind: kind,
		Port: state.Port{
			Protocol: proto,
			Number:   port,
			PID:      1234,
			Process:  "testd",
		},
	}
}

// TestNewWebhookNotifier_InvalidURL verifies that an empty or invalid URL
// causes NewWebhookNotifier to return an error.
func TestNewWebhookNotifier_InvalidURL(t *testing.T) {
	_, err := notifier.NewWebhookNotifier("", 5*time.Second)
	if err == nil {
		t.Fatal("expected error for empty URL, got nil")
	}

	_, err = notifier.NewWebhookNotifier("not-a-url", 5*time.Second)
	if err == nil {
		t.Fatal("expected error for invalid URL, got nil")
	}
}

// TestNewWebhookNotifier_ValidURL ensures a valid URL produces no error.
func TestNewWebhookNotifier_ValidURL(t *testing.T) {
	_, err := notifier.NewWebhookNotifier("http://example.com/hook", 5*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestWebhookNotifier_SendsPostRequest verifies that Notify fires a POST
// request to the configured endpoint.
func TestWebhookNotifier_SendsPostRequest(t *testing.T) {
	var receivedMethod string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	wn, err := notifier.NewWebhookNotifier(server.URL, 5*time.Second)
	if err != nil {
		t.Fatalf("NewWebhookNotifier: %v", err)
	}

	changes := []state.Change{makeChange("opened", "tcp", 8080)}
	if err := wn.Notify(changes); err != nil {
		t.Fatalf("Notify returned error: %v", err)
	}

	if receivedMethod != http.MethodPost {
		t.Errorf("expected POST, got %s", receivedMethod)
	}
}

// TestWebhookNotifier_PayloadStructure checks that the JSON payload contains
// the expected top-level fields and change data.
func TestWebhookNotifier_PayloadStructure(t *testing.T) {
	var body map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	wn, _ := notifier.NewWebhookNotifier(server.URL, 5*time.Second)
	changes := []state.Change{
		makeChange("opened", "tcp", 9090),
		makeChange("closed", "udp", 53),
	}

	if err := wn.Notify(changes); err != nil {
		t.Fatalf("Notify: %v", err)
	}

	if _, ok := body["timestamp"]; !ok {
		t.Error("payload missing 'timestamp' field")
	}
	if _, ok := body["changes"]; !ok {
		t.Error("payload missing 'changes' field")
	}

	rawChanges, ok := body["changes"].([]interface{})
	if !ok {
		t.Fatalf("'changes' is not an array")
	}
	if len(rawChanges) != 2 {
		t.Errorf("expected 2 changes in payload, got %d", len(rawChanges))
	}
}

// TestWebhookNotifier_ServerError ensures that a non-2xx response is treated
// as an error by Notify.
func TestWebhookNotifier_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	wn, _ := notifier.NewWebhookNotifier(server.URL, 5*time.Second)
	changes := []state.Change{makeChange("opened", "tcp", 443)}

	if err := wn.Notify(changes); err == nil {
		t.Error("expected error for 500 response, got nil")
	}
}

// TestWebhookNotifier_Timeout verifies that a slow server causes a timeout
// error when the configured deadline is exceeded.
func TestWebhookNotifier_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	wn, _ := notifier.NewWebhookNotifier(server.URL, 50*time.Millisecond)
	changes := []state.Change{makeChange("opened", "tcp", 22)}

	if err := wn.Notify(changes); err == nil {
		t.Error("expected timeout error, got nil")
	}
}
