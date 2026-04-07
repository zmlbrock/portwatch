package state_test

import (
	"testing"
	"time"

	"github.com/user/portwatch/internal/scanner"
	"github.com/user/portwatch/internal/state"
)

func makePort(proto string, port int) scanner.Port {
	return scanner.Port{
		Protocol: proto,
		Port:     port,
		State:    "open",
		ScannedAt: time.Now(),
	}
}

func TestNew(t *testing.T) {
	s := state.New()
	if s == nil {
		t.Fatal("expected non-nil state")
	}
}

func TestSnapshot_Empty(t *testing.T) {
	s := state.New()
	snap := s.Snapshot()
	if len(snap) != 0 {
		t.Errorf("expected empty snapshot, got %d entries", len(snap))
	}
}

func TestUpdate_StoresSnapshot(t *testing.T) {
	s := state.New()
	ports := []scanner.Port{
		makePort("tcp", 80),
		makePort("tcp", 443),
	}

	s.Update(ports)

	snap := s.Snapshot()
	if len(snap) != 2 {
		t.Errorf("expected 2 ports in snapshot, got %d", len(snap))
	}
}

func TestDiff_NoChanges(t *testing.T) {
	s := state.New()
	ports := []scanner.Port{
		makePort("tcp", 80),
	}

	s.Update(ports)
	opened, closed := s.Diff(ports)

	if len(opened) != 0 {
		t.Errorf("expected no opened ports, got %d", len(opened))
	}
	if len(closed) != 0 {
		t.Errorf("expected no closed ports, got %d", len(closed))
	}
}

func TestDiff_DetectsOpenedPort(t *testing.T) {
	s := state.New()
	initial := []scanner.Port{
		makePort("tcp", 80),
	}
	s.Update(initial)

	next := []scanner.Port{
		makePort("tcp", 80),
		makePort("tcp", 8080),
	}

	opened, closed := s.Diff(next)

	if len(opened) != 1 {
		t.Fatalf("expected 1 opened port, got %d", len(opened))
	}
	if opened[0].Port != 8080 {
		t.Errorf("expected opened port 8080, got %d", opened[0].Port)
	}
	if len(closed) != 0 {
		t.Errorf("expected no closed ports, got %d", len(closed))
	}
}

func TestDiff_DetectsClosedPort(t *testing.T) {
	s := state.New()
	initial := []scanner.Port{
		makePort("tcp", 80),
		makePort("tcp", 443),
	}
	s.Update(initial)

	next := []scanner.Port{
		makePort("tcp", 80),
	}

	opened, closed := s.Diff(next)

	if len(closed) != 1 {
		t.Fatalf("expected 1 closed port, got %d", len(closed))
	}
	if closed[0].Port != 443 {
		t.Errorf("expected closed port 443, got %d", closed[0].Port)
	}
	if len(opened) != 0 {
		t.Errorf("expected no opened ports, got %d", len(opened))
	}
}

func TestDiff_MultipleProtocols(t *testing.T) {
	s := state.New()
	initial := []scanner.Port{
		makePort("tcp", 53),
	}
	s.Update(initial)

	// Same port number but different protocol should be treated as distinct
	next := []scanner.Port{
		makePort("tcp", 53),
		makePort("udp", 53),
	}

	opened, closed := s.Diff(next)

	if len(opened) != 1 {
		t.Fatalf("expected 1 opened port, got %d", len(opened))
	}
	if opened[0].Protocol != "udp" {
		t.Errorf("expected opened udp port, got %s", opened[0].Protocol)
	}
}
