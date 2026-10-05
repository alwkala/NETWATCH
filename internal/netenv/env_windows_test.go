//go:build windows

package netenv

import (
	"context"
	"testing"
)

func TestWindowsNeighbors(t *testing.T) {
	env := New()
	ctx := context.Background()
	nbrs, err := env.Neighbors(ctx)
	if err != nil {
		t.Fatalf("Neighbors: %v", err)
	}
	if len(nbrs) == 0 {
		t.Log("warning: no neighbors found in ARP table")
	}
	for _, n := range nbrs {
		if !n.IP.IsValid() || n.MAC == "" {
			t.Errorf("invalid neighbor: %+v", n)
		}
	}
}
