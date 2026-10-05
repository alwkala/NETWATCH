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
)

// Per-device history types (frontend DeviceEvent['type']).
const (
	HistDiscovered      = "discovered"
	HistOnline          = "online"
	HistOffline         = "offline"
	HistIPChanged       = "ip_changed"
	HistServiceDetected = "service_detected"
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
	MAC         string          `json:"mac"`
	Vendor      string          `json:"vendor"`
	Type        string          `json:"type"`
	Status      string          `json:"status"`
	IsNew       bool            `json:"isNew,omitempty"`
	FirstSeen   time.Time       `json:"firstSeen"`
	LastSeen    time.Time       `json:"lastSeen"`
	LatencyMs   *int            `json:"latencyMs,omitempty"`
	OS          string          `json:"os,omitempty"`
	Notes       string          `json:"notes,omitempty"`
	Services    []DeviceService `json:"services,omitempty"`
	History     []DeviceEvent   `json:"history,omitempty"`
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
}
