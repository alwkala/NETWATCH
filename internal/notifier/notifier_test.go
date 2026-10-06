package notifier

import (
	"log/slog"
	"os"
	"testing"
)

func TestNotifierInterface(t *testing.T) {
	var n Notifier = New(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	if n == nil {
		t.Fatal("expected non-nil notifier")
	}

	// Should not panic or block
	n.NotifyNewDevice("Test Device", "192.168.1.100 joined network")
	n.NotifyDeviceOffline("Test Device", "192.168.1.100 went offline")
	n.NotifyNetworkChange("Network Changed", "Switched to 192.168.1.0/24")
}

func TestNoopNotifier(t *testing.T) {
	var n Notifier = NoopNotifier{}
	n.NotifyNewDevice("a", "b")
	n.NotifyDeviceOffline("a", "b")
	n.NotifyNetworkChange("a", "b")
}
