package netenv

import (
	"context"
	"encoding/binary"
	"net/netip"
	"testing"

	"netwatch/internal/model"
)

// buildSyntheticNBNSResponse builds a binary RFC 1002 NBSTAT response packet.
func buildSyntheticNBNSResponse(computerName, workgroup string, mac [6]byte) []byte {
	buf := make([]byte, 12)
	binary.BigEndian.PutUint16(buf[0:2], 0x1337) // Transaction ID
	binary.BigEndian.PutUint16(buf[2:4], 0x8400) // Flags: Response, Authoritative
	binary.BigEndian.PutUint16(buf[4:6], 0x0000) // Questions: 0
	binary.BigEndian.PutUint16(buf[6:8], 0x0001) // Answer RRs: 1
	binary.BigEndian.PutUint16(buf[8:10], 0x0000)
	binary.BigEndian.PutUint16(buf[10:12], 0x0000)

	// Answer Name: 0x00 (root / current)
	buf = append(buf, 0x00)
	// Type: NBSTAT (0x0021), Class: IN (0x0001), TTL: 0
	hdr := make([]byte, 10)
	binary.BigEndian.PutUint16(hdr[0:2], 0x0021)
	binary.BigEndian.PutUint16(hdr[2:4], 0x0001)
	binary.BigEndian.PutUint32(hdr[4:8], 0) // TTL

	// RDATA
	var rdata []byte
	numNames := byte(0)
	if computerName != "" {
		numNames++
	}
	if workgroup != "" {
		numNames++
	}
	rdata = append(rdata, numNames)

	if computerName != "" {
		entry := make([]byte, 18)
		copy(entry[:15], []byte(computerName))
		for i := len(computerName); i < 15; i++ {
			entry[i] = 0x20
		}
		entry[15] = 0x00 // Suffix: Workstation
		binary.BigEndian.PutUint16(entry[16:18], 0x0000) // Unique
		rdata = append(rdata, entry...)
	}

	if workgroup != "" {
		entry := make([]byte, 18)
		copy(entry[:15], []byte(workgroup))
		for i := len(workgroup); i < 15; i++ {
			entry[i] = 0x20
		}
		entry[15] = 0x00 // Suffix: Workgroup
		binary.BigEndian.PutUint16(entry[16:18], 0x8000) // Group flag
		rdata = append(rdata, entry...)
	}

	// Statistics: Unit ID (MAC)
	rdata = append(rdata, mac[:]...)
	// Pad extra statistics bytes (typically 46 bytes total statistics)
	rdata = append(rdata, make([]byte, 40)...)

	binary.BigEndian.PutUint16(hdr[8:10], uint16(len(rdata))) // RDLength
	buf = append(buf, hdr...)
	buf = append(buf, rdata...)
	return buf
}

func TestParseNBNSNodeStatus_Valid(t *testing.T) {
	mac := [6]byte{0x00, 0x1a, 0x2b, 0x3c, 0x4d, 0x5e}
	pkt := buildSyntheticNBNSResponse("MY-DESKTOP", "WORKGROUP", mac)

	info, err := ParseNBNSNodeStatus(pkt)
	if err != nil {
		t.Fatalf("ParseNBNSNodeStatus failed: %v", err)
	}

	if info.ComputerName != "MY-DESKTOP" {
		t.Errorf("got ComputerName %q, want %q", info.ComputerName, "MY-DESKTOP")
	}
	if info.Workgroup != "WORKGROUP" {
		t.Errorf("got Workgroup %q, want %q", info.Workgroup, "WORKGROUP")
	}
	if info.UnitID != "00:1a:2b:3c:4d:5e" {
		t.Errorf("got UnitID %q, want %q", info.UnitID, "00:1a:2b:3c:4d:5e")
	}
}

func TestParseNBNSNodeStatus_Malformed(t *testing.T) {
	// 1. Packet too short
	if _, err := ParseNBNSNodeStatus([]byte{0x13, 0x37}); err == nil {
		t.Error("expected error for packet < 12 bytes")
	}

	// 2. Query packet instead of response (QR bit = 0)
	query := make([]byte, 20)
	query[2] = 0x00 // QR = 0
	if _, err := ParseNBNSNodeStatus(query); err == nil {
		t.Error("expected error for query packet (QR=0)")
	}

	// 3. Zero answer RRs
	noAns := make([]byte, 12)
	noAns[2] = 0x80
	if _, err := ParseNBNSNodeStatus(noAns); err == nil {
		t.Error("expected error for ancount == 0")
	}

	// 4. Truncated RDATA
	mac := [6]byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55}
	full := buildSyntheticNBNSResponse("PC", "WG", mac)
	truncated := full[:len(full)-10]
	// Should either return error or gracefully handle without panic
	_, _ = ParseNBNSNodeStatus(truncated)
}

func TestNBNSProbe_DestinationPolicyDrop(t *testing.T) {
	probe := NewNBNSProbe()
	// Pass public WAN IP: Destination policy must silently drop it without dialing
	scope := model.DiscoveryScope{
		Hosts: []netip.Addr{
			netip.MustParseAddr("8.8.8.8"),
			netip.MustParseAddr("127.0.0.1"),
		},
	}

	ctx := context.Background()
	evidence, err := probe.Discover(ctx, scope)
	if err != nil {
		t.Fatalf("Discover unexpected error: %v", err)
	}
	if len(evidence) != 0 {
		t.Errorf("expected 0 evidence for public/loopback hosts, got %d", len(evidence))
	}
}
