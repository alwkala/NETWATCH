package netenv

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"net/netip"
	"strings"
	"time"

	"netwatch/internal/fingerprint"
	"netwatch/internal/model"
)

// Standard DNS types and classes for mDNS / DNS-SD.
const (
	dnsTypeA    = 1
	dnsTypePTR  = 12
	dnsTypeTXT  = 16
	dnsTypeAAAA = 28
	dnsTypeSRV  = 33
	dnsClassIN  = 1
)

// MDNSRecord represents a parsed DNS Resource Record.
type MDNSRecord struct {
	Name     string
	Type     uint16
	Class    uint16
	TTL      uint32
	IP       netip.Addr // For TypeA
	PTRName  string     // For TypePTR
	Target   string     // For TypeSRV
	Port     uint16     // For TypeSRV
	TXTAttrs map[string]string
}

// MDNSPacket represents a parsed mDNS message.
type MDNSPacket struct {
	Answers    []MDNSRecord
	Additionals []MDNSRecord
}

// ParseMDNSPacket defensively decodes an mDNS wire-format packet.
// Protects against compression pointer loops, truncated packets, and oversized labels.
func ParseMDNSPacket(buf []byte) (*MDNSPacket, error) {
	if len(buf) < 12 {
		return nil, fmt.Errorf("mdns: packet too short (%d bytes)", len(buf))
	}
	if len(buf) > MaxMDNSPacketSize {
		return nil, fmt.Errorf("mdns: packet exceeds maximum allowed size (%d bytes)", len(buf))
	}

	qdcount := int(binary.BigEndian.Uint16(buf[4:6]))
	ancount := int(binary.BigEndian.Uint16(buf[6:8]))
	nscount := int(binary.BigEndian.Uint16(buf[8:10]))
	arcount := int(binary.BigEndian.Uint16(buf[10:12]))

	totalRecords := ancount + nscount + arcount
	if totalRecords > MaxMDNSRecords {
		return nil, fmt.Errorf("mdns: packet contains too many records (%d)", totalRecords)
	}

	offset := 12

	// Skip Question section
	for i := 0; i < qdcount; i++ {
		var err error
		_, offset, err = readDNSName(buf, offset, 0)
		if err != nil {
			return nil, err
		}
		if offset+4 > len(buf) {
			return nil, fmt.Errorf("mdns: truncated question")
		}
		offset += 4 // QTYPE + QCLASS
	}

	pkt := &MDNSPacket{}

	// Helper to parse a record slice
	parseRecords := func(count int) ([]MDNSRecord, error) {
		var list []MDNSRecord
		for i := 0; i < count; i++ {
			if offset >= len(buf) {
				break
			}
			name, newOffset, err := readDNSName(buf, offset, 0)
			if err != nil {
				return list, err
			}
			offset = newOffset

			if offset+10 > len(buf) {
				return list, fmt.Errorf("mdns: truncated RR header")
			}

			rrType := binary.BigEndian.Uint16(buf[offset : offset+2])
			rrClass := binary.BigEndian.Uint16(buf[offset+2:offset+4]) & 0x7fff // mask out flush bit
			ttl := binary.BigEndian.Uint32(buf[offset+4 : offset+8])
			rdlen := int(binary.BigEndian.Uint16(buf[offset+8 : offset+10]))
			offset += 10

			if offset+rdlen > len(buf) {
				return list, fmt.Errorf("mdns: truncated RDATA")
			}

			rdata := buf[offset : offset+rdlen]
			rec := MDNSRecord{
				Name:  name,
				Type:  rrType,
				Class: rrClass,
				TTL:   ttl,
			}

			switch rrType {
			case dnsTypeA:
				if rdlen == 4 {
					rec.IP = netip.AddrFrom4([4]byte{rdata[0], rdata[1], rdata[2], rdata[3]})
				}
			case dnsTypePTR:
				ptrName, _, err := readDNSName(buf, offset, 0)
				if err == nil {
					rec.PTRName = ptrName
				}
			case dnsTypeSRV:
				if rdlen >= 6 {
					rec.Port = binary.BigEndian.Uint16(rdata[4:6])
					srvTarget, _, err := readDNSName(buf, offset+6, 0)
					if err == nil {
						rec.Target = srvTarget
					}
				}
			case dnsTypeTXT:
				rec.TXTAttrs = parseTXTAttributes(rdata)
			}

			offset += rdlen
			list = append(list, rec)
		}
		return list, nil
	}

	var err error
	pkt.Answers, err = parseRecords(ancount)
	if err != nil {
		return nil, err
	}

	// Skip authority records
	for i := 0; i < nscount; i++ {
		if offset >= len(buf) {
			break
		}
		_, newOffset, err := readDNSName(buf, offset, 0)
		if err != nil {
			break
		}
		offset = newOffset
		if offset+10 > len(buf) {
			break
		}
		rdlen := int(binary.BigEndian.Uint16(buf[offset+8 : offset+10]))
		offset += 10 + rdlen
	}

	pkt.Additionals, _ = parseRecords(arcount)

	return pkt, nil
}

