package engine

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"sync"
	"time"

	"netwatch/internal/netenv"
)

// fakeEnv simulates a /24 home network.
type fakeEnv struct {
	netenv.Portable
	mu        sync.Mutex
	adapter   netenv.Adapter
	ssid      string
	hosts     map[netip.Addr]fakeHost // present on the network
	ports     map[netip.Addr][]int
	arpOnly   map[netip.Addr]bool // answers ARP but not ICMP
	tcpOnly   map[netip.Addr]bool // answers TCP probe but not ICMP/ARP initially
	arpTable  map[netip.Addr]string
	rdns      map[netip.Addr]string
	pings     int
	wolTarget netip.Addr
}

type fakeHost struct {
	mac string
	rtt time.Duration
}

func newFakeEnv() *fakeEnv {
	f := &fakeEnv{
		adapter: netenv.Adapter{
			Name: "Wi-Fi", Type: "Wi-Fi", Desc: "Fake WLAN", MAC: "f0:18:98:c3:54:d2", Up: true,
			IP: netip.MustParsePrefix("192.168.1.24/24"), Gateway: netip.MustParseAddr("192.168.1.1"), Metric: 25, SpeedMbps: 866,
		},
		ssid:     "Home-5G",
		hosts:    map[netip.Addr]fakeHost{},
		ports:    map[netip.Addr][]int{},
		arpOnly:  map[netip.Addr]bool{},
		tcpOnly:  map[netip.Addr]bool{},
		arpTable: map[netip.Addr]string{},
		rdns:     map[netip.Addr]string{},
	}
	f.add("192.168.1.1", "e8:48:b8:31:7a:01", 2*time.Millisecond)
	f.add("192.168.1.12", "3c:06:30:4a:21:8f", 3*time.Millisecond)
	f.rdns[netip.MustParseAddr("192.168.1.12")] = "adrian-mbp.local"
	f.add("192.168.1.18", "aa:bb:cc:dd:ee:18", 4*time.Millisecond)
	f.ports[netip.MustParseAddr("192.168.1.18")] = []int{8001, 8002}
	return f
}

func (f *fakeEnv) add(ip, mac string, rtt time.Duration) {
	a := netip.MustParseAddr(ip)
	f.hosts[a] = fakeHost{mac, rtt}
	f.arpTable[a] = mac
}

// addTCPOnly adds a host that blocks ICMP and doesn't appear in ARP initially,
// but responds to TCP connect probes. This simulates Windows machines with
// firewalls that drop ICMP but accept connections on known ports.
func (f *fakeEnv) addTCPOnly(ip, mac string, openPorts []int) {
	a := netip.MustParseAddr(ip)
	f.hosts[a] = fakeHost{mac, 0}
	f.tcpOnly[a] = true
	f.ports[a] = openPorts
	// Not in arpTable initially — gets added when SendARP succeeds or
	// when the OS resolves the MAC during TCP connect.
}

func (f *fakeEnv) remove(ip string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	a := netip.MustParseAddr(ip)
	delete(f.hosts, a)
	delete(f.arpTable, a)
	delete(f.tcpOnly, a)
}

func (f *fakeEnv) Info(context.Context) (netenv.Info, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return netenv.Info{
		Adapters: []netenv.Adapter{f.adapter, {Name: "Ethernet", Type: "Ethernet", Desc: "Fake NIC", MAC: "70:b5:e8:2c:91:05"}},
		DNS:      []netip.Addr{netip.MustParseAddr("192.168.1.1")},
		SSID:     f.ssid,
	}, nil
}

func (f *fakeEnv) Neighbors(context.Context) ([]netenv.Neighbor, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []netenv.Neighbor
	for ip, mac := range f.arpTable {
		out = append(out, netenv.Neighbor{IP: ip, MAC: mac})
	}
	return out, nil
}

func (f *fakeEnv) Ping(_ context.Context, ip netip.Addr, _ time.Duration) (time.Duration, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.pings++
	h, ok := f.hosts[ip]
	if !ok || f.arpOnly[ip] || f.tcpOnly[ip] {
		return 0, false
	}
	return h.rtt, true
}

// SendARP simulates a directed ARP request. Returns the MAC from the hosts
// map if the host is present (regardless of arpOnly/tcpOnly flags, since ARP
// operates at L2). As a side-effect it populates the arpTable — matching
// real Windows behaviour where SendARP fills the OS cache.
func (f *fakeEnv) SendARP(_ context.Context, ip netip.Addr) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	h, ok := f.hosts[ip]
	if !ok {
		return "", fmt.Errorf("SendARP %s: host unreachable", ip)
	}
	// Populate the ARP cache as a side-effect.
	f.arpTable[ip] = h.mac
	return h.mac, nil
}

func (f *fakeEnv) TCPOpen(_ context.Context, ip netip.Addr, port int, _ time.Duration) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, p := range f.ports[ip] {
		if p == port {
			// TCP connect also populates ARP cache (OS resolves MAC during connect).
			if h, ok := f.hosts[ip]; ok {
				f.arpTable[ip] = h.mac
			}
			return true
		}
	}
	return false
}

func (f *fakeEnv) ReverseLookup(_ context.Context, ip netip.Addr, _ time.Duration) string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.rdns[ip]
}

func (f *fakeEnv) WakeOnLAN(_ context.Context, mac string, bc netip.Addr) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, err := net.ParseMAC(mac); err != nil {
		return err
	}
	f.wolTarget = bc
	return nil
}
