//go:build linux

package netenv

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"net/netip"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// New returns the Linux implementation. It exists so the engine can be
// developed and exercised on a non-Windows machine; Windows is the product
// target.
func New() Env { return &linuxEnv{} }

type linuxEnv struct{ Portable }

func (e *linuxEnv) Info(ctx context.Context) (Info, error) {
	var info Info
	gw, metric := defaultRoute()
	ifaces, err := net.Interfaces()
	if err != nil {
		return info, err
	}
	for _, ifc := range ifaces {
		if ifc.Flags&net.FlagLoopback != 0 {
			continue
		}
		ad := Adapter{Name: ifc.Name, Desc: ifc.Name, MAC: strings.ToLower(ifc.HardwareAddr.String()),
			Up: ifc.Flags&net.FlagUp != 0, Type: "Ethernet"}
		if _, err := os.Stat("/sys/class/net/" + ifc.Name + "/wireless"); err == nil {
			ad.Type = "Wi-Fi"
		}
		if strings.HasPrefix(ifc.Name, "tun") || strings.HasPrefix(ifc.Name, "wg") || strings.HasPrefix(ifc.Name, "tailscale") {
			ad.Type = "VPN"
		}
		addrs, _ := ifc.Addrs()
		for _, a := range addrs {
			if p, err := netip.ParsePrefix(a.String()); err == nil && p.Addr().Is4() {
				ad.IP = p
				break
			}
		}
		if g, ok := gw[ifc.Name]; ok {
			ad.Gateway = g
			ad.Metric = metric[ifc.Name]
		}
		info.Adapters = append(info.Adapters, ad)
	}
	if f, err := os.Open("/etc/resolv.conf"); err == nil {
		defer f.Close()
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			if fs := strings.Fields(sc.Text()); len(fs) == 2 && fs[0] == "nameserver" {
				if a, err := netip.ParseAddr(fs[1]); err == nil && a.Is4() {
					info.DNS = append(info.DNS, a)
				}
			}
		}
		if err := sc.Err(); err != nil {
			return info, err
		}
	}
	return info, nil
}

// defaultRoute parses /proc/net/route for per-interface IPv4 default gateways.
func defaultRoute() (map[string]netip.Addr, map[string]int) {
	gw, metric := map[string]netip.Addr{}, map[string]int{}
	f, err := os.Open("/proc/net/route")
	if err != nil {
		return gw, metric
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Scan() // header
	for sc.Scan() {
		fs := strings.Fields(sc.Text())
		if len(fs) < 7 || fs[1] != "00000000" {
			continue
		}
		g, err := strconv.ParseUint(fs[2], 16, 32)
		if err != nil {
			continue
		}
		var b [4]byte
		binary.LittleEndian.PutUint32(b[:], uint32(g))
		m, _ := strconv.Atoi(fs[6])
		gw[fs[0]], metric[fs[0]] = netip.AddrFrom4(b), m
	}
	if err := sc.Err(); err != nil {
		return gw, metric
	}
	return gw, metric
}

func (e *linuxEnv) Neighbors(ctx context.Context) ([]Neighbor, error) {
	f, err := os.Open("/proc/net/arp")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []Neighbor
	sc := bufio.NewScanner(f)
	sc.Scan()
	for sc.Scan() {
		fs := strings.Fields(sc.Text())
		if len(fs) < 4 || fs[2] != "0x2" { // 0x2 = complete
			continue
		}
		if ip, err := netip.ParseAddr(fs[0]); err == nil {
			out = append(out, Neighbor{IP: ip, MAC: fs[3]})
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	// Ingest IPv6 NDP neighbors via unprivileged `ip -6 neigh show`
	if bin, err := exec.LookPath("ip"); err == nil {
		cctx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		if out6, err := exec.CommandContext(cctx, bin, "-6", "neigh", "show").Output(); err == nil {
			for _, line := range strings.Split(string(out6), "\n") {
				fields := strings.Fields(line)
				// Format: <ipv6> dev <iface> lladdr <mac> <STATE>
				for i, f := range fields {
					if f == "lladdr" && i+1 < len(fields) && i > 0 {
						if ip, err := netip.ParseAddr(fields[0]); err == nil {
							out = append(out, Neighbor{IP: ip, MAC: fields[i+1]})
						}
						break
					}
				}
			}
		}
	}
	return FilterNeighbors(out), nil
}

var pingTime = regexp.MustCompile(`time[=<]([0-9.]+)`)

// Ping uses the system ping binary when present, otherwise falls back to TCP
// connect timing (a refused connection still proves the host is alive).
func (e *linuxEnv) Ping(ctx context.Context, ip netip.Addr, timeout time.Duration) (time.Duration, bool) {
	if bin, err := exec.LookPath("ping"); err == nil {
		secs := int(timeout.Seconds() + 0.999)
		cctx, cancel := context.WithTimeout(ctx, timeout+time.Second)
		defer cancel()
		out, err := exec.CommandContext(cctx, bin, "-n", "-c", "1", "-W", strconv.Itoa(secs), ip.String()).Output()
		if err != nil {
			return 0, false
		}
		if m := pingTime.FindSubmatch(out); m != nil {
			if v, err := strconv.ParseFloat(string(m[1]), 64); err == nil {
				return time.Duration(v * float64(time.Millisecond)), true
			}
		}
		return time.Millisecond, true
	}
	for _, port := range []int{443, 80, 22, 445} {
		start := time.Now()
		d := net.Dialer{Timeout: timeout}
		c, err := d.DialContext(ctx, "tcp", net.JoinHostPort(ip.String(), strconv.Itoa(port)))
		if err == nil {
			c.Close()
			return time.Since(start), true
		}
		if strings.Contains(err.Error(), "refused") {
			return time.Since(start), true
		}
	}
	return 0, false
}

var _ = fmt.Sprintf

func (e *linuxEnv) SendARP(_ context.Context, _ netip.Addr) (string, error) {
	return "", ErrUnsupported
}
