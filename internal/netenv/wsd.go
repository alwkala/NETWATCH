package netenv

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"net"
	"net/netip"
	"strings"
	"time"

	"netwatch/internal/fingerprint"
	"netwatch/internal/model"
)

// WSDProbePayload is the standard OASIS WS-Discovery 1.1 SOAP Probe request.
var wsdProbePayload = []byte(
	`<?xml version="1.0" encoding="utf-8"?>` +
		`<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope" ` +
		`xmlns:wsa="http://schemas.xmlsoap.org/ws/2004/08/addressing" ` +
		`xmlns:wsd="http://schemas.xmlsoap.org/ws/2005/04/discovery">` +
		`<soap:Header>` +
		`<wsa:To>urn:schemas-xmlsoap-org:ws:2005:04:discovery</wsa:To>` +
		`<wsa:Action>http://schemas.xmlsoap.org/ws/2005/04/discovery/Probe</wsa:Action>` +
		`<wsa:MessageID>urn:uuid:c18b76df-94bf-42f3-a261-netwatch-probe</wsa:MessageID>` +
		`</soap:Header>` +
		`<soap:Body>` +
		`<wsd:Probe><wsd:Types></wsd:Types></wsd:Probe>` +
		`</soap:Body>` +
		`</soap:Envelope>`,
)

// WSDMatch holds structured parameters extracted from a WS-Discovery ProbeMatches response.
type WSDMatch struct {
	IP        netip.Addr
	Types     []string
	Scopes    []string
	XAddrs    []string
	Endpoint  string
}

// SOAP ProbeMatches XML parsing types (defensive, minimal tree).
type soapEnvelope struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    soapBody `xml:"Body"`
}

type soapBody struct {
	ProbeMatches probeMatches `xml:"ProbeMatches"`
}

type probeMatches struct {
	ProbeMatch []probeMatch `xml:"ProbeMatch"`
}

type probeMatch struct {
	EndpointReference endpointReference `xml:"EndpointReference"`
	Types             string            `xml:"Types"`
	Scopes            string            `xml:"Scopes"`
	XAddrs            string            `xml:"XAddrs"`
}

type endpointReference struct {
	Address string `xml:"Address"`
}

// ParseWSDResponse parses a raw WS-Discovery SOAP response buffer.
// Defensive against oversized packets, XXE entities, and malformed XML.
func ParseWSDResponse(buf []byte, sender netip.Addr) ([]WSDMatch, error) {
	if len(buf) == 0 {
		return nil, fmt.Errorf("wsd: empty packet")
	}
	if len(buf) > 8192 {
		return nil, fmt.Errorf("wsd: packet exceeds maximum allowed size (8KB)")
	}

	decoder := xml.NewDecoder(bytes.NewReader(buf))
	decoder.Strict = false
	decoder.Entity = map[string]string{} // Disallow custom XML entities (XXE defense)

	var env soapEnvelope
	if err := decoder.Decode(&env); err != nil {
		return nil, fmt.Errorf("wsd: xml decode failed: %w", err)
	}

	var matches []WSDMatch
	for _, pm := range env.Body.ProbeMatches.ProbeMatch {
		m := WSDMatch{
			IP:       sender,
			Endpoint: fingerprint.SanitizeLANString(pm.EndpointReference.Address),
		}

		if pm.Types != "" {
			for _, t := range strings.Fields(pm.Types) {
				clean := fingerprint.SanitizeLANString(t)
				if clean != "" {
					m.Types = append(m.Types, clean)
				}
			}
		}

		if pm.Scopes != "" {
			for _, s := range strings.Fields(pm.Scopes) {
				clean := fingerprint.SanitizeLANString(s)
				if clean != "" {
					m.Scopes = append(m.Scopes, clean)
				}
			}
		}

		if pm.XAddrs != "" {
			for _, x := range strings.Fields(pm.XAddrs) {
				clean := fingerprint.SanitizeLANString(x)
				if clean != "" {
					m.XAddrs = append(m.XAddrs, clean)
				}
			}
		}

		matches = append(matches, m)
	}

	return matches, nil
}

// DiscoverWSD performs an unprivileged multicast WS-Discovery sweep over UDP 3702.
// STRICT INVARIANT: ZERO-FETCH on all XAddrs and Scopes (strings recorded as evidence only).
func DiscoverWSD(ctx context.Context) ([]model.DiscoveryEvidence, error) {
	dest := net.UDPAddrFromAddrPort(netip.AddrPortFrom(wsdMulticastAddr, 3702))
	if err := ValidateDestination(dest.AddrPort().Addr(), ProtocolWSD); err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		return nil, fmt.Errorf("wsd: listen udp: %w", err)
	}
	defer conn.Close()

	// Transmit Probe packet
	if _, err := conn.WriteTo(wsdProbePayload, dest); err != nil {
		return nil, fmt.Errorf("wsd: send probe: %w", err)
	}

	deadline := time.Now().Add(WSDTotalTimeout)
	_ = conn.SetDeadline(deadline)

	var (
		evidence []model.DiscoveryEvidence
		buf      = make([]byte, 8192)
		now      = time.Now().UTC()
	)

	for {
		if ctx.Err() != nil {
			break
		}

		n, remoteAddr, err := conn.ReadFrom(buf)
		if err != nil {
			break // timeout or closed
		}

		udpAddr, ok := remoteAddr.(*net.UDPAddr)
		if !ok {
			continue
		}

		senderIP, ok := netip.AddrFromSlice(udpAddr.IP)
		if !ok {
			continue
		}
		senderIP = senderIP.Unmap()

		matches, err := ParseWSDResponse(buf[:n], senderIP)
		if err != nil {
			continue
		}

		for _, m := range matches {
			if len(m.Types) > 0 {
				evidence = append(evidence, model.DiscoveryEvidence{
					Source:     model.SourceWSD,
					IP:         m.IP,
					Key:        "types",
					Value:      strings.Join(m.Types, " "),
					ObservedAt: now,
					LastSeen:   now,
				})
			}
			if len(m.Scopes) > 0 {
				evidence = append(evidence, model.DiscoveryEvidence{
					Source:     model.SourceWSD,
					IP:         m.IP,
					Key:        "scopes",
					Value:      strings.Join(m.Scopes, " "),
					ObservedAt: now,
					LastSeen:   now,
				})
			}
			if len(m.XAddrs) > 0 {
				evidence = append(evidence, model.DiscoveryEvidence{
					Source:     model.SourceWSD,
					IP:         m.IP,
					Key:        "xaddrs",
					Value:      strings.Join(m.XAddrs, " "),
					ObservedAt: now,
					LastSeen:   now,
				})
			}
			if m.Endpoint != "" {
				evidence = append(evidence, model.DiscoveryEvidence{
					Source:     model.SourceWSD,
					IP:         m.IP,
					Key:        "endpoint",
					Value:      m.Endpoint,
					ObservedAt: now,
					LastSeen:   now,
				})
			}
		}
	}

	return evidence, nil
}
