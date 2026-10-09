// Package model defines the wire types shared between the engine and the UI.
// JSON field names mirror frontend/src/types/*.ts exactly; timestamps are
// RFC 3339 and are formatted for display by the frontend adapter.
package model

import "time"

const (
	StatusOnline  = "online"
	StatusOffline = "offline"
)

// Device types (frontend DeviceType).
const (
	TypeRouter        = "Router"
	TypeComputer      = "Computer"
	TypePhone         = "Phone"
	TypeTablet        = "Tablet"
	TypeTV            = "TV"
	TypePrinter       = "Printer"
	TypeCamera        = "Camera"
	TypeIoT           = "IoT"
	TypeServer        = "Server"
	TypeNetworkDevice = "Network Device"
	TypeGameConsole   = "Game Console"
	TypeUnknown       = "Unknown"
)

// Network event types (frontend EventType).
const (
	EvNewDevice     = "new_device"
	EvOnline        = "online"
	EvOffline       = "offline"
	EvNetworkChange = "network_change"
	EvScan          = "scan"
	EvServiceChange = "service_change"
	EvDeviceMerged  = "device_merged"
)

// Trust status classifications (frontend TrustStatus).
const (
	TrustKnown   = "known"
	TrustGuest   = "guest"
	TrustUnknown = "unknown"
)

// Per-device history types (frontend DeviceEvent['type']).
const (
	HistDiscovered      = "discovered"
	HistOnline          = "online"
	HistOffline         = "offline"
	HistIPChanged       = "ip_changed"
	HistServiceDetected = "service_detected"
	HistDeviceMerged    = "device_merged"
)

type DeviceService struct {
	Port     int    `json:"port"`
	Protocol string `json:"protocol"` // TCP | UDP
	Service  string `json:"service"`
	Status   string `json:"status"` // Open | Closed | Filtered
}

type DeviceEvent struct {
	ID          string    `json:"id"`
	Timestamp   time.Time `json:"timestamp"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
}

type Device struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	CustomAlias string          `json:"customAlias,omitempty"`
	Hostname    string          `json:"hostname"`
	IP          string          `json:"ip"`
	IPv6        string          `json:"ipv6,omitempty"`
	MAC         string          `json:"mac"`
	Vendor      string          `json:"vendor"`
	Type            string          `json:"type"`
	Status          string          `json:"status"`
	IsNew           bool            `json:"isNew,omitempty"`
	IsRandomizedMAC bool            `json:"isRandomizedMac"`
	TrustStatus     string          `json:"trustStatus"` // known | guest | unknown
	MergedInto      string          `json:"mergedInto,omitempty"`
	FirstSeen       time.Time       `json:"firstSeen"`
	LastSeen        time.Time       `json:"lastSeen"`
	LatencyMs       *int            `json:"latencyMs,omitempty"`
	OS              string          `json:"os,omitempty"`
	Notes           string          `json:"notes,omitempty"`
	Services        []DeviceService     `json:"services,omitempty"`
	History         []DeviceEvent       `json:"history,omitempty"`
	Evidence        []DiscoveryEvidence `json:"evidence,omitempty"`
}

type NetworkEvent struct {
	ID         string    `json:"id"`
	Timestamp  time.Time `json:"timestamp"`
	Type       string    `json:"type"`
	Title      string    `json:"title"`
	DeviceName string    `json:"deviceName,omitempty"`
	DeviceID   string    `json:"deviceId,omitempty"`
	IP         string    `json:"ip,omitempty"`
	MAC        string    `json:"mac,omitempty"`
	Details    string    `json:"details,omitempty"`
}

type NetworkInterfaceItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"` // Wi-Fi | Ethernet | VPN
	AdapterName string `json:"adapterName"`
	IP          string `json:"ip"`
	Subnet      string `json:"subnet"`
	Gateway     string `json:"gateway"`
	MAC         string `json:"mac"`
	Status      string `json:"status"` // Connected | Disconnected
	SpeedMbps   int    `json:"speedMbps,omitempty"`
	IsDefault   bool   `json:"isDefault"`
}

