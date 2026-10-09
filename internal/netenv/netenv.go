// Package netenv is the boundary between the engine and the operating system.
// Everything that touches the network stack (ARP table, ICMP, adapters) sits
// behind the Env interface so the engine can be tested with a fake and the
// Windows implementation can be swapped without touching engine code.
package netenv

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"strings"
	"time"
)

// Adapter is one network interface as the UI presents it.
type Adapter struct {
	Name      string // friendly name ("Wi-Fi")
	Type      string // Wi-Fi | Ethernet | VPN
	Desc      string // driver description
	IP        netip.Prefix
	Gateway   netip.Addr
	MAC       string
	Up        bool
	SpeedMbps int
	Metric    int // lower = preferred route
}

// Info is a snapshot of the machine's network configuration.
type Info struct {
	Adapters []Adapter
	DNS      []netip.Addr
	SSID     string
}

// Active returns the adapter that carries the default route, if any.
func (i Info) Active() (Adapter, bool) {
	var best *Adapter
	for n := range i.Adapters {
		a := &i.Adapters[n]
		if !a.Up || !a.IP.IsValid() || !a.Gateway.IsValid() {
			continue
		}
		if best == nil || a.Metric < best.Metric {
			best = a
		}
	}
	if best == nil {
		return Adapter{}, false
	}
	return *best, true
}

// Neighbor is one ARP/NDP table entry.
type Neighbor struct {
	IP  netip.Addr
	MAC string // lower-case, colon separated
}

// Env abstracts the OS network stack.
type Env interface {
	Info(ctx context.Context) (Info, error)
	// Neighbors returns the OS ARP table (IPv4).
	Neighbors(ctx context.Context) ([]Neighbor, error)
	// Ping sends one ICMP echo and reports the round-trip time.
	Ping(ctx context.Context, ip netip.Addr, timeout time.Duration) (time.Duration, bool)
	// SendARP sends a directed ARP request and returns the resolved MAC.
	// It populates the OS ARP cache as a side-effect. On platforms that
	// do not support it, it returns ("", ErrUnsupported).
	SendARP(ctx context.Context, ip netip.Addr) (mac string, err error)
	// TCPOpen reports whether a TCP connect to ip:port succeeds.
	TCPOpen(ctx context.Context, ip netip.Addr, port int, timeout time.Duration) bool
	// ReverseLookup returns a hostname for ip, or "".
	ReverseLookup(ctx context.Context, ip netip.Addr, timeout time.Duration) string
	// WakeOnLAN broadcasts a magic packet.
	WakeOnLAN(ctx context.Context, mac string, subnetBroadcast netip.Addr) error
}

// DiscoveryPorts is the set of TCP ports probed to detect hosts that block
// ICMP but accept TCP connections. Kept small and focused on the most common
// services found on LAN devices.
var DiscoveryPorts = []int{80, 443, 445, 3389, 8080}

var ErrUnsupported = errors.New("not supported on this platform")

// Portable provides the parts of Env that are identical on every OS.
type Portable struct{}

func (Portable) TCPOpen(ctx context.Context, ip netip.Addr, port int, timeout time.Duration) bool {
	d := net.Dialer{Timeout: timeout}
	c, err := d.DialContext(ctx, "tcp", net.JoinHostPort(ip.String(), itoa(port)))
	if err != nil {
		return false
	}
	c.Close()
	return true
}

func (Portable) ReverseLookup(ctx context.Context, ip netip.Addr, timeout time.Duration) string {
	if !IsAllowedUnicastTarget(ip) {
		return ""
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	r := net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			host, _, err := net.SplitHostPort(address)
			if err != nil {
				return nil, err
			}
			addr, err := netip.ParseAddr(host)
			if err != nil || !IsAllowedUnicastTarget(addr) {
				// Zero-Egress Invariant: Strictly prohibit querying external/public DNS servers for LAN PTR queries.
				return nil, ErrDestinationNotAllowed
			}
			var d net.Dialer
			return d.DialContext(ctx, network, address)
		},
	}
	names, err := r.LookupAddr(ctx, ip.String())
	if err != nil || len(names) == 0 {
		return ""
	}
	return strings.TrimSuffix(names[0], ".")
}

func (Portable) WakeOnLAN(ctx context.Context, mac string, subnetBroadcast netip.Addr) error {
	hw, err := net.ParseMAC(mac)
	if err != nil || len(hw) != 6 {
		return errors.New("invalid MAC address")
	}
	pkt := make([]byte, 0, 102)
	for i := 0; i < 6; i++ {
		pkt = append(pkt, 0xFF)
	}
	for i := 0; i < 16; i++ {
		pkt = append(pkt, hw...)
	}
	// Send to the subnet-directed and the limited broadcast so the packet
	// leaves through the right adapter on multi-homed machines.
	targets := []string{"255.255.255.255:9"}
	if subnetBroadcast.IsValid() {
		targets = append([]string{net.JoinHostPort(subnetBroadcast.String(), "9")}, targets...)
	}
	var sent int
	var lastErr error
	for _, t := range targets {
		d := net.Dialer{Timeout: 2 * time.Second}
		c, err := d.DialContext(ctx, "udp4", t)
		if err != nil {
			lastErr = err
			continue
		}
		_, err = c.Write(pkt)
		c.Close()
		if err != nil {
			lastErr = err
			continue
		}
		sent++
	}
	if sent == 0 {
		return lastErr
	}
	return nil
}

// FilterNeighbors drops entries that are never real hosts: broadcast,
// multicast, link-local, invalid and all-zero MACs.
func FilterNeighbors(in []Neighbor) []Neighbor {
	out := in[:0:0]
	for _, n := range in {
		if !n.IP.Is4() || n.IP.IsMulticast() || n.IP.IsLinkLocalUnicast() || n.IP.IsUnspecified() {
			continue
		}
		hw, err := net.ParseMAC(n.MAC)
		if err != nil || len(hw) != 6 {
			continue
		}
		if hw[0]&0x01 != 0 { // multicast/broadcast MAC
			continue
		}
		zero := true
		for _, b := range hw {
			if b != 0 {
				zero = false
			}
		}
		if zero {
			continue
		}
		n.MAC = strings.ToLower(hw.String())
		out = append(out, n)
	}
	return out
}

// Hosts lists the usable host addresses to sweep for a local prefix. Prefixes
// larger than /22 are narrowed to the /24 containing the local address so a
// scan stays bounded and polite.
func Hosts(local netip.Prefix) []netip.Addr {
	p := local.Masked()
	if p.Bits() < 22 {
		p2, err := local.Addr().Prefix(24)
		if err != nil {
			return nil
		}
		p = p2
	}
	if p.Bits() >= 31 {
		return nil
	}
	var out []netip.Addr
	for a := p.Addr().Next(); p.Contains(a); a = a.Next() {
		// skip the broadcast address (last in range)
		if !p.Contains(a.Next()) {
			break
		}
		out = append(out, a)
	}
	return out
}

// IPv4Bytes returns the 4-byte form, or false for non-IPv4.
func IPv4Bytes(a netip.Addr) ([4]byte, bool) {
	if !a.Is4() && !a.Is4In6() {
		return [4]byte{}, false
	}
	return a.Unmap().As4(), true
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [12]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
