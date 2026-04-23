package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadEnvFileStringAndList(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	jsonPath := filepath.Join(dir, "mk.json")
	yamlPath := filepath.Join(dir, "mk.yaml")
	jsonData := []byte(`{"env_file":".env","commands":{"test":{"command":"go test","env_file":[".env.test"]}}}`)
	yamlData := []byte("env_file:\n  - .env\n  - .env.local\ncommands:\n  test:\n    command: go test\n    env_file: .env.test\n")
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
	if got, want := []string(jsonCfg.EnvFile), []string{".env"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("json EnvFile = %#v, want %#v", got, want)
	}
	if got, want := []string(jsonCfg.Commands["test"].EnvFile), []string{".env.test"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("json command EnvFile = %#v, want %#v", got, want)
	}

	yamlCfg, err := Load(yamlPath)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := []string(yamlCfg.EnvFile), []string{".env", ".env.local"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("yaml EnvFile = %#v, want %#v", got, want)
	}
	if got, want := []string(yamlCfg.Commands["test"].EnvFile), []string{".env.test"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("yaml command EnvFile = %#v, want %#v", got, want)
	}
}

func TestLoadEnvFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("A=1\nB=from-file\nQUOTED=\"hello world\"\nexport EXPORTED=yes\n# ignored\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".env.local"), []byte("B=from-local\nC='three'\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := LoadEnvFiles(dir, EnvFiles{".env", ".env.local"})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"A":        "1",
		"B":        "from-local",
		"C":        "three",
		"QUOTED":   "hello world",
		"EXPORTED": "yes",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("LoadEnvFiles() = %#v, want %#v", got, want)
	}
}
