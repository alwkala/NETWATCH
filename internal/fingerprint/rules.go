package fingerprint

import (
	"fmt"
	"strings"

	"netwatch/internal/model"
)

// ClassificationResult represents the determined device type and authentic observable reasons.
type ClassificationResult struct {
	Type    string
	Reasons []string
}

// ClassifyBag evaluates an EvidenceBag along with vendor and gateway info
// to produce a deterministic device classification with authentic verifiable evidence reasons.
func ClassifyBag(bag *model.EvidenceBag, vendor string, isGateway bool, openPorts []int) ClassificationResult {
	if isGateway {
		return ClassificationResult{
			Type:    model.TypeRouter,
			Reasons: []string{"Default Gateway"},
		}
	}

	var reasons []string

	if bag != nil {
		// 1. Multi-protocol evidence inspection (mDNS, SSDP, NBNS)
		for _, item := range bag.Current {
			switch item.Source {
			case model.SourceMDNS:
				if item.Key == "service" {
					switch item.Value {
					case "_ipp._tcp", "_printer._tcp":
						return ClassificationResult{
							Type:    model.TypePrinter,
							Reasons: []string{fmt.Sprintf("mDNS advertised service: %s", item.Value)},
						}
					case "_googlecast._tcp":
						return ClassificationResult{
							Type:    model.TypeTV,
							Reasons: []string{"mDNS advertised service: _googlecast._tcp"},
						}
					case "_airplay._tcp", "_raop._tcp":
						return ClassificationResult{
							Type:    model.TypeTV,
							Reasons: []string{fmt.Sprintf("mDNS advertised service: %s", item.Value)},
						}
					case "_spotify-connect._tcp":
						return ClassificationResult{
							Type:    model.TypeIoT,
							Reasons: []string{"mDNS advertised service: _spotify-connect._tcp"},
						}
					case "_hap._tcp":
						return ClassificationResult{
							Type:    model.TypeIoT,
							Reasons: []string{"mDNS advertised service: _hap._tcp (HomeKit)"},
						}
					}
				}
			case model.SourceSSDP:
				if item.Key == "st" {
					if strings.Contains(strings.ToLower(item.Value), "mediarenderer") {
						return ClassificationResult{
							Type:    model.TypeTV,
							Reasons: []string{fmt.Sprintf("SSDP target: %s", item.Value)},
						}
					}
					if strings.Contains(strings.ToLower(item.Value), "internetgatewaydevice") {
						return ClassificationResult{
							Type:    model.TypeRouter,
							Reasons: []string{fmt.Sprintf("SSDP target: %s", item.Value)},
						}
					}
				}
			case model.SourceNBNS:
				if item.Key == "hostname" && item.Value != "" {
					reasons = append(reasons, fmt.Sprintf("NetBIOS computer name: %s", item.Value))
				}
			}
		}
	}

	// 2. Fall back to existing port, hostname, and vendor heuristics
	hostname := ""
	if bag != nil && bag.CanonicalName != "" {
		hostname = bag.CanonicalName
	}

	legacyEvidence := Evidence{
		Vendor:    vendor,
		Hostname:  hostname,
		IsGateway: isGateway,
		OpenPorts: openPorts,
	}

	classifiedType := Classify(legacyEvidence)
	if classifiedType != model.TypeUnknown {
		if len(reasons) == 0 {
			if vendor != "" {
				reasons = append(reasons, fmt.Sprintf("Vendor match: %s", vendor))
			}
			if hostname != "" {
				reasons = append(reasons, fmt.Sprintf("Hostname pattern: %s", hostname))
			}
		}
		return ClassificationResult{
			Type:    classifiedType,
			Reasons: reasons,
		}
	}

	return ClassificationResult{
		Type:    model.TypeUnknown,
		Reasons: reasons,
	}
}
