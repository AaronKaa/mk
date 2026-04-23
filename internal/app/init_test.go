package app

import (
	"bytes"
	"testing"

	"github.com/AaronKaa/mk/internal/config"
)

func TestInitCreatesGenericConfig(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	var out bytes.Buffer
	if err := Run([]string{"init"}, &out, &out); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load("mk.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Commands) != 1 {
		t.Fatalf("expected one starter command, got %#v", cfg.Commands)
	}
	cmd, ok := cfg.Commands["ls"]
	if !ok {
		t.Fatal("expected generic ls command")
	}
	if cmd.Command != "ls" || !cmd.Open {
		t.Fatalf("unexpected ls command: %#v", cmd)
	}
}
