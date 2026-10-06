package netenv

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net"
	"net/netip"
	"strings"
	"time"

	"netwatch/internal/fingerprint"
	"netwatch/internal/model"
)

// SSDPPacket holds structured attributes extracted from an SSDP / HTTPU response packet.
type SSDPPacket struct {
	IP       netip.Addr
	ST       string
	USN      string
	Server   string
	Location string // STRICT INVARIANT: Record string only, NEVER fetch HTTP XML
}

// M-SEARCH query payload according to UPnP Device Architecture 1.0.
var ssdpMSearchPayload = []byte(
	"M-SEARCH * HTTP/1.1\r\n" +
		"HOST: 239.255.255.250:1900\r\n" +
		"MAN: \"ssdp:discover\"\r\n" +
		"MX: 1\r\n" +
		"ST: ssdp:all\r\n\r\n",
)

// ParseSSDPResponse parses an HTTPU response from an SSDP responder.
// Defensive against malformed headers, header bombs, and oversized lines.
func ParseSSDPResponse(buf []byte, sender netip.Addr) (*SSDPPacket, error) {
	if len(buf) == 0 {
		return nil, fmt.Errorf("ssdp: empty packet")
	}

	reader := bufio.NewReader(bytes.NewReader(buf))
	statusLine, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("ssdp: failed to read status line: %w", err)
	}

	statusLine = strings.TrimSpace(statusLine)
	// Must be an HTTP response (HTTP/1.1 200 OK or NOTIFY * HTTP/1.1)
	if !strings.HasPrefix(statusLine, "HTTP/1.1 200") && !strings.HasPrefix(statusLine, "NOTIFY") {
		return nil, fmt.Errorf("ssdp: invalid status line: %s", statusLine)
	}

	pkt := &SSDPPacket{
		IP: sender,
	}

	// Read headers defensively up to 64 lines max
	for lineCount := 0; lineCount < 64; lineCount++ {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break // End of HTTP headers
		}

		colonIdx := strings.IndexByte(line, ':')
		if colonIdx <= 0 {
			continue
		}

		key := strings.ToLower(strings.TrimSpace(line[:colonIdx]))
		val := strings.TrimSpace(line[colonIdx+1:])

		switch key {
		case "st":
			pkt.ST = fingerprint.SanitizeLANString(val)
		case "usn":
			pkt.USN = fingerprint.SanitizeLANString(val)
		case "server":
			pkt.Server = fingerprint.SanitizeLANString(val)
		case "location":
			// Record string only — never dialed or fetched
			pkt.Location = fingerprint.SanitizeLANString(val)
		}
	}

	return pkt, nil
}

// SSDPProbe implements model.DiscoveryProbe for SSDP/UPnP discovery.
type SSDPProbe struct {
	Timeout time.Duration
}

// NewSSDPProbe creates a new SSDPProbe with default timeout.
func NewSSDPProbe() *SSDPProbe {
	return &SSDPProbe{
		Timeout: SSDPTotalTimeout,
	}
}

func (p *SSDPProbe) Name() model.DiscoverySource {
	return model.SourceSSDP
}

// Discover sends a multicast M-SEARCH packet and collects unicast responses.
// Enforces destination policy and untrusted packet boundaries.
func (p *SSDPProbe) Discover(ctx context.Context, scope model.DiscoveryScope) ([]model.DiscoveryEvidence, error) {
	// Destination Policy Enforcement for SSDP multicast destination
	if err := ValidateDestination(ssdpMulticastAddr, ProtocolSSDP); err != nil {
		return nil, err
	}

	deadline := time.Now().Add(p.Timeout)
	probeCtx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	// Listen on an ephemeral UDP port for incoming unicast responses
	laddr := &net.UDPAddr{IP: net.IPv4zero, Port: 0}
	conn, err := net.ListenUDP("udp4", laddr)
	if err != nil {
		return nil, fmt.Errorf("ssdp: listen error: %w", err)
	}
	defer conn.Close()

	if d, ok := probeCtx.Deadline(); ok {
		_ = conn.SetDeadline(d)
	}

	dstAddr := &net.UDPAddr{
		IP:   net.ParseIP("239.255.255.250"),
		Port: 1900,
	}

	// Transmit M-SEARCH discovery query
	if _, err := conn.WriteTo(ssdpMSearchPayload, dstAddr); err != nil {
		return nil, fmt.Errorf("ssdp: write error: %w", err)
	}

	var evidenceList []model.DiscoveryEvidence
	seenResponses := make(map[string]bool)

	buf := make([]byte, 2048)
	now := time.Now().UTC()

	for {
		select {
		case <-probeCtx.Done():
			return evidenceList, nil
		default:
		}

		n, raddr, err := conn.ReadFrom(buf)
		if err != nil {
			break // Timeout or closed
		}

		udpAddr, ok := raddr.(*net.UDPAddr)
		if !ok || udpAddr == nil {
			continue
		}

		senderIP, ok := netip.AddrFromSlice(udpAddr.IP)
		if !ok {
			continue
		}
		senderIP = senderIP.Unmap()

		// Validate that the responding host is an allowed unicast address
		if !IsAllowedUnicastTarget(senderIP) {
			continue
		}

		pkt, err := ParseSSDPResponse(buf[:n], senderIP)
		if err != nil || pkt == nil {
			continue
		}

		dedupKey := fmt.Sprintf("%s|%s|%s", senderIP, pkt.ST, pkt.Server)
		if seenResponses[dedupKey] {
			continue
		}
		seenResponses[dedupKey] = true

		if pkt.ST != "" {
			evidenceList = append(evidenceList, model.DiscoveryEvidence{
				Source:     model.SourceSSDP,
				IP:         senderIP,
				Key:        "st",
				Value:      pkt.ST,
				ObservedAt: now,
				LastSeen:   now,
			})
		}
		if pkt.Server != "" {
			evidenceList = append(evidenceList, model.DiscoveryEvidence{
				Source:     model.SourceSSDP,
				IP:         senderIP,
				Key:        "server",
				Value:      pkt.Server,
				ObservedAt: now,
				LastSeen:   now,
			})
		}
		if pkt.USN != "" {
			evidenceList = append(evidenceList, model.DiscoveryEvidence{
				Source:     model.SourceSSDP,
				IP:         senderIP,
				Key:        "usn",
				Value:      pkt.USN,
				ObservedAt: now,
				LastSeen:   now,
			})
		}
		if pkt.Location != "" {
			evidenceList = append(evidenceList, model.DiscoveryEvidence{
				Source:     model.SourceSSDP,
				IP:         senderIP,
				Key:        "location",
				Value:      pkt.Location, // Recorded string only — NEVER FETCHED
				ObservedAt: now,
				LastSeen:   now,
			})
		}
	}

	return evidenceList, nil
}
