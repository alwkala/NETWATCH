// Package engine is the local network intelligence core: it sweeps the
// subnet, enriches what it finds, diffs the result against the inventory in
// SQLite and produces the events the UI shows.
package engine

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/netip"
	"sort"
	"strings"
	"sync"
	"time"

	"netwatch/internal/model"
	"netwatch/internal/netenv"
	"netwatch/internal/oui"
	"netwatch/internal/store"
)

var (
	ErrNoNetwork = errors.New("no active network interface with a default gateway")
	ErrNotFound  = store.ErrNotFound
	ErrInvalid   = errors.New("invalid request")
)

// OfflineAfterMisses is how many consecutive scans a device may be absent
// before it is marked offline. Phones and laptops routinely miss a single
// sweep while asleep; requiring two avoids online/offline flapping.
const OfflineAfterMisses = 2

type Options struct {
	Env    netenv.Env
	Store  *store.Store
	OUI    *oui.DB
	Logger *slog.Logger
	Now    func() time.Time
}

type Engine struct {
	env netenv.Env
	st  *store.Store
	oui *oui.DB
	log *slog.Logger
	now func() time.Time

	commitMu sync.Mutex // serializes inventory writes

	infoMu    sync.Mutex
	infoAt    time.Time
	infoCache netenv.Info

	scanMu sync.Mutex
	scan   *scanState

	mon monitor
}

func New(o Options) *Engine {
	if o.Logger == nil {
		o.Logger = slog.Default()
	}
	if o.Now == nil {
		o.Now = func() time.Time { return time.Now().UTC() }
	}
	if o.OUI == nil {
		o.OUI = oui.Default()
	}
	return &Engine{env: o.Env, st: o.Store, oui: o.OUI, log: o.Logger, now: o.Now}
}

// Start launches background work (gateway latency monitor). It stops when ctx ends.
func (e *Engine) Start(ctx context.Context) {
	go e.runMonitor(ctx)
	if err := e.checkNetworkChange(ctx); err != nil {
		e.log.Warn("network change check", "err", err)
	}
}

// info returns the OS network configuration, cached briefly because several
// API calls ask for it in quick succession.
func (e *Engine) info(ctx context.Context) (netenv.Info, error) {
	e.infoMu.Lock()
	defer e.infoMu.Unlock()
	if !e.infoAt.IsZero() && e.now().Sub(e.infoAt) < 2*time.Second {
		return e.infoCache, nil
	}
	in, err := e.env.Info(ctx)
	if err != nil {
		return netenv.Info{}, err
	}
	e.infoCache, e.infoAt = in, e.now()
	return in, nil
}

type activeNet struct {
	Adapter netenv.Adapter
	Info    netenv.Info
	Key     string // identifies the network, see netKey
}

// active resolves the default-route adapter and the key of its network.
func (e *Engine) active(ctx context.Context) (activeNet, error) {
	in, err := e.info(ctx)
	if err != nil {
		return activeNet{}, err
	}
	ad, ok := in.Active()
	if !ok {
		return activeNet{}, ErrNoNetwork
	}
	var gwMAC string
	if nb, err := e.env.Neighbors(ctx); err == nil {
		for _, n := range nb {
			if n.IP == ad.Gateway {
				gwMAC = n.MAC
				break
			}
		}
	}
	return activeNet{Adapter: ad, Info: in, Key: e.netKey(ctx, ad, gwMAC)}, nil
}

// netKey identifies a network by subnet + gateway MAC so a laptop moving
// between two 192.168.1.0/24 networks keeps two separate inventories. If the
// gateway MAC is momentarily unavailable the last known key for the same
// subnet and gateway is reused.
func (e *Engine) netKey(ctx context.Context, ad netenv.Adapter, gwMAC string) string {
	base := ad.IP.Masked().String() + "|" + ad.Gateway.String()
	if gwMAC != "" {
		key := base + "|" + gwMAC
		_ = e.st.SetMeta(ctx, "netkey:"+base, key)
		return key
	}
	if last, _ := e.st.Meta(ctx, "netkey:"+base); last != "" {
		return last
	}
	return base
}

func deviceID(netKey, mac string) string {
	sum := sha256.Sum256([]byte(netKey + "|" + strings.ToLower(mac)))
	return "dev-" + hex.EncodeToString(sum[:6])
}

