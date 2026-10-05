// Package fingerprint classifies devices from passive evidence (vendor,
// hostname, open ports). It is deliberately conservative: when the evidence
// is ambiguous the answer is "Unknown", never a guess.
package fingerprint

import (
	"fmt"
	"strings"

	"netwatch/internal/model"
	"netwatch/internal/oui"
)

// Evidence is everything known about a device at classification time.
type Evidence struct {
	Vendor    string // full IEEE vendor name, "" if unknown
	Hostname  string
	MAC       string
	IP        string
	IsGateway bool
	OpenPorts []int
}

// Classify returns a model.Type* value.
func Classify(e Evidence) string {
	if e.IsGateway {
		return model.TypeRouter
	}
	host := strings.ToLower(e.Hostname)
	vendor := strings.ToLower(e.Vendor)
	ports := make(map[int]bool, len(e.OpenPorts))
	for _, p := range e.OpenPorts {
		ports[p] = true
	}

	// 1. Services are the strongest evidence.
	switch {
	case ports[9100] || ports[515] || ports[631]:
		return model.TypePrinter
	case ports[62078]: // Apple lockdownd, present on iPhone/iPad
		if hasAny(host, "ipad") {
			return model.TypeTablet
		}
		return model.TypePhone
	case ports[8001] || ports[8002] || ports[8008] || ports[8009] || ports[8060]:
		return model.TypeTV // Samsung Tizen API, Chromecast/Google Cast, Roku ECP
	case ports[554] && !ports[445] && !ports[3389]:
		return model.TypeCamera // RTSP
	case ports[10001] || ports[10002]:
		return model.TypeNetworkDevice // UniFi device inform
	case ports[1400]:
		return model.TypeIoT // Sonos SOAP
	}

	// 2. Hostname hints.
	switch {
	case hasAny(host, "iphone", "galaxy-s", "pixel-", "redmi", "oneplus", "android",
		"huawei-", "oppo-", "vivo-", "poco-", "realme-", "nokia-", "moto-"):
		return model.TypePhone
	case hasAny(host, "ipad", "tablet", "galaxy-tab", "kindle", "surface-go"):
		return model.TypeTablet
	case hasAny(host, "-tv", "tv-", "bravia", "roku", "chromecast", "firetv",
		"appletv", "apple-tv", "shield", "mi-box", "nvidia-shield", "webos"):
		return model.TypeTV
	case hasAny(host, "printer", "laserjet", "officejet", "deskjet", "epson", "brother",
		"canon-printer", "pixma", "workforce"):
		return model.TypePrinter
	case hasAny(host, "cam", "ipcam", "doorbell", "camera"):
		return model.TypeCamera
	case hasAny(host, "macbook", "imac", "-mbp", "-mba", "mbp-", "mba-", "desktop",
		"laptop", "thinkpad", "-pc", "pc-", "workstation", "macmini", "mac-mini",
		"mac-pro", "mac-studio", "surface-pro", "surface-book",
		"dell-", "hp-", "lenovo-", "asus-"):
		return model.TypeComputer
	case hasAny(host, "nas", "synology", "diskstation", "qnap", "readynas",
		"freenas", "truenas", "unraid", "openmediavault"):
		return model.TypeServer
	case hasAny(host, "playstation", "ps4", "ps5", "xbox", "switch-",
		"nintendo", "steamdeck", "steam-deck"):
		return model.TypeGameConsole
	case hasAny(host, "echo", "echo-dot", "google-home", "homepod", "home-mini",
		"nest-hub", "alexa"):
		return model.TypeIoT
	case hasAny(host, "ubnt", "unifi", "ap-", "-ap", "access-point"):
		return model.TypeNetworkDevice
	}

	// 3. Vendor hints (only where the vendor makes one kind of device).
	switch {
	case hasAny(vendor, "synology", "qnap", "supermicro", "buffalo"):
		return model.TypeServer
	case hasAny(vendor, "brother", "lexmark", "xerox", "kyocera", "ricoh", "konica",
		"canon inc", "seiko epson", "hp inc"):
		return model.TypePrinter
	case hasAny(vendor, "hikvision", "dahua", "reolink", "axis comm", "amcrest",
		"wyze", "eufy", "arlo", "vivotek", "hanwha"):
		return model.TypeCamera
	case hasAny(vendor, "ubiquiti", "cisco", "netgear", "d-link", "mikrotik",
		"zyxel", "aruba", "ruckus", "tenda", "mercusys", "tp-link", "tplink",
		"huawei device", "linksys", "edimax", "draytek", "juniper", "fortinet",
		"palo alto", "sonicwall", "watchguard"):
		return model.TypeNetworkDevice
	case hasAny(vendor, "espressif", "tuya", "shelly", "itead",
		"signify", "philips lighting", "ring llc", "sonos", "ecobee",
		"nest labs", "amazon technologies", "belkin", "lifx", "tado",
		"meross", "switchbot", "wemo", "tp-link smart", "govee"):
		return model.TypeIoT
	case hasAny(vendor, "roku", "vizio", "tcl", "hisense", "lg electronics",
		"samsung electronics", "sony home"):
		return model.TypeTV
	case hasAny(vendor, "sony interactive"):
		return model.TypeGameConsole
	case hasAny(vendor, "nintendo"):
		return model.TypeGameConsole
	case hasAny(vendor, "valve"):
		return model.TypeGameConsole
	case hasAny(vendor, "microsoft corp"):
		// Microsoft makes Surfaces/Xboxes both — fall through to port check
		if hasAny(host, "xbox") {
			return model.TypeGameConsole
		}
		return model.TypeComputer
	case hasAny(vendor, "dell", "lenovo", "asustek", "micro-star", "gigabyte",
		"microsoft", "raspberry", "intel corporate", "framework", "acer",
		"hewlett packard enterprise", "fujitsu"):
		return model.TypeComputer
	case hasAny(vendor, "apple"):
		// Apple makes phones, tablets, computers, TVs — defer to hostname/ports
	case hasAny(vendor, "xiaomi", "oppo", "vivo", "oneplus", "realme",
		"motorola", "nokia"):
		if hasAny(host, "tablet", "pad") {
			return model.TypeTablet
		}
		return model.TypePhone
	}

	// 4. Desktop-OS services.
	if ports[3389] || ports[445] || ports[139] {
		return model.TypeComputer
	}
	return model.TypeUnknown
}

