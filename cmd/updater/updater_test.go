package main

import (
	"testing"
)

func TestIsNewer(t *testing.T) {
	cases := []struct {
		remote string
		local  string
		want   bool
	}{
		{"v1.2.3", "v1.2.2", true},
		{"v1.2.3", "v1.2.3", false},
		{"v1.2.3", "v1.3.0", false},
		{"v2.0.0", "v1.9.9", true},
		{"v1.0.0", "v2.0.0", false},
		{"v1.10.0", "v1.9.0", true},
	}
	for _, tc := range cases {
		got := isNewer(tc.remote, tc.local)
		if got != tc.want {
			t.Errorf("isNewer(%q, %q) = %v, want %v", tc.remote, tc.local, got, tc.want)
		}
	}
}

func TestFindChecksum(t *testing.T) {
	checksums := "abc123ef  home-guard.exe\ndef456ab  home-guard-updater.exe\n"

	hash, err := findChecksum(checksums, "home-guard.exe")
	if err != nil {
		t.Fatalf("findChecksum() error = %v", err)
	}
	if hash != "abc123ef" {
		t.Errorf("hash = %q, want %q", hash, "abc123ef")
	}

	hash2, err := findChecksum(checksums, "home-guard-updater.exe")
	if err != nil {
		t.Fatalf("findChecksum() error = %v", err)
	}
	if hash2 != "def456ab" {
		t.Errorf("hash = %q, want %q", hash2, "def456ab")
	}
}

func TestFindChecksumNotFound(t *testing.T) {
	checksums := "abc123ef  home-guard.exe\n"

	_, err := findChecksum(checksums, "missing.exe")
	if err == nil {
		t.Error("expected error for missing checksum, got nil")
	}
}

func TestParseVersion(t *testing.T) {
	cases := []struct {
		input string
		want  [3]int
	}{
		{"v1.2.3", [3]int{1, 2, 3}},
		{"1.2.3", [3]int{1, 2, 3}},
		{"v2.0.0", [3]int{2, 0, 0}},
		{"v1.10.5", [3]int{1, 10, 5}},
	}
	for _, tc := range cases {
		got := parseVersion(tc.input)
		if got != tc.want {
			t.Errorf("parseVersion(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}
