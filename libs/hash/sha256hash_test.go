package hash

import (
	"strings"
	"testing"
)

func TestGetSHA256Hash(t *testing.T) {
	tests := []struct {
		input    string
		wantLen  int
		wantHex  bool
	}{
		{"192.168.1.1", 64, true},
		{"::1", 64, true},
		{"", 64, true}, // hash of empty string is still 64 hex chars
	}

	for _, tt := range tests {
		got := GetSHA256Hash(tt.input)
		if len(got) != tt.wantLen {
			t.Errorf("GetSHA256Hash(%q): got len %d, want %d", tt.input, len(got), tt.wantLen)
		}
		if tt.wantHex && strings.ContainsAny(got, "ghijklmnopqrstuvwxyz") {
			t.Errorf("GetSHA256Hash(%q): result %q is not valid hex", tt.input, got)
		}
	}
}

func TestGetSHA256Hash_Deterministic(t *testing.T) {
	ip := "203.0.113.42"
	if GetSHA256Hash(ip) != GetSHA256Hash(ip) {
		t.Error("GetSHA256Hash should be deterministic")
	}
}

func TestGetSHA256Hash_Unique(t *testing.T) {
	if GetSHA256Hash("1.2.3.4") == GetSHA256Hash("5.6.7.8") {
		t.Error("different IPs should produce different hashes")
	}
}
