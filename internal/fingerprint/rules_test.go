package fingerprint

import (
	"net/netip"
	"testing"
	"time"

	"netwatch/internal/model"
)

func TestClassifyBag_MDNSPrinter(t *testing.T) {
	bag := &model.EvidenceBag{
		IP:  netip.MustParseAddr("192.168.1.55"),
		MAC: "00:11:22:33:44:55",
		Current: []model.DiscoveryEvidence{
			{
				Source:     model.SourceMDNS,
				Key:        "service",
				Value:      "_ipp._tcp",
				ObservedAt: time.Now(),
			},
		},
	}

	res := ClassifyBag(bag, "HP Inc", false, nil)
	if res.Type != model.TypePrinter {
		t.Errorf("expected TypePrinter, got %q", res.Type)
	}
	if len(res.Reasons) == 0 {
		t.Error("expected non-empty reasons list")
	}
}

func TestClassifyBag_SSDPMediaRenderer(t *testing.T) {
	bag := &model.EvidenceBag{
		IP:  netip.MustParseAddr("192.168.1.66"),
		MAC: "00:22:33:44:55:66",
		Current: []model.DiscoveryEvidence{
			{
				Source:     model.SourceSSDP,
				Key:        "st",
				Value:      "urn:schemas-upnp-org:device:MediaRenderer:1",
				ObservedAt: time.Now(),
			},
		},
	}

	res := ClassifyBag(bag, "Samsung Electronics", false, nil)
	if res.Type != model.TypeTV {
		t.Errorf("expected TypeTV, got %q", res.Type)
	}
}

func TestClassifyBag_Gateway(t *testing.T) {
	res := ClassifyBag(nil, "Netgear", true, nil)
	if res.Type != model.TypeRouter {
		t.Errorf("expected TypeRouter, got %q", res.Type)
	}
}
