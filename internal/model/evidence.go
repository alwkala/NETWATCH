package model

import (
	"context"
	"net/netip"
	"time"
)

// DiscoveryScope defines the execution parameters for a discovery pass.
type DiscoveryScope struct {
	Subnet netip.Prefix // Active subnet prefix (e.g. 192.168.1.0/24)
	Hosts  []netip.Addr // Known active host IP targets for directed unicast probes
	Iface  string       // Primary active network interface name
}

// DiscoveryProbe is the unified architectural contract for all network ingestion sensors.
type DiscoveryProbe interface {
	Name() DiscoverySource
	Discover(ctx context.Context, scope DiscoveryScope) ([]DiscoveryEvidence, error)
}


// DiscoverySource identifies the ingestion sensor or protocol that produced evidence.
type DiscoverySource string

const (
	SourceARP  DiscoverySource = "ARP"
	SourceICMP DiscoverySource = "ICMP"
	SourceTCP  DiscoverySource = "TCP"
	SourceNBNS DiscoverySource = "NBNS"
	SourceMDNS DiscoverySource = "mDNS"
	SourceSSDP DiscoverySource = "SSDP"
	SourceDNS  DiscoverySource = "rDNS"
	SourceNDP  DiscoverySource = "NDP"
	SourceWSD  DiscoverySource = "WSD"
)

// DiscoveryEvidence captures an atomic piece of network intelligence observed on the LAN.
type DiscoveryEvidence struct {
	ID         uint64          `json:"id,omitempty"`
	Source     DiscoverySource `json:"source"`
	IP         netip.Addr      `json:"ip"`
	MAC        string          `json:"mac,omitempty"`
	Key        string          `json:"key"`   // e.g. "hostname", "service", "model", "server", "unit_id", "st", "usn"
	Value      string          `json:"value"` // Sanitized string value
	ObservedAt time.Time       `json:"observedAt"`
	LastSeen   time.Time       `json:"lastSeen"`
}

// ObservedName represents a hostname or computer name announced by a specific discovery protocol.
type ObservedName struct {
	Source DiscoverySource `json:"source"`
	Name   string          `json:"name"`
}

// EvidenceBag contains the normalized, correlated intelligence aggregated for a single device identity.
type EvidenceBag struct {
	IP            netip.Addr          `json:"ip"`
	MAC           string              `json:"mac"`
	CanonicalName string              `json:"canonicalName"`
	ObservedNames []ObservedName      `json:"observedNames"`
	Current       []DiscoveryEvidence `json:"current"`
	Historical    []DiscoveryEvidence `json:"historical,omitempty"`
	Conflicts     []string            `json:"conflicts,omitempty"` // Diagnostic alerts (e.g. "NBNS Unit ID mismatch")
}

// NetworkContext captures environment and wireless network parameters decoupled from device ledger entries.
type NetworkContext struct {
	ID         int64     `json:"id,omitempty"`
	Interface  string    `json:"interface"`
	SSID       string    `json:"ssid"`
	BSSID      string    `json:"bssid"`
	Gateway    string    `json:"gateway"`
	Subnet     string    `json:"subnet"`
	IPv4       string    `json:"ipv4"`
	ObservedAt time.Time `json:"observedAt"`
}