// readDNSName reads a dotted domain name from DNS wire format with loop detection.
func readDNSName(buf []byte, offset int, depth int) (string, int, error) {
	if depth > 16 {
		return "", offset, fmt.Errorf("mdns: compression pointer cycle detected")
	}

	var labels []string
	origOffset := offset
	jumped := false

	for {
		if offset >= len(buf) {
			return "", origOffset, fmt.Errorf("mdns: name offset out of bounds")
		}

		length := int(buf[offset])
		if length == 0 {
			if !jumped {
				offset++
			}
			break
		}

		// Compression pointer (0xC0..)
		if length&0xc0 == 0xc0 {
			if offset+1 >= len(buf) {
				return "", origOffset, fmt.Errorf("mdns: truncated compression pointer")
			}
			ptrOffset := int(binary.BigEndian.Uint16(buf[offset:offset+2]) & 0x3fff)
			if ptrOffset >= len(buf) {
				return "", origOffset, fmt.Errorf("mdns: pointer points past buffer")
			}
			if !jumped {
				origOffset = offset + 2
				jumped = true
			}
			targetName, _, err := readDNSName(buf, ptrOffset, depth+1)
			if err != nil {
				return "", origOffset, err
			}
			if targetName != "" {
				labels = append(labels, targetName)
			}
			break
		}

		// Plain label
		offset++
		if offset+length > len(buf) {
			return "", origOffset, fmt.Errorf("mdns: label length exceeds buffer")
		}
		labelStr := string(buf[offset : offset+length])
		labels = append(labels, labelStr)
		offset += length
	}

	resOffset := offset
	if jumped {
		resOffset = origOffset
	}

	return strings.Join(labels, "."), resOffset, nil
}

func parseTXTAttributes(rdata []byte) map[string]string {
	attrs := make(map[string]string)
	pos := 0
	for pos < len(rdata) {
		strLen := int(rdata[pos])
		pos++
		if pos+strLen > len(rdata) {
			break
		}
		item := string(rdata[pos : pos+strLen])
		pos += strLen
		if eqIdx := strings.IndexByte(item, '='); eqIdx > 0 {
			k := strings.ToLower(strings.TrimSpace(item[:eqIdx]))
			v := strings.TrimSpace(item[eqIdx+1:])
			attrs[k] = v
		}
	}
	return attrs
}

// buildMDNSQuery builds a binary DNS query for PTR question.
func buildMDNSQuery(serviceName string) []byte {
	buf := make([]byte, 12)
	binary.BigEndian.PutUint16(buf[0:2], 0x0000) // Transaction ID (0 for mDNS)
	binary.BigEndian.PutUint16(buf[2:4], 0x0000) // Standard query
	binary.BigEndian.PutUint16(buf[4:6], 0x0001) // 1 question

	// Encode name into wire format labels
	parts := strings.Split(serviceName, ".")
	for _, p := range parts {
		if len(p) == 0 {
			continue
		}
		buf = append(buf, byte(len(p)))
		buf = append(buf, []byte(p)...)
	}
	buf = append(buf, 0x00) // root label

	// QTYPE: PTR (0x000c), QCLASS: IN (0x0001)
	qtail := make([]byte, 4)
	binary.BigEndian.PutUint16(qtail[0:2], dnsTypePTR)
	binary.BigEndian.PutUint16(qtail[2:4], dnsClassIN)
	buf = append(buf, qtail...)

	return buf
}

// Targeted high-value services for Pass C enumeration.
var targetedMDNSServices = []string{
	"_googlecast._tcp.local",
	"_airplay._tcp.local",
	"_raop._tcp.local",
	"_ipp._tcp.local",
	"_printer._tcp.local",
	"_smb._tcp.local",
	"_spotify-connect._tcp.local",
	"_hap._tcp.local",
}

// MDNSProbe implements model.DiscoveryProbe for mDNS / DNS-SD.
type MDNSProbe struct {
	Timeout time.Duration
}

// NewMDNSProbe constructs an MDNSProbe with default timeout bounds.
func NewMDNSProbe() *MDNSProbe {
	return &MDNSProbe{
		Timeout: MDNSTotalTimeout,
	}
}

func (p *MDNSProbe) Name() model.DiscoverySource {
	return model.SourceMDNS
}

