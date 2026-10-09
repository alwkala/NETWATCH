package netenv

import (
	"net/netip"
	"testing"
)

// FuzzParseMDNS validates that the mDNS packet parser handles arbitrary untrusted
// byte streams defensively without panicking, infinite recursion, or memory exhaustion.
func FuzzParseMDNS(f *testing.F) {
	// Seed corpus 1: Minimal valid DNS query header
	f.Add([]byte{
		0x00, 0x00, // ID
		0x00, 0x00, // Flags
		0x00, 0x01, // QDCOUNT: 1
		0x00, 0x00, // ANCOUNT: 0
		0x00, 0x00, // NSCOUNT: 0
		0x00, 0x00, // ARCOUNT: 0
		0x05, 'l', 'o', 'c', 'a', 'l', 0x00, // QNAME: local
		0x00, 0x0c, // QTYPE: PTR
		0x00, 0x01, // QCLASS: IN
	})

	// Seed corpus 2: Compressed pointer cycle attempt
	f.Add([]byte{
		0x00, 0x00, 0x84, 0x00,
		0x00, 0x00, 0x00, 0x01,
		0x00, 0x00, 0x00, 0x00,
		0xc0, 0x0c, // Pointer pointing to itself
		0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x3c, 0x00, 0x04,
		192, 168, 1, 10,
	})

	// Seed corpus 3: Truncated header
	f.Add([]byte{0x00, 0x01, 0x84})

	f.Fuzz(func(t *testing.T, data []byte) {
		// Target parser: must not panic or hang
		_, _ = ParseMDNSPacket(data)
	})
}

// FuzzParseSSDP validates that the SSDP HTTPU response parser resists
// malformed headers, header bombs, oversized strings, and control characters without panicking.
func FuzzParseSSDP(f *testing.F) {
	sender := netip.MustParseAddr("192.168.1.100")

	// Seed corpus 1: Valid SSDP HTTP response
	f.Add([]byte(
		"HTTP/1.1 200 OK\r\n" +
			"CACHE-CONTROL: max-age=1800\r\n" +
			"EXT:\r\n" +
			"LOCATION: http://192.168.1.100:8088/description.xml\r\n" +
			"SERVER: Linux/3.14.0 UPnP/1.0 SmartTV/1.0\r\n" +
			"ST: urn:schemas-upnp-org:device:MediaRenderer:1\r\n" +
			"USN: uuid:12345678-1234-1234-1234-123456789abc::urn:schemas-upnp-org:device:MediaRenderer:1\r\n\r\n",
	))

	// Seed corpus 2: NOTIFY alive
	f.Add([]byte(
		"NOTIFY * HTTP/1.1\r\n" +
			"HOST: 239.255.255.250:1900\r\n" +
			"NT: upnp:rootdevice\r\n" +
			"NTS: ssdp:alive\r\n\r\n",
	))

	// Seed corpus 3: Malformed status line
	f.Add([]byte("GARBAGE DATA WITHOUT HTTP STATUS\r\n\r\n"))

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = ParseSSDPResponse(data, sender)
	})
}

// FuzzParseNBNS validates that the NetBIOS Name Service (RFC 1002) Node Status parser
// handles corrupt or malicious node records, invalid name tables, and edge cases without panicking.
func FuzzParseNBNS(f *testing.F) {
	// Seed corpus 1: Valid minimal Node Status response header
	f.Add([]byte{
		0x13, 0x37, // TID
		0x84, 0x00, // Flags (Response)
		0x00, 0x00, // Questions: 0
		0x00, 0x01, // Answers: 1
		0x00, 0x00, // Authority: 0
		0x00, 0x00, // Additional: 0
		0x20, 'C', 'K', 'A', 'A', 'A', 'A', 'A', 'A', 'A', 'A', 'A', 'A', 'A', 'A', 'A', 'A',
		'A', 'A', 'A', 'A', 'A', 'A', 'A', 'A', 'A', 'A', 'A', 'A', 'A', 'A', 'A', 'A', 0x00,
		0x00, 0x21, // Type NBSTAT
		0x00, 0x01, // Class IN
		0x00, 0x00, 0x00, 0x00, // TTL
		0x00, 0x19, // RDLENGTH: 25 bytes
		0x01, // Num names: 1
		// 18 bytes name entry (15 chars + type + 2 flags)
		'M', 'Y', 'P', 'C', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
		0x00, 0x04, 0x00,
		// 6-byte Unit ID (MAC)
		0x00, 0x11, 0x22, 0x33, 0x44, 0x55,
	})

	// Seed corpus 2: Truncated buffer
	f.Add([]byte{0x13, 0x37, 0x84, 0x00})

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = ParseNBNSNodeStatus(data)
	})
}
