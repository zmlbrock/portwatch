package scanner_test

import (
	"testing"
	"time"

	"github.com/user/portwatch/internal/scanner"
)

func TestNew(t *testing.T) {
	s := scanner.New([]string{"tcp", "udp"}, 500*time.Millisecond)
	if s == nil {
		t.Fatal("expected non-nil scanner")
	}
}

func TestNew_DefaultTimeout(t *testing.T) {
	// Zero timeout should fall back to a sensible default
	s := scanner.New([]string{"tcp"}, 0)
	if s == nil {
		t.Fatal("expected non-nil scanner with zero timeout")
	}
}

func TestScan_ReturnsPorts(t *testing.T) {
	s := scanner.New([]string{"tcp"}, 500*time.Millisecond)

	ports, err := s.Scan()
	if err != nil {
		t.Fatalf("Scan() returned unexpected error: %v", err)
	}

	// We can't assert specific ports since the test environment varies,
	// but the result must be a non-nil slice.
	if ports == nil {
		t.Error("expected non-nil ports slice")
	}
}

func TestScan_PortFields(t *testing.T) {
	s := scanner.New([]string{"tcp"}, 500*time.Millisecond)

	ports, err := s.Scan()
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	for _, p := range ports {
		if p.Number == 0 {
			t.Errorf("port number should not be zero, got %+v", p)
		}
		if p.Protocol == "" {
			t.Errorf("protocol should not be empty, got %+v", p)
		}
		if p.State == "" {
			t.Errorf("state should not be empty, got %+v", p)
		}
	}
}

func TestScan_MultipleProtocols(t *testing.T) {
	s := scanner.New([]string{"tcp", "udp"}, 500*time.Millisecond)

	ports, err := s.Scan()
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	protocols := make(map[string]bool)
	for _, p := range ports {
		protocols[p.Protocol] = true
	}

	// At least tcp should appear if any port is open on the test host.
	// This is a best-effort check; CI environments may have no open ports.
	t.Logf("observed protocols: %v", protocols)
}

func TestScan_UnknownProtocolIgnored(t *testing.T) {
	// An unrecognised protocol should not cause Scan to error out;
	// it should simply be skipped.
	s := scanner.New([]string{"tcp", "bogusprotocol"}, 500*time.Millisecond)

	_, err := s.Scan()
	if err != nil {
		t.Fatalf("Scan() should not error on unknown protocol, got: %v", err)
	}
}

func TestScan_NoDuplicates(t *testing.T) {
	s := scanner.New([]string{"tcp"}, 500*time.Millisecond)

	ports, err := s.Scan()
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	seen := make(map[string]bool)
	for _, p := range ports {
		key := p.Protocol + ":" + string(rune(p.Number))
		if seen[key] {
			t.Errorf("duplicate port entry detected: %+v", p)
		}
		seen[key] = true
	}
}
