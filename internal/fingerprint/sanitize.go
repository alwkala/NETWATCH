package fingerprint

import (
	"strings"
	"unicode"
)

// SanitizeLANString cleans untrusted hostnames and strings discovered over the LAN
// (e.g., from rDNS, mDNS, NetBIOS, SSDP).
// It strips ASCII/Unicode control characters, ANSI escape sequences, and Unicode
// Bidirectional override characters (which can spoof display order in UIs and logs),
// enforces printable runes, trims leading/trailing whitespace, and clamps the length to 64 runes.
func SanitizeLANString(input string) string {
	if input == "" {
		return ""
	}

	var b strings.Builder
	b.Grow(len(input))

	for _, r := range input {
		// Strip C0, C1 control codes and unicode control categories
		if unicode.IsControl(r) {
			continue
		}
		// Strip Unicode bidirectional overrides and embedding marks:
		// U+202A to U+202E: LRE, RLE, PDF, LRO, RLO
		// U+2066 to U+2069: LRI, RLI, FSI, PDI
		if (r >= 0x202A && r <= 0x202E) || (r >= 0x2066 && r <= 0x2069) {
			continue
		}
		// Disallow non-printable or zero-width formatting runes
		if unicode.IsPrint(r) {
			b.WriteRune(r)
		}
	}

	s := strings.TrimSpace(b.String())
	runes := []rune(s)
	if len(runes) > 64 {
		return string(runes[:64])
	}
	return s
}

// EscapeCSVField prevents CSV Formula Injection (CWE-1236).
// If a string field begins with an execution trigger character ('=', '+', '-', '@', '\t', '\r'),
// it prefixes the field with a single quote (') so spreadsheet applications (Excel, Calc)
// treat it strictly as literal text rather than executable bytecode/macro.
func EscapeCSVField(field string) string {
	if len(field) == 0 {
		return ""
	}
	switch field[0] {
	case '=', '+', '-', '@', '\t', '\r':
		return "'" + field
	default:
		return field
	}
}
