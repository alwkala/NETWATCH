package netenv

import (
	"errors"
	"net/netip"
	"time"
)

// ErrDestinationNotAllowed indicates that an egress packet destination violates the Zero-Egress policy.
var ErrDestinationNotAllowed = errors.New("netenv: destination address not permitted by zero-egress policy")

// DiscoveryProtocol identifies the network protocol for destination enforcement.
type DiscoveryProtocol string

const (
	ProtocolARP  DiscoveryProtocol = "ARP"
	ProtocolICMP DiscoveryProtocol = "ICMP"
	ProtocolTCP  DiscoveryProtocol = "TCP"
	ProtocolNBNS DiscoveryProtocol = "NBNS"
	ProtocolMDNS DiscoveryProtocol = "mDNS"
	ProtocolSSDP DiscoveryProtocol = "SSDP"
	ProtocolDNS  DiscoveryProtocol = "rDNS"
	ProtocolNDP  DiscoveryProtocol = "NDP"
	ProtocolWSD  DiscoveryProtocol = "WSD"
)

// Untrusted LAN Input Guardrails (Denial-of-Service and memory poisoning defense).
const (
	MaxMDNSRecords      = 512
	MaxMDNSPacketSize   = 4096
	MaxServiceTypes     = 64
	MaxServiceInstances = 256
	MDNSTotalTimeout    = 500 * time.Millisecond
	SSDPTotalTimeout    = 1000 * time.Millisecond
	NBNSTotalTimeout    = 1500 * time.Millisecond
	WSDTotalTimeout     = 1000 * time.Millisecond
)

var (
	mdnsMulticastAddr = netip.MustParseAddr("224.0.0.251")
	ssdpMulticastAddr = netip.MustParseAddr("239.255.255.250")
	wsdMulticastAddr  = netip.MustParseAddr("239.255.255.250")
)

// IsAllowedUnicastTarget enforces that unicast discovery packets are strictly scoped
// to RFC 1918 private subnets and link-local IPv4 addresses.
//
// INVARIANT: Loopback (127.0.0.1) is explicitly rejected for LAN discovery probes
// to prevent confusing internal service endpoints with network devices.
func IsAllowedUnicastTarget(addr netip.Addr) bool {
	if !addr.IsValid() {
		return false
	}
	if addr.IsLoopback() {
		return false
	}
	if addr.IsPrivate() {
		return true // RFC 1918: 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16
	}
	if addr.IsLinkLocalUnicast() {
		return true // 169.254.0.0/16 IPv4 or fe80::/10 IPv6
	}
	// For IPv6 addresses, allow local and unique-local unicast (fc00::/7)
	if addr.Is6() && !addr.IsMulticast() && !addr.IsUnspecified() {
		return true
	}
	return false
}

// IsAllowedDiscoveryDestination checks whether the destination address is permitted
// for the given protocol under the NETWATCH zero-egress policy.
func IsAllowedDiscoveryDestination(addr netip.Addr, proto DiscoveryProtocol) bool {
	if !addr.IsValid() {
		return false
	}
	switch proto {
	case ProtocolARP, ProtocolICMP, ProtocolTCP, ProtocolNBNS, ProtocolNDP:
		return IsAllowedUnicastTarget(addr)
	case ProtocolMDNS:
		return addr == mdnsMulticastAddr
	case ProtocolSSDP:
		return addr == ssdpMulticastAddr
	case ProtocolWSD:
		return addr == wsdMulticastAddr
	case ProtocolDNS:
		return IsAllowedUnicastTarget(addr)
	default:
		return false
	}
}

// ValidateDestination returns ErrDestinationNotAllowed if addr is not permitted for proto.
func ValidateDestination(addr netip.Addr, proto DiscoveryProtocol) error {
	if !IsAllowedDiscoveryDestination(addr, proto) {
		return ErrDestinationNotAllowed
	}
	return nil
}
