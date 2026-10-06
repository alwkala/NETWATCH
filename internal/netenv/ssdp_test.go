package netenv

import (
	"net/netip"
	"strings"
	"testing"
)

func TestParseSSDPResponse_Valid(t *testing.T) {
	raw := "HTTP/1.1 200 OK\r\n" +
		"CACHE-CONTROL: max-age=1800\r\n" +
		"LOCATION: http://192.168.1.1:49152/rootdesc.xml\r\n" +
		"SERVER: Linux/3.14.0 UPnP/1.0 MediaRenderer/1.0\r\n" +
		"ST: urn:schemas-upnp-org:device:MediaRenderer:1\r\n" +
		"USN: uuid:12345678-1234-1234-1234-123456789abc::urn:schemas-upnp-org:device:MediaRenderer:1\r\n\r\n"

	sender := netip.MustParseAddr("192.168.1.50")
	pkt, err := ParseSSDPResponse([]byte(raw), sender)
	if err != nil {
		t.Fatalf("ParseSSDPResponse failed: %v", err)
	}

	if pkt.IP != sender {
		t.Errorf("got IP %v, want %v", pkt.IP, sender)
	}
	if pkt.ST != "urn:schemas-upnp-org:device:MediaRenderer:1" {
		t.Errorf("unexpected ST: %q", pkt.ST)
	}
	if pkt.Server != "Linux/3.14.0 UPnP/1.0 MediaRenderer/1.0" {
		t.Errorf("unexpected Server: %q", pkt.Server)
	}
	if pkt.Location != "http://192.168.1.1:49152/rootdesc.xml" {
		t.Errorf("unexpected Location: %q", pkt.Location)
	}
	if !strings.HasPrefix(pkt.USN, "uuid:12345678") {
		t.Errorf("unexpected USN: %q", pkt.USN)
	}
}

func TestParseSSDPResponse_Malformed(t *testing.T) {
	sender := netip.MustParseAddr("192.168.1.50")

	// 1. Empty buffer
	if _, err := ParseSSDPResponse([]byte(""), sender); err == nil {
		t.Error("expected error for empty packet")
	}

	// 2. Non-HTTP status line
	if _, err := ParseSSDPResponse([]byte("GARBAGE DATA\r\n\r\n"), sender); err == nil {
		t.Error("expected error for garbage status line")
	}

	// 3. Truncated without headers
	pkt, err := ParseSSDPResponse([]byte("HTTP/1.1 200 OK\r\n"), sender)
	if err != nil {
		t.Fatalf("expected graceful parse with empty fields, got %v", err)
	}
	if pkt.ST != "" || pkt.Server != "" {
		t.Errorf("expected empty headers, got ST=%q Server=%q", pkt.ST, pkt.Server)
	}
}

func TestSSDP_ZeroFetchInvariant(t *testing.T) {
	raw := "HTTP/1.1 200 OK\r\n" +
		"LOCATION: http://192.168.1.1:80/dangerous/path\r\n" +
		"ST: upnp:rootdevice\r\n\r\n"

	sender := netip.MustParseAddr("192.168.1.1")
	pkt, err := ParseSSDPResponse([]byte(raw), sender)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Invariant: The location is recorded solely as text
	if pkt.Location != "http://192.168.1.1:80/dangerous/path" {
		t.Errorf("Location not preserved as string: %q", pkt.Location)
	}
}
