//go:build windows

package netenv

import (
	"context"
	"encoding/binary"
	"fmt"
	"net/netip"
	"os/exec"
	"regexp"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// New returns the Windows implementation. It needs no administrator rights:
// adapters come from GetAdaptersAddresses, the ARP table from GetIpNetTable
// and ICMP from IcmpSendEcho (all unprivileged iphlpapi calls).
func New() Env { return &winEnv{} }

type winEnv struct{ Portable }

var (
	iphlpapi         = windows.NewLazySystemDLL("iphlpapi.dll")
	procGetIpNet     = iphlpapi.NewProc("GetIpNetTable")
	procGetIpNet2    = iphlpapi.NewProc("GetIpNetTable2")
	procFreeMibTable = iphlpapi.NewProc("FreeMibTable")
	procSendARP      = iphlpapi.NewProc("SendARP")
	procIcmpCreate   = iphlpapi.NewProc("IcmpCreateFile")
	procIcmpClose    = iphlpapi.NewProc("IcmpCloseHandle")
	procIcmpSendEcho = iphlpapi.NewProc("IcmpSendEcho")
)

// ---- adapters -------------------------------------------------------------

var virtualName = regexp.MustCompile(`(?i)(vethernet|vmware|virtualbox|hyper-v|loopback|bluetooth|teredo|isatap|pseudo|wi-fi direct|miniport|npcap|wfp)`)
var vpnName = regexp.MustCompile(`(?i)(tailscale|wireguard|vpn|openvpn|tap-windows|wintun|zerotier|nord|proton|cisco anyconnect|fortinet|globalprotect)`)

func (e *winEnv) Info(ctx context.Context) (Info, error) {
	var size uint32 = 16 * 1024
	var buf []byte
	const flags = windows.GAA_FLAG_INCLUDE_GATEWAYS | windows.GAA_FLAG_SKIP_ANYCAST | windows.GAA_FLAG_SKIP_MULTICAST
	for i := 0; i < 5; i++ {
		buf = make([]byte, size)
		err := windows.GetAdaptersAddresses(windows.AF_INET, flags, 0, (*windows.IpAdapterAddresses)(unsafe.Pointer(&buf[0])), &size)
		if err == nil {
			break
		}
		if err != windows.ERROR_BUFFER_OVERFLOW {
			return Info{}, fmt.Errorf("GetAdaptersAddresses: %w", err)
		}
	}

	var info Info
	seenDNS := map[netip.Addr]bool{}
	for aa := (*windows.IpAdapterAddresses)(unsafe.Pointer(&buf[0])); aa != nil; aa = aa.Next {
		name := windows.UTF16PtrToString(aa.FriendlyName)
		desc := windows.UTF16PtrToString(aa.Description)
		if aa.IfType == windows.IF_TYPE_SOFTWARE_LOOPBACK || virtualName.MatchString(name+" "+desc) {
			continue
		}
		ad := Adapter{
			Name:   name,
			Desc:   desc,
			Up:     aa.OperStatus == windows.IfOperStatusUp,
			Metric: int(aa.Ipv4Metric),
		}
		switch {
		case vpnName.MatchString(name+" "+desc) || aa.IfType == 131 /* tunnel */ || aa.IfType == 53:
			ad.Type = "VPN"
		case aa.IfType == windows.IF_TYPE_IEEE80211:
			ad.Type = "Wi-Fi"
		case aa.IfType == windows.IF_TYPE_ETHERNET_CSMACD:
			ad.Type = "Ethernet"
		default:
			continue
		}
		if n := int(aa.PhysicalAddressLength); n == 6 {
			ad.MAC = strings.ToLower(fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
				aa.PhysicalAddress[0], aa.PhysicalAddress[1], aa.PhysicalAddress[2],
				aa.PhysicalAddress[3], aa.PhysicalAddress[4], aa.PhysicalAddress[5]))
		}
		if aa.TransmitLinkSpeed != ^uint64(0) {
			ad.SpeedMbps = int(aa.TransmitLinkSpeed / 1_000_000)
		}
		for ua := aa.FirstUnicastAddress; ua != nil; ua = ua.Next {
			ip, ok := netip.AddrFromSlice(ua.Address.IP())
			if !ok || !ip.Unmap().Is4() {
				continue
			}
			ad.IP = netip.PrefixFrom(ip.Unmap(), int(ua.OnLinkPrefixLength))
			break
		}
		for ga := aa.FirstGatewayAddress; ga != nil; ga = ga.Next {
			if ip, ok := netip.AddrFromSlice(ga.Address.IP()); ok && ip.Unmap().Is4() {
				ad.Gateway = ip.Unmap()
				break
			}
		}
		if ad.Up {
			for da := aa.FirstDnsServerAddress; da != nil; da = da.Next {
				ip, ok := netip.AddrFromSlice(da.Address.IP())
				if !ok {
					continue
				}
				ip = ip.Unmap()
				if ip.Is4() && !seenDNS[ip] && ad.Gateway.IsValid() {
					seenDNS[ip] = true
					info.DNS = append(info.DNS, ip)
				}
			}
		}
		info.Adapters = append(info.Adapters, ad)
	}
	if act, ok := info.Active(); ok && act.Type == "Wi-Fi" {
		info.SSID = e.ssid(ctx)
	}
	info.IsPublicNetwork = (e.NetworkCategory(ctx) == "Public")
	return info, nil
}

