package engine

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/netip"
	"os"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"netwatch/internal/fingerprint"
	"netwatch/internal/model"
	"netwatch/internal/netenv"
	"netwatch/internal/store"
)

var ErrScanRunning = errors.New("a scan is already running")

// ScanSnapshot is the externally visible state of a scan.
type ScanSnapshot struct {
	ID       string            `json:"scanId"`
	Type     string            `json:"type"`
	Progress int               `json:"progress"` // 0..100
	Scanned  int               `json:"scanned"`
	Total    int               `json:"total"`
	Found    int               `json:"found"`
	Done     bool              `json:"done"`
	Result   *model.ScanResult `json:"result,omitempty"`
	Error    string            `json:"error,omitempty"`
}

type scanState struct {
	mu      sync.Mutex
	snap    ScanSnapshot
	changed chan struct{}
}

func newScanState(id, kind string) *scanState {
	return &scanState{snap: ScanSnapshot{ID: id, Type: kind}, changed: make(chan struct{})}
}

func (s *scanState) update(fn func(*ScanSnapshot)) {
	s.mu.Lock()
	fn(&s.snap)
	old := s.changed
	s.changed = make(chan struct{})
	s.mu.Unlock()
	close(old) // wake every waiting stream
}

// Watch returns the current snapshot and a channel closed on the next change.
func (s *scanState) Watch() (ScanSnapshot, <-chan struct{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.snap, s.changed
}

// ScanWatcher lets a caller observe a scan without owning it.
type ScanWatcher interface {
	// Watch returns the current snapshot and a channel closed on the next change.
	Watch() (ScanSnapshot, <-chan struct{})
}

// Scan returns the state of a scan by id, or nil.
func (e *Engine) Scan(id string) ScanWatcher {
	e.scanMu.Lock()
	defer e.scanMu.Unlock()
	if e.scan != nil && e.scan.snap.ID == id {
		return e.scan
	}
	return nil
}

// StartScan begins a scan in the background. If one is already running its id
// is returned together with ErrScanRunning.
func (e *Engine) StartScan(kind string) (string, error) {
	if kind == "" {
		kind = "quick"
	}
	if kind != "quick" && kind != "full" {
		return "", fmt.Errorf("%w: scan type must be quick or full", ErrInvalid)
	}
	e.scanMu.Lock()
	defer e.scanMu.Unlock()
	if e.scan != nil {
		if s, _ := e.scan.Watch(); !s.Done {
			return s.ID, ErrScanRunning
		}
	}
	var b [6]byte
	_, _ = rand.Read(b[:])
	st := newScanState("scan-"+hex.EncodeToString(b[:]), kind)
	e.scan = st
	go e.runScan(st)
	return st.snap.ID, nil
}

func (e *Engine) runScan(st *scanState) {
	// Scans outlive the HTTP request that started them.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	res, err := e.doScan(ctx, st)
	st.update(func(s *ScanSnapshot) {
		s.Done = true
		if err != nil {
			s.Error = err.Error()
			e.log.Error("scan failed", "id", s.ID, "err", err)
			return
		}
		s.Progress, s.Result = 100, &res
	})
}

type probe struct {
	rtt time.Duration
	ok  bool
}

func (e *Engine) doScan(ctx context.Context, st *scanState) (model.ScanResult, error) {
	started := e.now()
	snap, _ := st.Watch()
	kind := snap.Type

	an, err := e.active(ctx)
	if err != nil {
		return model.ScanResult{}, err
	}
	ad := an.Adapter
	hosts := hostsOf(ad.IP)
	total := len(hosts)
	if total == 0 {
		return model.ScanResult{}, fmt.Errorf("subnet %s is too small to scan", ad.IP.Masked())
	}
	st.update(func(s *ScanSnapshot) { s.Total = total })

	pingTimeout := 500 * time.Millisecond
	if kind == "full" {
		pingTimeout = 900 * time.Millisecond
	}

	var scanned, found atomic.Int64
	report := func(progress int) {
		st.update(func(s *ScanSnapshot) {
			s.Scanned, s.Found, s.Progress = int(scanned.Load()), int(found.Load()), progress
		})
	}

	// ---- 0. ARP warm-up: prime the OS cache for hosts that answer ARP ------
	// SendARP is fast (~5ms per call on Windows) and populates the OS ARP
	// table as a side-effect. We probe the gateway and a sample of hosts
	// before the ICMP sweep to catch devices that respond to ARP but drop ICMP.
	warmTargets := arpWarmTargets(hosts, ad.Gateway, 32)
	forEach(ctx, warmTargets, 16, func(ip netip.Addr) {
		// We discard the return value: the important side-effect is that
		// SendARP populates the OS ARP cache so the later Neighbors() call
		// picks these hosts up.
		_, _ = e.env.SendARP(ctx, ip)
	})
	// Ignore individual SendARP results — the important thing is that the
	// OS ARP cache is now primed. We read it after the ICMP sweep.

	// ---- 1. ICMP ping sweep ------------------------------------------------
	alive := map[netip.Addr]probe{}
	var amu sync.Mutex
	var lastReport atomic.Int64
	forEach(ctx, hosts, 96, func(ip netip.Addr) {
		rtt, ok := e.env.Ping(ctx, ip, pingTimeout)
		if ok {
			amu.Lock()
			alive[ip] = probe{rtt, true}
			amu.Unlock()
			found.Add(1)
		}
		n := scanned.Add(1)
		// Throttle progress updates to ~every 2% to keep SSE traffic light.
		if pct := int(n) * 60 / total; int64(pct) > lastReport.Load() || int(n) == total {
			lastReport.Store(int64(pct))
			report(pct)
		}
	})
	if err := ctx.Err(); err != nil {
		return model.ScanResult{}, err
	}
	if kind == "full" { // second chance for hosts that sleep or rate-limit ICMP
		var retry []netip.Addr
		for _, ip := range hosts {
			if _, ok := alive[ip]; !ok {
				retry = append(retry, ip)
			}
		}
		forEach(ctx, retry, 64, func(ip netip.Addr) {
			if rtt, ok := e.env.Ping(ctx, ip, 1500*time.Millisecond); ok {
				amu.Lock()
				alive[ip] = probe{rtt, true}
				amu.Unlock()
				found.Add(1)
			}
		})
	}
	report(65)

	// ---- 2. TCP fallback: detect hosts that block ICMP but accept TCP ------
	// Probe common service ports on hosts not yet discovered. This catches
	// Windows PCs with firewall enabled, IoT devices, etc.
	var tcpTargets []netip.Addr
	for _, ip := range hosts {
		if _, ok := alive[ip]; !ok {
			tcpTargets = append(tcpTargets, ip)
		}
	}
	tcpFound := map[netip.Addr]bool{}
	var tmu sync.Mutex
	if len(tcpTargets) > 0 {
		// Limit TCP discovery to a reasonable subset to keep scans fast.
		// On quick scans: probe only the gateway (if missed) + first 64 unresponding hosts.
		// On full scans: probe all.
		maxTCP := len(tcpTargets)
		if kind == "quick" && maxTCP > 64 {
			// Always include gateway if it was missed.
			limited := make([]netip.Addr, 0, 65)
			for _, ip := range tcpTargets {
				if ip == ad.Gateway {
					limited = append(limited, ip)
				}
			}
			for _, ip := range tcpTargets {
				if ip != ad.Gateway && len(limited) < 64 {
					limited = append(limited, ip)
				}
			}
			tcpTargets = limited
		}
		forEach(ctx, tcpTargets, 48, func(ip netip.Addr) {
			for _, port := range netenv.DiscoveryPorts {
				if ctx.Err() != nil {
					return
				}
				if e.env.TCPOpen(ctx, ip, port, 400*time.Millisecond) {
					tmu.Lock()
					tcpFound[ip] = true
					tmu.Unlock()
					found.Add(1)
					return // one open port is enough to confirm presence
				}
			}
		})
	}
	report(75)

	// ---- 3. Guaranteed gateway: if still not found, try SendARP directly ---
	if _, inAlive := alive[ad.Gateway]; !inAlive && !tcpFound[ad.Gateway] {
		if mac, err := e.env.SendARP(ctx, ad.Gateway); err == nil && mac != "" {
			e.log.Info("gateway found via SendARP", "ip", ad.Gateway, "mac", mac)
			_ = mac // side-effect is ARP cache population
		}
	}

	// ---- 4. ARP table: picks up hosts found by any method ------------------
	// Wait briefly for the OS to finish resolving MACs from the TCP connects
	// and ICMP echoes above, then read the full ARP table.
	time.Sleep(250 * time.Millisecond)
	nbrs, err := e.env.Neighbors(ctx)
	if err != nil {
		e.log.Warn("read ARP table", "err", err)
	}
	prefix := ad.IP.Masked()
	obsByMAC := map[string]*observation{}
	for _, n := range nbrs {
		if !prefix.Contains(n.IP) || n.IP == ad.IP.Addr() {
			continue
		}
		if !inHosts(hosts, n.IP) {
			continue
		}
		o := &observation{IP: n.IP, MAC: n.MAC}
		if p, ok := alive[n.IP]; ok {
			o.RTT, o.HasRTT = p.rtt, true
		}
		obsByMAC[n.MAC] = o
	}
	// Responders without an ARP row (e.g. gateway behind a proxy-ARP bridge)
	// cannot be identified without a MAC, so they are not inventoried.
	if mac := ad.MAC; mac != "" {
		obsByMAC[mac] = &observation{IP: ad.IP.Addr(), MAC: mac, IsSelf: true}
	}
	obs := make([]*observation, 0, len(obsByMAC))
	for _, o := range obsByMAC {
		obs = append(obs, o)
	}
	sort.Slice(obs, func(i, j int) bool { return obs[i].IP.Less(obs[j].IP) })

	// ---- 5. enrichment: hostnames (+ ports on Full scans) ------------------
	selfHost, _ := os.Hostname()
	var enriched atomic.Int64
	nObs := len(obs)
	forEach(ctx, obs, 24, func(o *observation) {
		if o.IsSelf {
			o.Hostname = selfHost
		} else {
			o.Hostname = e.env.ReverseLookup(ctx, o.IP, time.Second)
		}
		if kind == "full" {
			o.Ports, o.PortScanned = e.probePorts(ctx, o.IP), true
		}
		n := int(enriched.Add(1))
		report(80 + n*19/max(nObs, 1))
	})
	if err := ctx.Err(); err != nil {
		return model.ScanResult{}, err
	}

	// ---- 6. diff against the inventory and persist -------------------------
	e.commitMu.Lock()
	defer e.commitMu.Unlock()
	known, err := e.st.ListKnown(ctx, an.Key)
	if err != nil {
		return model.ScanResult{}, err
	}
	plain := make([]observation, len(obs))
	for i, o := range obs {
		plain[i] = *o
	}
	out := reconcile(reconcileIn{
		NetKey: an.Key, Now: e.now(), Gateway: ad.Gateway,
		Known: known, Obs: plain, Vendor: vendorFunc(e.oui),
	})
	dur := e.now().Sub(started)
	res := model.ScanResult{
		ScanID: snap.ID, Type: kind,
		ScannedAddresses: int(scanned.Load()), TotalAddresses: total,
		DevicesFound: out.Found, NewDevices: out.New, Errors: 0,
		DurationMs: dur.Milliseconds(), Timestamp: e.now(),
	}
	out.WS.Events = append(out.WS.Events, model.NetworkEvent{
		Timestamp: e.now(), Type: model.EvScan,
		Title:   fmt.Sprintf("%s discovery scan completed", strings.ToUpper(kind[:1])+kind[1:]),
		Details: fmt.Sprintf("%d addresses scanned · %d devices discovered · %d new", total, out.Found, out.New),
	})
	out.WS.Scan = &store.ScanRecord{ID: snap.ID, Type: kind, StartedAt: started, DurationMs: dur.Milliseconds(),
		Scanned: res.ScannedAddresses, Total: total, Found: out.Found, NewDevices: out.New}
	if err := e.st.Commit(ctx, out.WS); err != nil {
		return model.ScanResult{}, fmt.Errorf("save scan: %w", err)
	}
	_ = e.st.SetMeta(ctx, "last_net_key", an.Key)
	return res, nil
}

// arpWarmTargets returns a subset of hosts to pre-warm with SendARP,
// always including the gateway and prioritizing the low end of the range
// where routers, switches and IoT devices typically sit.
func arpWarmTargets(hosts []netip.Addr, gateway netip.Addr, maxCount int) []netip.Addr {
	out := make([]netip.Addr, 0, maxCount+1)
	// Always include gateway first.
	out = append(out, gateway)
	seen := map[netip.Addr]bool{gateway: true}
	for _, ip := range hosts {
		if len(out) >= maxCount {
			break
		}
		if !seen[ip] {
			out = append(out, ip)
			seen[ip] = true
		}
	}
	return out
}

// probePorts returns the open TCP ports among fingerprint.ScanPorts.
func (e *Engine) probePorts(ctx context.Context, ip netip.Addr) []int {
	var mu sync.Mutex
	var open []int
	forEach(ctx, fingerprint.ScanPorts, 12, func(p int) {
		if e.env.TCPOpen(ctx, ip, p, 500*time.Millisecond) {
			mu.Lock()
			open = append(open, p)
			mu.Unlock()
		}
	})
	sort.Ints(open)
	return open
}

// ScanDevicePorts probes one device and stores what is open.
func (e *Engine) ScanDevicePorts(ctx context.Context, id string) ([]model.DeviceService, error) {
	dev, err := e.st.GetDevice(ctx, id)
	if err != nil {
		return nil, err
	}
	ip, err := netip.ParseAddr(dev.IP)
	if err != nil {
		return nil, err
	}
	if dev.Status != model.StatusOnline {
		return []model.DeviceService{}, nil
	}
	open := e.probePorts(ctx, ip)

	e.commitMu.Lock()
	defer e.commitMu.Unlock()
	// Re-read inside the lock so we diff against the latest stored state.
	cur, err := e.st.GetDevice(ctx, id)
	if err != nil {
		return nil, err
	}
	an, err := e.active(ctx)
	if err != nil {
		return nil, err
	}
	known, err := e.st.ListKnown(ctx, an.Key)
	if err != nil {
		return nil, err
	}
	var prev *store.Known
	for i := range known {
		if known[i].ID == id {
			prev = &known[i]
		}
	}
	if prev == nil {
		return nil, ErrNotFound
	}
	out := reconcile(reconcileIn{
		NetKey: an.Key, Now: e.now(), Gateway: an.Adapter.Gateway, Known: known,
		Obs:    []observation{{IP: ip, MAC: cur.MAC, Hostname: cur.Hostname, Ports: open, PortScanned: true, HasRTT: false}},
		Vendor: vendorFunc(e.oui),
	})
	// Only apply this device's rows: do not count other hosts as missed.
	ws := store.Writeset{Services: out.WS.Services}
	for _, k := range out.WS.Devices {
		if k.ID == id {
			// A port scan is not a presence check: keep the stored status/latency.
			k.Status, k.LatencyMs, k.LastSeen, k.Missed = prev.Status, prev.LatencyMs, prev.LastSeen, prev.Missed
			ws.Devices = append(ws.Devices, k)
		}
	}
	for _, ev := range out.WS.Events {
		if ev.DeviceID == id && ev.Type == model.EvServiceChange {
			ws.Events = append(ws.Events, ev)
		}
	}
	for _, h := range out.WS.History {
		if h.DeviceID == id && h.Type == model.HistServiceDetected {
			ws.History = append(ws.History, h)
		}
	}
	if err := e.st.Commit(ctx, ws); err != nil {
		return nil, err
	}
	svcs := out.WS.Services[id]
	if svcs == nil {
		svcs = []model.DeviceService{}
	}
	return svcs, nil
}

// PingResult is the response of an on-demand ping.
type PingResult struct {
	Success   bool `json:"success"`
	LatencyMs int  `json:"latencyMs"`
}

// Ping pings one private IPv4 address. Public addresses are refused: this is
// a local-network tool, not a general-purpose probe.
func (e *Engine) Ping(ctx context.Context, ipStr string) (PingResult, error) {
	ip, err := netip.ParseAddr(ipStr)
	if err != nil || !ip.Is4() || !(ip.IsPrivate() || ip.IsLinkLocalUnicast() || isCGNAT(ip)) {
		return PingResult{}, fmt.Errorf("%w: only private IPv4 addresses can be pinged", ErrInvalid)
	}
	rtt, ok := e.env.Ping(ctx, ip, 2*time.Second)
	if !ok {
		return PingResult{}, nil
	}
	return PingResult{Success: true, LatencyMs: max(1, int((rtt + 500*time.Microsecond).Milliseconds()))}, nil
}

func isCGNAT(ip netip.Addr) bool {
	return netip.MustParsePrefix("100.64.0.0/10").Contains(ip)
}

// WakeResult is the response of a Wake-on-LAN request.
type WakeResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// WakeOnLAN sends a magic packet. A sent packet does not prove the device
// woke: that depends on the device's BIOS/NIC settings, and the message says so.
func (e *Engine) WakeOnLAN(ctx context.Context, mac string) (WakeResult, error) {
	var bc netip.Addr
	if an, err := e.active(ctx); err == nil {
		bc = broadcast(an.Adapter.IP)
	}
	if err := e.env.WakeOnLAN(ctx, mac, bc); err != nil {
		return WakeResult{}, fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	return WakeResult{Success: true, Message: fmt.Sprintf(
		"Magic packet (102 bytes) sent for %s to %s and 255.255.255.255 on UDP 9. The device wakes only if Wake-on-LAN is enabled in its BIOS/NIC settings.",
		strings.ToUpper(mac), bc)}, nil
}

func hostsOf(p netip.Prefix) []netip.Addr {
	all := netenv.Hosts(p)
	out := all[:0]
	for _, a := range all {
		if a != p.Addr() { // never ping ourselves
			out = append(out, a)
		}
	}
	return out
}

func inHosts(hosts []netip.Addr, a netip.Addr) bool {
	i := sort.Search(len(hosts), func(i int) bool { return !hosts[i].Less(a) })
	if i < len(hosts) && hosts[i] == a {
		return true
	}
	return false
}

// forEach runs fn over items with at most workers goroutines, stopping early
// when ctx is cancelled.
func forEach[T any](ctx context.Context, items []T, workers int, fn func(T)) {
	if len(items) == 0 {
		return
	}
	workers = min(workers, len(items))
	ch := make(chan T)
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for it := range ch {
				if ctx.Err() != nil {
					continue
				}
				fn(it)
			}
		}()
	}
	for _, it := range items {
		if ctx.Err() != nil {
			break
		}
		ch <- it
	}
	close(ch)
	wg.Wait()
}
