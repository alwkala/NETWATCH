package netenv

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"net/netip"
	"strings"
	"sync"
	"time"

	"netwatch/internal/fingerprint"
	"netwatch/internal/model"
)

// NBNS Node Status wildcard query packet (RFC 1002 Section 4.2.17).
// Encodes wildcard name "*" (0x2A) followed by 15 spaces (0x20).
var nbnsNodeStatusQuery = []byte{
	0x13, 0x37, // Transaction ID
	0x00, 0x00, // Flags: Standard Query
	0x00, 0x01, // Questions: 1
	0x00, 0x00, // Answer RRs: 0
	0x00, 0x00, // Authority RRs: 0
	0x00, 0x00, // Additional RRs: 0
	// Question Name: 32 bytes half-ASCII encoded wildcard "*" + 15 spaces
	0x20, // Length: 32
	'C', 'K', // 0x2A ('*')
	'A', 'A', 'A', 'A', 'A', 'A', 'A', 'A', // 4 spaces
	'A', 'A', 'A', 'A', 'A', 'A', 'A', 'A', // 4 spaces
	'A', 'A', 'A', 'A', 'A', 'A', 'A', 'A', // 4 spaces
	'A', 'A', 'A', 'A', 'A', 'A', // 3 spaces (total 15 spaces)
	0x00,       // Terminating zero byte
	0x00, 0x21, // Type: NBSTAT (0x0021)
	0x00, 0x01, // Class: IN (0x0001)
}

// NBNSNodeInfo holds the structured fields extracted from an RFC 1002 Node Status response.
type NBNSNodeInfo struct {
	ComputerName string
	Workgroup    string
	UnitID       string // 6-byte MAC address formatted as lowercase aa:bb:cc:dd:ee:ff
}

// ParseNBNSNodeStatus parses an RFC 1002 NBSTAT response payload.
// Returns (info, error). Defensively handles truncated, corrupt, or malicious packets.
func ParseNBNSNodeStatus(buf []byte) (*NBNSNodeInfo, error) {
	if len(buf) < 12 {
		return nil, fmt.Errorf("nbns: packet too short (%d bytes)", len(buf))
	}

	// Verify it's a response (QR bit = 1 in byte 2)
	if buf[2]&0x80 == 0 {
		return nil, fmt.Errorf("nbns: not a response packet")
	}

	ancount := binary.BigEndian.Uint16(buf[6:8])
	if ancount < 1 {
		return nil, fmt.Errorf("nbns: no answer records")
	}

	// Skip Question section if present
	offset := 12
	qdcount := binary.BigEndian.Uint16(buf[4:6])
	for i := 0; i < int(qdcount); i++ {
		offset = skipDNSName(buf, offset)
		if offset+4 > len(buf) {
			return nil, fmt.Errorf("nbns: truncated question section")
		}
		offset += 4 // QTYPE (2) + QCLASS (2)
	}

	// Read Answer section
	if offset >= len(buf) {
		return nil, fmt.Errorf("nbns: truncated before answer")
	}
	offset = skipDNSName(buf, offset)
	if offset+10 > len(buf) {
		return nil, fmt.Errorf("nbns: truncated answer header")
	}

	rrType := binary.BigEndian.Uint16(buf[offset : offset+2])
	rdLength := binary.BigEndian.Uint16(buf[offset+8 : offset+10])
	offset += 10

	if rrType != 0x0021 { // NBSTAT
		return nil, fmt.Errorf("nbns: expected NBSTAT answer type 0x0021, got 0x%04x", rrType)
	}

	if offset+int(rdLength) > len(buf) || int(rdLength) < 1 {
		return nil, fmt.Errorf("nbns: truncated RDATA")
	}

	rdata := buf[offset : offset+int(rdLength)]
	numNames := int(rdata[0])
	pos := 1

	info := &NBNSNodeInfo{}

	for i := 0; i < numNames; i++ {
		if pos+18 > len(rdata) {
			break // Guard against malformed count
		}
		rawName := rdata[pos : pos+15]
		suffix := rdata[pos+15]
		flags := binary.BigEndian.Uint16(rdata[pos+16 : pos+18])
		pos += 18

		isGroup := (flags & 0x8000) != 0
		cleanedName := strings.TrimSpace(string(rawName))
		if cleanedName == "" {
			continue
		}

		if isGroup {
			if suffix == 0x00 && info.Workgroup == "" {
				info.Workgroup = fingerprint.SanitizeLANString(cleanedName)
			}
		} else {
			// Unique names: 0x00 = Workstation/Host, 0x20 = Server
			if (suffix == 0x00 || suffix == 0x20) && info.ComputerName == "" {
				info.ComputerName = fingerprint.SanitizeLANString(cleanedName)
			}
		}
	}

	// Unit ID (Hardware MAC address) is the first 6 bytes of the node statistics table
	// directly following the names table (RFC 1002 section 4.2.18).
	if pos+6 <= len(rdata) {
		unitBytes := rdata[pos : pos+6]
		info.UnitID = fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
			unitBytes[0], unitBytes[1], unitBytes[2], unitBytes[3], unitBytes[4], unitBytes[5])
	}

	return info, nil
}