var ssidLine = regexp.MustCompile(`(?m)^\s*SSID\s*:\s*(.+?)\s*$`)

// ssid reads the connected network name. Windows 11 24H2+ may withhold it
// unless Location access is enabled; in that case "" is returned and the UI
// shows a dash.
func (e *winEnv) ssid(ctx context.Context) string {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "netsh", "wlan", "show", "interfaces")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	if m := ssidLine.FindSubmatch(out); m != nil {
		return string(m[1])
	}
	return ""
}

// NetworkCategory queries the Windows network connection profile ("Public", "Private", "Domain").
func (e *winEnv) NetworkCategory(ctx context.Context) string {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "powershell", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", "(Get-NetConnectionProfile | Select-Object -ExpandProperty NetworkCategory -First 1)")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
	out, err := cmd.Output()
	if err != nil {
		return "Private"
	}
	cat := strings.TrimSpace(string(out))
	if cat != "" {
		return cat
	}
	return "Private"
}

// ---- ARP / NDP table ------------------------------------------------------

type sockaddrInet struct {
	Family uint16
	Data   [26]byte
}

type mibIpNetRow2 struct {
	Address               sockaddrInet
	InterfaceIndex        uint32
	InterfaceLuid         uint64
	PhysicalAddress       [32]byte
	PhysicalAddressLength uint32
	State                 uint32
	Flags                 uint8
	Pad                   [3]byte
	ReachabilityTime      uint32
}

// MIB_IPNETROW: dwIndex(4), dwPhysAddrLen(4), bPhysAddr[8], dwAddr(4), dwType(4) = 24 bytes.
const ipNetRowSize = 24

