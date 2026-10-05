// Package oui resolves MAC address prefixes to vendor names entirely offline.
// The IEEE registry (MA-L, MA-M and MA-S blocks) is embedded in the binary,
// so no MAC address is ever sent to an external service.
package oui

import (
	"bufio"
	"bytes"
	_ "embed"
	"net"
	"strings"
	"sync"
)

//go:embed ieee-oui.txt
var embedded []byte

// DB maps hex prefixes (6, 7 or 9 hex digits) to vendor names.
type DB struct {
	m map[string]string
}

var (
	defaultOnce sync.Once
	defaultDB   *DB
)

// Default returns the database built from the embedded registry.
func Default() *DB {
	defaultOnce.Do(func() { defaultDB = Parse(embedded) })
	return defaultDB
}

// Parse reads "<HEXPREFIX>\t<Vendor>" lines; '#' starts a comment.
func Parse(data []byte) *DB {
	db := &DB{m: make(map[string]string, 40000)}
	sc := bufio.NewScanner(bytes.NewReader(data))
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || line[0] == '#' {
			continue
		}
		prefix, vendor, ok := strings.Cut(line, "\t")
		if !ok {
			continue
		}
		prefix = strings.ToUpper(strings.TrimSpace(prefix))
		if n := len(prefix); n != 6 && n != 7 && n != 9 {
			continue
		}
		db.m[prefix] = strings.TrimSpace(vendor)
	}
	return db
}

// Len reports the number of entries.
func (d *DB) Len() int { return len(d.m) }

// Lookup returns the vendor for a MAC using the longest matching prefix.
func (d *DB) Lookup(mac string) (string, bool) {
	hw, err := net.ParseMAC(mac)
	if err != nil || len(hw) < 6 {
		return "", false
	}
	hex := strings.ToUpper(strings.ReplaceAll(hw.String(), ":", ""))
	for _, n := range []int{9, 7, 6} {
		if v, ok := d.m[hex[:n]]; ok {
			return v, true
		}
	}
	return "", false
}

// IsLocallyAdministered reports a randomized/private MAC (bit 1 of first octet).
// Modern phones and laptops use these per network, so they have no vendor.
func IsLocallyAdministered(mac string) bool {
	hw, err := net.ParseMAC(mac)
	if err != nil || len(hw) < 1 {
		return false
	}
	return hw[0]&0x02 != 0
}

// Short trims legal suffixes for compact display ("Apple, Inc." -> "Apple").
func Short(vendor string) string {
	v := strings.TrimSpace(vendor)
	lower := strings.ToLower(v)
	cut := len(v)
	for _, tok := range []string{",", " co.", " co,", " co ", " corp", " inc", " ltd", " gmbh", " s.a", " limited", " technologies", " technology", " electronics", " international", " incorporated", " corporation", " company", " llc"} {
		if i := strings.Index(lower, tok); i > 0 && i < cut {
			cut = i
		}
	}
	v = strings.TrimSpace(v[:cut])
	if v == "" {
		return vendor
	}
	return v
}