// DisplayVendor is what the UI shows in the Vendor column.
func DisplayVendor(vendor, mac string) string {
	if vendor != "" {
		return vendor
	}
	if oui.IsLocallyAdministered(mac) {
		return "Private address (randomized MAC)"
	}
	return "Unknown vendor"
}

// Name derives a readable device name when the user has not set an alias.
func Name(e Evidence, typ string) string {
	if h := firstLabel(e.Hostname); h != "" {
		return h
	}
	short := ""
	if e.Vendor != "" {
		short = oui.Short(e.Vendor)
	}
	switch {
	case typ == model.TypeRouter && short != "":
		return short + " Router"
	case typ == model.TypeRouter:
		return "Router"
	case short != "" && typ != model.TypeUnknown:
		return fmt.Sprintf("%s %s", short, typ)
	case short != "":
		return short + " device"
	case oui.IsLocallyAdministered(e.MAC):
		return "Private-address device"
	}
	return "Unknown device"
}

func firstLabel(h string) string {
	h = strings.TrimSpace(strings.TrimSuffix(h, "."))
	if h == "" {
		return ""
	}
	if i := strings.IndexByte(h, '.'); i > 0 {
		h = h[:i]
	}
	return h
}

func hasAny(s string, subs ...string) bool {
	if s == "" {
		return false
	}
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

// ServiceName maps well-known TCP ports to a label for the Services table.
func ServiceName(port int) string {
	if n, ok := services[port]; ok {
		return n
	}
	return "TCP"
}

var services = map[int]string{
	21: "FTP", 22: "SSH", 23: "Telnet", 53: "DNS", 80: "HTTP", 139: "NetBIOS", 443: "HTTPS",
	445: "SMB", 515: "LPD", 554: "RTSP", 631: "IPP", 1400: "Sonos", 1883: "MQTT",
	3389: "RDP", 5000: "UPnP/AirPlay", 5353: "mDNS", 5900: "VNC",
	8001: "Samsung TV API", 8002: "Samsung TV API (TLS)", 8008: "Google Cast HTTP", 8009: "Google Cast",
	8060: "Roku ECP", 8080: "HTTP (alt)", 8443: "HTTPS (alt)", 9100: "JetDirect (RAW print)",
	10001: "UniFi Inform", 10002: "UniFi Inform (TLS)", 62078: "iOS lockdownd",
}

// ScanPorts is the TCP port list probed by Full scans and per-device port scans.
var ScanPorts = []int{
	21, 22, 23, 53, 80, 139, 443, 445, 515, 554, 631, 1400, 1883,
	3389, 5000, 5353, 5900,
	8001, 8002, 8008, 8009, 8060, 8080, 8443,
	9100, 10001, 10002, 62078,
}
