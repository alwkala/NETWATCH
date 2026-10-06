package netenv

import (
	"context"
	"testing"
	"time"
)

func TestDiscoverPublicIP_TimeoutOrOffline(t *testing.T) {
	// Context with already cancelled deadline should exit immediately and safely
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	ip := DiscoverPublicIP(ctx)
	if ip != "" {
		t.Fatalf("expected empty string for cancelled context, got %q", ip)
	}
}

func TestDiscoverPublicIP_Live(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	ip := DiscoverPublicIP(ctx)
	// If the host has active Internet, ip will be non-empty valid IPv4.
	// If offline/air-gapped, it returns empty string without panic.
	t.Logf("Discovered WAN Public IP: %q", ip)
}
