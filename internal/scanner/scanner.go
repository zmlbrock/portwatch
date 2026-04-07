// Package scanner provides functionality for scanning and detecting open ports
// on the local system using /proc/net or system calls.
package scanner

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

// Protocol represents a network protocol type.
type Protocol string

const (
	TCP  Protocol = "tcp"
	TCP6 Protocol = "tcp6"
	UDP  Protocol = "udp"
	UDP6 Protocol = "udp6"
)

// Port represents a single open port with its associated metadata.
type Port struct {
	Number    int
	Protocol  Protocol
	Address   string
	PID       int
	Process   string
	DiscoveredAt time.Time
}

// String returns a human-readable representation of the port.
func (p Port) String() string {
	if p.Process != "" {
		return fmt.Sprintf("%s/%d (%s) [pid:%d %s]", p.Protocol, p.Number, p.Address, p.PID, p.Process)
	}
	return fmt.Sprintf("%s/%d (%s)", p.Protocol, p.Number, p.Address)
}

// Key returns a unique identifier for this port entry.
func (p Port) Key() string {
	return fmt.Sprintf("%s:%d:%s", p.Protocol, p.Number, p.Address)
}

// Scanner scans the local system for open ports.
type Scanner struct {
	protocols []Protocol
}

// New creates a new Scanner that monitors the given protocols.
// If no protocols are specified, all supported protocols are monitored.
func New(protocols ...Protocol) *Scanner {
	if len(protocols) == 0 {
		protocols = []Protocol{TCP, TCP6, UDP, UDP6}
	}
	return &Scanner{protocols: protocols}
}

// Scan performs a port scan and returns all currently open ports.
func (s *Scanner) Scan() ([]Port, error) {
	var ports []Port

	for _, proto := range s.protocols {
		results, err := scanProtocol(proto)
		if err != nil {
			// Non-fatal: log and continue with other protocols
			continue
		}
		ports = append(ports, results...)
	}

	return ports, nil
}

// scanProtocol scans for open ports of a specific protocol using net.Listen probing.
// On Linux this uses /proc/net; here we use a portable approach via net package.
func scanProtocol(proto Protocol) ([]Port, error) {
	var ports []Port
	netProto := strings.TrimSuffix(string(proto), "6")

	// Scan well-known and registered port ranges (1–49151)
	// This is a lightweight sweep; privileged ports may require elevated access.
	for port := 1; port <= 49151; port++ {
		address := fmt.Sprintf(":%d", port)
		if isPortOpen(netProto, address) {
			ports = append(ports, Port{
				Number:       port,
				Protocol:     proto,
				Address:      resolveAddress(netProto, port),
				DiscoveredAt: time.Now(),
			})
		}
	}
	return ports, nil
}

// isPortOpen checks whether a given port is currently bound/listening.
func isPortOpen(proto, address string) bool {
	switch proto {
	case "tcp":
		ln, err := net.Listen("tcp", address)
		if err != nil {
			// Connection refused or port in use — port is open
			return isInUseError(err)
		}
		ln.Close()
		return false
	case "udp":
		ln, err := net.ListenPacket("udp", address)
		if err != nil {
			return isInUseError(err)
		}
		ln.Close()
		return false
	}
	return false
}

// isInUseError returns true if the error indicates the port is already in use.
func isInUseError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "address already in use")
}

// resolveAddress returns the string representation of the listening address.
func resolveAddress(proto string, port int) string {
	return "0.0.0.0:" + strconv.Itoa(port)
}
