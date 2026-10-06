package netenv

import (
	"encoding/binary"
	"net/netip"
	"strings"
	"testing"
)

// buildSyntheticMDNSPacket constructs a minimal mDNS packet with an A record and an SRV record.
func buildSyntheticMDNSPacket(hostname string, ip [4]byte, serviceInstance string, srvTarget string) []byte {
	buf := make([]byte, 12)
	binary.BigEndian.PutUint16(buf[0:2], 0x0000) // ID
	binary.BigEndian.PutUint16(buf[2:4], 0x8400) // Response
	binary.BigEndian.PutUint16(buf[4:6], 0x0000) // Questions
	binary.BigEndian.PutUint16(buf[6:8], 0x0002) // Answers (2)
	binary.BigEndian.PutUint16(buf[8:10], 0x0000)
	binary.BigEndian.PutUint16(buf[10:12], 0x0000)

	// Record 1: A record
	// Name: hostname + ".local"
	for _, part := range []string{hostname, "local"} {
		buf = append(buf, byte(len(part)))
		buf = append(buf, []byte(part)...)
	}
	buf = append(buf, 0x00) // root label

	aHdr := make([]byte, 10)
	binary.BigEndian.PutUint16(aHdr[0:2], dnsTypeA)
	binary.BigEndian.PutUint16(aHdr[2:4], dnsClassIN)
	binary.BigEndian.PutUint32(aHdr[4:8], 120) // TTL
	binary.BigEndian.PutUint16(aHdr[8:10], 4)   // RDLength
	buf = append(buf, aHdr...)
	buf = append(buf, ip[:]...)

	// Record 2: SRV record
	// Name: serviceInstance
	for _, part := range strings.Split(serviceInstance, ".") {
		if part == "" {
			continue
		}
		buf = append(buf, byte(len(part)))
		buf = append(buf, []byte(part)...)
	}
	buf = append(buf, 0x00)

	srvHdr := make([]byte, 10)
	binary.BigEndian.PutUint16(srvHdr[0:2], dnsTypeSRV)
	binary.BigEndian.PutUint16(srvHdr[2:4], dnsClassIN)
	binary.BigEndian.PutUint32(srvHdr[4:8], 120)

	var srvRdata []byte
	srvRdata = append(srvRdata, 0x00, 0x00) // Priority
	srvRdata = append(srvRdata, 0x00, 0x00) // Weight
	portBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(portBytes, 7000) // Port
	srvRdata = append(srvRdata, portBytes...)
	// Target
	for _, part := range []string{srvTarget, "local"} {
		srvRdata = append(srvRdata, byte(len(part)))
		srvRdata = append(srvRdata, []byte(part)...)
	}
	srvRdata = append(srvRdata, 0x00)

	binary.BigEndian.PutUint16(srvHdr[8:10], uint16(len(srvRdata)))
	buf = append(buf, srvHdr...)
	buf = append(buf, srvRdata...)

	return buf
}

func TestParseMDNSPacket_Valid(t *testing.T) {
	ipBytes := [4]byte{192, 168, 1, 45}
	pktBytes := buildSyntheticMDNSPacket("living-room", ipBytes, "tv._airplay._tcp.local", "living-room")

	pkt, err := ParseMDNSPacket(pktBytes)
	if err != nil {
		t.Fatalf("ParseMDNSPacket failed: %v", err)
	}

	if len(pkt.Answers) != 2 {
		t.Fatalf("expected 2 answers, got %d", len(pkt.Answers))
	}

	recA := pkt.Answers[0]
	if recA.Type != dnsTypeA {
		t.Errorf("expected TypeA, got %d", recA.Type)
	}
	expectedIP := netip.MustParseAddr("192.168.1.45")
	if recA.IP != expectedIP {
		t.Errorf("got IP %v, want %v", recA.IP, expectedIP)
	}

	recSRV := pkt.Answers[1]
	if recSRV.Type != dnsTypeSRV {
		t.Errorf("expected TypeSRV, got %d", recSRV.Type)
	}
	if recSRV.Port != 7000 {
		t.Errorf("expected port 7000, got %d", recSRV.Port)
	}
}

func TestParseMDNSPacket_CompressionPointerLoop(t *testing.T) {
	// Construct a malicious DNS packet with a compression pointer cycle:
	// Pointer at offset 12 points to offset 14; pointer at offset 14 points back to 12.
	buf := make([]byte, 16)
	binary.BigEndian.PutUint16(buf[0:2], 0x0000)
	binary.BigEndian.PutUint16(buf[2:4], 0x8400)
	binary.BigEndian.PutUint16(buf[4:6], 0x0001) // 1 question
	binary.BigEndian.PutUint16(buf[6:8], 0x0000)
	binary.BigEndian.PutUint16(buf[8:10], 0x0000)
	binary.BigEndian.PutUint16(buf[10:12], 0x0000)

	// Offset 12: Pointer to 14
	buf[12] = 0xc0
	buf[13] = 0x0e
	// Offset 14: Pointer back to 12
	buf[14] = 0xc0
	buf[15] = 0x0c

	// Parser MUST return error and NOT hang in an infinite loop
	_, err := ParseMDNSPacket(buf)
	if err == nil {
		t.Error("expected error on cyclic compression pointer, got nil")
	}
}

func TestExtractServiceType(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Living Room._airplay._tcp.local", "_airplay._tcp"},
		{"_googlecast._tcp.local", "_googlecast._tcp"},
		{"DeskJet._ipp._tcp.local", "_ipp._tcp"},
		{"random.host.local", ""},
	}

	for _, tc := range tests {
		got := extractServiceType(tc.input)
		if got != tc.expected {
			t.Errorf("extractServiceType(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}
