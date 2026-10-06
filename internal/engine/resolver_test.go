package engine

import (
	"net/netip"
	"testing"
	"time"

	"netwatch/internal/model"
)

func TestResolveEvidence_MultiProtocolCorrelation(t *testing.T) {
	ip := netip.MustParseAddr("192.168.1.50")
	mac := "00:11:22:33:44:55"
	now := time.Now().UTC()

	obs := []observation{
		{IP: ip, MAC: mac},
	}

	raw := []model.DiscoveryEvidence{
		// NBNS
		{Source: model.SourceNBNS, IP: ip, Key: "hostname", Value: "LIVING-ROOM-PC", ObservedAt: now},
		{Source: model.SourceNBNS, IP: ip, Key: "unit_id", Value: "00:11:22:33:44:55", ObservedAt: now},
		// mDNS
		{Source: model.SourceMDNS, IP: ip, Key: "hostname", Value: "Living-Room-PC.local", ObservedAt: now},
		{Source: model.SourceMDNS, IP: ip, Key: "service", Value: "_airplay._tcp", ObservedAt: now},
		// SSDP
		{Source: model.SourceSSDP, IP: ip, Key: "st", Value: "urn:schemas-upnp-org:device:MediaRenderer:1", ObservedAt: now},
		{Source: model.SourceSSDP, IP: ip, Key: "server", Value: "Linux UPnP/1.0", ObservedAt: now},
	}

	bags := ResolveEvidence(obs, raw, nil)
	if len(bags) != 1 {
		t.Fatalf("expected 1 bag, got %d", len(bags))
	}

	bag := bags[mac]
	if bag == nil {
		t.Fatalf("expected bag for MAC %s", mac)
	}

	// Canonical Name should prefer mDNS non-generic name
	if bag.CanonicalName != "Living-Room-PC" {
		t.Errorf("CanonicalName = %q, want %q", bag.CanonicalName, "Living-Room-PC")
	}

	// Observed names should include both mDNS and NBNS
	if len(bag.ObservedNames) != 2 {
		t.Errorf("expected 2 observed names, got %d", len(bag.ObservedNames))
	}

	// MAC match should be verified without conflict
	if len(bag.Conflicts) != 0 {
		t.Errorf("expected 0 conflicts, got %v", bag.Conflicts)
	}

	var foundMACMatch bool
	for _, item := range bag.Current {
		if item.Key == "mac_match" && item.Value == "true" {
			foundMACMatch = true
			break
		}
	}
	if !foundMACMatch {
		t.Error("expected mac_match=true evidence item")
	}
}

func TestResolveEvidence_MACMismatchConflict(t *testing.T) {
	ip := netip.MustParseAddr("192.168.1.60")
	arpMAC := "aa:bb:cc:dd:ee:01"
	fakeNBNSMAC := "aa:bb:cc:dd:ee:99"
	now := time.Now().UTC()

	obs := []observation{
		{IP: ip, MAC: arpMAC},
	}

	raw := []model.DiscoveryEvidence{
		{Source: model.SourceNBNS, IP: ip, Key: "hostname", Value: "STRANGE-DEVICE", ObservedAt: now},
		{Source: model.SourceNBNS, IP: ip, Key: "unit_id", Value: fakeNBNSMAC, ObservedAt: now},
	}

	bags := ResolveEvidence(obs, raw, nil)
	bag := bags[arpMAC]
	if bag == nil {
		t.Fatalf("expected bag for %s", arpMAC)
	}

	// Conflict must be flagged!
	if len(bag.Conflicts) == 0 {
		t.Fatal("expected conflict note for mismatched MAC, got none")
	}

	var foundMismatch bool
	for _, item := range bag.Current {
		if item.Key == "mac_match" && item.Value == "mismatch" {
			foundMismatch = true
			break
		}
	}
	if !foundMismatch {
		t.Error("expected mac_match=mismatch evidence item")
	}
}

func TestResolveEvidence_CustomAliasPrecedence(t *testing.T) {
	ip := netip.MustParseAddr("192.168.1.70")
	mac := "11:22:33:44:55:66"
	now := time.Now().UTC()

	obs := []observation{
		{IP: ip, MAC: mac},
	}

	raw := []model.DiscoveryEvidence{
		{Source: model.SourceMDNS, IP: ip, Key: "hostname", Value: "device.local", ObservedAt: now},
	}

	customAliases := map[string]string{
		mac: "My Custom Server",
	}

	bags := ResolveEvidence(obs, raw, customAliases)
	bag := bags[mac]
	if bag == nil {
		t.Fatalf("expected bag for %s", mac)
	}

	if bag.CanonicalName != "My Custom Server" {
		t.Errorf("CanonicalName = %q, want %q", bag.CanonicalName, "My Custom Server")
	}
}
