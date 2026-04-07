// Package notifier provides additional notification backends for portwatch.
package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/user/portwatch/internal/state"
)

// WebhookPayload is the JSON body sent to the webhook endpoint.
type WebhookPayload struct {
	Timestamp string         `json:"timestamp"`
	Changes   []ChangeDetail `json:"changes"`
}

// ChangeDetail describes a single port change event.
type ChangeDetail struct {
	Type     string `json:"type"` // "opened" or "closed"
	Protocol string `json:"protocol"`
	Port     int    `json:"port"`
	Process  string `json:"process,omitempty"`
	PID      int    `json:"pid,omitempty"`
}

// WebhookNotifier sends port change alerts to an HTTP endpoint.
type WebhookNotifier struct {
	URL     string
	Timeout time.Duration
	client  *http.Client
}

// NewWebhookNotifier creates a WebhookNotifier that posts to the given URL.
// If timeout is zero, a default of 10 seconds is used.
func NewWebhookNotifier(url string, timeout time.Duration) *WebhookNotifier {
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return &WebhookNotifier{
		URL:     url,
		Timeout: timeout,
		client:  &http.Client{Timeout: timeout},
	}
}

// Notify serialises the diff as JSON and POSTs it to the configured endpoint.
// It returns an error if the HTTP request fails or the server responds with a
// non-2xx status code.
func (w *WebhookNotifier) Notify(diff state.Diff) error {
	if len(diff.Opened)+len(diff.Closed) == 0 {
		return nil
	}

	payload := buildPayload(diff)

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("webhook: marshal payload: %w", err)
	}

	resp, err := w.client.Post(w.URL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("webhook: post to %s: %w", w.URL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook: server returned %d for %s", resp.StatusCode, w.URL)
	}

	return nil
}

// buildPayload converts a state.Diff into a WebhookPayload.
func buildPayload(diff state.Diff) WebhookPayload {
	details := make([]ChangeDetail, 0, len(diff.Opened)+len(diff.Closed))

	for _, p := range diff.Opened {
		details = append(details, ChangeDetail{
			Type:     "opened",
			Protocol: p.Protocol,
			Port:     p.Port,
			Process:  p.Process,
			PID:      p.PID,
		})
	}

	for _, p := range diff.Closed {
		details = append(details, ChangeDetail{
			Type:     "closed",
			Protocol: p.Protocol,
			Port:     p.Port,
			Process:  p.Process,
			PID:      p.PID,
		})
	}

	return WebhookPayload{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Changes:   details,
	}
}
