package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadJSONAndYAML(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	jsonPath := filepath.Join(dir, "mk.json")
	yamlPath := filepath.Join(dir, "mk.yaml")

	jsonData := []byte(`{"commands":{"test":{"command":"go test ./...","open":true,"help":"Run tests","usage":"mk test [pkg]"}}}`)
	yamlData := []byte("commands:\n  shell:\n    command: bash\n    open: false\n")

	if err := os.WriteFile(jsonPath, jsonData, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(yamlPath, yamlData, 0o644); err != nil {
		t.Fatal(err)
	}

	jsonCfg, err := Load(jsonPath)
	if err != nil {
		t.Fatal(err)
	}
	if !jsonCfg.Commands["test"].Open {
		t.Fatal("expected json command to be open")
	}

	yamlCfg, err := Load(yamlPath)
	if err != nil {
		t.Fatal(err)
	}
	if yamlCfg.Commands["shell"].Command != "bash" {
		t.Fatalf("unexpected yaml command: %q", yamlCfg.Commands["shell"].Command)
	}
}

func TestLoadMultiCommand(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "mk.json")
	data := []byte(`{"commands":{"test":{"commands":["go test ./...","go vet ./..."],"open":true}}}`)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	got := cfg.Commands["test"].CommandList()
	if len(got) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(got))
	}
	if got[0] != "go test ./..." || got[1] != "go vet ./..." {
		t.Fatalf("unexpected commands: %#v", got)
	}
}
