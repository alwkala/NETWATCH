package engine

import (
	"fmt"
	"net/netip"
	"strings"

	"netwatch/internal/fingerprint"
	"netwatch/internal/model"
)

// normalizeMAC standardizes MAC strings to lowercase colon-separated hex (e.g. "aa:bb:cc:dd:ee:ff").
func normalizeMAC(raw string) string {
	s := strings.ToLower(strings.TrimSpace(raw))
	s = strings.ReplaceAll(s, "-", ":")
	return s
}

// ResolveEvidence correlates raw evidence items collected across probes
// by device IP and MAC, checks MAC Unit ID integrity, deduplicates observed names,
// and derives the canonical display name.
func ResolveEvidence(
	obsList []observation,
	rawEvidence []model.DiscoveryEvidence,
	customAliases map[string]string,
) map[string]*model.EvidenceBag {
	// Index observations by IP and MAC
	ipToMAC := make(map[netip.Addr]string)
	macToIP := make(map[string]netip.Addr)
	for _, o := range obsList {
		normMAC := normalizeMAC(o.MAC)
		if normMAC != "" {
			ipToMAC[o.IP] = normMAC
			macToIP[normMAC] = o.IP
		}
	}

	bags := make(map[string]*model.EvidenceBag) // Keyed by MAC

	getOrCreateBag := func(ip netip.Addr, mac string) *model.EvidenceBag {
		normMAC := normalizeMAC(mac)
		if normMAC == "" && ip.IsValid() {
			if m, ok := ipToMAC[ip]; ok {
				normMAC = m
			}
		}
		if normMAC == "" {
			return nil // Untracked host without MAC
		}
		bag, exists := bags[normMAC]
		if !exists {
			bag = &model.EvidenceBag{
				IP:  ip,
				MAC: normMAC,
			}
			bags[normMAC] = bag
		}
		if !bag.IP.IsValid() && ip.IsValid() {
			bag.IP = ip
		}
		return bag
	}

	// Initialize bags for all known observations
	for _, o := range obsList {
		normMAC := normalizeMAC(o.MAC)
		if normMAC != "" {
			getOrCreateBag(o.IP, normMAC)
		}
	}

	// Correlate raw evidence items into bags
	for _, ev := range rawEvidence {
		bag := getOrCreateBag(ev.IP, ev.MAC)
		if bag == nil {
			continue
		}
		bag.Current = append(bag.Current, ev)
	}

	// Resolve each device bag: verify MACs, collect observed names, choose canonical name
	for mac, bag := range bags {
		var mdnsHost, nbnsHost, rdnsHost string
		seenNames := make(map[string]bool)

		for _, item := range bag.Current {
			switch item.Source {
			case model.SourceMDNS:
				if item.Key == "hostname" {
					h := strings.TrimSuffix(item.Value, ".local")
					h = strings.TrimSpace(h)
					if h != "" {
						mdnsHost = h
						nameKey := fmt.Sprintf("mDNS|%s", item.Value)
						if !seenNames[nameKey] {
							seenNames[nameKey] = true
							bag.ObservedNames = append(bag.ObservedNames, model.ObservedName{
								Source: model.SourceMDNS,
								Name:   item.Value,
							})
						}
					}
				}
			case model.SourceNBNS:
				switch item.Key {
				case "hostname":
					h := strings.TrimSpace(item.Value)
					if h != "" {
						nbnsHost = h
						nameKey := fmt.Sprintf("NBNS|%s", h)
						if !seenNames[nameKey] {
							seenNames[nameKey] = true
							bag.ObservedNames = append(bag.ObservedNames, model.ObservedName{
								Source: model.SourceNBNS,
								Name:   h,
							})
						}
					}
				case "unit_id":
					// Correlate NBNS Unit ID (hardware MAC) with ARP MAC
					nbnsMAC := normalizeMAC(item.Value)
					if nbnsMAC != "" && nbnsMAC != "00:00:00:00:00:00" {
						if nbnsMAC == mac {
							// Corroborated match
							bag.Current = append(bag.Current, model.DiscoveryEvidence{
								Source:     model.SourceNBNS,
								IP:         bag.IP,
								MAC:        mac,
								Key:        "mac_match",
								Value:      "true",
								ObservedAt: item.ObservedAt,
								LastSeen:   item.LastSeen,
							})
						} else {
							// Mismatch detected: emit diagnostic conflict note, do NOT merge
							conflictNote := fmt.Sprintf("NBNS Unit ID (%s) does not match ARP MAC (%s)", nbnsMAC, mac)
							bag.Conflicts = append(bag.Conflicts, conflictNote)
							bag.Current = append(bag.Current, model.DiscoveryEvidence{
								Source:     model.SourceNBNS,
								IP:         bag.IP,
								MAC:        mac,
								Key:        "mac_match",
								Value:      "mismatch",
								ObservedAt: item.ObservedAt,
								LastSeen:   item.LastSeen,
							})
						}
					}
				}
			case model.SourceDNS:
				if item.Key == "hostname" {
					h := strings.TrimSpace(item.Value)
					if h != "" {
						rdnsHost = h
						nameKey := fmt.Sprintf("rDNS|%s", h)
						if !seenNames[nameKey] {
							seenNames[nameKey] = true
							bag.ObservedNames = append(bag.ObservedNames, model.ObservedName{
								Source: model.SourceDNS,
								Name:   h,
							})
						}
					}
				}
			}
		}

		// Determine Canonical Name
		if alias, hasAlias := customAliases[mac]; hasAlias && strings.TrimSpace(alias) != "" {
			bag.CanonicalName = fingerprint.SanitizeLANString(alias)
		} else if mdnsHost != "" {
			bag.CanonicalName = fingerprint.SanitizeLANString(mdnsHost)
		} else if nbnsHost != "" {
			bag.CanonicalName = fingerprint.SanitizeLANString(nbnsHost)
		} else if rdnsHost != "" {
			bag.CanonicalName = fingerprint.SanitizeLANString(rdnsHost)
		}
	}

	return bags
}
