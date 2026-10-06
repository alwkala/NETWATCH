package netenv

import (
	"errors"
	"net/netip"
	"testing"
)

func TestIsAllowedUnicastTarget(t *testing.T) {
	tests := []struct {
		ip      string
		allowed bool
		reason  string
	}{
		// RFC 1918
		{"192.168.1.1", true, "RFC 1918 192.168/16"},
		{"10.0.0.254", true, "RFC 1918 10/8"},
		{"172.16.0.5", true, "RFC 1918 172.16/12"},
		{"172.31.255.254", true, "RFC 1918 172.16/12 upper boundary"},
		// Link-Local
		{"169.254.1.20", true, "Link-local 169.254/16"},
		// Invariant: Loopback must be rejected for discovery probes
		{"127.0.0.1", false, "IPv4 loopback rejected for probes"},
		{"::1", false, "IPv6 loopback rejected for probes"},
		// Public IPv4 (Zero-Egress strictly forbidden)
		{"8.8.8.8", false, "Public Google DNS"},
		{"1.1.1.1", false, "Public Cloudflare DNS"},
		{"142.250.190.46", false, "Public Google WAN IP"},
		{"208.67.222.222", false, "Public OpenDNS"},
		// Multicast addresses should NOT be treated as unicast targets
		{"224.0.0.251", false, "mDNS multicast is not unicast"},
		{"239.255.255.250", false, "SSDP multicast is not unicast"},
		// Invalid / unspecified
		{"0.0.0.0", false, "Zero address"},
		{"", false, "Empty string"},
	}

	for _, tc := range tests {
		addr, _ := netip.ParseAddr(tc.ip)
		got := IsAllowedUnicastTarget(addr)
		if got != tc.allowed {
			t.Errorf("IsAllowedUnicastTarget(%q) = %v, want %v (%s)", tc.ip, got, tc.allowed, tc.reason)
		}
	}
}

func TestIsAllowedDiscoveryDestination(t *testing.T) {
	mDNSAddr := netip.MustParseAddr("224.0.0.251")
	ssdpAddr := netip.MustParseAddr("239.255.255.250")
	privateIP := netip.MustParseAddr("192.168.1.50")
	publicIP := netip.MustParseAddr("8.8.8.8")
	loopbackIP := netip.MustParseAddr("127.0.0.1")

	// NBNS: Unicast private only
	if !IsAllowedDiscoveryDestination(privateIP, ProtocolNBNS) {
		t.Error("NBNS should allow private unicast target")
	}
	if IsAllowedDiscoveryDestination(publicIP, ProtocolNBNS) {
		t.Error("NBNS must reject public WAN IP")
	}
	if IsAllowedDiscoveryDestination(mDNSAddr, ProtocolNBNS) {
		t.Error("NBNS must reject mDNS multicast destination")
	}
	if IsAllowedDiscoveryDestination(loopbackIP, ProtocolNBNS) {
		t.Error("NBNS must reject loopback")
	}

	// mDNS: 224.0.0.251 only
	if !IsAllowedDiscoveryDestination(mDNSAddr, ProtocolMDNS) {
		t.Error("mDNS must allow 224.0.0.251")
	}
	if IsAllowedDiscoveryDestination(ssdpAddr, ProtocolMDNS) {
		t.Error("mDNS must reject SSDP multicast address")
	}
	if IsAllowedDiscoveryDestination(privateIP, ProtocolMDNS) {
		t.Error("mDNS must reject unicast IP as destination")
	}
	if IsAllowedDiscoveryDestination(publicIP, ProtocolMDNS) {
		t.Error("mDNS must reject public IP")
	}

	// SSDP: 239.255.255.250 only
	if !IsAllowedDiscoveryDestination(ssdpAddr, ProtocolSSDP) {
		t.Error("SSDP must allow 239.255.255.250")
	}
	if IsAllowedDiscoveryDestination(mDNSAddr, ProtocolSSDP) {
		t.Error("SSDP must reject mDNS multicast address")
	}
	if IsAllowedDiscoveryDestination(privateIP, ProtocolSSDP) {
		t.Error("SSDP must reject unicast IP as destination")
	}
	if IsAllowedDiscoveryDestination(publicIP, ProtocolSSDP) {
		t.Error("SSDP must reject public IP")
	}
}

func TestValidateDestination(t *testing.T) {
	valid := netip.MustParseAddr("192.168.1.1")
	invalid := netip.MustParseAddr("8.8.8.8")

	if err := ValidateDestination(valid, ProtocolNBNS); err != nil {
		t.Errorf("ValidateDestination(valid) = %v, want nil", err)
	}

	err := ValidateDestination(invalid, ProtocolNBNS)
	if !errors.Is(err, ErrDestinationNotAllowed) {
		t.Errorf("ValidateDestination(invalid) = %v, want ErrDestinationNotAllowed", err)
	}
}