// Discover executes multi-pass mDNS observation and service enumeration:
// Pass A (Passive listen) + Pass B (Meta-query) + Pass C (Targeted services).
func (p *MDNSProbe) Discover(ctx context.Context, scope model.DiscoveryScope) ([]model.DiscoveryEvidence, error) {
	// Destination Policy Enforcement for mDNS destination
	if err := ValidateDestination(mdnsMulticastAddr, ProtocolMDNS); err != nil {
		return nil, err
	}

	deadline := time.Now().Add(p.Timeout)
	probeCtx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	laddr := &net.UDPAddr{IP: net.IPv4zero, Port: 0}
	conn, err := net.ListenUDP("udp4", laddr)
	if err != nil {
		return nil, fmt.Errorf("mdns: listen error: %w", err)
	}
	defer conn.Close()

	if d, ok := probeCtx.Deadline(); ok {
		_ = conn.SetDeadline(d)
	}

	dstAddr := &net.UDPAddr{
		IP:   net.ParseIP("224.0.0.251"),
		Port: 5353,
	}

	// Pass B: DNS-SD meta query (_services._dns-sd._udp.local)
	metaQuery := buildMDNSQuery("_services._dns-sd._udp.local")
	_, _ = conn.WriteTo(metaQuery, dstAddr)

	// Pass C: Targeted high-value service queries
	for _, svc := range targetedMDNSServices {
		query := buildMDNSQuery(svc)
		_, _ = conn.WriteTo(query, dstAddr)
	}

	var evidenceList []model.DiscoveryEvidence
	seenKeys := make(map[string]bool)

	// Hostname-to-IP correlation map built from A records
	hostToIP := make(map[string]netip.Addr)
	// Pending service records to correlate once A records are resolved
	type pendingSvc struct {
		target  string
		service string
		model   string
	}
	var pendings []pendingSvc

	buf := make([]byte, 4096)
	now := time.Now().UTC()

	for {
		select {
		case <-probeCtx.Done():
			goto Correlate
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

		// Sender must be valid unicast private target
		if !IsAllowedUnicastTarget(senderIP) {
			continue
		}

		pkt, err := ParseMDNSPacket(buf[:n])
		if err != nil || pkt == nil {
			continue
		}

		allRecords := append(pkt.Answers, pkt.Additionals...)
		for _, rec := range allRecords {
			if rec.Type == dnsTypeA && rec.IP.IsValid() && IsAllowedUnicastTarget(rec.IP) {
				cleanHost := fingerprint.SanitizeLANString(strings.TrimSuffix(rec.Name, ".local"))
				if cleanHost != "" {
					hostToIP[strings.ToLower(rec.Name)] = rec.IP
					hostToIP[strings.ToLower(cleanHost)] = rec.IP

					key := fmt.Sprintf("%s|hostname|%s", rec.IP, cleanHost)
					if !seenKeys[key] {
						seenKeys[key] = true
						evidenceList = append(evidenceList, model.DiscoveryEvidence{
							Source:     model.SourceMDNS,
							IP:         rec.IP,
							Key:        "hostname",
							Value:      cleanHost + ".local",
							ObservedAt: now,
							LastSeen:   now,
						})
					}
				}
			} else if rec.Type == dnsTypeSRV && rec.Target != "" {
				svcType := extractServiceType(rec.Name)
				pendings = append(pendings, pendingSvc{
					target:  strings.ToLower(rec.Target),
					service: svcType,
				})
			} else if rec.Type == dnsTypeTXT && len(rec.TXTAttrs) > 0 {
				modelName := rec.TXTAttrs["md"]
				if modelName == "" {
					modelName = rec.TXTAttrs["model"]
				}
				if modelName != "" {
					pendings = append(pendings, pendingSvc{
						target:  strings.ToLower(rec.Name),
						model:   fingerprint.SanitizeLANString(modelName),
					})
				}
			}
		}
	}

Correlate:
	// Correlate pending service types and models with resolved host IPs
	for _, p := range pendings {
		targetClean := strings.TrimSuffix(p.target, ".")
		targetCleanNoLocal := strings.TrimSuffix(targetClean, ".local")

		var resolvedIP netip.Addr
		if ip, ok := hostToIP[targetClean]; ok {
			resolvedIP = ip
		} else if ip, ok := hostToIP[targetCleanNoLocal]; ok {
			resolvedIP = ip
		}

		if resolvedIP.IsValid() {
			if p.service != "" {
				key := fmt.Sprintf("%s|service|%s", resolvedIP, p.service)
				if !seenKeys[key] {
					seenKeys[key] = true
					evidenceList = append(evidenceList, model.DiscoveryEvidence{
						Source:     model.SourceMDNS,
						IP:         resolvedIP,
						Key:        "service",
						Value:      p.service,
						ObservedAt: now,
						LastSeen:   now,
					})
				}
			}
			if p.model != "" {
				key := fmt.Sprintf("%s|model|%s", resolvedIP, p.model)
				if !seenKeys[key] {
					seenKeys[key] = true
					evidenceList = append(evidenceList, model.DiscoveryEvidence{
						Source:     model.SourceMDNS,
						IP:         resolvedIP,
						Key:        "model",
						Value:      p.model,
						ObservedAt: now,
						LastSeen:   now,
					})
				}
			}
		}
	}

	return evidenceList, nil
}

// extractServiceType parses e.g. "_airplay._tcp.local" or "Living Room._airplay._tcp.local" into "_airplay._tcp"
func extractServiceType(name string) string {
	parts := strings.Split(name, ".")
	for i := 0; i < len(parts)-1; i++ {
		if strings.HasPrefix(parts[i], "_") && strings.HasPrefix(parts[i+1], "_") {
			return parts[i] + "." + parts[i+1]
		}
	}
	return ""
}
