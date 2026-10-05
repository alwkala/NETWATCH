package engine

import (
	"fmt"
	"net/netip"
	"sort"
	"strings"
	"time"

	"netwatch/internal/fingerprint"
	"netwatch/internal/model"
	"netwatch/internal/oui"
	"netwatch/internal/store"
)

// observation is what one scan learned about one host.
type observation struct {
	IP       netip.Addr
	MAC      string
	Hostname string
	RTT      time.Duration
	HasRTT   bool
	Ports    []int // open TCP ports; meaningful only when PortScanned
	// PortScanned is true when this scan probed ports (Full scans).
	PortScanned bool
	IsSelf      bool
}

type reconcileIn struct {
	NetKey  string
	Now     time.Time
	Gateway netip.Addr
	Known   []store.Known
	Obs     []observation
	Vendor  func(mac string) string
}

type reconcileOut struct {
	WS    store.Writeset
	Found int
	New   int
}

// reconcile diffs a scan against the stored inventory. It is pure: no I/O,
// no clock, no randomness, so every rule below is unit-tested.
func reconcile(in reconcileIn) reconcileOut {
	known := make(map[string]*store.Known, len(in.Known))
	for i := range in.Known {
		known[in.Known[i].ID] = &in.Known[i]
	}
	var out reconcileOut
	seen := map[string]bool{}
	ws := &out.WS
	ws.Services = map[string][]model.DeviceService{}

	// Stable order -> stable event order.
	obs := append([]observation(nil), in.Obs...)
	sort.Slice(obs, func(i, j int) bool { return obs[i].IP.Less(obs[j].IP) })

	for _, o := range obs {
		id := deviceID(in.NetKey, o.MAC)
		if seen[id] {
			continue // duplicate ARP row for the same MAC
		}
		seen[id] = true
		out.Found++

		vendor := in.Vendor(o.MAC)
		prev, existed := known[id]

		ports := o.Ports
		if !o.PortScanned && existed {
			for _, sv := range prev.Services {
				ports = append(ports, sv.Port)
			}
		}
		ev := fingerprint.Evidence{
			Vendor: vendor, Hostname: o.Hostname, MAC: o.MAC, IP: o.IP.String(),
			IsGateway: o.IP == in.Gateway, OpenPorts: ports,
		}
		typ := fingerprint.Classify(ev)
		if o.IsSelf && typ == model.TypeUnknown {
			typ = model.TypeComputer
		}
		name := fingerprint.Name(ev, typ)
		if o.IsSelf {
			name += " (This PC)"
		}
		var lat *int
		if o.HasRTT {
			v := max(1, int((o.RTT + 500*time.Microsecond).Milliseconds()))
			lat = &v
		}

		k := store.Known{Net: in.NetKey}
		k.ID, k.MAC, k.IP = id, o.MAC, o.IP.String()
		k.Hostname, k.Vendor = o.Hostname, fingerprint.DisplayVendor(vendor, o.MAC)
		k.Type, k.Name, k.Status = typ, name, model.StatusOnline
		k.LatencyMs, k.LastSeen, k.Missed = lat, in.Now, 0

		detail := "Seen in ARP table"
		if lat != nil {
			detail = fmt.Sprintf("Responded to ICMP echo (%d ms)", *lat)
		}

		if !existed {
			k.FirstSeen, k.IsNew = in.Now, true
			out.New++
			ws.Events = append(ws.Events, model.NetworkEvent{
				Timestamp: in.Now, Type: model.EvNewDevice, Title: "New device discovered",
				DeviceName: k.Name, DeviceID: id, IP: k.IP, MAC: strings.ToUpper(k.MAC),
				Details: fmt.Sprintf("New host on the subnet (%s)", k.Vendor),
			})
			ws.History = append(ws.History, store.HistoryRow{DeviceID: id, Time: in.Now, Type: model.HistDiscovered,
				Desc: "First discovered at " + k.IP})
		} else {
			// Preserve user-owned fields and identity continuity.
			k.CustomAlias, k.Notes, k.OS = prev.CustomAlias, prev.Notes, prev.OS
			k.FirstSeen, k.IsNew = prev.FirstSeen, prev.IsNew
			if k.Type == model.TypeUnknown && prev.Type != model.TypeUnknown {
				k.Type = prev.Type // never forget a classification because this scan saw fewer ports
			}
			if k.Hostname == "" {
				k.Hostname = prev.Hostname
			}
			if prev.Status == model.StatusOffline {
				ws.Events = append(ws.Events, model.NetworkEvent{
					Timestamp: in.Now, Type: model.EvOnline, Title: "Device came online",
					DeviceName: displayName(k), DeviceID: id, IP: k.IP, MAC: strings.ToUpper(k.MAC), Details: detail,
				})
				ws.History = append(ws.History, store.HistoryRow{DeviceID: id, Time: in.Now, Type: model.HistOnline, Desc: "Device came online"})
			}
			if prev.IP != k.IP {
				ws.History = append(ws.History, store.HistoryRow{DeviceID: id, Time: in.Now, Type: model.HistIPChanged,
					Desc: fmt.Sprintf("IP address changed %s → %s", prev.IP, k.IP)})
			}
		}
		ws.Devices = append(ws.Devices, k)

		// Services: only a port-scanning pass may change the stored set.
		if o.PortScanned {
			svcs := make([]model.DeviceService, 0, len(o.Ports))
			ps := append([]int(nil), o.Ports...)
			sort.Ints(ps)
			for _, p := range ps {
				svcs = append(svcs, model.DeviceService{Port: p, Protocol: "TCP", Service: fingerprint.ServiceName(p), Status: "Open"})
			}
			ws.Services[id] = svcs
			addedNow := newPorts(prev, svcs)
			for _, sv := range addedNow {
				ws.History = append(ws.History, store.HistoryRow{DeviceID: id, Time: in.Now, Type: model.HistServiceDetected,
					Desc: fmt.Sprintf("Open port %d/%s (%s)", sv.Port, strings.ToLower(sv.Protocol), sv.Service)})
			}
			if existed && (len(addedNow) > 0 || closedPorts(prev, svcs) > 0) {
				ws.Events = append(ws.Events, model.NetworkEvent{
					Timestamp: in.Now, Type: model.EvServiceChange, Title: "Exposed services changed",
					DeviceName: displayName(k), DeviceID: id, IP: k.IP, MAC: strings.ToUpper(k.MAC),
					Details: fmt.Sprintf("%d open port(s) now, was %d", len(svcs), len(prev.Services)),
				})
			}
		}
	}

	// Devices of this network that did not answer.
	for _, prev := range in.Known {
		if seen[prev.ID] {
			continue
		}
		k := prev
		if prev.Status == model.StatusOnline {
			k.Missed++
			if k.Missed >= OfflineAfterMisses {
				k.Status, k.LatencyMs = model.StatusOffline, nil
				ws.Events = append(ws.Events, model.NetworkEvent{
					Timestamp: in.Now, Type: model.EvOffline, Title: "Device went offline",
					DeviceName: displayName(k), DeviceID: k.ID, IP: k.IP, MAC: strings.ToUpper(k.MAC),
					Details: fmt.Sprintf("No response in %d consecutive scans", k.Missed),
				})
				ws.History = append(ws.History, store.HistoryRow{DeviceID: k.ID, Time: in.Now, Type: model.HistOffline, Desc: "Device went offline"})
			}
			ws.Devices = append(ws.Devices, k)
		}
	}
	return out
}

func displayName(k store.Known) string {
	if k.CustomAlias != "" {
		return k.CustomAlias
	}
	return k.Name
}

func newPorts(prev *store.Known, now []model.DeviceService) []model.DeviceService {
	old := map[int]bool{}
	if prev != nil {
		for _, s := range prev.Services {
			old[s.Port] = true
		}
	}
	var out []model.DeviceService
	for _, s := range now {
		if !old[s.Port] {
			out = append(out, s)
		}
	}
	return out
}

func closedPorts(prev *store.Known, now []model.DeviceService) int {
	if prev == nil {
		return 0
	}
	cur := map[int]bool{}
	for _, s := range now {
		cur[s.Port] = true
	}
	n := 0
	for _, s := range prev.Services {
		if !cur[s.Port] {
			n++
		}
	}
	return n
}

// vendorFunc adapts the OUI database to the reconcile input.
func vendorFunc(db *oui.DB) func(string) string {
	return func(mac string) string {
		v, _ := db.Lookup(mac)
		return v
	}
}
