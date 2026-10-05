//go:build !windows && !linux

package netenv

import (
	"context"
	"net/netip"
	"time"
)

// New returns an Env that reports ErrUnsupported; NETWATCH targets Windows.
func New() Env { return unsupported{} }

type unsupported struct{ Portable }

func (unsupported) Info(context.Context) (Info, error)            { return Info{}, ErrUnsupported }
func (unsupported) Neighbors(context.Context) ([]Neighbor, error) { return nil, ErrUnsupported }
func (unsupported) Ping(context.Context, netip.Addr, time.Duration) (time.Duration, bool) {
	return 0, false
}
func (unsupported) SendARP(context.Context, netip.Addr) (string, error) {
	return "", ErrUnsupported
}