// NetworkInfo builds the Network page / sidebar model.
func (e *Engine) NetworkInfo(ctx context.Context) (model.NetworkInfo, error) {
	out := model.NetworkInfo{
		DNS:        []string{},
		Interfaces: []model.NetworkInterfaceItem{},
		Status:     "Disconnected",
		PingStats:  e.mon.stats(),
	}
	in, err := e.info(ctx)
	if err != nil {
		return out, err
	}
	act, hasActive := in.Active()
	for _, d := range in.DNS {
		out.DNS = append(out.DNS, d.String())
	}
	for _, ad := range in.Adapters {
		item := model.NetworkInterfaceItem{
			ID:          "if-" + slug(ad.Name),
			Name:        ad.Name,
			Type:        ad.Type,
			AdapterName: ad.Desc,
			MAC:         strings.ToUpper(ad.MAC),
			Status:      "Disconnected",
			Subnet:      "—",
			Gateway:     "—",
			IP:          "0.0.0.0",
			SpeedMbps:   ad.SpeedMbps,
			IsDefault:   hasActive && ad.Name == act.Name,
		}
		if ad.Up && ad.IP.IsValid() {
			item.Status = "Connected"
			item.IP = ad.IP.Addr().String()
			item.Subnet = ad.IP.Masked().String()
			if ad.Gateway.IsValid() {
				item.Gateway = ad.Gateway.String()
			}
		}
		out.Interfaces = append(out.Interfaces, item)
	}
	sort.SliceStable(out.Interfaces, func(i, j int) bool {
		a, b := out.Interfaces[i], out.Interfaces[j]
		if a.IsDefault != b.IsDefault {
			return a.IsDefault
		}
		return a.Status == "Connected" && b.Status != "Connected"
	})
	if !hasActive {
		return out, nil
	}
	p := act.IP.Masked()
	out.Status = "Connected"
	out.SSID = in.SSID
	out.NetworkName = act.Name
	if in.SSID != "" {
		out.NetworkName = in.SSID
	}
	out.InterfaceName, out.InterfaceType = act.Name, act.Type
	out.Gateway = act.Gateway.String()
	out.Subnet = p.String()
	out.LocalIP = act.IP.Addr().String()
	out.Netmask = netmask(p.Bits())
	out.Broadcast = broadcast(p).String()
	if hostBits := 32 - p.Bits(); hostBits >= 2 && hostBits <= 30 {
		out.TotalAddresses = (1 << hostBits) - 2
	}
	if an, err := e.active(ctx); err == nil {
		if ks, err := e.st.ListKnown(ctx, an.Key); err == nil {
			for _, k := range ks {
				if k.Status == model.StatusOnline {
					out.ActiveAddresses++
				}
			}
		}
	}
	return out, nil
}

// Devices lists the inventory of the network the machine is on now.
func (e *Engine) Devices(ctx context.Context) ([]model.Device, error) {
	an, err := e.active(ctx)
	if errors.Is(err, ErrNoNetwork) {
		return []model.Device{}, nil
	} else if err != nil {
		return nil, err
	}
	ks, err := e.st.ListKnown(ctx, an.Key)
	if err != nil {
		return nil, err
	}
	out := make([]model.Device, 0, len(ks))
	for _, k := range ks {
		out = append(out, k.Device)
	}
	sort.Slice(out, func(i, j int) bool {
		a, _ := netip.ParseAddr(out[i].IP)
		b, _ := netip.ParseAddr(out[j].IP)
		return a.Less(b)
	})
	return out, nil
}

func (e *Engine) Device(ctx context.Context, id string) (*model.Device, error) {
	return e.st.GetDevice(ctx, id)
}

func (e *Engine) History(ctx context.Context, id string) ([]model.DeviceEvent, error) {
	if _, err := e.st.GetDevice(ctx, id); err != nil {
		return nil, err
	}
	return e.st.DeviceHistory(ctx, id, 200)
}

func (e *Engine) Events(ctx context.Context) ([]model.NetworkEvent, error) {
	return e.st.Events(ctx, 300)
}

func (e *Engine) UpdateDevice(ctx context.Context, id string, p model.DevicePatch) (*model.Device, error) {
	if p.CustomAlias != nil {
		a := strings.TrimSpace(*p.CustomAlias)
		if len(a) > 80 {
			return nil, fmt.Errorf("%w: alias longer than 80 characters", ErrInvalid)
		}
		p.CustomAlias = &a
	}
	if p.Notes != nil && len(*p.Notes) > 2000 {
		return nil, fmt.Errorf("%w: notes longer than 2000 characters", ErrInvalid)
	}
	e.commitMu.Lock()
	defer e.commitMu.Unlock()
	if err := e.st.UpdateDevice(ctx, id, p); err != nil {
		return nil, err
	}
	return e.st.GetDevice(ctx, id)
}

// ClearHistory wipes the inventory, events and scan log.
func (e *Engine) ClearHistory(ctx context.Context) error {
	e.commitMu.Lock()
	defer e.commitMu.Unlock()
	return e.st.ClearHistory(ctx)
}

// checkNetworkChange records a network_change event when the machine is on a
// different network than the last time NETWATCH looked.
func (e *Engine) checkNetworkChange(ctx context.Context) error {
	an, err := e.active(ctx)
	if err != nil {
		if errors.Is(err, ErrNoNetwork) {
			return nil
		}
		return err
	}
	e.commitMu.Lock()
	defer e.commitMu.Unlock()
	last, err := e.st.Meta(ctx, "last_net_key")
	if err != nil {
		return err
	}
	if last == an.Key {
		return nil
	}
	if last != "" {
		name := an.Adapter.Name
		if an.Info.SSID != "" {
			name = an.Info.SSID
		}
		err := e.st.Commit(ctx, store.Writeset{Events: []model.NetworkEvent{{
			Timestamp: e.now(),
			Type:      model.EvNetworkChange,
			Title:     "Network changed",
			Details:   fmt.Sprintf("Now on %s (%s via %s)", name, an.Adapter.IP.Masked(), an.Adapter.Name),
		}}})
		if err != nil {
			return err
		}
	}
	return e.st.SetMeta(ctx, "last_net_key", an.Key)
}

func slug(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		default:
			b.WriteByte('-')
		}
	}
	return strings.Trim(b.String(), "-")
}

func netmask(bits int) string {
	var m uint32
	if bits > 0 {
		m = ^uint32(0) << (32 - bits)
	}
	return fmt.Sprintf("%d.%d.%d.%d", byte(m>>24), byte(m>>16), byte(m>>8), byte(m))
}

func broadcast(p netip.Prefix) netip.Addr {
	a := p.Masked().Addr().As4()
	hostBits := 32 - p.Bits()
	v := uint32(a[0])<<24 | uint32(a[1])<<16 | uint32(a[2])<<8 | uint32(a[3])
	if hostBits > 0 {
		v |= (uint32(1) << hostBits) - 1
	}
	return netip.AddrFrom4([4]byte{byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v)})
}
