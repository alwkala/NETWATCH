package store_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"netwatch/internal/model"
	"netwatch/internal/store"
)

func TestStore_OpenAndMigrate(t *testing.T) {
	ctx := context.Background()
	// Test memory mode
	sMem, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open(:memory:) failed: %v", err)
	}
	defer sMem.Close()

	// Test file mode in temporary directory
	dbPath := filepath.Join(t.TempDir(), "test.db")
	sFile, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("Open(%q) failed: %v", dbPath, err)
	}
	defer sFile.Close()

	ok, err := sFile.IntegrityCheck(ctx)
	if err != nil || ok != "ok" {
		t.Fatalf("IntegrityCheck: got (%v, %v), want 'ok', nil", ok, err)
	}
}

func TestStore_DevicesAndReconciliation(t *testing.T) {
	ctx := context.Background()
	st, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	now := time.Now().UTC().Truncate(time.Millisecond)
	netKey := "192.168.1.0/24:00:11:22:33:44:55"

	// 1. Commit initial device
	dev1 := store.Known{
		Device: model.Device{
			ID:        "dev-1",
			MAC:       "00:11:22:33:44:55",
			IP:        "192.168.1.50",
			IPv6:      "fe80::1",
			Hostname:  "test-pc",
			Vendor:    "Intel",
			Type:      "Computer",
			Status:    model.StatusOnline,
			FirstSeen: now,
			LastSeen:  now,
			IsNew:     true,
		},
		Net:    netKey,
		Missed: 0,
	}

	ws := store.Writeset{
		Devices: []store.Known{dev1},
		Events: []model.NetworkEvent{
			{
				ID:         "ev-1",
				Timestamp:  now,
				Type:       model.EvNewDevice,
				Title:      "New Device Detected",
				DeviceID:   "dev-1",
				DeviceName: "test-pc",
				IP:         "192.168.1.50",
				MAC:        "00:11:22:33:44:55",
			},
		},
		History: []store.HistoryRow{
			{
				DeviceID: "dev-1",
				Time:     now,
				Type:     model.HistDiscovered,
				Desc:     "Discovered on network",
			},
		},
		Services: map[string][]model.DeviceService{
			"dev-1": {
				{Port: 80, Protocol: "TCP", Service: "HTTP", Status: "Open"},
			},
		},
	}

	if err := st.Commit(ctx, ws); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	// 2. Query known devices
	known, err := st.ListKnown(ctx, netKey)
	if err != nil {
		t.Fatalf("ListKnown failed: %v", err)
	}
	if len(known) != 1 || known[0].ID != "dev-1" || known[0].IPv6 != "fe80::1" {
		t.Fatalf("unexpected known devices: %+v", known)
	}
	if len(known[0].Services) != 1 || known[0].Services[0].Port != 80 {
		t.Fatalf("services not preserved: %+v", known[0].Services)
	}

	// 3. Patch device
	customAlias := "My Workstation"
	patch := model.DevicePatch{
		CustomAlias: &customAlias,
		IsNew:       func(b bool) *bool { return &b }(false),
	}
	if err := st.UpdateDevice(ctx, "dev-1", patch); err != nil {
		t.Fatalf("UpdateDevice failed: %v", err)
	}

	// 4. Retrieve single device
	d, err := st.GetDevice(ctx, "dev-1")
	if err != nil {
		t.Fatalf("GetDevice failed: %v", err)
	}
	if d.CustomAlias != customAlias || d.IsNew {
		t.Fatalf("expected custom alias %q and IsNew=false, got %+v", customAlias, d)
	}

	// 5. Check history and events
	history, err := st.DeviceHistory(ctx, "dev-1", 10)
	if err != nil {
		t.Fatalf("DeviceHistory failed: %v", err)
	}
	if len(history) != 1 || history[0].Type != model.HistDiscovered {
		t.Fatalf("unexpected device history: %+v", history)
	}

	events, err := st.Events(ctx, 10)
	if err != nil {
		t.Fatalf("Events failed: %v", err)
	}
	if len(events) != 1 || events[0].Type != model.EvNewDevice {
		t.Fatalf("unexpected events: %+v", events)
	}
}