// skipDNSName skips a wire-format DNS/NetBIOS name (supporting simple label sequences and 0xC0 pointers).
func skipDNSName(buf []byte, offset int) int {
	for offset < len(buf) {
		b := buf[offset]
		if b == 0 {
			return offset + 1
		}
		if b&0xc0 == 0xc0 { // Pointer
			return offset + 2
		}
		offset += int(b) + 1
	}
	return len(buf)
}

// NBNSProbe implements engine.DiscoveryProbe for RFC 1002 NetBIOS Name Service.
type NBNSProbe struct {
	Timeout time.Duration
}

// NewNBNSProbe constructs an NBNSProbe with default timeout.
func NewNBNSProbe() *NBNSProbe {
	return &NBNSProbe{
		Timeout: 250 * time.Millisecond,
	}
}

func (p *NBNSProbe) Name() model.DiscoverySource {
	return model.SourceNBNS
}

// Discover probes the provided active target hosts over directed UDP 137.
// Enforces IsAllowedDiscoveryDestination on every destination before sending.
func (p *NBNSProbe) Discover(ctx context.Context, scope model.DiscoveryScope) ([]model.DiscoveryEvidence, error) {
	if len(scope.Hosts) == 0 {
		return nil, nil
	}

	deadline := time.Now().Add(NBNSTotalTimeout)
	probeCtx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	var evidenceList []model.DiscoveryEvidence
	var mu sync.Mutex

	// Concurrency worker pool (max 32 concurrent hosts)
	sem := make(chan struct{}, 32)
	var wg sync.WaitGroup

	now := time.Now().UTC()

	for _, host := range scope.Hosts {
		// Strict Destination Policy Enforcement at connection boundary
		if err := ValidateDestination(host, ProtocolNBNS); err != nil {
			continue // Drop public, loopback, or invalid targets
		}

		if probeCtx.Err() != nil {
			break
		}

		wg.Add(1)
		go func(target netip.Addr) {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
			case <-probeCtx.Done():
				return
			}
			defer func() { <-sem }()

			info, err := p.probeHost(probeCtx, target)
			if err != nil || info == nil {
				return
			}

			mu.Lock()
			defer mu.Unlock()

			if info.ComputerName != "" {
				evidenceList = append(evidenceList, model.DiscoveryEvidence{
					Source:     model.SourceNBNS,
					IP:         target,
					Key:        "hostname",
					Value:      info.ComputerName,
					ObservedAt: now,
					LastSeen:   now,
				})
			}
			if info.Workgroup != "" {
				evidenceList = append(evidenceList, model.DiscoveryEvidence{
					Source:     model.SourceNBNS,
					IP:         target,
					Key:        "workgroup",
					Value:      info.Workgroup,
					ObservedAt: now,
					LastSeen:   now,
				})
			}
			if info.UnitID != "" && info.UnitID != "00:00:00:00:00:00" {
				evidenceList = append(evidenceList, model.DiscoveryEvidence{
					Source:     model.SourceNBNS,
					IP:         target,
					Key:        "unit_id",
					Value:      info.UnitID,
					ObservedAt: now,
					LastSeen:   now,
				})
			}
		}(host)
	}

	wg.Wait()
	return evidenceList, nil
}

func (p *NBNSProbe) probeHost(ctx context.Context, target netip.Addr) (*NBNSNodeInfo, error) {
	dest := net.JoinHostPort(target.String(), "137")
	conn, err := net.DialTimeout("udp4", dest, p.Timeout)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	} else {
		_ = conn.SetDeadline(time.Now().Add(p.Timeout))
	}

	if _, err := conn.Write(nbnsNodeStatusQuery); err != nil {
		return nil, err
	}

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}

	return ParseNBNSNodeStatus(buf[:n])
}
