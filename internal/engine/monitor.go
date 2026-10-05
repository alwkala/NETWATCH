package engine

import (
	"context"
	"sync"
	"time"

	"netwatch/internal/model"
)

const (
	monitorInterval = 5 * time.Second
	monitorSamples  = 20
)

// monitor pings the default gateway on a timer so the dashboard's latency
// panel reflects real measurements (and real packet loss), not a placeholder.
type monitor struct {
	mu      sync.Mutex
	samples []int // ms; -1 = lost
}

func (m *monitor) add(ms int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.samples = append(m.samples, ms)
	if len(m.samples) > monitorSamples {
		m.samples = m.samples[len(m.samples)-monitorSamples:]
	}
}

func (m *monitor) reset() {
	m.mu.Lock()
	m.samples = nil
	m.mu.Unlock()
}

func (m *monitor) stats() model.NetworkPingStats {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := model.NetworkPingStats{History: []int{}}
	var sum, ok, lost int
	for _, s := range m.samples {
		if s < 0 {
			lost++
			out.History = append(out.History, 0)
			continue
		}
		ok++
		sum += s
		if s > out.PeakMs {
			out.PeakMs = s
		}
		out.CurrentMs = s
		out.History = append(out.History, s)
	}
	if ok > 0 {
		out.AvgMs = (sum + ok/2) / ok
	}
	if n := len(m.samples); n > 0 {
		out.PacketLossPercent = float64(lost) * 100 / float64(n)
		if m.samples[n-1] < 0 {
			out.CurrentMs = 0
		}
	}
	return out
}

func (e *Engine) runMonitor(ctx context.Context) {
	t := time.NewTicker(monitorInterval)
	defer t.Stop()
	var lastGW string
	for {
		in, err := e.info(ctx)
		if err == nil {
			if ad, ok := in.Active(); ok {
				if gw := ad.Gateway.String(); gw != lastGW {
					e.mon.reset()
					lastGW = gw
				}
				if rtt, ok := e.env.Ping(ctx, ad.Gateway, time.Second); ok {
					e.mon.add(max(1, int((rtt + 500*time.Microsecond).Milliseconds())))
				} else {
					e.mon.add(-1)
				}
			}
		}
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}
