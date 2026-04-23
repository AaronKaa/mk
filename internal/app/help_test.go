package app

import (
	"bytes"
	"os"
	"testing"
)

func TestHelpReturnsInvalidConfigError(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if err := os.WriteFile("mk.json", []byte(`{"commands":{"bad":{"command":""}}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := Run([]string{"--help"}, &out, &out); err == nil {
		t.Fatal("expected invalid config error")
	}
}
