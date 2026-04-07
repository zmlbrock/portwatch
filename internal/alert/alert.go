// Package alert provides notification mechanisms for port change events.
package alert

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/user/portwatch/internal/scanner"
)

// Level represents the severity of an alert.
type Level string

const (
	LevelInfo    Level = "INFO"
	LevelWarning Level = "WARNING"
	LevelCritical Level = "CRITICAL"
)

// Event describes a detected port change.
type Event struct {
	Timestamp time.Time
	Level     Level
	Message   string
	Port      scanner.PortState
}

// Notifier is the interface that wraps the Notify method.
type Notifier interface {
	Notify(event Event) error
}

// Manager dispatches alerts to one or more notifiers.
type Manager struct {
	notifiers []Notifier
}

// NewManager creates a new alert Manager with the given notifiers.
func NewManager(notifiers ...Notifier) *Manager {
	return &Manager{notifiers: notifiers}
}

// Send dispatches an event to all registered notifiers.
// It collects and returns any errors encountered.
func (m *Manager) Send(event Event) error {
	var errs []error
	for _, n := range m.notifiers {
		if err := n.Notify(event); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("alert dispatch errors: %v", errs)
	}
	return nil
}

// ConsoleNotifier writes alert events to an io.Writer (typically stdout/stderr).
type ConsoleNotifier struct {
	out io.Writer
}

// NewConsoleNotifier creates a ConsoleNotifier that writes to the given writer.
// If w is nil, os.Stdout is used.
func NewConsoleNotifier(w io.Writer) *ConsoleNotifier {
	if w == nil {
		w = os.Stdout
	}
	return &ConsoleNotifier{out: w}
}

// Notify formats and writes the event to the configured writer.
func (c *ConsoleNotifier) Notify(event Event) error {
	_, err := fmt.Fprintf(
		c.out,
		"[%s] %s %-8s %s\n",
		event.Timestamp.Format(time.RFC3339),
		event.Level,
		formatPort(event.Port),
		event.Message,
	)
	return err
}

// NewPortEvent constructs an Event for a newly opened port.
func NewPortEvent(port scanner.PortState) Event {
	return Event{
		Timestamp: time.Now(),
		Level:     LevelWarning,
		Message:   "new port detected",
		Port:      port,
	}
}

// ClosedPortEvent constructs an Event for a port that is no longer open.
func ClosedPortEvent(port scanner.PortState) Event {
	return Event{
		Timestamp: time.Now(),
		Level:     LevelInfo,
		Message:   "port closed",
		Port:      port,
	}
}

// formatPort returns a human-readable representation of a PortState.
func formatPort(p scanner.PortState) string {
	return fmt.Sprintf("%s/%d", p.Protocol, p.Port)
}
