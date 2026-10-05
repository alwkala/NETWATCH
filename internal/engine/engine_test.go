package engine

import (
	"context"
	"errors"
	"net/netip"
	"testing"
	"time"

	"netwatch/internal/model"
	"netwatch/internal/store"
)

func newTestEngine(t *testing.T) (*Engine, *fakeEnv) {
	t.Helper()
	st, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	env := newFakeEnv()
	return New(Options{Env: env, Store: st}), env
}

func runScan(t *testing.T, e *Engine, kind string) model.ScanResult {
	t.Helper()
	id, err := e.StartScan(kind)
	if err != nil {
		t.Fatalf("StartScan: %v", err)
	}
	sc := e.Scan(id)
	deadline := time.After(30 * time.Second)
	for {
		snap, ch := sc.Watch()
		if snap.Done {
			if snap.Error != "" {
				t.Fatalf("scan error: %s", snap.Error)
			}
			return *snap.Result
		}
		select {
		case <-ch:
		case <-deadline:
			t.Fatal("scan timed out")
		}
	}
}

func byIP(t *testing.T, e *Engine, ip string) model.Device {
	t.Helper()
	ds, err := e.Devices(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range ds {
		if d.IP == ip {
			return d
		}
	}
	t.Fatalf("no device at %s in %d devices", ip, len(ds))
	return model.Device{}
}

func eventTypes(t *testing.T, e *Engine) map[string]int {
	t.Helper()
	evs, err := e.Events(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	m := map[string]int{}
	for _, ev := range evs {
		m[ev.Type]++
	}
	return m
}

func TestFirstScanDiscoversEverything(t *testing.T) {
	e, _ := newTestEngine(t)
	res := runScan(t, e, "quick")

	// gateway + two hosts + this PC
	if res.DevicesFound != 4 || res.NewDevices != 4 {
		t.Fatalf("found=%d new=%d, want 4/4", res.DevicesFound, res.NewDevices)
	}
	if res.TotalAddresses != 253 { // /24 minus network, broadcast and ourselves
		t.Errorf("total=%d, want 253", res.TotalAddresses)
	}
	gw := byIP(t, e, "192.168.1.1")
	if gw.Type != model.TypeRouter || gw.Status != model.StatusOnline || !gw.IsNew {
		t.Errorf("gateway = %+v", gw)
	}
	if gw.LatencyMs == nil || *gw.LatencyMs != 2 {
		t.Errorf("gateway latency = %v", gw.LatencyMs)
	}
	mbp := byIP(t, e, "192.168.1.12")
	if mbp.Hostname != "adrian-mbp.local" || mbp.Name != "adrian-mbp" || mbp.Type != model.TypeComputer {
		t.Errorf("macbook = %+v", mbp)
	}
	if mbp.Vendor == "" || mbp.Vendor == "Unknown vendor" {
		t.Errorf("OUI lookup failed for 3c:06:30 (Apple): %q", mbp.Vendor)
	}
	self := byIP(t, e, "192.168.1.24")
	if self.Name == "" || self.Type != model.TypeComputer {
		t.Errorf("self = %+v", self)
	}

	ev := eventTypes(t, e)
	if ev[model.EvNewDevice] != 4 || ev[model.EvScan] != 1 {
		t.Errorf("events = %v", ev)
	}
}

func TestOfflineNeedsConsecutiveMisses(t *testing.T) {
	e, env := newTestEngine(t)
	runScan(t, e, "quick")

	env.remove("192.168.1.12")
	runScan(t, e, "quick")
	if d := byIP(t, e, "192.168.1.12"); d.Status != model.StatusOnline {
		t.Fatalf("one missed scan must not flip status, got %s", d.Status)
	}
	runScan(t, e, "quick")
	d := byIP(t, e, "192.168.1.12")
	if d.Status != model.StatusOffline || d.LatencyMs != nil {
		t.Fatalf("after %d misses want offline, got %+v", OfflineAfterMisses, d)
	}
	if n := eventTypes(t, e)[model.EvOffline]; n != 1 {
		t.Errorf("offline events = %d, want 1", n)
	}

	// Another scan while it is still gone must not repeat the event.
	runScan(t, e, "quick")
	if n := eventTypes(t, e)[model.EvOffline]; n != 1 {
		t.Errorf("offline event repeated: %d", n)
	}

	// It comes back.
	env.add("192.168.1.12", "3c:06:30:4a:21:8f", 5*time.Millisecond)
	runScan(t, e, "quick")
	if d := byIP(t, e, "192.168.1.12"); d.Status != model.StatusOnline {
		t.Fatalf("want online again, got %s", d.Status)
	}
	if n := eventTypes(t, e)[model.EvOnline]; n != 1 {
		t.Errorf("online events = %d, want 1", n)
	}
	hist, err := e.History(context.Background(), byIP(t, e, "192.168.1.12").ID)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, h := range hist {
		got = append(got, h.Type)
	}
	want := []string{model.HistOnline, model.HistOffline, model.HistDiscovered} // newest first
	if len(got) != 3 || got[0] != want[0] || got[1] != want[1] || got[2] != want[2] {
		t.Errorf("history = %v, want %v", got, want)
	}
}

func TestUserEditsSurviveRescans(t *testing.T) {
	e, _ := newTestEngine(t)
	runScan(t, e, "quick")
	d := byIP(t, e, "192.168.1.18")
	if !d.IsNew {
		t.Fatal("expected new flag")
	}
	alias, notes, no := "Living Room TV", "bought 2024", false
	if _, err := e.UpdateDevice(context.Background(), d.ID, model.DevicePatch{CustomAlias: &alias, Notes: &notes, IsNew: &no}); err != nil {
		t.Fatal(err)
	}
	runScan(t, e, "quick")
	d = byIP(t, e, "192.168.1.18")
	if d.CustomAlias != alias || d.Notes != notes || d.IsNew {
		t.Errorf("edits lost after rescan: %+v", d)
	}
	if _, err := e.UpdateDevice(context.Background(), "dev-nope", model.DevicePatch{Notes: &notes}); !errors.Is(err, ErrNotFound) {
		t.Errorf("unknown id: err=%v, want ErrNotFound", err)
	}
	long := string(make([]byte, 81))
	if _, err := e.UpdateDevice(context.Background(), d.ID, model.DevicePatch{CustomAlias: &long}); !errors.Is(err, ErrInvalid) {
		t.Errorf("long alias: err=%v, want ErrInvalid", err)
	}
}

func TestFullScanDetectsServicesAndClassifies(t *testing.T) {
	e, _ := newTestEngine(t)
	runScan(t, e, "quick")
	if d := byIP(t, e, "192.168.1.18"); d.Type != model.TypeUnknown {
		t.Fatalf("with no evidence the type must stay Unknown, got %s", d.Type)
	}
	runScan(t, e, "full")
	d := byIP(t, e, "192.168.1.18")
	if len(d.Services) != 2 || d.Services[0].Port != 8001 || d.Services[0].Status != "Open" {
		t.Errorf("services = %+v", d.Services)
	}
	if d.Type != model.TypeTV {
		t.Errorf("Samsung TV API ports should classify as TV, got %s", d.Type)
	}
	// A following quick scan must not forget services or type.
	runScan(t, e, "quick")
	d = byIP(t, e, "192.168.1.18")
	if len(d.Services) != 2 || d.Type != model.TypeTV {
		t.Errorf("quick scan clobbered port data: %+v", d)
	}
}

func TestPerDevicePortScanReportsChanges(t *testing.T) {
	e, env := newTestEngine(t)
	runScan(t, e, "quick")
	d := byIP(t, e, "192.168.1.12")
	svcs, err := e.ScanDevicePorts(context.Background(), d.ID)
	if err != nil || len(svcs) != 0 {
		t.Fatalf("svcs=%v err=%v", svcs, err)
	}
	env.ports[netip.MustParseAddr("192.168.1.12")] = []int{22, 5000}
	svcs, err = e.ScanDevicePorts(context.Background(), d.ID)
	if err != nil || len(svcs) != 2 {
		t.Fatalf("svcs=%v err=%v", svcs, err)
	}
	if n := eventTypes(t, e)[model.EvServiceChange]; n != 1 {
		t.Errorf("service_change events = %d, want 1", n)
	}
	// Probing must not touch presence state.
	after := byIP(t, e, "192.168.1.12")
	if after.Status != model.StatusOnline || after.LatencyMs == nil {
		t.Errorf("port scan altered presence: %+v", after)
	}
}

func TestARPOnlyHostIsInventoried(t *testing.T) {
	e, env := newTestEngine(t)
	env.add("192.168.1.37", "64:90:c1:28:fe:84", 0)
	env.arpOnly[netip.MustParseAddr("192.168.1.37")] = true // drops ICMP
	runScan(t, e, "quick")
	d := byIP(t, e, "192.168.1.37")
	if d.Status != model.StatusOnline || d.LatencyMs != nil {
		t.Errorf("arp-only host: %+v", d)
	}
}

func TestNetworksAreKeptSeparate(t *testing.T) {
	e, env := newTestEngine(t)
	if err := e.checkNetworkChange(context.Background()); err != nil {
		t.Fatal(err)
	}
	runScan(t, e, "quick")
	home, _ := e.Devices(context.Background())

	// Same subnet numbers, different gateway hardware = different network.
	env.mu.Lock()
	env.arpTable[netip.MustParseAddr("192.168.1.1")] = "00:11:22:33:44:55"
	env.hosts[netip.MustParseAddr("192.168.1.1")] = fakeHost{"00:11:22:33:44:55", time.Millisecond}
	env.mu.Unlock()
	e.infoMu.Lock()
	e.infoAt = time.Time{}
	e.infoMu.Unlock()
	if err := e.checkNetworkChange(context.Background()); err != nil {
		t.Fatal(err)
	}
	office, _ := e.Devices(context.Background())
	if len(office) != 0 {
		t.Fatalf("new network should start with an empty inventory, got %d devices", len(office))
	}
	if n := eventTypes(t, e)[model.EvNetworkChange]; n != 1 {
		t.Errorf("network_change events = %d, want 1", n)
	}
	runScan(t, e, "quick")
	if office, _ = e.Devices(context.Background()); len(office) == 0 {
		t.Fatal("office scan found nothing")
	}

	// Back home: the original inventory is still there, untouched.
	env.mu.Lock()
	env.arpTable[netip.MustParseAddr("192.168.1.1")] = "e8:48:b8:31:7a:01"
	env.hosts[netip.MustParseAddr("192.168.1.1")] = fakeHost{"e8:48:b8:31:7a:01", 2 * time.Millisecond}
	env.mu.Unlock()
	e.infoMu.Lock()
	e.infoAt = time.Time{}
	e.infoMu.Unlock()
	again, _ := e.Devices(context.Background())
	if len(again) != len(home) {
		t.Errorf("home inventory changed: %d -> %d devices", len(home), len(again))
	}
}

func TestSecondScanWhileRunningAttaches(t *testing.T) {
	e, _ := newTestEngine(t)
	id1, err := e.StartScan("quick")
	if err != nil {
		t.Fatal(err)
	}
	id2, err := e.StartScan("quick")
	if !errors.Is(err, ErrScanRunning) || id1 != id2 {
		// The scan can legitimately finish before the second call on a fast machine.
		if snap, _ := e.Scan(id1).Watch(); !snap.Done {
			t.Fatalf("id1=%s id2=%s err=%v", id1, id2, err)
		}
	}
	if _, err := e.StartScan("bogus"); !errors.Is(err, ErrInvalid) {
		t.Errorf("bogus scan type: %v", err)
	}
}

func TestPingRefusesPublicAddresses(t *testing.T) {
	e, _ := newTestEngine(t)
	for _, ip := range []string{"8.8.8.8", "1.1.1.1", "not-an-ip", "::1", "example.com"} {
		if _, err := e.Ping(context.Background(), ip); !errors.Is(err, ErrInvalid) {
			t.Errorf("Ping(%q) err=%v, want ErrInvalid", ip, err)
		}
	}
	r, err := e.Ping(context.Background(), "192.168.1.12")
	if err != nil || !r.Success || r.LatencyMs != 3 {
		t.Errorf("ping private host: %+v err=%v", r, err)
	}
	if r, _ := e.Ping(context.Background(), "192.168.1.200"); r.Success {
		t.Error("absent host reported success")
	}
}

func TestNetworkInfo(t *testing.T) {
	e, _ := newTestEngine(t)
	runScan(t, e, "quick")
	info, err := e.NetworkInfo(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if info.Subnet != "192.168.1.0/24" || info.Gateway != "192.168.1.1" || info.Netmask != "255.255.255.0" ||
		info.Broadcast != "192.168.1.255" || info.TotalAddresses != 254 || info.SSID != "Home-5G" ||
		info.NetworkName != "Home-5G" || info.InterfaceType != "Wi-Fi" || info.LocalIP != "192.168.1.24" {
		t.Errorf("info = %+v", info)
	}
	if info.ActiveAddresses != 4 {
		t.Errorf("active = %d, want 4", info.ActiveAddresses)
	}
	if len(info.Interfaces) != 2 || !info.Interfaces[0].IsDefault || info.Interfaces[0].Status != "Connected" ||
		info.Interfaces[1].Status != "Disconnected" {
		t.Errorf("interfaces = %+v", info.Interfaces)
	}
	if info.DNS == nil || info.PingStats.History == nil {
		t.Error("slices must serialize as [] not null")
	}
}
