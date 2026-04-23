package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadVarsAcceptsStringsAndShellBlocks(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "mk.yaml")
	data := []byte(`vars:
  PROGRAM: foo
  C_FILES:
    shell: printf '*.c'
commands:
  build:
    command: cc -o {{PROGRAM}} {{C_FILES}}
`)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := cfg.Vars["PROGRAM"].Value; got != "foo" {
		t.Fatalf("PROGRAM = %q, want foo", got)
	}
	if got := cfg.Vars["C_FILES"].Shell; got != "printf '*.c'" {
		t.Fatalf("C_FILES shell = %q", got)
	}
}

func TestLoadRejectsReservedVarName(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "mk.json")
	data := []byte(`{"vars":{"args":"bad"},"commands":{"test":{"command":"echo hi"}}}`)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Load(path); err == nil {
		t.Fatal("expected reserved var error")
	}
}
