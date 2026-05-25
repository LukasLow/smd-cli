package main

import (
	"testing"
)

func TestDetectTemplateFromCommand(t *testing.T) {
	tests := []struct {
		command string
		want    string
		wantOk  bool
	}{
		{"npm install", "javascript/node", true},
		{"node server.js", "javascript/node", true},
		{"python script.py", "python", true},
		{"go run .", "go", true},
		{"cargo build", "rust", true},
		{"", "", false},
		{"unknowncmd", "", false},
	}
	for _, tc := range tests {
		tpl, ok := DetectTemplateFromCommand(tc.command)
		if ok != tc.wantOk {
			t.Errorf("DetectTemplateFromCommand(%q) ok=%v, want %v", tc.command, ok, tc.wantOk)
			continue
		}
		if ok && tpl.Name != tc.want {
			t.Errorf("DetectTemplateFromCommand(%q) = %q, want %q", tc.command, tpl.Name, tc.want)
		}
	}
}

func TestParseSelection(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"1", []string{"nodejs"}},
		{"1,2", []string{"nodejs", "python"}},
		{"3,4", []string{"go", "rust"}},
		{"", []string{}},
		{"5", []string{}},
		{"1,bad,2", []string{"nodejs", "python"}},
	}
	for _, tc := range tests {
		got := parseSelection(tc.input)
		if !stringSliceEqual(got, tc.want) {
			t.Errorf("parseSelection(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

func TestIsEnvFileBlocked(t *testing.T) {
	cfg := &Config{
		Security: SecurityConfig{
			AllowEnv: []string{".env.local"},
		},
	}

	if !IsEnvFileBlocked(".env", cfg, true) {
		t.Error(".env should be blocked in temp mode")
	}
	if IsEnvFileBlocked(".env.local", cfg, true) {
		t.Error(".env.local should be allowed in temp mode")
	}
	if IsEnvFileBlocked(".env", cfg, false) {
		t.Error("nothing should be blocked in live mode")
	}
	if IsEnvFileBlocked("foo.txt", cfg, true) {
		t.Error("non-env files should not be blocked")
	}
}

func TestIsDynamicMode(t *testing.T) {
	if !IsDynamicMode(&Config{Container: ContainerConfig{Image: "docker.io/nixos/nix:latest"}}) {
		t.Error("nix image should be dynamic")
	}
	if IsDynamicMode(&Config{Container: ContainerConfig{Image: "docker.io/node:20-alpine"}}) {
		t.Error("node image should not be dynamic")
	}
}

func TestDefaultBlockFilesForTypes(t *testing.T) {
	got := defaultBlockFilesForTypes([]string{"nodejs"})
	if len(got) == 0 {
		t.Error("expected blocks for nodejs")
	}

	got = defaultBlockFilesForTypes([]string{"javascript"})
	if len(got) == 0 {
		t.Error("expected blocks for javascript")
	}

	got = defaultBlockFilesForTypes([]string{"unknown"})
	if len(got) != 0 {
		t.Error("expected no blocks for unknown type")
	}
}

func TestDefaultPortsDedup(t *testing.T) {
	ports := defaultPorts([]string{"nodejs", "go"})
	count := 0
	for _, p := range ports {
		if p == "8080:8080" {
			count++
		}
	}
	if count > 1 {
		t.Errorf("8080:8080 should be deduped, got %d occurrences", count)
	}
}

func TestDefaultPortsUnknown(t *testing.T) {
	ports := defaultPorts([]string{"unknown"})
	if len(ports) != 0 {
		t.Errorf("expected no ports for unknown type, got %v", ports)
	}
}

func TestIsAllowedEnvFile(t *testing.T) {
	cfg := &Config{
		Security: SecurityConfig{
			AllowEnv: []string{".env.local", ".env.production"},
		},
	}
	if !isAllowedEnvFile(".env.local", cfg) {
		t.Error(".env.local should be allowed")
	}
	if !isAllowedEnvFile(".env.production", cfg) {
		t.Error(".env.production should be allowed")
	}
	if isAllowedEnvFile(".env", cfg) {
		t.Error(".env should not be allowed")
	}
}

func TestGetProjectName(t *testing.T) {
	name := GetProjectName()
	if name == "" {
		t.Error("expected non-empty project name")
	}
}

func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
