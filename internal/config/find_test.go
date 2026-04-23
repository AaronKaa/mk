package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindWalksParents(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	child := filepath.Join(dir, "a", "b")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(dir, "mk.json")
	if err := os.WriteFile(want, []byte(`{"commands":{}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := Find(child)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("Find() = %q, want %q", got, want)
	}
}

func TestFindForConversion(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	child := filepath.Join(dir, "a", "b")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatal(err)
	}
	jsonPath := filepath.Join(dir, "mk.json")
	yamlPath := filepath.Join(dir, "mk.yaml")
	if err := os.WriteFile(jsonPath, []byte(`{"commands":{}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(yamlPath, []byte("commands: {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := FindForConversion(child, "yaml")
	if err != nil {
		t.Fatal(err)
	}
	if got != jsonPath {
		t.Fatalf("FindForConversion(yaml) = %q, want %q", got, jsonPath)
	}

	got, err = FindForConversion(child, "json")
	if err != nil {
		t.Fatal(err)
	}
	if got != yamlPath {
		t.Fatalf("FindForConversion(json) = %q, want %q", got, yamlPath)
	}
}

func TestConvertPath(t *testing.T) {
	t.Parallel()

	if got, want := ConvertPath("/tmp/mk.json", "yaml"), "/tmp/mk.yaml"; got != want {
		t.Fatalf("ConvertPath() = %q, want %q", got, want)
	}
	if got, want := ConvertPath("/tmp/mk.yaml", "json"), "/tmp/mk.json"; got != want {
		t.Fatalf("ConvertPath() = %q, want %q", got, want)
	}
}
