package autostart

import (
	"testing"
)

func TestAutostartManager(t *testing.T) {
	mgr := New()
	if mgr == nil {
		t.Fatal("expected non-nil autostart manager")
	}

	// Read state without modifying
	_, err := mgr.IsEnabled()
	if err != nil {
		t.Logf("IsEnabled returned: %v (expected if registry key absent)", err)
	}
}