type NetworkPingStats struct {
	CurrentMs         int     `json:"currentMs"`
	AvgMs             int     `json:"avgMs"`
	PeakMs            int     `json:"peakMs"`
	PacketLossPercent float64 `json:"packetLossPercent"`
	History           []int   `json:"history"`
}

type NetworkInfo struct {
	NetworkName     string                 `json:"networkName"`
	SSID            string                 `json:"ssid"`
	InterfaceName   string                 `json:"interfaceName"`
	InterfaceType   string                 `json:"interfaceType"`
	Gateway         string                 `json:"gateway"`
	Subnet          string                 `json:"subnet"`
	DNS             []string               `json:"dns"`
	LocalIP         string                 `json:"localIp"`
	PublicIP        string                 `json:"publicIp,omitempty"`
	IsPublicNetwork bool                   `json:"isPublicNetwork,omitempty"`
	Broadcast       string                 `json:"broadcast"`
	Netmask         string                 `json:"netmask"`
	TotalAddresses  int                    `json:"totalAddresses"`
	ActiveAddresses int                    `json:"activeAddresses"`
	Status          string                 `json:"status"` // Connected | Disconnected | Error
	PingStats       NetworkPingStats       `json:"pingStats"`
	Interfaces      []NetworkInterfaceItem `json:"interfaces"`
}

type ScanResult struct {
	ScanID           string    `json:"scanId"`
	Type             string    `json:"type"` // quick | full
	ScannedAddresses int       `json:"scannedAddresses"`
	TotalAddresses   int       `json:"totalAddresses"`
	DevicesFound     int       `json:"devicesFound"`
	NewDevices       int       `json:"newDevices"`
	Errors           int       `json:"errors"`
	DurationMs       int64     `json:"durationMs"`
	Timestamp        time.Time `json:"timestamp"`
}

// DevicePatch is the set of user-editable fields (PATCH /devices/{id}).
// Nil means "leave unchanged".
type DevicePatch struct {
	CustomAlias *string `json:"customAlias,omitempty"`
	Notes       *string `json:"notes,omitempty"`
	IsNew       *bool   `json:"isNew,omitempty"`
	TrustStatus *string `json:"trustStatus,omitempty"`
	Type        *string `json:"type,omitempty"`
}

// MergeDevicesRequest defines the payload for POST /v1/devices/{targetId}/merge.
type MergeDevicesRequest struct {
	SourceID string `json:"sourceId"`
}

// Settings represents persistent user preferences.
type Settings struct {
	AutoDiscovery       bool   `json:"autoDiscovery"`
	ScanInterval        string `json:"scanInterval"` // "1m" | "5m" | "15m" | "1h" | "manual"
	NotifyNewDevice     bool   `json:"notifyNewDevice"`
	NotifyDeviceOffline bool   `json:"notifyDeviceOffline"`
	NotifyNetworkChange bool   `json:"notifyNetworkChange"`
	LaunchAtStartup     bool   `json:"launchAtStartup"`
	StartMinimized      bool   `json:"startMinimized"`
	Language            string `json:"language,omitempty"`
}

// DefaultSettings returns safe initial preferences.
func DefaultSettings() Settings {
	return Settings{
		AutoDiscovery:       true,
		ScanInterval:        "5m",
		NotifyNewDevice:     true,
		NotifyDeviceOffline: false,
		NotifyNetworkChange: true,
		LaunchAtStartup:     false,
		StartMinimized:      false,
		Language:            "en",
	}
}

// DatabaseStats represents SQLite status and record counts.
type DatabaseStats struct {
	DBPath        string `json:"dbPath"`
	FileSizeBytes int64  `json:"fileSizeBytes"`
	DeviceCount   int    `json:"deviceCount"`
	EventCount    int    `json:"eventCount"`
	ScanCount     int    `json:"scanCount"`
	WALEnabled    bool   `json:"walEnabled"`
}

// MaintenanceResult returns outcome for operations like VACUUM or integrity check.
type MaintenanceResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