func TestStore_Settings(t *testing.T) {
	ctx := context.Background()
	st, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	// Default settings when unset
	defs, err := st.GetSettings(ctx)
	if err != nil {
		t.Fatalf("GetSettings (defaults) failed: %v", err)
	}
	if defs.ScanInterval != "5m" || !defs.NotifyNewDevice {
		t.Fatalf("unexpected default settings: %+v", defs)
	}

	// Save custom settings
	custom := defs
	custom.ScanInterval = "15m"
	custom.NotifyNewDevice = false
	custom.StartMinimized = true

	if err := st.SaveSettings(ctx, custom); err != nil {
		t.Fatalf("SaveSettings failed: %v", err)
	}

	saved, err := st.GetSettings(ctx)
	if err != nil {
		t.Fatalf("GetSettings failed: %v", err)
	}
	if saved.ScanInterval != "15m" || saved.NotifyNewDevice != false || !saved.StartMinimized {
		t.Fatalf("saved settings mismatch: %+v", saved)
	}
}

func TestStore_StatsVacuumIntegrity(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "metrics.db")
	st, err := store.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	stats, err := st.Stats(ctx, dbPath)
	if err != nil {
		t.Fatalf("Stats failed: %v", err)
	}
	if stats.DeviceCount != 0 || stats.FileSizeBytes == 0 {
		t.Fatalf("unexpected stats: %+v", stats)
	}

	// Check Vacuum
	if err := st.Vacuum(ctx); err != nil {
		t.Fatalf("Vacuum failed: %v", err)
	}

	// Check Integrity
	status, err := st.IntegrityCheck(ctx)
	if err != nil || status != "ok" {
		t.Fatalf("IntegrityCheck failed: status=%s, err=%v", status, err)
	}
}

