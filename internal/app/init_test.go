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
	if _, ok := cfg.Commands["test"]; !ok {
		t.Fatal("expected generic test command")
	}
}
