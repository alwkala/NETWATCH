package engine

import (
	"net/netip"
	"testing"
	"time"

	"netwatch/internal/model"
	"netwatch/internal/store"
)

func vendorOf(string) string { return "Acme Corp" }

func TestReconcileClassifiesAndNamesDevices(t *testing.T) {
	now := time.Unix(1_800_000_000, 0).UTC()
	out := reconcile(reconcileIn{
		NetKey: "n", Now: now, Gateway: netip.MustParseAddr("10.0.0.1"), Vendor: vendorOf,
		Obs: []observation{
			{IP: netip.MustParseAddr("10.0.0.1"), MAC: "02:00:00:00:00:01"},
			{IP: netip.MustParseAddr("10.0.0.2"), MAC: "02:00:00:00:00:02", Ports: []int{9100}, PortScanned: true},
			{IP: netip.MustParseAddr("10.0.0.3"), MAC: "02:00:00:00:00:03"},
			{IP: netip.MustParseAddr("10.0.0.3"), MAC: "02:00:00:00:00:03"}, // duplicate ARP row
		},
	})
	if out.Found != 3 || out.New != 3 {
		t.Fatalf("found=%d new=%d", out.Found, out.New)
	}
	types := map[string]string{}
	for _, d := range out.WS.Devices {
		types[d.IP] = d.Type
	}
	if types["10.0.0.1"] != model.TypeRouter || types["10.0.0.2"] != model.TypePrinter || types["10.0.0.3"] != model.TypeUnknown {
		t.Errorf("types = %v", types)
	}
}

func TestReconcileDoesNotDowngradeKnownType(t *testing.T) {
	now := time.Unix(1_800_000_000, 0).UTC()
	id := deviceID("n", "aa:bb:cc:00:00:01")
	prev := store.Known{Net: "n"}
	prev.ID, prev.MAC, prev.IP, prev.Type, prev.Status, prev.FirstSeen = id, "aa:bb:cc:00:00:01", "10.0.0.5", model.TypeTV, model.StatusOnline, now.Add(-time.Hour)
	out := reconcile(reconcileIn{
		NetKey: "n", Now: now, Vendor: vendorOf, Known: []store.Known{prev},
		Obs: []observation{{IP: netip.MustParseAddr("10.0.0.5"), MAC: "aa:bb:cc:00:00:01"}},
	})
	if got := out.WS.Devices[0].Type; got != model.TypeTV {
		t.Errorf("type = %s, want TV kept", got)
	}
	if out.New != 0 {
		t.Errorf("existing device counted as new")
	}
}

func TestReconcileRecordsIPChange(t *testing.T) {
	now := time.Unix(1_800_000_000, 0).UTC()
	id := deviceID("n", "aa:bb:cc:00:00:01")
	prev := store.Known{Net: "n"}
	prev.ID, prev.MAC, prev.IP, prev.Status, prev.FirstSeen = id, "aa:bb:cc:00:00:01", "10.0.0.5", model.StatusOnline, now
	out := reconcile(reconcileIn{
		NetKey: "n", Now: now, Vendor: vendorOf, Known: []store.Known{prev},
		Obs: []observation{{IP: netip.MustParseAddr("10.0.0.9"), MAC: "aa:bb:cc:00:00:01"}},
	})
	var found bool
	for _, h := range out.WS.History {
		if h.Type == model.HistIPChanged {
			found = true
		}
	}
	if !found {
		t.Error("ip_changed history row missing")
	}
}

func TestReconcileLeavesOfflineDevicesAlone(t *testing.T) {
	now := time.Unix(1_800_000_000, 0).UTC()
	prev := store.Known{Net: "n"}
	prev.ID, prev.Status = "dev-x", model.StatusOffline
	out := reconcile(reconcileIn{NetKey: "n", Now: now, Vendor: vendorOf, Known: []store.Known{prev}})
	if len(out.WS.Devices) != 0 || len(out.WS.Events) != 0 {
		t.Errorf("offline device generated writes: %+v", out.WS)
	}
}
