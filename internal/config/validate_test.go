package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadRejectsCommandAndCommandsTogether(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "mk.json")
	data := []byte(`{"commands":{"bad":{"command":"go test ./...","commands":["go vet ./..."]}}}`)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Load(path); err == nil {
		t.Fatal("expected invalid config error")
	}
}

func TestLoadRejectsMissingAliasTarget(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "mk.json")
	data := []byte(`{"commands":{"bad":{"alias":"missing"}}}`)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Load(path); err == nil {
		t.Fatal("expected invalid alias error")
	}
}

func TestLoadNormalizesLegacyAliasCommand(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "mk.json")
	data := []byte(`{"commands":{"test":{"command":"go test ./..."},"t":{"alias":"test"}}}`)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := cfg.Commands["t"]; ok {
		t.Fatal("legacy alias command should be removed")
	}
	if got := cfg.Commands["test"].Aliases; len(got) != 1 || got[0] != "t" {
		t.Fatalf("aliases = %#v, want [t]", got)
	}
}

func TestLoadRejectsAliasCommandNameConflict(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "mk.json")
	data := []byte(`{"commands":{"test":{"command":"go test ./...","aliases":["build"]},"build":{"command":"go build ./..."}}}`)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Load(path); err == nil {
		t.Fatal("expected alias conflict error")
	}
}
