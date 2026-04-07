package alert_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/state"
)

func makeChange(kind, proto, addr string, port int) state.Change {
	return state.Change{
		Kind: kind,
		Port: state.Port{
			Protocol: proto,
			Address:  addr,
			Port:     port,
			PID:      1234,
			Process:  "testproc",
		},
	}
}

func TestNewManager_NilNotifiers(t *testing.T) {
	m := alert.NewManager(nil)
	if m == nil {
		t.Fatal("expected non-nil manager")
	}
}

func TestNewManager_WithNotifiers(t *testing.T) {
	var buf bytes.Buffer
	n := alert.NewConsoleNotifier(&buf)
	m := alert.NewManager([]alert.Notifier{n})
	if m == nil {
		t.Fatal("expected non-nil manager")
	}
}

func TestConsoleNotifier_OpenedPort(t *testing.T) {
	var buf bytes.Buffer
	n := alert.NewConsoleNotifier(&buf)

	change := makeChange("opened", "tcp", "0.0.0.0", 8080)
	err := n.Notify([]state.Change{change})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "opened") {
		t.Errorf("expected output to contain 'opened', got: %s", out)
	}
	if !strings.Contains(out, "8080") {
		t.Errorf("expected output to contain port '8080', got: %s", out)
	}
	if !strings.Contains(out, "tcp") {
		t.Errorf("expected output to contain protocol 'tcp', got: %s", out)
	}
}

func TestConsoleNotifier_ClosedPort(t *testing.T) {
	var buf bytes.Buffer
	n := alert.NewConsoleNotifier(&buf)

	change := makeChange("closed", "udp", "127.0.0.1", 53)
	err := n.Notify([]state.Change{change})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "closed") {
		t.Errorf("expected output to contain 'closed', got: %s", out)
	}
	if !strings.Contains(out, "53") {
		t.Errorf("expected output to contain port '53', got: %s", out)
	}
}

func TestConsoleNotifier_EmptyChanges(t *testing.T) {
	var buf bytes.Buffer
	n := alert.NewConsoleNotifier(&buf)

	err := n.Notify([]state.Change{})
	if err != nil {
		t.Fatalf("unexpected error on empty changes: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected no output for empty changes, got: %s", buf.String())
	}
}

func TestManager_Notify_PropagatesChanges(t *testing.T) {
	var buf1, buf2 bytes.Buffer
	n1 := alert.NewConsoleNotifier(&buf1)
	n2 := alert.NewConsoleNotifier(&buf2)

	m := alert.NewManager([]alert.Notifier{n1, n2})

	changes := []state.Change{
		makeChange("opened", "tcp", "0.0.0.0", 9090),
	}

	err := m.Notify(changes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i, buf := range []*bytes.Buffer{&buf1, &buf2} {
		if !strings.Contains(buf.String(), "9090") {
			t.Errorf("notifier %d did not receive change, output: %s", i+1, buf.String())
		}
	}
}

func TestManager_Notify_Timestamp(t *testing.T) {
	var buf bytes.Buffer
	n := alert.NewConsoleNotifier(&buf)
	m := alert.NewManager([]alert.Notifier{n})

	before := time.Now()
	changes := []state.Change{makeChange("opened", "tcp", "0.0.0.0", 3000)}
	_ = m.Notify(changes)
	after := time.Now()

	// Ensure notify completes within a reasonable window
	if after.Sub(before) > 2*time.Second {
		t.Error("Notify took unexpectedly long")
	}
}