func (e *winEnv) Neighbors(ctx context.Context) ([]Neighbor, error) {
	// Try modern GetIpNetTable2 (supports both IPv4 ARP and IPv6 NDP without elevation)
	if procGetIpNet2.Find() == nil && procFreeMibTable.Find() == nil {
		var pTable unsafe.Pointer
		r, _, _ := procGetIpNet2.Call(0, uintptr(unsafe.Pointer(&pTable)))
		if r == 0 && pTable != nil {
			defer procFreeMibTable.Call(uintptr(pTable))
			numEntries := *(*uint32)(pTable)
			offset := uintptr(8)
			rowSize := unsafe.Sizeof(mibIpNetRow2{})

			var out []Neighbor
			for i := uint32(0); i < numEntries; i++ {
				rowPtr := (*mibIpNetRow2)(unsafe.Add(pTable, offset+uintptr(i)*rowSize))
				if rowPtr.PhysicalAddressLength == 6 && rowPtr.State != 0 && rowPtr.State != 1 {
					mac := fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
						rowPtr.PhysicalAddress[0], rowPtr.PhysicalAddress[1], rowPtr.PhysicalAddress[2],
						rowPtr.PhysicalAddress[3], rowPtr.PhysicalAddress[4], rowPtr.PhysicalAddress[5])

					var ip netip.Addr
					switch rowPtr.Address.Family {
					case 2: // AF_INET
						ip = netip.AddrFrom4(*(*[4]byte)(unsafe.Pointer(&rowPtr.Address.Data[2])))
					case 23: // AF_INET6
						ip = netip.AddrFrom16(*(*[16]byte)(unsafe.Pointer(&rowPtr.Address.Data[6])))
					}

					if ip.IsValid() {
						out = append(out, Neighbor{IP: ip, MAC: mac})
					}
				}
			}
			return FilterNeighbors(out), nil
		}
	}

	// Fallback to legacy GetIpNetTable (IPv4 only)
	var size uint32
	r, _, _ := procGetIpNet.Call(0, uintptr(unsafe.Pointer(&size)), 0)
	if r != uintptr(windows.ERROR_INSUFFICIENT_BUFFER) && r != 0 {
		return nil, fmt.Errorf("GetIpNetTable: %w", syscall.Errno(r))
	}
	if size == 0 {
		return nil, nil
	}
	buf := make([]byte, size)
	r, _, _ = procGetIpNet.Call(uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size)), 0)
	if r != 0 {
		return nil, fmt.Errorf("GetIpNetTable: %w", syscall.Errno(r))
	}
	n := int(binary.LittleEndian.Uint32(buf[0:4]))
	var out []Neighbor
	for i := 0; i < n; i++ {
		off := 4 + i*ipNetRowSize
		if off+ipNetRowSize > len(buf) {
			break
		}
		row := buf[off : off+ipNetRowSize]
		physLen := binary.LittleEndian.Uint32(row[4:8])
		typ := binary.LittleEndian.Uint32(row[20:24])
		if physLen != 6 || (typ != 3 && typ != 4) { // dynamic or static only
			continue
		}
		mac := fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", row[8], row[9], row[10], row[11], row[12], row[13])
		ip := netip.AddrFrom4([4]byte{row[16], row[17], row[18], row[19]})
		out = append(out, Neighbor{IP: ip, MAC: mac})
	}
	return FilterNeighbors(out), nil
}

// SendARP sends a directed ARP request via iphlpapi!SendARP. It is
// unprivileged and populates the OS ARP cache as a side-effect, which is
// exactly what we want: even if the function itself succeeds only partially,
// a following GetIpNetTable call will pick the entry up.
func (e *winEnv) SendARP(_ context.Context, ip netip.Addr) (string, error) {
	b, ok := IPv4Bytes(ip)
	if !ok {
		return "", fmt.Errorf("SendARP: not IPv4")
	}
	dest := binary.LittleEndian.Uint32(b[:])
	var mac [6]byte
	macLen := uint32(6)
	r, _, _ := procSendARP.Call(
		uintptr(dest), 0, // destIP, srcIP (0 = auto)
		uintptr(unsafe.Pointer(&mac[0])),
		uintptr(unsafe.Pointer(&macLen)),
	)
	if r != 0 {
		return "", fmt.Errorf("SendARP %s: %w", ip, syscall.Errno(r))
	}
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
		mac[0], mac[1], mac[2], mac[3], mac[4], mac[5]), nil
}

// ---- ICMP -----------------------------------------------------------------

func (e *winEnv) Ping(ctx context.Context, ip netip.Addr, timeout time.Duration) (time.Duration, bool) {
	b, ok := IPv4Bytes(ip)
	if !ok || ctx.Err() != nil {
		return 0, false
	}
	h, _, _ := procIcmpCreate.Call()
	if h == 0 || h == ^uintptr(0) {
		return 0, false
	}
	defer procIcmpClose.Call(h)

	dest := binary.LittleEndian.Uint32(b[:])
	payload := []byte("NETWATCH")
	reply := make([]byte, 128) // ICMP_ECHO_REPLY + payload + ICMP_ERROR_INFO
	n, _, _ := procIcmpSendEcho.Call(h, uintptr(dest),
		uintptr(unsafe.Pointer(&payload[0])), uintptr(len(payload)), 0,
		uintptr(unsafe.Pointer(&reply[0])), uintptr(len(reply)), uintptr(timeout.Milliseconds()))
	if n == 0 {
		return 0, false
	}
	// ICMP_ECHO_REPLY: Address(4) Status(4) RoundTripTime(4) ...
	if binary.LittleEndian.Uint32(reply[0:4]) != dest || binary.LittleEndian.Uint32(reply[4:8]) != 0 {
		return 0, false
	}
	return time.Duration(binary.LittleEndian.Uint32(reply[8:12])) * time.Millisecond, true
}