func TestStore_ClearHistory(t *testing.T) {
	ctx := context.Background()
	st, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	now := time.Now().UTC()
	ws := store.Writeset{
		Devices: []store.Known{
			{Device: model.Device{ID: "d1", MAC: "00:00:00:00:00:01", IP: "10.0.0.1", FirstSeen: now, LastSeen: now}},
		},
		Events: []model.NetworkEvent{
			{ID: "e1", Timestamp: now, Type: model.EvNewDevice, Title: "test"},
		},
	}
	if err := st.Commit(ctx, ws); err != nil {
		t.Fatal(err)
	}

	// Clear
	if err := st.ClearHistory(ctx); err != nil {
		t.Fatalf("ClearHistory failed: %v", err)
	}

	stats, err := st.Stats(ctx, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	if stats.DeviceCount != 0 || stats.EventCount != 0 {
		t.Fatalf("tables not emptied: %+v", stats)
	}

	clearedAt, err := st.Meta(ctx, "history_cleared_at")
	if err != nil || clearedAt == "" {
		t.Fatalf("expected history_cleared_at in meta table, got %q, err=%v", clearedAt, err)
	}
}

func TestStore_PruneEvents(t *testing.T) {
	ctx := context.Background()
	st, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	oldTime := time.Now().UTC().AddDate(0, 0, -40)
	recentTime := time.Now().UTC().AddDate(0, 0, -5)

	ws := store.Writeset{
		Devices: []store.Known{
			{Device: model.Device{ID: "d1", MAC: "00:11:22:33:44:55", IP: "192.168.1.50", FirstSeen: oldTime, LastSeen: recentTime}},
		},
		Events: []model.NetworkEvent{
			{ID: "e-old", Timestamp: oldTime, Type: model.EvNewDevice, Title: "old event"},
			{ID: "e-recent", Timestamp: recentTime, Type: model.EvNewDevice, Title: "recent event"},
		},
		History: []store.HistoryRow{
			{DeviceID: "d1", Time: oldTime, Type: model.HistDiscovered, Desc: "old history"},
			{DeviceID: "d1", Time: recentTime, Type: model.HistOnline, Desc: "recent history"},
		},
	}
	if err := st.Commit(ctx, ws); err != nil {
		t.Fatal(err)
	}

	deleted, err := st.PruneEvents(ctx, 30)
	if err != nil {
		t.Fatalf("PruneEvents failed: %v", err)
	}
	if deleted != 2 {
		t.Fatalf("expected 2 pruned items (1 event + 1 history), got %d", deleted)
	}

	evs, err := st.Events(ctx, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(evs) != 1 || evs[0].Title != "recent event" {
		t.Fatalf("expected only recent event left, got: %+v", evs)
	}
}

func TestStore_MergeDevicesAndTrustStatus(t *testing.T) {
	ctx := context.Background()
	st, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	now := time.Now().UTC()
	ws := store.Writeset{
		Devices: []store.Known{
			{Device: model.Device{ID: "dev-target", MAC: "00:11:22:33:44:55", IP: "192.168.1.10", Name: "My Phone", TrustStatus: model.TrustKnown, FirstSeen: now, LastSeen: now}},
			{Device: model.Device{ID: "dev-source", MAC: "02:AA:BB:CC:DD:EE", IP: "192.168.1.11", Name: "Private MAC Phone", IsRandomizedMAC: true, TrustStatus: model.TrustUnknown, FirstSeen: now, LastSeen: now}},
		},
		History: []store.HistoryRow{
			{DeviceID: "dev-target", Time: now, Type: model.HistDiscovered, Desc: "Target first discovered"},
			{DeviceID: "dev-source", Time: now, Type: model.HistDiscovered, Desc: "Source first discovered"},
		},
	}
	if err := st.Commit(ctx, ws); err != nil {
		t.Fatal(err)
	}

	// 1. Verify TrustStatus update
	guest := model.TrustGuest
	if err := st.UpdateDevice(ctx, "dev-target", model.DevicePatch{TrustStatus: &guest}); err != nil {
		t.Fatalf("UpdateDevice trustStatus: %v", err)
	}
	d, err := st.GetDevice(ctx, "dev-target")
	if err != nil {
		t.Fatal(err)
	}
	if d.TrustStatus != model.TrustGuest {
		t.Fatalf("expected TrustStatus to be %q, got %q", model.TrustGuest, d.TrustStatus)
	}

	// 2. Perform merge
	if err := st.MergeDevices(ctx, "dev-target", "dev-source"); err != nil {
		t.Fatalf("MergeDevices: %v", err)
	}

	// Source should be deleted
	_, err = st.GetDevice(ctx, "dev-source")
	if err != store.ErrNotFound {
		t.Fatalf("expected source device to be deleted, got err: %v", err)
	}

	// Target history should contain merged events
	hist, err := st.DeviceHistory(ctx, "dev-target", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(hist) != 3 { // target discovered + source discovered + device_merged
		t.Fatalf("expected 3 history items after merge, got %d", len(hist))
	}

	// MACAliases should map source MAC to target
	aliases, err := st.MACAliases(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if aliases["02:AA:BB:CC:DD:EE"] != "dev-target" {
		t.Fatalf("expected alias 02:AA:BB:CC:DD:EE -> dev-target, got %q", aliases["02:AA:BB:CC:DD:EE"])
	}
}

func TestStore_EvidenceAndNetworkContext(t *testing.T) {
	ctx := context.Background()
	st, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	now := time.Now().UTC()
	ws := store.Writeset{
		Devices: []store.Known{
			{Device: model.Device{ID: "dev-test", MAC: "00:11:22:33:44:55", IP: "192.168.1.50", FirstSeen: now, LastSeen: now}},
		},
	}
	if err := st.Commit(ctx, ws); err != nil {
		t.Fatal(err)
	}

	// 1. Evidence persistence
	evItems := []model.DiscoveryEvidence{
		{Source: model.SourceMDNS, Key: "hostname", Value: "MyHost.local", ObservedAt: now, LastSeen: now},
		{Source: model.SourceSSDP, Key: "st", Value: "urn:schemas-upnp-org:device:MediaRenderer:1", ObservedAt: now, LastSeen: now},
	}
	if err := st.SaveEvidence(ctx, "dev-test", evItems); err != nil {
		t.Fatalf("SaveEvidence failed: %v", err)
	}

	retrieved, err := st.GetDeviceEvidence(ctx, "dev-test")
	if err != nil {
		t.Fatalf("GetDeviceEvidence failed: %v", err)
	}
	if len(retrieved) != 2 {
		t.Fatalf("expected 2 evidence items, got %d", len(retrieved))
	}

	// 2. Network Context persistence
	nc := model.NetworkContext{
		Interface:  "Wi-Fi",
		SSID:       "HomeNet-5G",
		BSSID:      "aa:bb:cc:dd:ee:ff",
		Gateway:    "192.168.1.1",
		Subnet:     "192.168.1.0/24",
		IPv4:       "192.168.1.100",
		ObservedAt: now,
	}
	if err := st.SaveNetworkContext(ctx, nc); err != nil {
		t.Fatalf("SaveNetworkContext failed: %v", err)
	}

	latestNC, err := st.GetLatestNetworkContext(ctx)
	if err != nil {
		t.Fatalf("GetLatestNetworkContext failed: %v", err)
	}
	if latestNC == nil || latestNC.SSID != "HomeNet-5G" {
		t.Fatalf("unexpected latest network context: %+v", latestNC)
	}
}

