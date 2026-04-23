package app

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestConvertMakefileCommandWritesOutput(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if err := os.WriteFile("Makefile", []byte("test:\n\tgo test ./...\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := Run([]string{"convert-make", "-o", "mk.yaml"}, &out, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "created mk.yaml") {
		t.Fatalf("unexpected output: %s", out.String())
	}
	data, err := os.ReadFile("mk.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "command: go test ./...") {
		t.Fatalf("unexpected converted config: %s", string(data))
	}
}
