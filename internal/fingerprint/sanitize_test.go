package fingerprint

import (
	"strings"
	"testing"
)

func TestSanitizeLANString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "clean hostname",
			input:    "desktop-gaming-pc.local",
			expected: "desktop-gaming-pc.local",
		},
		{
			name:     "strips null bytes and control chars",
			input:    "printer\x00\x07\x1b[31m-device\r\n",
			expected: "printer[31m-device",
		},
		{
			name:     "strips unicode bidi overrides",
			input:    "safe-host\u202Ecod-ed", // U+202E is RLO
			expected: "safe-hostcod-ed",
		},
		{
			name:     "strips directional isolates",
			input:    "host\u2066test\u2069name",
			expected: "hosttestname",
		},
		{
			name:     "trims whitespace",
			input:    "   my-phone   ",
			expected: "my-phone",
		},
		{
			name:     "preserves valid arabic and unicode names",
			input:    "جهاز-أحمد-المكتبي",
			expected: "جهاز-أحمد-المكتبي",
		},
		{
			name:     "clamps long string to 64 runes",
			input:    strings.Repeat("a", 100),
			expected: strings.Repeat("a", 64),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := SanitizeLANString(tc.input)
			if got != tc.expected {
				t.Errorf("SanitizeLANString(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}

func TestEscapeCSVField(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "plain string",
			input:    "iPhone-14",
			expected: "iPhone-14",
		},
		{
			name:     "formula equals",
			input:    "=cmd|'/C calc'!A0",
			expected: "'=cmd|'/C calc'!A0",
		},
		{
			name:     "formula plus",
			input:    "+123456",
			expected: "'+123456",
		},
		{
			name:     "formula minus",
			input:    "-2+5+cmd",
			expected: "'-2+5+cmd",
		},
		{
			name:     "formula at sign",
			input:    "@SUM(1+1)",
			expected: "'@SUM(1+1)",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := EscapeCSVField(tc.input)
			if got != tc.expected {
				t.Errorf("EscapeCSVField(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}
